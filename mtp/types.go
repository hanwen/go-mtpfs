// The MTP packages defines data types and procedures for
// communicating with an MTP device.  Beyond the communication
// primitive, it implements many useful operations in the file ops.go.
// These may serve as an example how to implement further operations.
package mtp

import (
	"io"
	"time"
)

// Container is the data type for sending/receiving MTP requests and
// responses.
type Container struct {
	Code          uint16
	SessionID     uint32
	TransactionID uint32
	Param         []uint32
}

type DeviceInfo struct {
	StandardVersion           uint16
	MTPVendorExtensionID      uint32
	MTPVersion                uint16
	MTPExtension              string
	FunctionalMode            uint16
	OperationsSupported       []uint16
	EventsSupported           []uint16
	DevicePropertiesSupported []uint16
	CaptureFormats            []uint16
	PlaybackFormats           []uint16
	Manufacturer              string
	Model                     string
	DeviceVersion             string
	SerialNumber              string
}

// DataTypeSelector is the special type to indicate the actual type of
// fields of DataDependentType.
type DataTypeSelector uint16
type DataDependentType interface{}

// The Decoder interface is for types that need special decoding
// support, eg. the ones using DataDependentType.
type Decoder interface {
	Decode(r io.Reader) error
}

type Encoder interface {
	Encode(w io.Writer) error
}

type PropDescRangeForm struct {
	MinimumValue DataDependentType
	MaximumValue DataDependentType
	StepSize     DataDependentType
}

type PropDescEnumForm struct {
	Values []DataDependentType
}

type DevicePropDescFixed struct {
	DevicePropertyCode  uint16
	DataType            DataTypeSelector
	GetSet              uint8
	FactoryDefaultValue DataDependentType
	CurrentValue        DataDependentType
	FormFlag            uint8
}

type DevicePropDesc struct {
	DevicePropDescFixed
	Form interface{}
}

type ObjectPropDescFixed struct {
	ObjectPropertyCode  uint16
	DataType            DataTypeSelector
	GetSet              uint8
	FactoryDefaultValue DataDependentType
	GroupCode           uint32
	FormFlag            uint8
}

type ObjectPropDesc struct {
	ObjectPropDescFixed
	Form interface{}
}

type Uint32Array struct {
	Values []uint32
}

type Uint16Array struct {
	Values []uint16
}

type Uint64Value struct {
	Value uint64
}

type StringValue struct {
	Value string
}

type StorageInfo struct {
	StorageType        uint16
	FilesystemType     uint16
	AccessCapability   uint16
	MaxCapability      uint64
	FreeSpaceInBytes   uint64
	FreeSpaceInImages  uint32
	StorageDescription string
	VolumeLabel        string
}

func (d *StorageInfo) IsHierarchical() bool {
	return d.FilesystemType == FST_GenericHierarchical
}

func (d *StorageInfo) IsRemovable() bool {
	return (d.StorageType == ST_RemovableROM ||
		d.StorageType == ST_RemovableRAM)
}

type ObjectInfo struct {
	StorageID           uint32
	ObjectFormat        uint16
	ProtectionStatus    uint16
	CompressedSize      uint32
	ThumbFormat         uint16
	ThumbCompressedSize uint32
	ThumbPixWidth       uint32
	ThumbPixHeight      uint32
	ImagePixWidth       uint32
	ImagePixHeight      uint32
	ImageBitDepth       uint32
	ParentObject        uint32
	AssociationType     uint16
	AssociationDesc     uint32
	SequenceNumber      uint32
	Filename            string
	CaptureDate         time.Time
	ModificationDate    time.Time
	Keywords            string
}

// USB stuff.

type usbBulkHeader struct {
	Length        uint32
	Type          uint16
	Code          uint16
	TransactionID uint32
}

type usbBulkContainer struct {
	usbBulkHeader
	Param [5]uint32
}

const usbHdrLen = 2*2 + 2*4
const usbBulkLen = 5*4 + usbHdrLen
