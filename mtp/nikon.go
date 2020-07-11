package mtp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
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

func (d *Device2) NikonGetLiveViewStatus() (error, bool) {
	val := StringValue{}
	err := d.GetDevicePropValue(DPC_NIKON_LiveViewStatus, &val)

	if err != nil && err != io.EOF {
		return err, false
	}

	return nil, err == io.EOF
}

/*
func (d *Device) RunTransactionWithNoParams(code uint16) error {
	var req, rep Container
	req.Code = code
	req.Param = []uint32{}
	return d.RunTransaction(&req, &rep, nil, nil, 0)
}
*/

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

func (d *Device2) NikonGetLiveViewImg() (LiveView, error) {
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
