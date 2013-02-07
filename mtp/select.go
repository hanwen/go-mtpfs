package mtp

import (
	"fmt"
	"regexp"

	"github.com/hanwen/go-mtpfs/usb"
)
 
func candidateFromDeviceDescriptor(d *usb.Device) *Device {
	dd, err := d.GetDeviceDescriptor()
	if err != nil {
		return nil
	}
	for i := byte(0); i < dd.NumConfigurations; i++ {
		cdecs, err := d.GetConfigDescriptor(i)
		if err != nil {
			return nil
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
					m.configValue = cdecs.ConfigurationValue
					return &m
				}
			}
		}
	}

	return nil
}

// FindDevices finds likely MTP devices without opening them.
func FindDevices(c *usb.Context) ([]*Device, error) {
	l, err := c.GetDeviceList()
	if err != nil {
		return nil, err
	}

	var cands []*Device
	for _, d := range l {
		cand := candidateFromDeviceDescriptor(d)
		if cand != nil {
			cands = append(cands, cand)
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
		id, err := cand.ID()
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

	if len(cands) == 0 {
		return nil, fmt.Errorf("no device matched")
	}

	cand := cands[0]
	config, err := cand.h.GetConfiguration()
	if err != nil {
		return nil, fmt.Errorf("could not get configuration of %v: %v",
			ids[0], err)
	}
	acd, err := cand.dev.GetActiveConfigDescriptor()
	if config != cand.configValue {
		err := cand.h.SetConfiguration(cand.configValue)
		if err != nil {
			return nil, fmt.Errorf("could not set configuration of %v: %v",
				ids[0], err)
		}
	}
	return cands[0], nil
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
