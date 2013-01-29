package usb

func ClassToStr(c byte) string {
	switch c {
	case CLASS_PER_INTERFACE:
		return "PER_INTERFACE"
	case CLASS_AUDIO:
		return "AUDIO"
	case CLASS_COMM:
		return "COMM"
	case CLASS_HID:
		return "HID"
	case CLASS_PHYSICAL:
		return "PHYSICAL"
	case CLASS_PRINTER:
		return "PRINTER"
	case CLASS_IMAGE:
		return "IMAGE"
	case CLASS_MASS_STORAGE:
		return "MASS_STORAGE"
	case CLASS_HUB:
		return "HUB"
	case CLASS_DATA:
		return "DATA"
	case CLASS_SMART_CARD:
		return "SMART_CARD"
	case CLASS_CONTENT_SECURITY:
		return "CONTENT_SECURITY"
	case CLASS_VIDEO:
		return "VIDEO"
	case CLASS_PERSONAL_HEALTHCARE:
		return "PERSONAL_HEALTHCARE"
	case CLASS_DIAGNOSTIC_DEVICE:
		return "DIAGNOSTIC_DEVICE"
	case CLASS_WIRELESS:
		return "WIRELESS"
	case CLASS_APPLICATION:
		return "APPLICATION"
	case CLASS_VENDOR_SPEC:
		return "VENDOR_SPEC"
	}
	return "other"
}

func dtToStr(c int) string {
	switch c {
	case DT_DEVICE:
		return "DEVICE"
	case DT_CONFIG:
		return "CONFIG"
	case DT_STRING:
		return "STRING"
	case DT_INTERFACE:
		return "INTERFACE"
	case DT_ENDPOINT:
		return "ENDPOINT"
	case DT_HID:
		return "HID"
	case DT_REPORT:
		return "REPORT"
	case DT_PHYSICAL:
		return "PHYSICAL"
	case DT_HUB:
		return "HUB"
	}
	return "other"
}

func requestToStr(c int) string {
	switch c {
	case REQUEST_GET_STATUS:
		return "GET_STATUS"
	case REQUEST_CLEAR_FEATURE:
		return "CLEAR_FEATURE"
		/** Set or enable a specific feature */
	case REQUEST_SET_FEATURE:
		return "SET_FEATURE"
		/** Set device address for all future accesses */
	case REQUEST_SET_ADDRESS:
		return "SET_ADDRESS"
		/** Get the specified descriptor */
	case REQUEST_GET_DESCRIPTOR:
		return "GET_DESCRIPTOR"
		/** Used to update existing descriptors or add new descriptors */
	case REQUEST_SET_DESCRIPTOR:
		return "SET_DESCRIPTOR"
		/** Get the current device configuration value */
	case REQUEST_GET_CONFIGURATION:
		return "GET_CONFIGURATION"
		/** Set device configuration */
	case REQUEST_SET_CONFIGURATION:
		return "SET_CONFIGURATION"
		/** Return the selected alternate setting for the specified interface */
	case REQUEST_GET_INTERFACE:
		return "GET_INTERFACE"
		/** Select an alternate interface for the specified interface */
	case REQUEST_SET_INTERFACE:
		return "SET_INTERFACE"
		/** Set then report an endpoint's synchronization frame */
	case REQUEST_SYNCH_FRAME:
		return "SYNCH_FRAME"
	}
	return "other"
}

func transferTypeString(t byte) string {
	switch t {
	case TRANSFER_TYPE_CONTROL:
		return "control"
	case TRANSFER_TYPE_ISOCHRONOUS:
		return "iso"
	case TRANSFER_TYPE_BULK:
		return "bulk"
	case TRANSFER_TYPE_INTERRUPT:
		return "int"
	}
	panic("unknown")
}
