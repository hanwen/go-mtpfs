package mtp

import (
	"fmt"
	"os"
	"strings"
)

// Print hex on stderr.
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

// extract single name.
func getName(m map[int]string, val int) string {
	n, ok := m[val]
	if !ok {
		n = fmt.Sprintf("0x%x", val)
	}
	return n
}

// Extract names from name map.
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
