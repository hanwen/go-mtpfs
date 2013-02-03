package mtp

import (
	"io"
)

// Android MTP extensions

// Same as GetPartialObject, but with 64 bit offset
const OC_ANDROID_GET_PARTIAL_OBJECT64 = 0x95C1

// Same as GetPartialObject64, but copying host to device
const OC_ANDROID_SEND_PARTIAL_OBJECT = 0x95C2

// Truncates file to 64 bit length
const OC_ANDROID_TRUNCATE_OBJECT = 0x95C3

// Must be called before using SendPartialObject and TruncateObject
const OC_ANDROID_BEGIN_EDIT_OBJECT = 0x95C4

// Called to commit changes made by SendPartialObject and TruncateObject
const OC_ANDROID_END_EDIT_OBJECT = 0x95C5

func init() {
	OC_names[0x95C1] = "ANDROID_GET_PARTIAL_OBJECT64"
	OC_names[0x95C2] = "ANDROID_SEND_PARTIAL_OBJECT"
	OC_names[0x95C3] = "ANDROID_TRUNCATE_OBJECT"
	OC_names[0x95C4] = "ANDROID_BEGIN_EDIT_OBJECT"
	OC_names[0x95C5] = "ANDROID_END_EDIT_OBJECT"
}

func (d *Device) AndroidGetPartialObject64(handle uint32, w io.Writer, offset int64, size uint32) error {
	var req, rep Container
	req.Code = OC_ANDROID_GET_PARTIAL_OBJECT64
	req.Param = []uint32{handle, uint32(offset & 0xFFFFFFFF), uint32(offset >> 32), size}
	return d.RunTransaction(&req, &rep, w, nil, 0)
}

func (d *Device) AndroidBeginEditObject(handle uint32) error {
	var req, rep Container

	req.Code = OC_ANDROID_BEGIN_EDIT_OBJECT
	req.Param = []uint32{handle}
	return d.RunTransaction(&req, &rep, nil, nil, 0)
}

func (d *Device) AndroidTruncate(handle uint32, offset int64) error {
	var req, rep Container

	req.Code = OC_ANDROID_TRUNCATE_OBJECT
	req.Param = []uint32{handle, uint32(offset & 0xFFFFFFFF), uint32(offset >> 32)}
	return d.RunTransaction(&req, &rep, nil, nil, 0)
}

func (d *Device) AndroidSendPartialObject(handle uint32, offset int64, size uint32, r io.Reader) error {
	var req, rep Container

	req.Code = OC_ANDROID_SEND_PARTIAL_OBJECT
	req.Param = []uint32{handle, uint32(offset & 0xFFFFFFFF), uint32(offset >> 32), size}

	// MtpServer.cpp is buggy: it uses write() without offset
	// rather than pwrite to send the data for data coming with
	// the header packet
	d.SeparateHeader = true
	err := d.RunTransaction(&req, &rep, nil, r, int64(size))
	d.SeparateHeader = false
	return err
}

func (d *Device) AndroidEndEditObject(handle uint32) error {
	var req, rep Container

	req.Code = OC_ANDROID_END_EDIT_OBJECT
	req.Param = []uint32{handle}
	return d.RunTransaction(&req, &rep, nil, nil, 0)
}
