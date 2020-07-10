package mtp

import (
	"time"

	"go.uber.org/atomic"
)

type MutableTicker struct {
	C <-chan bool
	d *atomic.Int64
	e *atomic.Bool
	i chan bool
}

func NewMutableTicker(d time.Duration) *MutableTicker {
	c := make(chan bool, 1)
	mt := &MutableTicker{
		C: c,
		d: atomic.NewInt64(int64(d)),
		e: atomic.NewBool(true),
		i: make(chan bool, 1),
	}

	go func() {
		for {
			if mt.e.Load() {
				select {
				case c <- true:
				default:
				}
			}

			t := time.NewTimer(time.Duration(mt.d.Load()))
			select {
			case <-t.C:
			case <-mt.i:
			}
		}
	}()

	return mt
}

func (mt *MutableTicker) SetInterval(d time.Duration) {
	mt.d.Store(int64(d))
	mt.interrupt()
}

func (mt *MutableTicker) Stop() {
	mt.e.Store(false)
	mt.interrupt()
}

func (mt *MutableTicker) Start() {
	mt.e.Store(true)
	mt.interrupt()
}

func (mt *MutableTicker) interrupt() {
	select {
	case mt.i <- true:
	default:
	}
}
