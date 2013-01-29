package mtp

// Android MTP extensions

// Same as GetPartialObject, but with 64 bit offset
const OC_ANDROID_GET_PARTIAL_OBJECT64 = 0x95C1

// Same as GetPartialObject64, but copying host to device
const OC_ANDROID_SEND_PARTIAL_OBJECT = 0x95C2

// Truncates file to 64 bit length
const OC_ANDROID_TRUNCATE_OBJECT = 0x95C3

// Must be called before using SendPartialObject and TruncateObject
const OC_ANDROID_BEGIN_EDIT_OBJECT = 0x95C4

// Called to commit changes made by SendPartialObject and TruncateObject
const OC_ANDROID_END_EDIT_OBJECT = 0x95C5

func init() {
	OC_names[0x95C1] = "ANDROID_GET_PARTIAL_OBJECT64"
	OC_names[0x95C2] = "ANDROID_SEND_PARTIAL_OBJECT"
	OC_names[0x95C3] = "ANDROID_TRUNCATE_OBJECT"
	OC_names[0x95C4] = "ANDROID_BEGIN_EDIT_OBJECT"
	OC_names[0x95C5] = "ANDROID_END_EDIT_OBJECT"
}
