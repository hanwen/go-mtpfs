package mtp

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
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
		s.log.WithField("prefix", "LV.HandleStream").Errorf("Failed to upgrade: %s", err)
	}
	defer ws.Close()

	s.registerStreamClient(ws)
	for {
		var mes struct{}
		err := ws.ReadJSON(&mes)
		if err != nil {
			s.log.WithField("prefix", "LV.HandleStream").Errorf("Failed to read a message: %s", err)
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
}

type InfoPayload struct {
	Width  int    `json:"width"`
	Height int    `json:"height"`
	FPS    int    `json:"fps"`
	Frame  []byte `json:"frame"`
}

func (s *LVServer) HandleControl(w http.ResponseWriter, r *http.Request) {
	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.log.WithField("prefix", "LV.HandleControl").Errorf("Failed to upgrade: %s", err)
	}
	defer ws.Close()

	s.registerControlClient(ws)
	for {
		var p ControlPayload
		err := ws.ReadJSON(&p)
		if err != nil {
			s.log.WithField("prefix", "LV.HandleControl").Errorf("Failed to read a message: %s", err)
			s.unregisterControlClient(ws)
			return
		}

		if p.AFEnable != nil && p.AFInterval != nil {
			if *p.AFEnable {
				s.afTicker.Start()
			} else {
				s.afTicker.Stop()
				continue
			}

			if *p.AFInterval < 1 {
				s.log.WithField("prefix", "LV.HandleControl").Errorf("Invalid AF interval: %d", *p.AFInterval)
				continue
			}

			s.afInterval = *p.AFInterval
			s.afTicker.SetInterval(time.Duration(*p.AFInterval) * time.Second)
			if err != nil {
				s.log.WithField("prefix", "LV.HandleControl").Errorf("Failed to set interval: %d", *p.AFInterval)
			}
		}

		if p.AFFocusNow != nil && *p.AFFocusNow {
			select {
			case s.afNowChan <- true:
			default:
			}
		}

		if p.LRFPS != nil {
			s.lrFPS.Store(*p.LRFPS)
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

	s.eg.Go(s.workerLV)
	//s.eg.Go(s.workerAF)
	time.Sleep(500 * time.Millisecond)
	s.eg.Go(s.frameCaptorSakura)
	s.eg.Go(s.workerBroadcastFrame)
	//s.eg.Go(s.workerBroadcastInfo)
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
			s.log.WithField("prefix", "LV.workerLV").Warning(err)
			continue
		} else if status {
			continue
		}

		err = s.startLiveView()
		if err != nil {
			s.log.WithField("prefix", "LV.workerLV").Warning(err)
		}
		return nil
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
			s.log.WithField("prefix", "LV.workerAF").Warning(err)
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
				s.log.WithField("prefix", "LV.captor").Warning(err)
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
				s.log.WithField("prefix", "LV.workerBroadcastFrame").Errorf("Failed to send a frame: %s", err)
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
				s.log.WithField("prefix", "LV.workerBroadcastInfo").Errorf("Failed to marshal payload: %s", err)
				continue
			}
			err = c.WriteMessage(websocket.TextMessage, j)
			if err != nil {
				s.log.WithField("prefix", "LV.workerBroadcastInfo").Errorf("Failed to send a frame: %s", err)
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

	err, status := s.NikonGetLiveViewStatus()
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

	lv, err := s.NikonGetLiveViewImg()
	if err != nil {
		return LiveView{}, err
	}
	return lv, nil
}
