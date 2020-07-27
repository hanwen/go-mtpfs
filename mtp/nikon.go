package mtp

// Nikon MTP extensions

const (
	OC_NIKON_AfDrive = 0x90C1
)

const (
	LVHeaderSize = 384
)

type Rotation int

const (
	Rotation0       Rotation = 0
	Rotation90      Rotation = 90
	RotationMinus90 Rotation = -90
	Rotation180     Rotation = 180
)

type AF int

const (
	AFNotActive AF = 0
	AFFail      AF = 1
	AFSuccess   AF = 2
)
