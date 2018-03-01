package mtp

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math/rand"
	"time"
)

var _ = log.Println

func init() {
	rand.Seed(time.Now().UnixNano())
}

// OpenSession opens a session, which is necesary for any command that
// queries or modifies storage. It is an error to open a session
// twice.  If OpenSession() fails, it will not attempt to close the device.
func (d *Device) OpenSession() error {
	if d.session != nil {
		return fmt.Errorf("session already open")
	}
	var req, rep Container
	req.Code = OC_OpenSession

	// avoid 0xFFFFFFFF and 0x00000000 for session IDs.
	sid := uint32(rand.Int31()) | 1
	req.Param = []uint32{sid} // session

	// If opening the session fails, we want to be able to reset
	// the device, so don't do sanity checks afterwards.
	if err := d.runTransaction(&req, &rep, nil, nil, 0); err != nil {
		return err
	}

	d.session = &sessionData{
		tid: 1,
		sid: sid,
	}
	return nil
}

// Closes a sessions. This is done automatically if the device is closed.
func (d *Device) CloseSession() error {
	var req, rep Container
	req.Code = OC_CloseSession
	err := d.RunTransaction(&req, &rep, nil, nil, 0)
	d.session = nil
	return err
}

func (d *Device) GetData(req *Container, info interface{}) error {
	var buf bytes.Buffer
	var rep Container
	if err := d.RunTransaction(req, &rep, &buf, nil, 0); err != nil {
		return err
	}
	err := Decode(&buf, info)
	if d.MTPDebug && err == nil {
		log.Printf("MTP decoded %#v", info)
	}
	return err
}

func (d *Device) GetDeviceInfo(info *DeviceInfo) error {
	var req Container
	req.Code = OC_GetDeviceInfo
	return d.GetData(&req, info)
}

func (d *Device) GetStorageIDs(info *Uint32Array) error {
	var req Container
	req.Code = OC_GetStorageIDs
	return d.GetData(&req, info)
}

func (d *Device) GetObjectPropDesc(objPropCode, objFormatCode uint16, info *ObjectPropDesc) error {
	var req Container
	req.Code = OC_MTP_GetObjectPropDesc
	req.Param = []uint32{uint32(objPropCode), uint32(objFormatCode)}
	return d.GetData(&req, info)
}

func (d *Device) GetObjectPropValue(objHandle uint32, objPropCode uint16, value interface{}) error {
	var req Container
	req.Code = OC_MTP_GetObjectPropValue
	req.Param = []uint32{objHandle, uint32(objPropCode)}
	return d.GetData(&req, value)
}

func (d *Device) SetObjectPropValue(objHandle uint32, objPropCode uint16, value interface{}) error {
	var req, rep Container
	req.Code = OC_MTP_SetObjectPropValue
	req.Param = []uint32{objHandle, uint32(objPropCode)}
	return d.SendData(&req, &rep, value)
}

func (d *Device) SendData(req *Container, rep *Container, value interface{}) error {
	var buf bytes.Buffer
	if err := Encode(&buf, value); err != nil {
		return err
	}
	if d.MTPDebug {
		log.Printf("MTP encoded %#v", value)
	}
	return d.RunTransaction(req, rep, nil, &buf, int64(buf.Len()))
}

func (d *Device) GetObjectPropsSupported(objFormatCode uint16, props *Uint16Array) error {
	var req Container

	req.Code = OC_MTP_GetObjectPropsSupported
	req.Param = []uint32{uint32(objFormatCode)}
	return d.GetData(&req, props)
}

func (d *Device) GetDevicePropDesc(propCode uint16, info *DevicePropDesc) error {
	var req Container
	req.Code = OC_GetDevicePropDesc
	req.Param = append(req.Param, uint32(propCode))
	return d.GetData(&req, info)
}

func (d *Device) SetDevicePropValue(propCode uint32, src interface{}) error {
	var req, rep Container
	req.Code = OC_SetDevicePropValue
	req.Param = []uint32{propCode}
	return d.SendData(&req, &rep, src)
}

func (d *Device) GetDevicePropValue(propCode uint32, dest interface{}) error {
	var req Container
	req.Code = OC_GetDevicePropValue
	req.Param = []uint32{propCode}
	return d.GetData(&req, dest)
}

func (d *Device) ResetDevicePropValue(propCode uint32) error {
	var req, rep Container
	req.Code = OC_ResetDevicePropValue
	req.Param = []uint32{propCode}
	return d.RunTransaction(&req, &rep, nil, nil, 0)
}

func (d *Device) GetStorageInfo(ID uint32, info *StorageInfo) error {
	var req Container
	req.Code = OC_GetStorageInfo
	req.Param = []uint32{ID}
	return d.GetData(&req, info)
}

func (d *Device) GetObjectHandles(storageID, objFormatCode, parent uint32, info *Uint32Array) error {
	var req Container
	req.Code = OC_GetObjectHandles
	req.Param = []uint32{storageID, objFormatCode, parent}
	return d.GetData(&req, info)
}

func (d *Device) GetObjectInfo(handle uint32, info *ObjectInfo) error {
	var req Container
	req.Code = OC_GetObjectInfo
	req.Param = []uint32{handle}
	return d.GetData(&req, info)
}

func (d *Device) GetNumObjects(storageId uint32, formatCode uint16, parent uint32) (uint32, error) {
	var req, rep Container
	req.Code = OC_GetNumObjects
	req.Param = []uint32{storageId, uint32(formatCode), parent}
	if err := d.RunTransaction(&req, &rep, nil, nil, 0); err != nil {
		return 0, err
	}
	return rep.Param[0], nil
}

func (d *Device) DeleteObject(handle uint32) error {
	var req, rep Container
	req.Code = OC_DeleteObject
	req.Param = []uint32{handle, 0x0}

	return d.RunTransaction(&req, &rep, nil, nil, 0)
}

func (d *Device) SendObjectInfo(wantStorageID, wantParent uint32, info *ObjectInfo) (storageID, parent, handle uint32, err error) {
	var req, rep Container
	req.Code = OC_SendObjectInfo
	req.Param = []uint32{wantStorageID, wantParent}

	if err = d.SendData(&req, &rep, info); err != nil {
		return
	}

	if len(rep.Param) < 3 {
		err = fmt.Errorf("SendObjectInfo: got %v, need 3 response parameters", rep.Param)
		return
	}

	return rep.Param[0], rep.Param[1], rep.Param[2], nil
}

func (d *Device) SendObject(r io.Reader, size int64) error {
	var req, rep Container
	req.Code = OC_SendObject
	return d.RunTransaction(&req, &rep, nil, r, size)
}

func (d *Device) GetObject(handle uint32, w io.Writer) error {
	var req, rep Container
	req.Code = OC_GetObject
	req.Param = []uint32{handle}

	return d.RunTransaction(&req, &rep, w, nil, 0)
}

func (d *Device) GetPartialObject(handle uint32, w io.Writer, offset uint32, size uint32) error {
	var req, rep Container
	req.Code = OC_GetPartialObject
	req.Param = []uint32{handle, offset, size}
	return d.RunTransaction(&req, &rep, w, nil, 0)
}
