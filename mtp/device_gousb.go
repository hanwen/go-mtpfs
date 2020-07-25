package mtp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/google/gousb"

	"github.com/hanwen/usb"
)

// DeviceGoUSB implements mtp.Device.
// It accesses libusb driver via gousb.
type DeviceGoUSB struct {
	dev         *gousb.Device
	devDesc     *gousb.DeviceDesc
	configDesc  gousb.ConfigDesc
	ifaceDesc   gousb.InterfaceDesc
	sendEPDesc  gousb.EndpointDesc
	fetchEPDesc gousb.EndpointDesc
	eventEPDesc gousb.EndpointDesc

	iConfiguration int
	iInterface     int
	iAltSetting    int

	config  gousb.Config
	iface   gousb.Interface
	sendEP  *gousb.OutEndpoint
	fetchEP *gousb.InEndpoint
	eventEP *gousb.InEndpoint

	session *sessionData

	Debug struct {
		MTP  bool
		USB  bool
		Data bool
	}
}

func (d *DeviceGoUSB) connected() bool {
	return d.sendEP != nil
}

// Close releases the interface, and closes the device.
func (d *DeviceGoUSB) Close() error {
	if !d.connected() {
		return nil // or error?
	}

	if d.session != nil {
		var req, rep Container
		req.Code = OC_CloseSession
		// RunTransaction runs close, so can't use CloseSession().

		err := d.runTransaction(&req, &rep, nil, nil, 0)
		if err != nil && d.Debug.USB {
			log.WithField("prefix", "usb").Errorf("failed to close session")
		}
		d.session = nil
	}

	err := d.config.Close()
	if err != nil && d.Debug.USB {
		log.WithField("prefix", "usb").Errorf("failed to close configuration: %s", err)
	}
	d.iface.Close()

	d.sendEP = nil
	d.fetchEP = nil
	d.eventEP = nil
	return nil
}

// Open opens an MTP device.
func (d *DeviceGoUSB) Open() error {
	// Unusual closing...
	cfg, err := d.dev.Config(d.iConfiguration)
	if err != nil {
		return fmt.Errorf("failed to open configuration: %s", err)
	}

	iface, err := cfg.Interface(d.iInterface, d.iAltSetting)
	if err != nil {
		cfg.Close()
		return fmt.Errorf("failed to open interface: %s", err)
	}

	d.sendEP, err = iface.OutEndpoint(int(d.sendEPDesc.Address))
	if err != nil {
		cfg.Close()
		iface.Close()
		return fmt.Errorf("failed to open send EP: %s", err)
	}

	d.fetchEP, err = iface.InEndpoint(int(d.fetchEPDesc.Address))
	if err != nil {
		cfg.Close()
		iface.Close()
		return fmt.Errorf("failed to open fetch EP: %s", err)
	}

	d.eventEP, err = iface.InEndpoint(int(d.eventEPDesc.Number))
	if err != nil {
		cfg.Close()
		iface.Close()
		return fmt.Errorf("failed to open event EP: %s", err)
	}

	info := DeviceInfo{}
	err = d.GetDeviceInfo(&info)

	// Some of the win8phones have no interface field.
	if len(d.ifaceDesc.AltSettings) == 0 {
		info := DeviceInfo{}
		err = d.GetDeviceInfo(&info)
		if err != nil && d.Debug.USB {
			log.WithField("prefix", "usb").Errorf("failed to get device info: %s", err)
		}

		if !strings.Contains(info.MTPExtension, "microsoft") {
			err = d.Close()
			if err != nil && d.Debug.USB {
				log.WithField("prefix", "usb").Errorf("failed to close device: %s", err)
			}
			return fmt.Errorf("no MTP extensions in %s", info.MTPExtension)
		}
	} else {
		if iface.Setting.Class != gousb.ClassPTP {
			err = d.Close()
			if err != nil && d.Debug.USB {
				log.WithField("prefix", "usb").Errorf("failed to close device: %s", err)
			}
			return fmt.Errorf("interface has no MTP/PTP/Image class")
		}
	}

	return nil
}

func (d *DeviceGoUSB) sendReq(req *Container) error {
	c := usbBulkContainer{
		usbBulkHeader: usbBulkHeader{
			Length:        uint32(usbHdrLen + 4*len(req.Param)),
			Type:          USB_CONTAINER_COMMAND,
			Code:          req.Code,
			TransactionID: req.TransactionID,
		},
	}
	for i := range req.Param {
		c.Param[i] = req.Param[i]
	}

	var wData [usbBulkLen]byte
	buf := bytes.NewBuffer(wData[:0])

	binary.Write(buf, binary.LittleEndian, c.usbBulkHeader)
	if err := binary.Write(buf, binary.LittleEndian, c.Param[:len(req.Param)]); err != nil {
		panic(err)
	}

	d.dataPrint(d.sendEPDesc, buf.Bytes())
	_, err := d.bulkTransferOut(d.sendEP, buf.Bytes())
	if err != nil {
		return err
	}
	return nil
}

// Fetches one USB packet. The header is split off, and the remainder is returned.
// dest should be at least 512bytes.
func (d *DeviceGoUSB) fetchPacket(dest []byte, header *usbBulkHeader) (rest []byte, err error) {
	n, err := d.bulkTransferIn(d.fetchEP, dest[:d.fetchEPDesc.MaxPacketSize])
	if n > 0 {
		d.dataPrint(d.fetchEPDesc, dest[:n])
	}

	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(dest[:n])
	if err = binary.Read(buf, binary.LittleEndian, header); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (d *DeviceGoUSB) decodeRep(h *usbBulkHeader, rest []byte, rep *Container) error {
	if h.Type != USB_CONTAINER_RESPONSE {
		return SyncError(fmt.Sprintf("got type %d (%s) in response, want CONTAINER_RESPONSE.", h.Type, USB_names[int(h.Type)]))
	}

	rep.Code = h.Code
	rep.TransactionID = h.TransactionID

	restLen := int(h.Length) - usbHdrLen
	if restLen > len(rest) {
		return fmt.Errorf("header specified 0x%x bytes, but have 0x%x",
			restLen, len(rest))
	}
	nParam := restLen / 4
	for i := 0; i < nParam; i++ {
		rep.Param = append(rep.Param, byteOrder.Uint32(rest[4*i:]))
	}

	if rep.Code != RC_OK {
		return RCError(rep.Code)
	}
	return nil
}

func (d *DeviceGoUSB) RunTransactionWithNoParams(code uint16) error {
	var req, rep Container
	req.Code = code
	req.Param = []uint32{}
	return d.RunTransaction(&req, &rep, nil, nil, 0)
}

// Runs a single MTP transaction. dest and src cannot be specified at
// the same time.  The request should fill out Code and Param as
// necessary. The response is provided here, but usually only the
// return code is of interest.  If the return code is an error, this
// function will return an RCError instance.
//
// Errors that are likely to affect future transactions lead to
// closing the connection. Such errors include: invalid transaction
// IDs, USB errors (BUSY, IO, ACCESS etc.), and receiving data for
// operations that expect no data.
func (d *DeviceGoUSB) RunTransaction(req *Container, rep *Container,
	dest io.Writer, src io.Reader, writeSize int64) error {
	if err := d.runTransaction(req, rep, dest, src, writeSize); err != nil {
		_, ok2 := err.(SyncError)
		_, ok1 := err.(usb.Error)
		if ok1 || ok2 {
			return Catastrophic(fmt.Sprintf("fatal error: %s", err))
		}
		return err
	}
	return nil
}

// runTransaction is like RunTransaction, but without sanity checking
// before and after the call.
func (d *DeviceGoUSB) runTransaction(req *Container, rep *Container,
	dest io.Writer, src io.Reader, writeSize int64) error {
	var finalPacket []byte
	if d.session != nil {
		req.SessionID = d.session.sid
		req.TransactionID = d.session.tid
		d.session.tid++
	}

	if d.Debug.MTP {
		log.Printf("MTP request %s %v\n", OC_names[int(req.Code)], req.Param)
	}

	if err := d.sendReq(req); err != nil {
		if d.Debug.MTP {
			log.Printf("MTP sendreq failed: %v\n", err)
		}
		return err
	}

	if src != nil {
		hdr := usbBulkHeader{
			Type:          USB_CONTAINER_DATA,
			Code:          req.Code,
			Length:        uint32(writeSize),
			TransactionID: req.TransactionID,
		}

		_, err := d.bulkWrite(&hdr, src, writeSize)
		if err != nil {
			return err
		}
	}
	fetchPacketSize := d.fetchEPDesc.MaxPacketSize
	data := make([]byte, fetchPacketSize)
	h := &usbBulkHeader{}
	rest, err := d.fetchPacket(data[:], h)
	if err != nil {
		return err
	}
	var unexpectedData bool
	if h.Type == USB_CONTAINER_DATA {
		if dest == nil {
			dest = &NullWriter{}
			unexpectedData = true
			if d.Debug.MTP {
				log.Printf("MTP discarding unexpected data 0x%x bytes", h.Length)
			}
		}
		if d.Debug.MTP {
			log.Printf("MTP data 0x%x bytes", h.Length)
		}

		dest.Write(rest)

		if len(rest)+usbHdrLen == fetchPacketSize {
			// If this was a full packet, read until we
			// have a short read.
			_, finalPacket, err = d.bulkRead(dest)
			if err != nil {
				return err
			}
		}

		h = &usbBulkHeader{}
		if len(finalPacket) > 0 {
			if d.Debug.MTP {
				log.Printf("Reusing final packet")
			}
			rest = finalPacket
			finalBuf := bytes.NewBuffer(finalPacket[:len(finalPacket)])
			err = binary.Read(finalBuf, binary.LittleEndian, h)
		} else {
			rest, err = d.fetchPacket(data[:], h)
		}
	}

	err = d.decodeRep(h, rest, rep)
	if d.Debug.MTP {
		log.Printf("MTP response %s %v", getName(RC_names, int(rep.Code)), rep.Param)
	}
	if unexpectedData {
		return SyncError(fmt.Sprintf("unexpected data for code %s", getName(RC_names, int(req.Code))))
	}

	if err != nil {
		return err
	}
	if d.session != nil && rep.TransactionID != req.TransactionID {
		return SyncError(fmt.Sprintf("transaction ID mismatch got %x want %x",
			rep.TransactionID, req.TransactionID))
	}
	rep.SessionID = req.SessionID
	return nil
}

// Prints data going over the USB connection.
func (d *DeviceGoUSB) dataPrint(epDesc gousb.EndpointDesc, data []byte) {
	if !d.Debug.Data {
		return
	}
	ep := uint8(epDesc.Address)
	dir := "send"
	if 0 != ep&usb.ENDPOINT_IN {
		dir = "recv"
	}
	fmt.Fprintf(os.Stderr, "%s: 0x%x bytes with ep 0x%x:\n", dir, len(data), ep)
	hexDump(data)
}

// bulkWrite returns the number of non-header bytes written.
func (d *DeviceGoUSB) bulkWrite(hdr *usbBulkHeader, r io.Reader, size int64) (n int64, err error) {
	packetSize := d.sendEPDesc.MaxPacketSize
	if hdr != nil {
		if size+usbHdrLen > 0xFFFFFFFF {
			hdr.Length = 0xFFFFFFFF
		} else {
			hdr.Length = uint32(size + usbHdrLen)
		}

		packetArr := make([]byte, packetSize)
		var packet []byte
		packet = packetArr[:]

		buf := bytes.NewBuffer(packet[:0])
		binary.Write(buf, byteOrder, hdr)
		cpSize := int64(len(packet) - usbHdrLen)
		if cpSize > size {
			cpSize = size
		}

		_, err = io.CopyN(buf, r, cpSize)
		d.dataPrint(d.sendEPDesc, buf.Bytes())
		_, err = d.bulkTransferOut(d.sendEP, buf.Bytes())
		if err != nil {
			return cpSize, err
		}
		size -= cpSize
		n += cpSize
	}

	var buf [rwBufSize]byte
	var lastTransfer int
	for size > 0 {
		var m int
		toread := buf[:]
		if int64(len(toread)) > size {
			toread = buf[:int(size)]
		}

		m, err = r.Read(toread)
		if err != nil {
			break
		}
		size -= int64(m)

		d.dataPrint(d.sendEPDesc, buf[:m])
		lastTransfer, err = d.bulkTransferOut(d.sendEP, buf[:m])
		n += int64(lastTransfer)

		if err != nil || lastTransfer == 0 {
			break
		}
	}
	if lastTransfer%packetSize == 0 {
		// write a short packet just to be sure.
		d.bulkTransferOut(d.sendEP, buf[:0])
	}

	return n, err
}

func (d *DeviceGoUSB) bulkRead(w io.Writer) (n int64, lastPacket []byte, err error) {
	var buf [rwBufSize]byte
	var lastRead int
	for {
		toread := buf[:]
		lastRead, err = d.bulkTransferIn(d.fetchEP, toread)
		if err != nil {
			break
		}
		if lastRead > 0 {
			d.dataPrint(d.fetchEPDesc, buf[:lastRead])

			w, err := w.Write(buf[:lastRead])
			n += int64(w)
			if err != nil {
				break
			}
		}
		if d.Debug.MTP {
			log.Printf("MTP bulk read 0x%x bytes.", lastRead)
		}
		if lastRead < len(toread) {
			// short read.
			break
		}
	}
	packetSize := d.fetchEPDesc.MaxPacketSize
	if lastRead%packetSize == 0 {
		// This should be a null packet, but on Linux + XHCI it's actually
		// CONTAINER_OK instead. To be liberal with the XHCI behavior, return
		// the final packet and inspect it in the calling function.
		var nullReadSize int
		nullReadSize, err = d.bulkTransferIn(d.fetchEP, buf[:])
		if d.Debug.MTP {
			log.Printf("Expected null packet, read %d bytes", nullReadSize)
		}
		return n, buf[:nullReadSize], err
	}
	return n, buf[:0], err
}

func (d *DeviceGoUSB) bulkTransferIn(ep *gousb.InEndpoint, buf []byte) (int, error) {
	return ep.Read(buf)
}

func (d *DeviceGoUSB) bulkTransferOut(ep *gousb.OutEndpoint, buf []byte) (int, error) {
	return ep.Write(buf)
}

// Configure is a robust version of OpenSession. On failure, it resets
// the device and reopens the device and the session.
func (d *DeviceGoUSB) Configure() error {
	if err := d.Open(); err != nil {
		return err
	}

	err := d.OpenSession()
	if err == RCError(RC_SessionAlreadyOpened) {
		// It's open, so close the session. Fortunately, this
		// even works without a transaction ID, at least on Android.
		d.CloseSession()
		err = d.OpenSession()
	}

	if err != nil {
		log.Printf("OpenSession failed: %v; attempting reset", err)
		d.Close()

		// Give the device some rest.
		time.Sleep(1000 * time.Millisecond)
		if err := d.Open(); err != nil {
			return fmt.Errorf("opening after reset: %v", err)
		}
		if err := d.OpenSession(); err != nil {
			return fmt.Errorf("OpenSession after reset: %v", err)
		}
	}
	return nil
}
