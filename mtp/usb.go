package mtp

import "github.com/google/gousb"

func (d *Device2) bulkTransferIn(ep *gousb.InEndpoint, buf []byte) (int, error) {
	return ep.Read(buf)
}

func (d *Device2) bulkTransferOut(ep *gousb.OutEndpoint, buf []byte) (int, error) {
	return ep.Write(buf)
}
