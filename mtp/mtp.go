package mtp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/hanwen/go-mtpfs/usb"
)

// An MTP device.
type Device struct {
	h   *usb.DeviceHandle
	dev *usb.Device

	claimed bool

	// split off descriptor?
	devDescr    usb.DeviceDescriptor
	ifaceDescr  usb.InterfaceDescriptor
	sendEp      byte
	fetchEp     byte
	eventEp     byte
	configIndex byte

	// In milliseconds. Defaults to 2 seconds.
	Timeout int

	// Print request/response codes.
	DebugPrint bool

	// Print data as it passes over the USB connection.
	DataPrint bool

	// If set, send header in separate write.
	SeparateHeader bool

	session *sessionData
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

// Close releases the interface, and closes the device.
func (d *Device) Close() error {
	if d.h == nil {
		return nil // or error?
	}

	if d.session != nil {
		err := d.CloseSession()
		if err != nil {
 			d.h.Reset()
		}
	}
	
	if d.claimed {
		d.h.ReleaseInterface(d.ifaceDescr.InterfaceNumber)
	}
	err := d.h.Close()
	d.h = nil
	return err
}

// Done releases the libusb device reference.
func (d *Device) Done() {
	d.dev.Unref()
	d.dev = nil
}

// Open opens an MTP device.
func (d *Device) Open() error {
	if d.Timeout == 0 {
		d.Timeout = 2000
	}

	if d.h != nil {
		return fmt.Errorf("already open")
	}

	var err error
	d.h, err = d.dev.Open()
	if err != nil {
		return err
	}

	iface, err := d.h.GetStringDescriptorASCII(d.ifaceDescr.InterfaceStringIndex)
	if err != nil {
		d.Close()
		return err
	}

	if !strings.Contains(iface, "MTP") {
		d.Close()
		return fmt.Errorf("has no MTP in interface string")
	}
	return nil
}

// Id is the manufacturer + product + serial
func (d *Device) Id() (string, error) {
	if d.h == nil {
		return "", fmt.Errorf("device not open")
	}

	var ids []string
	for _, b := range []byte{
		d.devDescr.Manufacturer,
		d.devDescr.Product,
		d.devDescr.SerialNumber} {
		s, err := d.h.GetStringDescriptorASCII(b)
		if err != nil {
			return "", err
		}
		ids = append(ids, s)
	}

	return strings.Join(ids, " "), nil
}

func (d *Device) sendReq(req *Container) error {
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

	err := binary.Write(buf, binary.LittleEndian, c)
	if err != nil {
		panic(err)
	}

	d.dataPrint(d.sendEp, buf.Bytes())
	_, err = d.h.BulkTransfer(d.sendEp, buf.Bytes(), d.Timeout)
	if err != nil {
		return err
	}
	return nil
}

const packetSize = 512

// Fetches one USB packet. The header is split off, and the remainder is returned.
// dest should be at least 512bytes.
func (d *Device) fetchPacket(dest []byte, header *usbBulkHeader) (rest []byte, err error) {
	n, err := d.h.BulkTransfer(d.fetchEp, dest[:packetSize], d.Timeout)
	if n > 0 {
		d.dataPrint(d.fetchEp, dest[:n])
	}

	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(dest[:n])
	err = binary.Read(buf, binary.LittleEndian, header)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (d *Device) decodeRep(h *usbBulkHeader, rest []byte, rep *Container) error {
	if h.Type != USB_CONTAINER_RESPONSE {
		return fmt.Errorf("got type %d in response", h.Type)
	}

	rep.Code = h.Code
	if rep.Code != RC_OK {
		return RCError(rep.Code)
	}
	rep.TransactionID = h.TransactionID

	restLen := int(h.Length) - usbHdrLen
	if restLen != len(rest) {
		return fmt.Errorf("header specified 0x%x bytes, but have 0x%x",
			restLen, len(rest))
	}
	nParam := restLen / 4
	for i := 0; i < nParam; i++ {
		rep.Param = append(rep.Param, byteOrder.Uint32(rest[4*i:]))
	}

	return nil
}

// Runs a single MTP transaction. dest and src cannot be specified at
// the same time.  The request should fill out Code and Param as
// necessary. The response is provided here, but usually only the
// return code is of interest.  If the return code is an error, this
// function will return an RCError instance.
func (d *Device) RunTransaction(req *Container, rep *Container,
	dest io.Writer, src io.Reader, writeSize int64) error {
	if d.session != nil {
		req.SessionID = d.session.sid
		req.TransactionID = d.session.tid
		d.session.tid++
	}

	if d.DebugPrint {
		log.Printf("MTP request %s %v\n", OC_names[int(req.Code)], req.Param)
	}

	err := d.sendReq(req)
	if err != nil {
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
			d.Close()
			return err
		}
	}
	var data [packetSize]byte
	h := &usbBulkHeader{}
	rest, err := d.fetchPacket(data[:], h)
	if err != nil {
		return err
	}
	if h.Type == USB_CONTAINER_DATA {
		if dest == nil {
			d.Close()
			return fmt.Errorf("no sink for data")
		}
		if d.DebugPrint {
			log.Printf("MTP data 0x%x bytes", h.Length)
		}

		size := int(h.Length)
		dest.Write(rest)
		size -= len(rest) + usbHdrLen
		if size > 0 {
			_, err = d.bulkRead(dest)
			if err != nil {
				return err
			}
		}

		h = &usbBulkHeader{}
		rest, err = d.fetchPacket(data[:], h)
	}

	err = d.decodeRep(h, rest, rep)
	if d.DebugPrint {
		log.Printf("MTP response %s %v", RC_names[int(rep.Code)], rep.Param)
	}
	if err != nil {
		return err
	}
	if d.session != nil && rep.TransactionID != req.TransactionID {
		return fmt.Errorf("transaction ID mismatch got %x want %x",
			rep.TransactionID, req.TransactionID)
	}
	rep.SessionID = req.SessionID
	return nil
}

// Prints data going over the USB connection.
func (d *Device) dataPrint(ep byte, data []byte) {
	if !d.DataPrint {
		return
	}
	dir := "send"
	if 0 != ep&usb.ENDPOINT_IN {
		dir = "recv"
	}
	fmt.Fprintf(os.Stderr, "%s: 0x%x bytes with ep 0x%x:\n", dir, len(data), ep)
	hexDump(data)
}

// The linux usb stack can send 16kb per call, according to libusb.
const rwBufSize = 0x4000

// bulkWrite returns the number of non-header bytes written.
func (d *Device) bulkWrite(hdr *usbBulkHeader, r io.Reader, size int64) (n int64, err error) {
	if hdr != nil {
		if size+usbHdrLen > 0xFFFFFFFF {
			hdr.Length = 0xFFFFFFFF
		} else {
			hdr.Length = uint32(size + usbHdrLen)
		}

		var packetArr [packetSize]byte
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
		d.dataPrint(d.sendEp, buf.Bytes())
		_, err = d.h.BulkTransfer(d.sendEp, buf.Bytes(), d.Timeout)
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

		d.dataPrint(d.sendEp, buf[:m])
		lastTransfer, err = d.h.BulkTransfer(d.sendEp, buf[:m], d.Timeout)
		n += int64(lastTransfer)

		if err != nil || lastTransfer == 0 {
			break
		}
	}
	if lastTransfer%packetSize == 0 {
		// write a short packet just to be sure.
		d.h.BulkTransfer(d.sendEp, buf[:0], d.Timeout)
	}

	return n, err
}

func (d *Device) bulkRead(w io.Writer) (n int64, err error) {
	var buf [rwBufSize]byte
	var lastRead int
	for {
		toread := buf[:]
		lastRead, err = d.h.BulkTransfer(d.fetchEp, toread, d.Timeout)
		if err != nil {
			break
		}
		if lastRead > 0 {
			d.dataPrint(d.fetchEp, buf[:lastRead])

			w, err := w.Write(buf[:lastRead])
			n += int64(w)
			if err != nil {
				break
			}
		}
		if lastRead < len(toread) {
			// short read.
			break
		}
	}
	if lastRead%packetSize == 0 {
		// Null read.
		d.h.BulkTransfer(d.fetchEp, buf[:0], d.Timeout)
	}
	return n, err
}
