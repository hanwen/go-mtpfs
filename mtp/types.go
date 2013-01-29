package mtp

import (
	"time"
)

type Container struct {
	Code          uint16
	SessionID     uint32
	TransactionID uint32
	Param         []uint32
}

type USBBulkHeader struct {
	Length        uint32
	Type          uint16
	Code          uint16
	TransactionID uint32
}

type USBBulkContainer struct {
	USBBulkHeader
	Param [5]uint32
}

const HdrLen = 2*2 + 2*4
const BulkLen = 5*4 + HdrLen

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

type DataTypeSelector uint16
type DataDependentType interface{}

type PropDescRangeForm struct {
	MinimumValue DataDependentType
	MaximumValue DataDependentType
	StepSize     DataDependentType
}

type PropDescEnumForm struct {
	Values []DataDependentType
}

type PropString struct {
	Value string
}

type DevicePropDesc struct {
	DevicePropertyCode  uint16
	DataType            DataTypeSelector
	GetSet              uint8
	FactoryDefaultValue DataDependentType
	CurrentValue        DataDependentType
	FormFlag            uint8
	Form                interface{}
}

type ObjectPropDesc struct {
	ObjectPropertyCode  uint16
	DataType            uint16
	GetSet              uint8
	FactoryDefaultValue DataDependentType
	GroupCode           uint32
	FormFlag            uint8
	Form                interface{}
}

type StorageIDs struct {
	IDs []uint32
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
	return d.FilesystemType  == FST_GenericHierarchical
}


func (d *StorageInfo) IsRemovable() bool {
       return (d.StorageType == ST_RemovableROM ||
		d.StorageType == ST_RemovableRAM)
}

type ObjectHandles struct {
	Handles []uint32
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
