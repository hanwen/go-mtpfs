package mtp

// These tests require a single Android MTP device that is connected
// and unlocked.

import (
	"bytes"
	"flag"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

// VerboseTest returns true if the testing framework is run with -v.
func VerboseTest() bool {
	flag := flag.Lookup("test.v")
	return flag != nil && flag.Value.String() == "true"
}

func setDebug(dev *DeviceDirect) {
	dev.Debug.Data = VerboseTest()
	dev.Debug.MTP = VerboseTest()
	dev.Debug.USB = VerboseTest()
}

func TestDeviceProperties(t *testing.T) {
	dev, err := SelectDeviceDirect(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer dev.Close()

	setDebug(dev)
	err = dev.Configure()
	if err != nil {
		t.Fatal("configure failed:", err)
	}

	// Test non-supported device property first.
	var battery DevicePropDesc
	err = dev.GetDevicePropDesc(DPC_BatteryLevel, &battery)
	if err != nil {
		// Not an error; not supported on Android.
		t.Log("battery failed:", err)
	} else {
		t.Logf("%#v\n", battery)
	}

	var friendly DevicePropDesc
	err = dev.GetDevicePropDesc(DPC_MTP_DeviceFriendlyName, &friendly)
	if err != nil {
		t.Fatal("getDevicePropDesc FriendlyName failed:", err)
	} else {
		t.Logf("%s: %#v\n", DPC_names[DPC_MTP_DeviceFriendlyName], friendly)
	}
	before := friendly.CurrentValue.(string)

	newVal := fmt.Sprintf("gomtp device_test %x", rand.Int31())
	str := StringValue{newVal}
	err = dev.SetDevicePropValue(DPC_MTP_DeviceFriendlyName, &str)
	if err != nil {
		t.Error("setDevicePropValue failed:", err)
	}

	str.Value = ""
	err = dev.GetDevicePropValue(DPC_MTP_DeviceFriendlyName, &str)
	if err != nil {
		t.Error("getDevicePropValue failed:", err)
	}
	if str.Value != newVal {
		t.Logf("got %q for property value, want %q\n", str.Value, newVal)
	}

	err = dev.ResetDevicePropValue(DPC_MTP_DeviceFriendlyName)
	if err != nil {
		// For some reason, this is not supported? Returns
		// unknown error code 0xffff
		t.Log("ResetDevicePropValue failed:", err)
	}

	str = StringValue{before}
	err = dev.SetDevicePropValue(DPC_MTP_DeviceFriendlyName, &str)
	if err != nil {
		t.Error("setDevicePropValue failed:", err)
	}

	// Test object properties.
	props := Uint16Array{}
	err = dev.GetObjectPropsSupported(OFC_Undefined, &props)
	if err != nil {
		t.Errorf("getObjectPropsSupported failed: %v\n", err)
	} else {
		t.Logf("GetObjectPropsSupported (OFC_Undefined) value: %s\n", getNames(OPC_names, props.Values))
	}

	for _, p := range props.Values {
		var objPropDesc ObjectPropDesc
		if p == OPC_PersistantUniqueObjectIdentifier {
			// can't deal with int128.
			continue
		}

		err = dev.GetObjectPropDesc(p, OFC_Undefined, &objPropDesc)
		name := OPC_names[int(p)]
		if err != nil {
			t.Errorf("getObjectPropDesc(%s) failed: %v\n", name, err)
		} else {
			t.Logf("GetObjectPropDesc(%s) value: %#v %T\n", name, objPropDesc,
				InstantiateType(objPropDesc.DataType).Interface())
		}
	}

}

func TestDeviceInfo(t *testing.T) {
	dev, err := SelectDeviceDirect(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer dev.Close()
	setDebug(dev)

	i, _ := dev.ID()
	t.Log("device:", i)
	info := DeviceInfo{}
	err = dev.GetDeviceInfo(&info)
	if err != nil {
		t.Error("getDeviceInfo failed:", err)
	} else {
		t.Logf("%v\n", &info)
	}
}

func TestDeviceStorage(t *testing.T) {
	dev, err := SelectDeviceDirect(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer dev.Close()
	setDebug(dev)

	i, _ := dev.ID()
	t.Log("device:", i)

	info := DeviceInfo{}
	err = dev.GetDeviceInfo(&info)
	if err != nil {
		t.Log("GetDeviceInfo failed:", err)
	} else {
		t.Logf("device info %v\n", &info)
	}

	err = dev.Configure()
	if err != nil {
		t.Fatal("configure failed:", err)
	}

	sids := Uint32Array{}
	err = dev.GetStorageIDs(&sids)
	if err != nil {
		t.Fatalf("getStorageIDs failed: %v", err)
	} else {
		t.Logf("%#v\n", sids)
	}

	if len(sids.Values) == 0 {
		t.Fatalf("no storages")
	}

	id := sids.Values[0]
	var storageInfo StorageInfo
	dev.GetStorageInfo(id, &storageInfo)
	if err != nil {
		t.Fatalf("getStorageInfo failed: %s", err)
	} else {
		t.Logf("%#v\n", storageInfo)
	}

	resp, err := dev.GetNumObjects(id, 0x0, 0x0)
	if err != nil {
		t.Fatalf("genericRPC failed: %s", err)
	} else {
		t.Logf("num objects %#v\n", resp)
	}

	// 500 + 512 triggers the null read case on both sides.
	const testSize = 500 + 512
	data := make([]byte, testSize)
	for i := range data {
		data[i] = byte(i % 256)
	}

	name := fmt.Sprintf("go-mtp-test%x", rand.Int31())
	buf := bytes.NewBuffer(data)
	send := ObjectInfo{
		StorageID:        id,
		ObjectFormat:     OFC_Undefined,
		ParentObject:     0xFFFFFFFF,
		Filename:         name,
		CompressedSize:   uint32(len(data)),
		ModificationDate: time.Now(),
		Keywords:         "bla",
	}

	_, _, handle, err := dev.SendObjectInfo(id, 0xFFFFFFFF, &send)
	if err != nil {
		t.Fatalf("sendObjectInfo failed: %s", err)
	} else {
		t.Logf("Sent objectinfo handle: 0x%x\n", handle)
		err = dev.SendObject(buf, int64(len(data)))
		if err != nil {
			t.Log("SendObject failed:", err)
		}
	}

	hs := Uint32Array{}
	err = dev.GetObjectHandles(id,
		OFC_Undefined,
		//OFC_Association,
		0xFFFFFFFF, &hs)

	if err != nil {
		t.Fatalf("getObjectHandles failed: %v", err)
	} else {
		found := false
		for _, h := range hs.Values {
			if h == handle {
				found = true
				break
			}
		}

		if !found {
			t.Fatalf("uploaded file is not there")
		}
	}

	var backInfo ObjectInfo
	err = dev.GetObjectInfo(handle, &backInfo)
	if err != nil {
		t.Fatalf("getObjectInfo failed: %v", err)
	} else {
		t.Logf("info %#v\n", backInfo)
	}

	var objSize Uint64Value
	err = dev.GetObjectPropValue(handle, OPC_ObjectSize, &objSize)
	if err != nil {
		t.Fatalf("getObjectPropValue failed: %v", err)
	} else {
		t.Logf("info %#v\n", objSize)
	}

	if objSize.Value != testSize {
		t.Errorf("object size error: got %v, want %v", objSize.Value, testSize)
	}

	backBuf := &bytes.Buffer{}
	err = dev.GetObject(handle, backBuf)
	if err != nil {
		t.Fatalf("getObject failed: %v", err)
	} else {
		if bytes.Compare(backBuf.Bytes(), data) != 0 {
			t.Fatalf("back comparison failed.")
		}
	}

	newName := fmt.Sprintf("mtp-doodle-test%x", rand.Int31())
	err = dev.SetObjectPropValue(handle, OPC_ObjectFileName, &StringValue{newName})
	if err != nil {
		t.Errorf("error renaming object: %v", err)
	}

	err = dev.DeleteObject(handle)
	if err != nil {
		t.Fatalf("deleteObject failed: %v", err)
	}
}
