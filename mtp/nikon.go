package mtp

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/sync/errgroup"
)

// Nikon MTP extensions

const (
	OC_NIKON_AfDrive = 0x90C1
)

const (
	LVHeaderSize = 384
)

type Rotation int

const (
	Rotation0       Rotation = 0
	Rotation90      Rotation = 90
	RotationMinus90 Rotation = -90
	Rotation180     Rotation = 180
)

type AF int

const (
	AFNotActive AF = 0
	AFFail      AF = 1
	AFSuccess   AF = 2
)

func (d *Device) NikonGetLiveViewStatus() (error, bool) {
	val := StringValue{}
	err := d.GetDevicePropValue(DPC_NIKON_LiveViewStatus, &val)

	if err != nil && err != io.EOF {
		return err, false
	}

	return nil, err == io.EOF
}

func (d *Device) RunTransactionWithNoParams(code uint16) error {
	var req, rep Container
	req.Code = code
	req.Param = []uint32{}
	return d.RunTransaction(&req, &rep, nil, nil, 0)
}

type liveViewRaw struct {
	LVWidth             int16
	LVHeight            int16
	Width               int16
	Height              int16
	Dummy1              [8]byte
	FocusFrameWidth     int16
	FocusFrameHeight    int16
	FocusX              int16
	FocusY              int16
	Dummy2              [5]byte
	Rotation            int8
	Dummy3              [10]byte
	AutoFocus           int8
	Dummy4              [15]byte
	MovieTimeRemainInt  int16
	MovieTimeRemainFrac int16
	Recording           int8
}

type LiveView struct {
	LVWidth          int16
	LVHeight         int16
	Width            int16
	Height           int16
	FocusFrameWidth  int16
	FocusFrameHeight int16
	FocusX           int16
	FocusY           int16
	Rotation         Rotation
	AutoFocus        AF
	MovieTimeRemain  float64
	Recording        bool

	JPEG []byte
}

func (d *Device) NikonGetLiveViewImg() (LiveView, error) {
	var req, rep Container
	buf := bytes.NewBuffer([]byte{})

	req.Code = OC_NIKON_GetLiveViewImg
	req.Param = []uint32{}
	err := d.RunTransaction(&req, &rep, buf, nil, 0)
	if err != nil {
		if casted, ok := err.(RCError); ok && uint16(casted) == RC_NIKON_NotLiveView {
			return LiveView{}, fmt.Errorf("failed to obtain an image: live view is not activated")
		}
		return LiveView{}, fmt.Errorf("failed to obtain an image: %s", err)
	} else if buf.Len() <= LVHeaderSize {
		return LiveView{}, fmt.Errorf("failed to obtain an image: the data has insufficient length")
	}

	raw := buf.Bytes()

	lvr := liveViewRaw{}
	err = binary.Read(bytes.NewReader(raw[8:LVHeaderSize]), binary.BigEndian, &lvr)
	if err != nil {
		return LiveView{}, fmt.Errorf("failed to decode header")
	}

	remain, err := strconv.ParseFloat(fmt.Sprintf("%d.%d", lvr.MovieTimeRemainInt, lvr.MovieTimeRemainFrac), 64)
	if err != nil {
		return LiveView{}, fmt.Errorf("failed to parse MovieTimeRemain: %s", err)
	}

	rot := Rotation0
	if lvr.Rotation == 1 {
		rot = RotationMinus90
	} else if lvr.Rotation == 2 {
		rot = Rotation90
	} else if lvr.Rotation == 3 {
		rot = Rotation180
	}

	af := AFNotActive
	if lvr.AutoFocus == 1 {
		af = AFFail
	} else if lvr.AutoFocus == 2 {
		af = AFSuccess
	}

	return LiveView{
		LVWidth:          lvr.LVWidth,
		LVHeight:         lvr.LVHeight,
		Width:            lvr.Width,
		Height:           lvr.Height,
		FocusFrameWidth:  lvr.FocusFrameWidth,
		FocusFrameHeight: lvr.FocusFrameHeight,
		FocusX:           lvr.FocusX,
		FocusY:           lvr.FocusY,
		Rotation:         rot,
		AutoFocus:        af,
		MovieTimeRemain:  remain,
		Recording:        lvr.Recording == 1,
		JPEG:             raw[LVHeaderSize:],
	}, nil
}

// LVServer captures LV images and serves the images asynchronously.

type LVServer struct {
	Frame     []byte
	frameHash [16]byte
	frameLock sync.Mutex

	upgrader websocket.Upgrader
	clients  map[*websocket.Conn]bool
	wsLock   sync.Mutex

	dev     *Device
	mtpLock sync.Mutex

	eg  *errgroup.Group
	ctx context.Context
}

func NewLVServer(dev *Device, ctx context.Context) *LVServer {
	eg, egCtx := errgroup.WithContext(ctx)

	return &LVServer{
		Frame:   nil,
		clients: map[*websocket.Conn]bool{},
		dev:     dev,

		eg:  eg,
		ctx: egCtx,
	}
}

// HTTP handler / WebSocket

func (s *LVServer) HandleClient(w http.ResponseWriter, r *http.Request) {
	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("failed to upgrade: %s", err)
	}
	defer ws.Close()

	s.registerClient(ws)
	for {
		var mes struct{}
		err := ws.ReadJSON(&mes)
		if err != nil {
			log.Printf("failed to read a message: %s", err)
			s.unregisterClient(ws)
			return
		}
	}
}

func (s *LVServer) registerClient(c *websocket.Conn) {
	s.wsLock.Lock()
	defer s.wsLock.Unlock()
	s.clients[c] = true
}

func (s *LVServer) unregisterClient(c *websocket.Conn) {
	s.wsLock.Lock()
	defer s.wsLock.Unlock()
	delete(s.clients, c)
}

// Workers

func (s *LVServer) Run() error {
	defer func() {
		_ = s.endLiveView()
	}()

	s.eg.Go(s.workerLV)
	s.eg.Go(s.workerAF)
	time.Sleep(500 * time.Millisecond)
	s.eg.Go(s.frameCaptorSakura)
	s.eg.Go(s.workerBroadcast)
	return s.eg.Wait()
}

func (s *LVServer) workerLV() error {
	tick := time.NewTicker(time.Second)

	for {
		select {
		case <-tick.C:
			// let's go!
		case <-s.ctx.Done():
			return nil
		}

		status, err := s.getLiveViewStatus()
		if err != nil {
			return err
		} else if status {
			continue
		}

		err = s.startLiveView()
		if err != nil {
			return err
		}
	}
}

func (s *LVServer) workerAF() error {
	tick := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-tick.C:
			// Let's go!
		case <-s.ctx.Done():
			return nil
		}

		err := s.autoFocus()
		if err != nil {
			return err
		}
	}
}

func (s *LVServer) frameCaptorSakura() error {
	set := func(lv LiveView) {
		s.frameLock.Lock()
		defer s.frameLock.Unlock()
		s.Frame = lv.JPEG
		s.frameHash = md5.Sum(lv.JPEG)
	}

	for {
		select {
		case <-s.ctx.Done():
			return nil
		default:
			// Let's go!
		}

		lv, err := s.getLiveViewImg()
		if err != nil {
			if err.Error() == "failed to obtain an image: live view is not activated" {
				time.Sleep(time.Second)
			} else {
				log.Printf("Failed to capture, wait for 1s: %s", err)
				time.Sleep(time.Second)
			}
		}
		set(lv)
	}
}

func (s *LVServer) workerBroadcast() error {
	copyFrame := func() ([]byte, [16]byte) {
		s.frameLock.Lock()
		defer s.frameLock.Unlock()

		newHash := [16]byte{}
		for i, b := range s.frameHash {
			newHash[i] = b
		}
		return s.Frame[:], newHash
	}

	broadcast := func(jpeg []byte) {
		s.wsLock.Lock()
		defer s.wsLock.Unlock()

		b64 := base64.StdEncoding.EncodeToString(jpeg)

		for c := range s.clients {
			err := c.WriteMessage(websocket.TextMessage, []byte(b64))
			if err != nil {
				log.Printf("failed to send a frame: %s", err)
			}
		}
	}

	lastHash := [16]byte{}
	for {
		select {
		case <-s.ctx.Done():
			return nil
		default:
			// Let's go!
		}

		if s.frameHash == lastHash {
			time.Sleep(time.Millisecond)
			continue
		}

		var jpeg []byte
		jpeg, lastHash = copyFrame()
		if len(jpeg) == 0 {
			continue
		}
		broadcast(jpeg)
	}
}

// Thread-safe MTP communication

func (s *LVServer) startLiveView() error {
	s.mtpLock.Lock()
	defer s.mtpLock.Unlock()

	err := s.dev.RunTransactionWithNoParams(OC_NIKON_StartLiveView)
	if err != nil {
		if casted, ok := err.(RCError); ok && uint16(casted) == RC_NIKON_InvalidStatus {
			return fmt.Errorf("failed to start live view: InvalidStatus (battery level is low?)")
		}
		return fmt.Errorf("failed to start live view: %s", err)
	}
	return nil
}

func (s *LVServer) endLiveView() error {
	s.mtpLock.Lock()
	defer s.mtpLock.Unlock()

	err := s.dev.RunTransactionWithNoParams(OC_NIKON_EndLiveView)
	if err != nil {
		return fmt.Errorf("failed to end live view: %s", err)
	}
	return nil
}

func (s *LVServer) getLiveViewStatus() (bool, error) {
	s.mtpLock.Lock()
	defer s.mtpLock.Unlock()

	err, status := s.dev.NikonGetLiveViewStatus()
	if err != nil {
		return false, fmt.Errorf("failed to get live view status: %s", err)
	}
	return status, nil
}

func (s *LVServer) autoFocus() error {
	s.mtpLock.Lock()
	defer s.mtpLock.Unlock()

	err := s.dev.RunTransactionWithNoParams(OC_NIKON_AfDrive)
	if err != nil {
		return fmt.Errorf("failed to do auto focus: %s", err)
	}
	return nil
}

func (s *LVServer) getLiveViewImg() (LiveView, error) {
	s.mtpLock.Lock()
	defer s.mtpLock.Unlock()

	lv, err := s.dev.NikonGetLiveViewImg()
	if err != nil {
		return LiveView{}, err
	}
	return lv, nil
}
