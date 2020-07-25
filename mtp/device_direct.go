package mtp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/hanwen/usb"
)

// DeviceDirect implements mtp.Device.
// It accesses libusb driver via hanwen/usb.
type DeviceDirect struct {
	h   *usb.DeviceHandle
	dev *usb.Device

	claimed bool

	// split off descriptor?
	devDescr    usb.DeviceDescriptor
	ifaceDescr  usb.InterfaceDescriptor
	sendEP      byte
	fetchEP     byte
	eventEP     byte
	configValue byte

	// In milliseconds. Defaults to 2 seconds.
	Timeout int

	// Print request/response codes.
	MTPDebug bool

	// Print USB calls.
	USBDebug bool

	// Print data as it passes over the USB connection.
	DataDebug bool

	// If set, send header in separate write.
	SeparateHeader bool

	session *sessionData
}

func (d *DeviceDirect) fetchMaxPacketSize() int {
	return d.dev.GetMaxPacketSize(d.fetchEP)
}

func (d *DeviceDirect) sendMaxPacketSize() int {
	return d.dev.GetMaxPacketSize(d.sendEP)
}

// Close releases the interface, and closes the device.
func (d *DeviceDirect) Close() error {
	if d.h == nil {
		return nil // or error?
	}

	if d.session != nil {
		var req, rep Container
		req.Code = OC_CloseSession
		// RunTransaction runs close, so can't use CloseSession().

		if err := d.runTransaction(&req, &rep, nil, nil, 0); err != nil {
			err := d.h.Reset()
			if d.USBDebug {
				log.Printf("USB: Reset, err: %v", err)
			}
		}
	}

	if d.claimed {
		err := d.h.ReleaseInterface(d.ifaceDescr.InterfaceNumber)
		if d.USBDebug {
			log.Printf("USB: ReleaseInterface 0x%x, err: %v", d.ifaceDescr.InterfaceNumber, err)
		}
	}
	err := d.h.Close()
	d.h = nil

	if d.USBDebug {
		log.Printf("USB: Close, err: %v", err)
	}
	return err
}

// Done releases the libusb device reference.
func (d *DeviceDirect) Done() {
	d.dev.Unref()
	d.dev = nil
}

// Claims the USB interface of the device.
func (d *DeviceDirect) claim() error {
	if d.h == nil {
		return fmt.Errorf("mtp: claim: device not open")
	}

	err := d.h.ClaimInterface(d.ifaceDescr.InterfaceNumber)
	if d.USBDebug {
		log.Printf("USB: ClaimInterface 0x%x, err: %v", d.ifaceDescr.InterfaceNumber, err)
	}
	if err == nil {
		d.claimed = true
	}

	return err
}

// Open opens an MTP device.
func (d *DeviceDirect) Open() error {
	if d.Timeout == 0 {
		d.Timeout = 2000
	}

	if d.h != nil {
		return fmt.Errorf("already open")
	}

	var err error
	d.h, err = d.dev.Open()
	if d.USBDebug {
		log.Printf("USB: Open, err: %v", err)
	}
	if err != nil {
		return err
	}

	d.claim()
	if d.ifaceDescr.InterfaceStringIndex == 0 {
		// Some of the win8phones have no interface field.
		info := DeviceInfo{}
		d.GetDeviceInfo(&info)

		if !strings.Contains(info.MTPExtension, "microsoft") {
			d.Close()
			return fmt.Errorf("mtp: no MTP extensions in %s", info.MTPExtension)
		}
	} else {
		iface, err := d.h.GetStringDescriptorASCII(d.ifaceDescr.InterfaceStringIndex)
		if err != nil {
			d.Close()
			return err
		}

		if !strings.Contains(iface, "MTP") {
			d.Close()
			return fmt.Errorf("has no MTP in interface string")
		}
	}

	return nil
}

// ID is the manufacturer + product + serial
func (d *DeviceDirect) ID() (string, error) {
	if d.h == nil {
		return "", fmt.Errorf("mtp: ID: device not open")
	}

	var ids []string
	for _, b := range []byte{
		d.devDescr.Manufacturer,
		d.devDescr.Product,
		d.devDescr.SerialNumber} {
		s, err := d.h.GetStringDescriptorASCII(b)
		if err != nil {
			if d.USBDebug {
				log.Printf("USB: GetStringDescriptorASCII, err: %v", err)
			}
			return "", err
		}
		ids = append(ids, s)
	}

	return strings.Join(ids, " "), nil
}

func (d *DeviceDirect) sendReq(req *Container) error {
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

	d.dataPrint(d.sendEP, buf.Bytes())
	_, err := d.h.BulkTransfer(d.sendEP, buf.Bytes(), d.Timeout)
	if err != nil {
		return err
	}
	return nil
}

// Fetches one USB packet. The header is split off, and the remainder is returned.
// dest should be at least 512bytes.
func (d *DeviceDirect) fetchPacket(dest []byte, header *usbBulkHeader) (rest []byte, err error) {
	n, err := d.h.BulkTransfer(d.fetchEP, dest[:d.fetchMaxPacketSize()], d.Timeout)
	if n > 0 {
		d.dataPrint(d.fetchEP, dest[:n])
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

func (d *DeviceDirect) decodeRep(h *usbBulkHeader, rest []byte, rep *Container) error {
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

func (d *DeviceDirect) RunTransactionWithNoParams(code uint16) error {
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
func (d *DeviceDirect) RunTransaction(req *Container, rep *Container,
	dest io.Writer, src io.Reader, writeSize int64) error {
	if d.h == nil {
		return fmt.Errorf("mtp: cannot run operation %v, device is not open",
			OC_names[int(req.Code)])
	}
	if err := d.runTransaction(req, rep, dest, src, writeSize); err != nil {
		_, ok2 := err.(SyncError)
		_, ok1 := err.(usb.Error)
		if ok1 || ok2 {
			log.Printf("fatal error %v; closing connection.", err)
			d.Close()
		}
		return err
	}
	return nil
}

// runTransaction is like RunTransaction, but without sanity checking
// before and after the call.
func (d *DeviceDirect) runTransaction(req *Container, rep *Container,
	dest io.Writer, src io.Reader, writeSize int64) error {
	var finalPacket []byte
	if d.session != nil {
		req.SessionID = d.session.sid
		req.TransactionID = d.session.tid
		d.session.tid++
	}

	if d.MTPDebug {
		log.Printf("MTP request %s %v\n", OC_names[int(req.Code)], req.Param)
	}

	if err := d.sendReq(req); err != nil {
		if d.MTPDebug {
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
	fetchPacketSize := d.fetchMaxPacketSize()
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
			if d.MTPDebug {
				log.Printf("MTP discarding unexpected data 0x%x bytes", h.Length)
			}
		}
		if d.MTPDebug {
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
			if d.MTPDebug {
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
	if d.MTPDebug {
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
func (d *DeviceDirect) dataPrint(ep byte, data []byte) {
	if !d.DataDebug {
		return
	}
	dir := "send"
	if 0 != ep&usb.ENDPOINT_IN {
		dir = "recv"
	}
	fmt.Fprintf(os.Stderr, "%s: 0x%x bytes with ep 0x%x:\n", dir, len(data), ep)
	hexDump(data)
}

// bulkWrite returns the number of non-header bytes written.
func (d *DeviceDirect) bulkWrite(hdr *usbBulkHeader, r io.Reader, size int64) (n int64, err error) {
	packetSize := d.sendMaxPacketSize()
	if hdr != nil {
		if size+usbHdrLen > 0xFFFFFFFF {
			hdr.Length = 0xFFFFFFFF
		} else {
			hdr.Length = uint32(size + usbHdrLen)
		}

		packetArr := make([]byte, packetSize)
		var packet []byte
		if d.SeparateHeader {
			packet = packetArr[:usbHdrLen]
		} else {
			packet = packetArr[:]
		}

		buf := bytes.NewBuffer(packet[:0])
		binary.Write(buf, byteOrder, hdr)
		cpSize := int64(len(packet) - usbHdrLen)
		if cpSize > size {
			cpSize = size
		}

		_, err = io.CopyN(buf, r, cpSize)
		d.dataPrint(d.sendEP, buf.Bytes())
		_, err = d.h.BulkTransfer(d.sendEP, buf.Bytes(), d.Timeout)
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

		d.dataPrint(d.sendEP, buf[:m])
		lastTransfer, err = d.h.BulkTransfer(d.sendEP, buf[:m], d.Timeout)
		n += int64(lastTransfer)

		if err != nil || lastTransfer == 0 {
			break
		}
	}
	if lastTransfer%packetSize == 0 {
		// write a short packet just to be sure.
		d.h.BulkTransfer(d.sendEP, buf[:0], d.Timeout)
	}

	return n, err
}

func (d *DeviceDirect) bulkRead(w io.Writer) (n int64, lastPacket []byte, err error) {
	var buf [rwBufSize]byte
	var lastRead int
	for {
		toread := buf[:]
		lastRead, err = d.h.BulkTransfer(d.fetchEP, toread, d.Timeout)
		if err != nil {
			break
		}
		if lastRead > 0 {
			d.dataPrint(d.fetchEP, buf[:lastRead])

			w, err := w.Write(buf[:lastRead])
			n += int64(w)
			if err != nil {
				break
			}
		}
		if d.MTPDebug {
			log.Printf("MTP bulk read 0x%x bytes.", lastRead)
		}
		if lastRead < len(toread) {
			// short read.
			break
		}
	}
	packetSize := d.fetchMaxPacketSize()
	if lastRead%packetSize == 0 {
		// This should be a null packet, but on Linux + XHCI it's actually
		// CONTAINER_OK instead. To be liberal with the XHCI behavior, return
		// the final packet and inspect it in the calling function.
		var nullReadSize int
		nullReadSize, err = d.h.BulkTransfer(d.fetchEP, buf[:], d.Timeout)
		if d.MTPDebug {
			log.Printf("Expected null packet, read %d bytes", nullReadSize)
		}
		return n, buf[:nullReadSize], err
	}
	return n, buf[:0], err
}

// Configure is a robust version of OpenSession. On failure, it resets
// the device and reopens the device and the session.
func (d *DeviceDirect) Configure() error {
	if d.h == nil {
		if err := d.Open(); err != nil {
			return err
		}
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
		if d.h != nil {
			d.h.Reset()
		}
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
