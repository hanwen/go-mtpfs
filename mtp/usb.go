package mtp

import "github.com/google/gousb"

func (d *DeviceGoUSB) bulkTransferIn(ep *gousb.InEndpoint, buf []byte) (int, error) {
	return ep.Read(buf)
}

func (d *DeviceGoUSB) bulkTransferOut(ep *gousb.OutEndpoint, buf []byte) (int, error) {
	return ep.Write(buf)
}
