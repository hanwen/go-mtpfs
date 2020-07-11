package mtp

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/gousb"
	"github.com/hanwen/usb"
)

func FindDevice(ctx *gousb.Context, idv, idp uint16) (*DeviceGoUSB, error) {
	var mtpDev []*DeviceGoUSB

	devs, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		v, p := uint16(desc.Vendor), uint16(desc.Product)
		if idv != 0 && idp != 0 && (v != idv || p != idp) {
			return false
		}

		for _, conf := range desc.Configs {
			for _, iface := range conf.Interfaces {
				hasImageClass := false
				for _, alt := range iface.AltSettings {
					hasImageClass = hasImageClass || alt.Class == gousb.ClassPTP
				}
				if !hasImageClass {
					continue
				}

				for _, alt := range iface.AltSettings {
					if len(alt.Endpoints) != 3 {
						continue
					}

					var ev, fe, se gousb.EndpointDesc
					for _, ep := range alt.Endpoints {
						switch {
						case ep.Direction == gousb.EndpointDirectionIn && ep.TransferType == gousb.TransferTypeInterrupt:
							ev = ep
						case ep.Direction == gousb.EndpointDirectionIn && ep.TransferType == gousb.TransferTypeBulk:
							fe = ep
						case ep.Direction == gousb.EndpointDirectionOut && ep.TransferType == gousb.TransferTypeBulk:
							se = ep
						}
					}

					if se.Address > 0 && fe.Address > 0 && ev.Address > 0 {
						d := &DeviceGoUSB{
							devDesc:     desc,
							ifaceDesc:   iface,
							sendEPDesc:  se,
							fetchEPDesc: fe,
							eventEPDesc: ev,
							configDesc:  conf,

							iConfiguration: conf.Number,
							iInterface:     iface.Number,
							iAltSetting:    alt.Number,
						}
						mtpDev = append(mtpDev, d)
						return true
					}
				}
			}
		}
		return false
	})

	if err != nil {
		return nil, fmt.Errorf("failed to iterate USB devices: %s", err)
	}

	if len(mtpDev) == 0 {
		return nil, fmt.Errorf("found no MTP devices")
	} else if len(mtpDev) > 1 {
		var s []string
		for i, d := range mtpDev {
			s = append(s, fmt.Sprintf("%d. %04x:%04x", i+1, d.devDesc.Vendor, d.devDesc.Product))
		}
		return nil, fmt.Errorf("found multiple MTP devices: %s", strings.Join(s, ", "))
	}

	found := mtpDev[0]
	found.dev = devs[0]
	return found, nil
}

func candidateFromDeviceDescriptor(d *usb.Device) *DeviceDirect {
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
				m := DeviceDirect{}
				for _, s := range a.EndPoints {
					switch {
					case s.Direction() == usb.ENDPOINT_IN && s.TransferType() == usb.TRANSFER_TYPE_INTERRUPT:
						m.eventEP = s.EndpointAddress
					case s.Direction() == usb.ENDPOINT_IN && s.TransferType() == usb.TRANSFER_TYPE_BULK:
						m.fetchEP = s.EndpointAddress
					case s.Direction() == usb.ENDPOINT_OUT && s.TransferType() == usb.TRANSFER_TYPE_BULK:
						m.sendEP = s.EndpointAddress
					}
				}
				if m.sendEP > 0 && m.fetchEP > 0 && m.eventEP > 0 {
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
func FindDevices(c *usb.Context, vid, pid uint16) ([]*DeviceDirect, error) {
	l, err := c.GetDeviceList()
	if err != nil {
		return nil, err
	}

	var cands []*DeviceDirect
	for _, d := range l {
		v, _ := d.GetDeviceDescriptor()
		if vid != 0 && v.IdVendor != vid {
			continue
		} else if pid != 0 && v.IdProduct != pid {
			continue
		}
		cand := candidateFromDeviceDescriptor(d)
		if cand != nil {
			cands = append(cands, cand)
		}
	}
	l.Done()

	return cands, nil
}

// selectDevice finds a device that matches given pattern
func selectDevice(cands []*DeviceDirect, pattern string) (*DeviceDirect, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	var found []*DeviceDirect
	for _, cand := range cands {
		if err := cand.Open(); err != nil {
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

	if len(found) == 0 {
		return nil, fmt.Errorf("no device matched")
	}

	if len(found) > 1 {
		return nil, fmt.Errorf("mtp: more than 1 device: %s", strings.Join(ids, ","))
	}

	cand := found[0]
	config, err := cand.h.GetConfiguration()
	if err != nil {
		return nil, fmt.Errorf("could not get configuration of %v: %v",
			ids[0], err)
	}
	if config != cand.configValue {

		if err := cand.h.SetConfiguration(cand.configValue); err != nil {
			return nil, fmt.Errorf("could not set configuration of %v: %v",
				ids[0], err)
		}
	}
	return found[0], nil
}

// SelectDevice returns opened MTP device that matches the given pattern.
func SelectDevice(vid, pid uint16) (*DeviceDirect, error) {
	c := usb.NewContext()

	devs, err := FindDevices(c, vid, pid)
	if err != nil {
		return nil, err
	}
	if len(devs) == 0 {
		return nil, fmt.Errorf("no MTP devices found")
	}

	return selectDevice(devs, "")
}
