package usb

import (
	"testing"
)

func TestDevice(t *testing.T) {
	c := NewContext()
	defer c.Exit()
	l, err := c.GetDeviceList()
	if err != nil {
		t.Fatal("GetDeviceList failed:", err)
	}

	for _, dev := range l {
		t.Logf("bus 0x%x addr 0x%x speed 0x%x\n",
			dev.GetBusNumber(), dev.GetDeviceAddress(), dev.GetDeviceSpeed())
		dd, err := dev.GetDeviceDescriptor()
		if err != nil {
			t.Logf("GetDeviceDescriptor failed: %v", err)
			continue
		}

		t.Logf("Vendor/Product %x:%x Class/subclass/protocol %x:%x:%x: %s\n",
			dd.IdVendor, dd.IdProduct, dd.DeviceClass, dd.DeviceSubClass, dd.DeviceProtocol, CLASS_names[dd.DeviceClass])

		stringDescs := []byte{
			dd.Manufacturer, dd.Product, dd.SerialNumber,
		}

		for i := 0; i < int(dd.NumConfigurations); i++ {
			cd, err := dev.GetConfigDescriptor(byte(i))
			if err != nil {
				t.Logf("GetConfigDescriptor failed: %v", err)
				continue
			}
			stringDescs = append(stringDescs, cd.ConfigurationIndex)
			t.Logf(" config value %x, attributes %x power %d\n", cd.ConfigurationValue,
				cd.Attributes, cd.MaxPower)
			for idx, iface := range cd.Interfaces {
				t.Logf("  iface %d\n", idx)
				for _, alt := range iface.AltSetting {
					t.Logf("   num %d class/subclass/protocol %x/%x/%x\n",
						alt.InterfaceNumber, alt.InterfaceClass, alt.InterfaceSubClass, alt.InterfaceProtocol)
					for _, ep := range alt.EndPoints {
						t.Logf("    %v", &ep)
					}
					stringDescs = append(stringDescs, alt.InterfaceStringIndex)
				}
			}
		}

		dh, err := dev.Open()
		if err != nil {
			t.Logf("can't open: %v", err)
			continue
		}

		for _, c := range stringDescs {
			str, err := dh.GetStringDescriptorASCII(c)
			if err != nil {
				t.Logf("GetStringDescriptorASCII %d failed: %v", c, err)
				continue
			}
			t.Logf(" desc %d: %s", c, str)
		}
		dh.Close()
	}
}
