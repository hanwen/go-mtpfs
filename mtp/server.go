package mtp

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"go.uber.org/atomic"

	"github.com/gorilla/websocket"
	"github.com/paulbellamy/ratecounter"
	"golang.org/x/sync/errgroup"
)

// LVServer captures LV images and serves the images asynchronously.

type LVServer struct {
	Frame        []byte
	newFrameChan chan bool
	frameLock    sync.Mutex

	fpsRate  *ratecounter.RateCounter
	info     InfoPayload
	infoLock sync.Mutex

	upgrader       websocket.Upgrader
	streamClients  map[*websocket.Conn]bool
	streamLock     sync.Mutex
	controlClients map[*websocket.Conn]bool
	controlLock    sync.Mutex

	dev     Device
	mtpLock sync.Mutex
	dummy   bool

	afInterval int
	afTicker   *MutableTicker
	afNowChan  chan bool

	lrFPS *atomic.Int64

	eg  *errgroup.Group
	ctx context.Context
	log *logrus.Logger
}

func NewLVServer(dev Device, log *logrus.Logger, ctx context.Context) *LVServer {
	eg, egCtx := errgroup.WithContext(ctx)

	return &LVServer{
		Frame:        nil,
		newFrameChan: make(chan bool, 1),

		fpsRate: ratecounter.NewRateCounter(time.Second),

		streamClients:  map[*websocket.Conn]bool{},
		controlClients: map[*websocket.Conn]bool{},

		dev:   dev,
		dummy: dev == nil,

		afInterval: 5,
		afTicker:   NewMutableTicker(5 * time.Second),
		afNowChan:  make(chan bool),

		lrFPS: atomic.NewInt64(0),

		eg:  eg,
		ctx: egCtx,
		log: log,
	}
}

// HTTP handler / WebSocket

func (s *LVServer) HandleStream(w http.ResponseWriter, r *http.Request) {
	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.log.WithField("prefix", "lv.HandleStream").Errorf("failed to upgrade: %s", err)
	}
	defer ws.Close()

	s.registerStreamClient(ws)
	for {
		var mes struct{}
		err := ws.ReadJSON(&mes)
		if err != nil {
			s.log.WithField("prefix", "lv.HandleStream").Errorf("failed to read a message: %s", err)
			s.unregisterStreamClient(ws)
			return
		}
	}
}

func (s *LVServer) registerStreamClient(c *websocket.Conn) {
	s.streamLock.Lock()
	defer s.streamLock.Unlock()
	s.streamClients[c] = true
}

func (s *LVServer) unregisterStreamClient(c *websocket.Conn) {
	s.streamLock.Lock()
	defer s.streamLock.Unlock()
	delete(s.streamClients, c)
}

type ControlPayload struct {
	AFEnable   *bool  `json:"af_enable,omitempty"`
	AFInterval *int   `json:"af_interval,omitempty"`
	AFFocusNow *bool  `json:"af_focus_now,omitempty"`
	LRFPS      *int64 `json:"lr_fps,omitempty"`
	ISO        *int   `json:"iso,omitempty"`
}

type InfoPayload struct {
	ISOs   []int  `json:"isos"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	FPS    int    `json:"fps"`
	Frame  []byte `json:"frame"`
}

func (s *LVServer) HandleControl(w http.ResponseWriter, r *http.Request) {
	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.log.WithField("prefix", "lv.HandleControl").Errorf("failed to upgrade: %s", err)
	}
	defer ws.Close()

	s.registerControlClient(ws)
	for {
		var p ControlPayload
		err := ws.ReadJSON(&p)
		if err != nil {
			s.log.WithField("prefix", "lv.HandleControl").Errorf("failed to read a message: %s", err)
			s.unregisterControlClient(ws)
			return
		}

		if p.AFEnable != nil && p.AFInterval != nil {
			if *p.AFEnable {
				s.log.WithField("prefix", "lv.HandleControl").Debug("enable AF")
				s.afTicker.Start()
			} else {
				s.log.WithField("prefix", "lv.HandleControl").Debug("disable AF")
				s.afTicker.Stop()
				continue
			}

			if *p.AFInterval < 1 {
				s.log.WithField("prefix", "lv.HandleControl").Errorf("invalid AF interval: %d", *p.AFInterval)
				continue
			}

			s.afInterval = *p.AFInterval
			s.afTicker.SetInterval(time.Duration(*p.AFInterval) * time.Second)
			if err != nil {
				s.log.WithField("prefix", "lv.HandleControl").Errorf("failed to set interval: %d", *p.AFInterval)
			}
			s.log.WithField("prefix", "lv.HandleControl").Debugf("set AF interval: %d", *p.AFInterval)
		}

		if p.AFFocusNow != nil && *p.AFFocusNow {
			s.log.WithField("prefix", "lv.HandleControl").Debug("focus now")
			select {
			case s.afNowChan <- true:
			default:
			}
		}

		if p.LRFPS != nil {
			s.lrFPS.Store(*p.LRFPS)
		}

		if p.ISO != nil {
			s.log.WithField("prefix", "lv.HandleControl").Debugf("set ISO: %d", *p.ISO)
			err = s.setISO(*p.ISO)
			if err != nil {
				s.log.WithField("prefix", "lv.HandleControl").Errorf("failed to set ISO: %s", err)
			}
		}
	}
}

func (s *LVServer) registerControlClient(c *websocket.Conn) {
	s.controlLock.Lock()
	defer s.controlLock.Unlock()
	s.controlClients[c] = true
}

func (s *LVServer) unregisterControlClient(c *websocket.Conn) {
	s.controlLock.Lock()
	defer s.controlLock.Unlock()
	delete(s.controlClients, c)
}

// Workers

func (s *LVServer) Run() error {
	defer func() {
		_ = s.endLiveView()
	}()

	isos, err := s.getISOs()
	if err != nil {
		return fmt.Errorf("failed to obtain ISO list: %s", err)
	}
	s.info.ISOs = isos

	s.eg.Go(s.workerLV)
	s.eg.Go(s.workerAF)
	time.Sleep(500 * time.Millisecond)
	s.eg.Go(s.frameCaptorSakura)
	s.eg.Go(s.workerBroadcastFrame)
	s.eg.Go(s.workerBroadcastInfo)
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
			s.log.WithField("prefix", "lv.workerLV").Warning(err)
			continue
		} else if status {
			continue
		}

		err = s.startLiveView()
		if err != nil {
			s.log.WithField("prefix", "lv.workerLV").Warning(err)
		}
	}
}

func (s *LVServer) workerAF() error {
	for {
		select {
		case <-s.afTicker.C:
			// Let's go!
		case <-s.afNowChan:
			// Do it now
		case <-s.ctx.Done():
			return nil
		}

		err := s.autoFocus()
		if err != nil {
			s.log.WithField("prefix", "lv.workerAF").Warning(err)
		}
	}
}

func (s *LVServer) frameCaptorSakura() error {
	set := func(lv LiveView) {
		s.frameLock.Lock()
		s.infoLock.Lock()
		defer s.frameLock.Unlock()
		defer s.infoLock.Unlock()
		s.Frame = lv.JPEG
		s.info.Width = int(lv.LVWidth)
		s.info.Height = int(lv.LVHeight)
		select {
		case s.newFrameChan <- true:
		default:
		}
	}

	last := time.Now()

	for {
		select {
		case <-s.ctx.Done():
			return nil
		default:
			// Let's go!
		}

		if s.dummy {
			time.Sleep(time.Second)
			continue
		}

		if s.lrFPS.Load() > 0 {
			time.Sleep(last.Add(time.Second / time.Duration(s.lrFPS.Load())).Sub(time.Now()))
		}
		last = time.Now()

		lv, err := s.getLiveViewImg()
		if err != nil {
			if err.Error() == "failed to obtain an image: live view is not activated" {
				time.Sleep(time.Second)
			} else {
				s.log.WithField("prefix", "lv.frameCaptor").Warning(err)
				time.Sleep(time.Second)
			}
		}
		set(lv)
		s.fpsRate.Incr(1)
	}
}

func (s *LVServer) copyFrame() []byte {
	s.frameLock.Lock()
	defer s.frameLock.Unlock()
	return s.Frame[:]
}

func (s *LVServer) workerBroadcastFrame() error {
	broadcast := func(jpeg []byte) {
		s.streamLock.Lock()
		defer s.streamLock.Unlock()

		b64 := base64.StdEncoding.EncodeToString(jpeg)

		for c := range s.streamClients {
			err := c.WriteMessage(websocket.TextMessage, []byte(b64))
			if err != nil {
				s.log.WithField("prefix", "lv.workerBroadcastFrame").Errorf("failed to send a frame: %s", err)
			}
		}
	}

	for {
		select {
		case <-s.ctx.Done():
			return nil
		case <-s.newFrameChan:
		}

		var jpeg []byte
		jpeg = s.copyFrame()
		if len(jpeg) == 0 {
			continue
		}
		broadcast(jpeg)
	}
}

func (s *LVServer) workerBroadcastInfo() error {
	tick := time.NewTicker(time.Second)

	broadcast := func() {
		s.controlLock.Lock()
		s.infoLock.Lock()
		defer s.controlLock.Unlock()
		defer s.infoLock.Unlock()

		s.info.Frame = s.copyFrame()
		s.info.FPS = int(s.fpsRate.Rate())

		for c := range s.controlClients {
			j, err := json.Marshal(s.info)
			if err != nil {
				s.log.WithField("prefix", "lv.workerBroadcastInfo").Errorf("failed to marshal payload: %s", err)
				continue
			}
			err = c.WriteMessage(websocket.TextMessage, j)
			if err != nil {
				s.log.WithField("prefix", "lv.workerBroadcastInfo").Errorf("failed to send a frame: %s", err)
			}
		}
	}

	for {
		select {
		case <-s.ctx.Done():
			return nil
		case <-tick.C:
			// Let's go!
		}

		broadcast()
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

	if s.dummy {
		return nil
	}

	err := s.dev.RunTransactionWithNoParams(OC_NIKON_EndLiveView)
	if err != nil {
		return fmt.Errorf("failed to end live view: %s", err)
	}
	return nil
}

func (s *LVServer) getLiveViewStatus() (bool, error) {
	s.mtpLock.Lock()
	defer s.mtpLock.Unlock()

	if s.dummy {
		return true, nil
	}

	err, status := s.getLiveViewStatusInner()
	if err != nil {
		return false, fmt.Errorf("failed to get live view status: %s", err)
	}
	return status, nil
}

func (s *LVServer) autoFocus() error {
	s.mtpLock.Lock()
	defer s.mtpLock.Unlock()

	if s.dummy {
		return nil
	}

	err := s.dev.RunTransactionWithNoParams(OC_NIKON_AfDrive)
	if err != nil {
		return fmt.Errorf("failed to do auto focus: %s", err)
	}
	return nil
}

func (s *LVServer) getLiveViewImg() (LiveView, error) {
	s.mtpLock.Lock()
	defer s.mtpLock.Unlock()

	if s.dummy {
		return LiveView{}, nil
	}

	lv, err := s.getLiveViewImgInner()
	if err != nil {
		return LiveView{}, err
	}
	return lv, nil
}

// Plain MTP communication

func (s *LVServer) getLiveViewStatusInner() (error, bool) {
	val := StringValue{}
	err := s.dev.GetDevicePropValue(DPC_NIKON_LiveViewStatus, &val)

	if err != nil && err != io.EOF {
		return err, false
	}

	return nil, err == io.EOF
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

func (s *LVServer) getLiveViewImgInner() (LiveView, error) {
	var req, rep Container
	buf := bytes.NewBuffer([]byte{})

	req.Code = OC_NIKON_GetLiveViewImg
	req.Param = []uint32{}
	err := s.dev.RunTransaction(&req, &rep, buf, nil, 0)
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

func (s *LVServer) getISOs() ([]int, error) {
	s.mtpLock.Lock()
	defer s.mtpLock.Unlock()

	if s.dummy {
		return []int{100, 1000, 10000}, nil
	}

	isoi := make([]int, 0)

	val := DevicePropDesc{}
	err := s.dev.GetDevicePropDesc(DPC_ExposureIndex, &val)

	if err != nil && err != io.EOF {
		return isoi, err
	}

	asserted, ok := val.Form.(*PropDescEnumForm)
	if !ok {
		return isoi, fmt.Errorf("unexpedted type: could not assert that returned prop is enum form")
	}

	for _, raw := range asserted.Values {
		iso, ok := raw.(uint64)
		if !ok {
			return isoi, fmt.Errorf("unexpedted type: could not assert that form value is uint64")
		}
		isoi = append(isoi, int(iso))
	}

	return isoi, nil
}

func (s *LVServer) setISO(iso int) error {
	s.mtpLock.Lock()
	defer s.mtpLock.Unlock()

	if s.dummy {
		return nil
	}

	err := s.dev.SetDevicePropValue(DPC_ExposureIndex, &struct {
		ISO uint64
	}{
		ISO: uint64(iso),
	})
	if err != nil {
		return fmt.Errorf("failed to set ISO: %s", err)
	}
	return nil
}
