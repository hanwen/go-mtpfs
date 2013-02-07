// The usb package is a straighforward cgo wrapping of the libusb 1.0
// API. It only supports the synchronous API, since Goroutines can be
// used for asynchronous use-cases.

package usb

// #cgo LDFLAGS: -L/lib64 -lusb-1.0
// #cgo CFLAGS: -I/usr/include/libusb-1.0
// #include <libusb.h>
import "C"
import (
	"fmt"
	"reflect"
	"unsafe"
)

const SPEED_UNKNOWN = C.LIBUSB_SPEED_UNKNOWN
const SPEED_LOW = C.LIBUSB_SPEED_LOW
const SPEED_FULL = C.LIBUSB_SPEED_FULL
const SPEED_HIGH = C.LIBUSB_SPEED_HIGH
const SPEED_SUPER = C.LIBUSB_SPEED_SUPER

type ControlSetup C.struct_libusb_control_setup
type Transfer C.struct_libusb_transfer

// Device and/or interface class codes.
const CLASS_PER_INTERFACE = 0
const CLASS_AUDIO = 1
const CLASS_COMM = 2
const CLASS_HID = 3
const CLASS_PHYSICAL = 5
const CLASS_PRINTER = 7
const CLASS_IMAGE = 6
const CLASS_MASS_STORAGE = 8
const CLASS_HUB = 9
const CLASS_DATA = 10
const CLASS_SMART_CARD = 0x0b
const CLASS_CONTENT_SECURITY = 0x0d
const CLASS_VIDEO = 0x0e
const CLASS_PERSONAL_HEALTHCARE = 0x0f
const CLASS_DIAGNOSTIC_DEVICE = 0xdc
const CLASS_WIRELESS = 0xe0
const CLASS_APPLICATION = 0xfe
const CLASS_VENDOR_SPEC = 0xff

// Descriptor types as defined by the USB specification.
const DT_DEVICE = 0x01
const DT_CONFIG = 0x02
const DT_STRING = 0x03
const DT_INTERFACE = 0x04
const DT_ENDPOINT = 0x05
const DT_HID = 0x21
const DT_REPORT = 0x22
const DT_PHYSICAL = 0x23
const DT_HUB = 0x29

// Standard request types, as defined in table 9-3 of the USB2 specifications
const REQUEST_GET_STATUS = 0x00
const REQUEST_CLEAR_FEATURE = 0x01

// Set or enable a specific feature
const REQUEST_SET_FEATURE = 0x03

// Set device address for all future accesses
const REQUEST_SET_ADDRESS = 0x05

// Get the specified descriptor
const REQUEST_GET_DESCRIPTOR = 0x06

// Used to update existing descriptors or add new descriptors
const REQUEST_SET_DESCRIPTOR = 0x07

// Get the current device configuration value
const REQUEST_GET_CONFIGURATION = 0x08

// Set device configuration
const REQUEST_SET_CONFIGURATION = 0x09

// Return the selected alternate setting for the specified interface.
const REQUEST_GET_INTERFACE = 0x0A

// Select an alternate interface for the specified interface
const REQUEST_SET_INTERFACE = 0x0B

// Set then report an endpoint's synchronization frame
const REQUEST_SYNCH_FRAME = 0x0C


// The error codes returned by libusb.
type Error int

func (e Error) Error() string {
	return C.GoString(C.libusb_error_name(C.int(e)))
}

func toErr(e C.int) error {
	if e < 0 {
		return Error(e)
	}
	return nil
}

const SUCCESS = Error(0)
const ERROR_IO = Error(-1)
const ERROR_INVALID_PARAM = Error(-2)
const ERROR_ACCESS = Error(-3)
const ERROR_NO_DEVICE = Error(-4)
const ERROR_NOT_FOUND = Error(-5)
const ERROR_BUSY = Error(-6)
const ERROR_TIMEOUT = Error(-7)
const ERROR_OVERFLOW = Error(-8)
const ERROR_PIPE = Error(-9)
const ERROR_INTERRUPTED = Error(-10)
const ERROR_NO_MEM = Error(-11)
const ERROR_NOT_SUPPORTED = Error(-12)
const ERROR_OTHER = Error(-99)

const TRANSFER_COMPLETED = 0
const TRANSFER_ERROR = 1
const TRANSFER_TIMED_OUT = 2
const TRANSFER_CANCELLED = 3
const TRANSFER_STALL = 4
const TRANSFER_NO_DEVICE = 5
const TRANSFER_OVERFLOW = 6

const TRANSFER_SHORT_NOT_OK = 1 << 0
const TRANSFER_FREE_BUFFER = 1 << 1
const TRANSFER_FREE_TRANSFER = 1 << 2

// Request types to use in ControlTransfer().
const REQUEST_TYPE_STANDARD = (0x00 << 5)
const REQUEST_TYPE_CLASS = (0x01 << 5)
const REQUEST_TYPE_VENDOR = (0x02 << 5)
const REQUEST_TYPE_RESERVED = (0x03 << 5)

// Recipient bits for the reqType of ControlTransfer(). Values 4 - 31
// are reserved.
const RECIPIENT_DEVICE = 0x00
const RECIPIENT_INTERFACE = 0x01
const RECIPIENT_ENDPOINT = 0x02
const RECIPIENT_OTHER = 0x03

// Synchronization types for isochronous endpoints, used in
// EndpointDescriptor.Attributes, bits 2:3.
const ISO_SYNC_TYPE_NONE = 0
const ISO_SYNC_TYPE_ASYNC = 1
const ISO_SYNC_TYPE_ADAPTIVE = 2
const ISO_SYNC_TYPE_SYNC = 3

// Usage types used in EndpointDescriptor.Attributes, bits 4:5.
const ISO_USAGE_TYPE_DATA = 0
const ISO_USAGE_TYPE_FEEDBACK = 1
const ISO_USAGE_TYPE_IMPLICIT = 2

// DeviceDescriptor is the standard USB device descriptor as
// documented in section 9.6.1 of the USB 2.0 specification.
type DeviceDescriptor struct {
	// Size of this descriptor (in bytes)
	Length byte
	// Descriptor type.
	DescriptorType byte
	// USB specification release number in binary-coded decimal.
	USBRelease uint16

	// USB-IF class code for the device.
	DeviceClass byte
	// USB-IF subclass code for the device, qualified by the
	// DeviceClass value.
	DeviceSubClass byte
	// USB-IF protocol code for the device, qualified by the
	// DeviceClass and DeviceSubClass values.
	DeviceProtocol byte
	// Maximum packet size for endpoint 0.
	MaxPacketSize0 byte
	// USB-IF vendor ID.
	IdVendor uint16
	// USB-IF product ID.
	IdProduct uint16
	// Device release number in binary-coded decimal.
	Device uint16

	// Index of string descriptor describing manufacturer.
	Manufacturer byte
	// Index of string descriptor describing product.
	Product byte
	// Index of string descriptor containing device serial number.
	SerialNumber byte

	// Number of possible configurations.
	NumConfigurations byte
}

// A collection of alternate settings for a USB interface.
type Interface struct {
	AltSetting []InterfaceDescriptor
}

func (f *Interface) fromC(c *C.struct_libusb_interface) {
	f.AltSetting = make([]InterfaceDescriptor, c.num_altsetting)

	ds := []C.struct_libusb_interface_descriptor{}
	cSlice(unsafe.Pointer(&ds), unsafe.Pointer(c.altsetting), c.num_altsetting)
	for i, s := range ds {
		f.AltSetting[i].fromC(&s)
	}
}

// EndpointDescriptor represents the standard USB endpoint
// descriptor. This descriptor is documented in section 9.6.3 of the
// USB 2.0 specification.
type EndpointDescriptor struct {
	// Size of this descriptor (in bytes)
	Length byte

	// Descriptor type. Will have value LIBUSB_DT_ENDPOINT in this
	// context.
	DescriptorType byte

	// The address of the endpoint described by this descriptor. Bits 0:3 are
	// the endpoint number. Bits 4:6 are reserved. Bit 7 indicates direction.
	EndpointAddress byte

	// Attributes which apply to the endpoint when it is configured using
	// the ConfigurationValue. Bits 0:1 determine the transfer type and
	// correspond to libusb_transfer_type. Bits 2:3 are only used for
	// isochronous endpoints and correspond to libusb_iso_sync_type.
	// Bits 4:5 are also only used for isochronous endpoints and correspond to
	// libusb_iso_usage_type. Bits 6:7 are reserved.
	Attributes byte

	// Maximum packet size this endpoint is capable of sending/receiving.
	MaxPacketSize uint16

	// Interval for polling endpoint for data transfers.
	Interval byte

	// For audio devices only: the rate at which synchronization feedback
	// is provided.
	Refresh byte

	// For audio devices only: the address if the synch endpoint
	SynchAddress byte

	// Extra descriptors. If libusb encounters unknown endpoint
	// descriptors, it will store them here, should you wish to
	// parse them.
	Extra []byte
}

// Endpoint transfer types, for bits 0:1 of
// EndpointDescriptor.Attributes
const TRANSFER_TYPE_CONTROL = 0
const TRANSFER_TYPE_ISOCHRONOUS = 1
const TRANSFER_TYPE_BULK = 2
const TRANSFER_TYPE_INTERRUPT = 3

// in: device-to-host
const ENDPOINT_IN = 0x80

// out: host-to-device
const ENDPOINT_OUT = 0x00

func (e *EndpointDescriptor) TransferType() byte {
	return e.Attributes & 0x3
}

func (e *EndpointDescriptor) Direction() byte {
	return e.EndpointAddress & ENDPOINT_IN
}

func (e *EndpointDescriptor) Number() byte {
	return e.EndpointAddress & 0x0f
}

func (e *EndpointDescriptor) String() string {
	tDir := "out"
	if e.EndpointAddress&ENDPOINT_IN != 0 {
		tDir = "in"
	}

	// TODO - print isochronous data too.
	s := fmt.Sprintf("ep num %x dir %s ttype %s maxpacket %d",
		e.Number(), tDir, transferTypeString(e.TransferType()),
		e.MaxPacketSize)

	return s
}

func byteArrToSlice(a *C.uchar, n C.int) []byte {
	var g []C.uchar

	b := make([]byte, int(n))
	for i, c := range g {
		b[i] = byte(c)
	}
	return b
}

func cSlice(slice unsafe.Pointer, arr unsafe.Pointer, n C.int) {
	h := (*reflect.SliceHeader)(slice)
	h.Cap = int(n)
	h.Len = int(n)
	h.Data = uintptr(unsafe.Pointer(arr))
}

func (d *EndpointDescriptor) fromC(c *C.struct_libusb_endpoint_descriptor) {
	d.Length = byte(c.bLength)
	d.DescriptorType = byte(c.bDescriptorType)
	d.EndpointAddress = byte(c.bEndpointAddress)
	d.Attributes = byte(c.bmAttributes)
	d.MaxPacketSize = uint16(c.wMaxPacketSize)
	d.Interval = byte(c.bInterval)
	d.Refresh = byte(c.bRefresh)
	d.SynchAddress = byte(c.bSynchAddress)

	d.Extra = byteArrToSlice(c.extra, c.extra_length)
}

// InterfaceDescriptor contains the standard USB interface descriptor,
// according to section 9.6.5 of the USB 2.0 specification.
type InterfaceDescriptor struct {
	// Size of this descriptor (in bytes)
	Length byte

	// Descriptor type. Will have value DT_INTERFACE
	// LIBUSB_DT_INTERFACE in this context.
	DescriptorType byte

	// Number of this interface
	InterfaceNumber byte

	// Value used to select this alternate setting for this interface
	AlternateSetting byte

	// USB-IF class code for this interface.
	InterfaceClass byte

	// USB-IF subclass code for this interface, qualified by the
	// InterfaceClass value
	InterfaceSubClass byte

	// USB-IF protocol code for this interface, qualified by the
	// InterfaceClass and InterfaceSubClass values
	InterfaceProtocol byte

	// Index of string descriptor describing this interface
	InterfaceStringIndex byte

	// Array of endpoint descriptors.
	EndPoints []EndpointDescriptor

	// Extra descriptors. If libusb encounters unknown interface
	// descriptors, it will store them here, should you wish to
	// parse them.
	Extra []byte
}

func (d *InterfaceDescriptor) fromC(c *C.struct_libusb_interface_descriptor) {
	d.Length = byte(c.bLength)
	d.DescriptorType = byte(c.bDescriptorType)
	d.InterfaceNumber = byte(c.bInterfaceNumber)
	d.AlternateSetting = byte(c.bAlternateSetting)
	d.InterfaceClass = byte(c.bInterfaceClass)
	d.InterfaceSubClass = byte(c.bInterfaceSubClass)
	d.InterfaceProtocol = byte(c.bInterfaceProtocol)
	d.InterfaceStringIndex = byte(c.iInterface)
	d.EndPoints = make([]EndpointDescriptor, c.bNumEndpoints)

	cs := []C.struct_libusb_endpoint_descriptor{}
	cSlice(unsafe.Pointer(&cs), unsafe.Pointer(c.endpoint), C.int(c.bNumEndpoints))
	for i, s := range cs {
		d.EndPoints[i].fromC(&s)
	}

	d.Extra = byteArrToSlice(c.extra, c.extra_length)
}

type IsoPacketDescriptor struct {
	// Length of data to request in this packet
	Length uint

	// Amount of data that was actually transferred
	ActualLength uint

	// Status code for this packet
	Status int
}

type ConfigDescriptor struct {
	// Size of this descriptor (in bytes)
	Length byte

	// Descriptor type. Will have value DT_CONFIG LIBUSB_DT_CONFIG
	// in this context.
	DescriptorType byte

	// Total length of data returned for this configuration
	TotalLength uint16

	// Identifier value for this configuration
	ConfigurationValue byte

	// Index of string descriptor describing this configuration
	ConfigurationIndex byte

	// Configuration characteristics
	Attributes byte

	// Maximum power consumption of the USB device from this bus in this
	// configuration when the device is fully opreation. Expressed in units
	// of 2 mA.
	MaxPower byte

	// Array of interfaces supported by this configuration.
	Interfaces []Interface

	// Extra descriptors. If libusb encounters unknown configuration
	// descriptors, it will store them here, should you wish to parse them.
	Extra []byte
}

func (d *ConfigDescriptor) fromC(c *C.struct_libusb_config_descriptor) {
	d.Length = byte(c.bLength)
	d.DescriptorType = byte(c.bDescriptorType)
	d.TotalLength = uint16(c.wTotalLength)
	d.ConfigurationValue = byte(c.bConfigurationValue)
	d.ConfigurationIndex = byte(c.iConfiguration)
	d.Attributes = byte(c.bmAttributes)
	d.MaxPower = byte(c.MaxPower)

	d.Interfaces = make([]Interface, c.bNumInterfaces)

	cis := []C.struct_libusb_interface{}
	cSlice(unsafe.Pointer(&cis), unsafe.Pointer(c._interface), C.int(c.bNumInterfaces))
	for i, iface := range cis {
		d.Interfaces[i].fromC(&iface)
	}
	d.Extra = byteArrToSlice(c.extra, c.extra_length)
}

type Context C.struct_libusb_context
type Device C.struct_libusb_device
type DeviceHandle C.struct_libusb_device_handle

func NewContext() *Context {
	var r *C.struct_libusb_context
	C.libusb_init(&r)
	return (*Context)(r)
}

func (c *Context) me() *C.struct_libusb_context {
	return (*C.struct_libusb_context)(c)
}

func (c *Context) SetDebug(level int) {
	C.libusb_set_debug(c.me(), C.int(level))
}

func (c *Context) Exit() {
	C.libusb_exit(c.me())
}

type DeviceList []*Device

func (d DeviceList) Done() {
	C.libusb_free_device_list((**C.libusb_device)(unsafe.Pointer((&d[0]))), 1)
}

func (c *Context) GetDeviceList() (DeviceList, error) {
	var devs **C.libusb_device
	count := C.libusb_get_device_list(c.me(), &devs)
	if count < 0 {
		return nil, Error(count)
	}
	slice := &reflect.SliceHeader{uintptr(unsafe.Pointer(devs)), int(count), int(count)}
	rdevs := *(*[]*Device)(unsafe.Pointer(slice))
	return DeviceList(rdevs), nil
}

func (d *Device) me() *C.libusb_device {
	return (*C.libusb_device)(d)
}

// Get the number of the bus that a device is connected to.
func (d *Device) GetBusNumber() uint8 {
	return uint8(C.libusb_get_bus_number(d.me()))
}

// Get the address of the device on the bus it is connected to.
func (d *Device) GetDeviceAddress() uint8 {
	return uint8(C.libusb_get_device_address(d.me()))
}

// Get the negotiated connection speed for a device.
func (d *Device) GetDeviceSpeed() int {
	return int(C.libusb_get_device_speed(d.me()))
}

// Convenience function to retrieve the MaxPacketSize value for a particular endpoint in the active device configuration.
func (d *Device) GetMaxPacketSize(endpoint byte) int {
	return int(C.libusb_get_max_packet_size(d.me(), C.uchar(endpoint)))
}

// Calculate the maximum packet size which a specific endpoint is capable is sending or receiving in the duration of 1 microframe.
func (d *Device) GetMaxIsoPacketSize(endpoint byte) int {
	return int(C.libusb_get_max_iso_packet_size(d.me(), C.uchar(endpoint)))
}

func (d *Device) Ref() *Device {
	return (*Device)(C.libusb_ref_device((*C.libusb_device)(d.me())))
}

// Decrement the reference count of a device.
func (d *Device) Unref() {
	C.libusb_unref_device((*C.libusb_device)(d.me()))
}

func (d *Device) GetDeviceDescriptor() (*DeviceDescriptor, error) {
	// this relies on struct packing being equal.
	var dd DeviceDescriptor
	r := C.libusb_get_device_descriptor(d.me(), (*C.struct_libusb_device_descriptor)(unsafe.Pointer(&dd)))
	return &dd, toErr(r)
}

func (d *Device) GetActiveConfigDescriptor() (*ConfigDescriptor, error) {
	var desc *C.struct_libusb_config_descriptor
	r := C.libusb_get_active_config_descriptor(d.me(), &desc)
	if r < 0 {
		return nil, toErr(r)
	}

	var cd ConfigDescriptor
	cd.fromC(desc)
	C.libusb_free_config_descriptor(desc)
	return &cd, nil
}

func (d *Device) GetConfigDescriptor(config byte) (*ConfigDescriptor, error) {
	var desc *C.struct_libusb_config_descriptor
	r := C.libusb_get_config_descriptor(d.me(), C.uint8_t(config), &desc)
	if r < 0 {
		return nil, toErr(r)
	}

	var cd ConfigDescriptor
	cd.fromC(desc)
	C.libusb_free_config_descriptor(desc)
	return &cd, nil
}

func (d *Device) GetConfigDescriptorByValue(value byte) (*ConfigDescriptor, error) {
	var desc *C.struct_libusb_config_descriptor
	r := C.libusb_get_config_descriptor_by_value(d.me(), C.uint8_t(value), &desc)
	if r < 0 {
		return nil, toErr(r)
	}

	var cd ConfigDescriptor
	cd.fromC(desc)
	C.libusb_free_config_descriptor(desc)
	return &cd, nil
}

// Determine the ConfigurationValue of the currently active configuration.
func (h *DeviceHandle) GetConfiguration() (byte, error) {
	var r C.int
	err := C.libusb_get_configuration(h.me(), &r)
 	return byte(r), toErr(err)
}

// Set the active configuration for a device. The argument should be
// a ConfigurationValue, as given in the ConfigDescriptor.
func (h *DeviceHandle) SetConfiguration(c byte) error {
	err := C.libusb_set_configuration(h.me(), C.int(c))
	return toErr(err)
}

// Open a device and obtain a device handle.
func (d *Device) Open() (*DeviceHandle, error) {
	var h *C.libusb_device_handle
	r := C.libusb_open(d.me(), &h)
	return (*DeviceHandle)(h), toErr(r)
}

func (h *DeviceHandle) me() *C.libusb_device_handle {
	return (*C.libusb_device_handle)(h)
}

// Close a device handle.
func (h *DeviceHandle) Close() error {
	C.libusb_close(h.me())
	return nil
}

// Get the underlying device for a handle.
func (h *DeviceHandle) Device() *Device {
	return (*Device)(C.libusb_get_device(h.me()))
}

// Claim an interface on a given device handle.
func (h *DeviceHandle) ClaimInterface(num byte) error {
	return toErr(C.libusb_claim_interface(h.me(), C.int(num)))
}

// Release an interface previously claimed with libusb_claim_interface().
func (h *DeviceHandle) ReleaseInterface(num byte) error {
	return toErr(C.libusb_release_interface(h.me(), C.int(num)))
}

// Activate an alternate setting for an interface.
func (h *DeviceHandle) SetInterfaceAltSetting(num int, alternate int) error {
	return toErr(C.libusb_set_interface_alt_setting(h.me(), C.int(num), C.int(alternate)))
}

func (h *DeviceHandle) GetStringDescriptorASCII(descIndex byte) (string, error) {
	data := make([]byte, 1024)
	start := (*C.uchar)(unsafe.Pointer(&data[0]))
	r := C.libusb_get_string_descriptor_ascii(h.me(),
		C.uint8_t(descIndex), start, C.int(len(data)))
	return C.GoString((*C.char)(unsafe.Pointer(&data[0]))), toErr(r)
}

func (h *DeviceHandle) ControlTransfer(reqType, req byte, value, index uint16,
	data []byte, timeout int) error {
	var ptr *byte
	if len(data) > 0 {
		ptr = &data[0]
	}
	if len(data) > 0xffff {
		return fmt.Errorf("overflow")
	}
	err := C.libusb_control_transfer(h.me(),
		C.uint8_t(reqType), C.uint8_t(req), C.uint16_t(value), C.uint16_t(index),
		(*C.uchar)(ptr), C.uint16_t(len(data)), C.uint(timeout))
	return toErr(err)
}

func (h *DeviceHandle) BulkTransfer(endpoint byte, data []byte, timeout int) (actual int, err error) {
	var ptr *byte
	if len(data) > 0 {
		ptr = &data[0]
	}

	var n C.int
	e := C.libusb_bulk_transfer(h.me(), C.uchar(endpoint), (*C.uchar)(ptr),
		C.int(len(data)), &n, C.uint(timeout))
	return int(n), toErr(e)
}

func (h *DeviceHandle) InterruptTransfer(endpoint byte, data []byte, timeout int) (actual int, err error) {
	var ptr *byte
	if len(data) > 0 {
		ptr = &data[0]
	}

	var n C.int
	e := C.libusb_bulk_transfer(h.me(), C.uchar(endpoint), (*C.uchar)(ptr),
		C.int(len(data)), &n, C.uint(timeout))
	return int(n), toErr(e)
}

// Perform a USB port reset to reinitialize a device.
func (h *DeviceHandle) Reset() error {
	return toErr(C.libusb_reset_device(h.me()))
}

// Clear an halt/stall for a endpoint.
func (h *DeviceHandle) ClearHalt(endpoint byte) error {
	return toErr(C.libusb_clear_halt(h.me(), C.uchar(endpoint)))
}

// Determine if a kernel driver is active on an interface.
func (h *DeviceHandle) KernelDriverActive(ifaceNum int) (bool, error) {
	ret := C.libusb_kernel_driver_active(h.me(), C.int(ifaceNum))
	if ret == 0 {
		return false, nil
	}
	if ret == 1 {
		return true, nil
	}
	return false, toErr(ret)
}

// Detach a kernel driver from an interface.
func (h *DeviceHandle) DetachKernelDriver(ifaceNum int) error {
	return toErr(C.libusb_detach_kernel_driver(h.me(), C.int(ifaceNum)))
}

// Re-attach an interface's kernel driver, which was previously detached using libusb_detach_kernel_driver().
func (h *DeviceHandle) AttachKernelDriver(ifaceNum int) error {
	return toErr(C.libusb_attach_kernel_driver(h.me(), C.int(ifaceNum)))
}
