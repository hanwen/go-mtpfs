package mtp

import (
	"fmt"
	"io"
)

type Device interface {
	Configure() error
	SetDebug(flags DebugFlags)
	RunTransactionWithNoParams(code uint16) error
	RunTransaction(req *Container, rep *Container, dest io.Writer, src io.Reader, writeSize int64) error
	GetDevicePropValue(propCode uint32, dest interface{}) error
}

type sessionData struct {
	tid uint32
	sid uint32
}

// RCError are return codes from the Container.Code field.
type RCError uint16

func (e RCError) Error() string {
	n, ok := RC_names[int(e)]
	if ok {
		return n
	}
	return fmt.Sprintf("RetCode %x", uint16(e))
}

// SyncError is an error type that indicates lost transaction
// synchronization in the protocol.
type SyncError string

func (s SyncError) Error() string {
	return string(s)
}

type Catastrophic string

func (f Catastrophic) Error() string {
	return string(f)
}

// The linux usb stack can send 16kb per call, according to libusb.
const rwBufSize = 0x4000

type DebugFlags struct {
	MTP  bool
	USB  bool
	Data bool
}
