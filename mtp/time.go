package mtp

import (
	"sync"
	"time"

	"go.uber.org/atomic"
)

type atomicTime struct {
	l sync.Mutex
	t time.Time
}

func newAtomicTime(initial time.Time) *atomicTime {
	return &atomicTime{
		t: initial,
	}
}

func (a *atomicTime) Store(v time.Time) {
	a.l.Lock()
	defer a.l.Unlock()
	a.t = v
}

func (a *atomicTime) Load() time.Time {
	a.l.Lock()
	defer a.l.Unlock()
	return a.t.Add(0)
}

type MutableTicker struct {
	C <-chan bool
	d *atomic.Int64
	e *atomic.Bool
	n *atomicTime
}

func NewMutableTicker(d time.Duration) *MutableTicker {
	c := make(chan bool, 1)
	mt := &MutableTicker{
		C: c,
		d: atomic.NewInt64(int64(d)),
		e: atomic.NewBool(true),
		n: newAtomicTime(time.Now()),
	}

	go func() {
		next := func() time.Time {
			return mt.n.Load().Add(time.Duration(mt.d.Load()))
		}

		for {
			mt.n.Store(time.Now())

			if mt.e.Load() {
				select {
				case c <- true:
				default:
				}
			}

			for time.Now().Before(next()) {
				time.Sleep(time.Millisecond)
			}
		}
	}()

	return mt
}

func (mt *MutableTicker) SetInterval(d time.Duration) {
	mt.n.Store(time.Now())
	mt.d.Store(int64(d))
}

func (mt *MutableTicker) Stop() {
	mt.e.Store(false)
}

func (mt *MutableTicker) Start() {
	mt.n.Store(time.Now())
	mt.e.Store(true)
}
