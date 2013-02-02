package mtp

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hanwen/go-mtpfs/usb"
)

// Finds likely MTP devices without opening them.
func FindDevices(c *usb.Context) ([]*Device, error) {
	l, err := c.GetDeviceList()
	if err != nil {
		return nil, err
	}

	var cands []*Device
	for _, d := range l {
		dd, err := d.GetDeviceDescriptor()
		if err != nil {
			continue
		}

		for i := byte(0); i < dd.NumConfigurations; i++ {
			cdecs, err := d.GetConfigDescriptor(i)
			if err != nil {
				return nil, fmt.Errorf("GetConfigDescriptor %d: %v", i, err)
			}
			for _, iface := range cdecs.Interfaces {
				for _, a := range iface.AltSetting {
					if len(a.EndPoints) != 3 {
						continue
					}

					m := Device{}
					for _, s := range a.EndPoints {
						switch {
						case s.Direction() == usb.ENDPOINT_IN && s.TransferType() == usb.TRANSFER_TYPE_INTERRUPT:
							m.eventEp = s.EndpointAddress
						case s.Direction() == usb.ENDPOINT_IN && s.TransferType() == usb.TRANSFER_TYPE_BULK:
							m.fetchEp = s.EndpointAddress
						case s.Direction() == usb.ENDPOINT_OUT && s.TransferType() == usb.TRANSFER_TYPE_BULK:
							m.sendEp = s.EndpointAddress
						}
					}
					if m.sendEp > 0 && m.fetchEp > 0 && m.eventEp > 0 {
						m.devDescr = *dd
						m.ifaceDescr = a
						m.dev = d.Ref()
						m.configIndex = i
						cands = append(cands, &m)
					}
				}
			}
		}
	}
	l.Done()

	return cands, nil
}

// Finds a device that matches given pattern
func selectDevice(cands []*Device, pattern string) (*Device, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	var found []*Device
	for _, cand := range cands {
		err := cand.Open()
		if err != nil {
			continue
		}

		found = append(found, cand)
	}

	if len(found) == 0 {
		return nil, fmt.Errorf("no MTP devices found")
	}

	cands = found
	found = nil
	var ids []string
	for i, cand := range cands {
		id, err := cand.Id()
		if err != nil {
			// TODO - close cands
			return nil, fmt.Errorf("Id dev %d: %v", i, err)
		}

		if pattern == "" || re.FindString(id) != "" {
			found = append(found, cand)
			ids = append(ids, id)
		} else {
			cand.Close()
			cand.Done()
		}
	}

	if len(cands) > 1 {
		return nil, fmt.Errorf("ambiguous devices: %s",
			strings.Join(ids, ", "))
	}

	if len(cands) == 0 {
		return nil, fmt.Errorf("no device matched")
	}

	cand := cands[0]

	// For some reason, always have to reset
	if err := cand.h.Reset(); err != nil {
		return nil, err
	}

	// TODO - set active configuration
	return cand, nil
}

// Return opened MTP device that matches given pattern.
func SelectDevice(pattern string) (*Device, error) {
	c := usb.NewContext()
	
	devs, err := FindDevices(c)
	if err != nil {
		return nil, err
	}
	if len(devs) == 0 {
		return nil, fmt.Errorf("no MTP devices found")
	}

	return selectDevice(devs, pattern)
}
