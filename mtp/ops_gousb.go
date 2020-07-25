package mtp

import (
	"bytes"
	"fmt"
	"math/rand"
)

func (d *DeviceGoUSB) OpenSession() error {
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
func (d *DeviceGoUSB) CloseSession() error {
	var req, rep Container
	req.Code = OC_CloseSession
	err := d.RunTransaction(&req, &rep, nil, nil, 0)
	d.session = nil
	return err
}

func (d *DeviceGoUSB) GetData(req *Container, info interface{}) error {
	var buf bytes.Buffer
	var rep Container
	if err := d.RunTransaction(req, &rep, &buf, nil, 0); err != nil {
		return err
	}
	err := Decode(&buf, info)
	if d.Debug.MTP && err == nil {
		log.WithField("prefix", "mtp").Debugf("MTP decoded %#v", info)
	}
	return err
}

func (d *DeviceGoUSB) GetDeviceInfo(info *DeviceInfo) error {
	var req Container
	req.Code = OC_GetDeviceInfo
	return d.GetData(&req, info)
}

func (d *DeviceGoUSB) GetDevicePropValue(propCode uint32, dest interface{}) error {
	var req Container
	req.Code = OC_GetDevicePropValue
	req.Param = []uint32{propCode}
	return d.GetData(&req, dest)
}
