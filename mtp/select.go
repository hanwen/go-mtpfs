package mtp

import (
	"fmt"
	"strings"

	"github.com/google/gousb"
	"github.com/hanwen/usb"
)

func SelectDeviceGoUSB(ctx *gousb.Context, vid, pid uint16) (*DeviceGoUSB, error) {
	var mtpDev []*DeviceGoUSB

	if vid != 0 && pid != 0 {
		log.WithField("prefix", "usb").Infof("searching %04d:%04d", vid, pid)
	}

	devs, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		v, p := uint16(desc.Vendor), uint16(desc.Product)
		if vid != 0 && pid != 0 && (v != vid || p != pid) {
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

						log.WithField("prefix", "usb").Infof("found: %04x:%04x", v, p)
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

// SelectDeviceDirect returns opened MTP device that matches the given pattern.
func SelectDeviceDirect(vid, pid uint16) (*DeviceDirect, error) {
	c := usb.NewContext()

	l, err := c.GetDeviceList()
	if err != nil {
		return nil, err
	}
	defer l.Done()

	var devs []*DeviceDirect
	for _, d := range l {
		v, _ := d.GetDeviceDescriptor()
		if vid != 0 && v.IdVendor != vid {
			continue
		} else if pid != 0 && v.IdProduct != pid {
			continue
		}
		cand := candidateFromDeviceDescriptor(d)
		if cand != nil {
			log.WithField("prefix", "usb").Infof("found: %04x:%04x", v.IdVendor, v.IdProduct)
			devs = append(devs, cand)
		}
	}

	if len(devs) == 0 {
		return nil, fmt.Errorf("no MTP devices found")
	}

	dev := devs[0]
	vendor, product := dev.devDescr.IdVendor, dev.devDescr.IdProduct

	if len(devs) > 1 {
		log.WithField("prefix", "mtp").Warningf("detected more than 1 device, opening the first device: %04x:%04x", vendor, product)
	}

	if err := dev.Open(); err != nil {
		return nil, fmt.Errorf("could not open %04x:%04x", dev.devDescr.IdVendor, dev.devDescr.IdProduct)
	}

	config, err := dev.h.GetConfiguration()
	if err != nil {
		return nil, fmt.Errorf("could not get configuration of %04x:%04x: %v", vendor, product, err)
	}

	if config != dev.configValue {
		if err := dev.h.SetConfiguration(dev.configValue); err != nil {
			return nil, fmt.Errorf("could not set configuration of %04x:%04x: %v", vendor, product, err)
		}
	}

	dev.Timeout = 1000
	return dev, nil
}
