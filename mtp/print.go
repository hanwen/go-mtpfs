package mtp

import (
	"fmt"
	"strings"
)

func getNames(m map[int]string, vals []uint16) string {
	r := []string{}
	for _, v := range vals {
		n, ok := m[int(v)]
		if !ok {
			n = fmt.Sprintf("0x%x", v)
		}
		r = append(r, n)
	}
	return strings.Join(r, ", ")
}

func (i *DeviceInfo) String() string {
	return fmt.Sprintf("stdv: %x, ext: %x, mtp: v%x, mtp ext: %q fmod: %x ops: %s evs: %s "+
		"dprops: %s fmts: %s capfmts: %s manu: %q model: %q devv: %q serno: %q",
		i.StandardVersion,
		i.MTPVendorExtensionID,
		i.MTPVersion,
		i.MTPExtension,
		i.FunctionalMode,
		getNames(OC_names, i.OperationsSupported),
		getNames(EC_names, i.EventsSupported),
		getNames(DPC_names, i.DevicePropertiesSupported),
		getNames(OFC_names, i.CaptureFormats),
		getNames(OFC_names, i.PlaybackFormats),

		i.Manufacturer,
		i.Model,
		i.DeviceVersion,
		i.SerialNumber)
}
