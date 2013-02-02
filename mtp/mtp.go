package mtp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/hanwen/go-mtpfs/usb"
)

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
	timeout     int

	DebugPrint bool

	// If set, send header in separate write.
	SeparateHeader bool

	session *Session
}

type Session struct {
	tid uint32
	sid uint32
}

type RCError uint16

func (e RCError) Error() string {
	return fmt.Sprintf("RC %s", RC_names[int(e)])
}

func (d *Device) Close() error {
	if d.h == nil {
		return nil // or error?
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

// Open an MTP device.
func (d *Device) Open() error {
	d.timeout = 60000
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
	c := USBBulkContainer{
		USBBulkHeader: USBBulkHeader{
			Length:        uint32(HdrLen + 4*len(req.Param)),
			Type:          USB_CONTAINER_COMMAND,
			Code:          req.Code,
			TransactionID: req.TransactionID,
		},
	}
	for i := range req.Param {
		c.Param[i] = req.Param[i]
	}

	var wData [BulkLen]byte
	buf := bytes.NewBuffer(wData[:0])

	err := binary.Write(buf, binary.LittleEndian, c)
	if err != nil {
		panic(err)
	}

	d.debugPrint(d.sendEp, buf.Bytes())
	_, err = d.h.BulkTransfer(d.sendEp, buf.Bytes(), d.timeout)
	if err != nil {
		return err
	}
	return nil
}

const packetSize = 512

// Fetches one USB packet. The header is split off, and the remainder is returned.
// dest should be at least 512bytes.
func (d *Device) fetchPacket(dest []byte, header *USBBulkHeader) (rest []byte, err error) {
	n, err := d.h.BulkTransfer(d.fetchEp, dest[:packetSize], d.timeout)
	if n > 0 {
		d.debugPrint(d.fetchEp, dest[:n])
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

func (d *Device) decodeRep(h *USBBulkHeader, rest []byte, rep *Container) error {
	if h.Type != USB_CONTAINER_RESPONSE {
		return fmt.Errorf("got type %d in response", h.Type)
	}

	rep.Code = h.Code
	if rep.Code != RC_OK {
		return RCError(rep.Code)
	}
	rep.TransactionID = h.TransactionID

	restLen := int(h.Length) - HdrLen
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

// Runs a single MTP transaction. dest and src cannot be specified at the same time.
func (d *Device) RPC(req *Container, rep *Container,
	dest io.Writer, src io.Reader, writeSize int64) error {
	if d.session != nil {
		req.SessionID = d.session.sid
		req.TransactionID = d.session.tid
		d.session.tid++
	}

	err := d.sendReq(req)
	if err != nil {
		return err
	}

	if src != nil {
		hdr := USBBulkHeader{
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
	h := &USBBulkHeader{}
	rest, err := d.fetchPacket(data[:], h)
	if err != nil {
		return err
	}
	if h.Type == USB_CONTAINER_DATA {
		if dest == nil {
			d.Close()
			return fmt.Errorf("no sink for data")
		}

		size := int(h.Length)
		dest.Write(rest)
		size -= len(rest) + HdrLen
		if size > 0 {
			_, err = d.bulkRead(dest)
			if err != nil {
				return err
			}
		}

		h = &USBBulkHeader{}
		rest, err = d.fetchPacket(data[:], h)
	}

	err = d.decodeRep(h, rest, rep)
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

func (d *Device) debugPrint(ep byte, data []byte) {
	if !d.DebugPrint {
		return
	}
	dir := "send"
	if 0 != ep&usb.ENDPOINT_IN {
		dir = "recv"
	}
	fmt.Fprintf(os.Stderr, "%s: 0x%x bytes with ep 0x%x:\n", dir, len(data), ep)
	hexDump(data)
}

func hexDump(data []byte) {
	i := 0
	for i < len(data) {
		next := i + 16
		if next > len(data) {
			next = len(data)
		}
		ss := []string{}
		s := fmt.Sprintf("%x", data[i:next])
		for j := 0; j < len(s); j += 4 {
			e := j + 4
			if len(s) < e {
				e = len(s)
			}
			ss = append(ss, s[j:e])
		}
		chars := make([]byte, next-i)
		for j, c := range data[i:next] {
			if c < 32 || c > 127 {
				c = '.'
			}
			chars[j] = c
		}
		fmt.Fprintf(os.Stderr, "%04x: %-40s %s\n", i, strings.Join(ss, " "), chars)
		i = next
	}
}

// The linux usb stack can send 16kb per call, according to libusb.
const rwBufSize = 0x4000

// bulkWrite returns the number of non-header bytes written.
func (d *Device) bulkWrite(hdr *USBBulkHeader, r io.Reader, size int64) (n int64, err error) {
	if hdr != nil {
		if size+HdrLen > 0xFFFFFFFF {
			hdr.Length = 0xFFFFFFFF
		} else {
			hdr.Length = uint32(size + HdrLen)
		}

		var packetArr [packetSize]byte
		var packet []byte
		if d.SeparateHeader {
			packet = packetArr[:HdrLen]
		} else {
			packet = packetArr[:]
		}

		buf := bytes.NewBuffer(packet[:0])
		binary.Write(buf, byteOrder, hdr)
		cpSize := int64(len(packet) - HdrLen)
		if cpSize > size {
			cpSize = size
		}

		_, err = io.CopyN(buf, r, cpSize)
		d.debugPrint(d.sendEp, buf.Bytes())
		_, err = d.h.BulkTransfer(d.sendEp, buf.Bytes(), d.timeout)
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

		d.debugPrint(d.sendEp, buf[:m])
		lastTransfer, err = d.h.BulkTransfer(d.sendEp, buf[:m], d.timeout)
		n += int64(lastTransfer)

		if err != nil || lastTransfer == 0 {
			break
		}
	}
	if lastTransfer%packetSize == 0 {
		// write a short packet just to be sure.
		d.h.BulkTransfer(d.sendEp, buf[:0], d.timeout)
	}

	return n, err
}

func (d *Device) bulkRead(w io.Writer) (n int64, err error) {
	var buf [rwBufSize]byte
	var lastRead int
	for {
		toread := buf[:]
		lastRead, err = d.h.BulkTransfer(d.fetchEp, toread, d.timeout)
		if err != nil {
			break
		}
		if lastRead > 0 {
			d.debugPrint(d.fetchEp, buf[:lastRead])

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
		d.h.BulkTransfer(d.fetchEp, buf[:0], d.timeout)
	}
	return n, err
}
