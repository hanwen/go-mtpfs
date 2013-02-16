package mtp

// These tests require a single Android MTP device that is connected
// and unlocked.

import (
	"bytes"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"
)


// VerboseTest returns true if the testing framework is run with -v.
func VerboseTest() bool {
	flag := flag.Lookup("test.v")
	return flag != nil && flag.Value.String() == "true"
}

func TestAndroid(t *testing.T) {
	dev, err := SelectDevice("")
	if err != nil {
		t.Fatal(err)
	}
	defer dev.Close()

	info := DeviceInfo{}
	err = dev.GetDeviceInfo(&info)
	if err != nil {
		t.Fatal("GetDeviceInfo failed:", err)
	}

	if !strings.Contains(info.MTPExtension, "android.com:") {
		t.Log("no android extensions", info.MTPExtension)
		return
	}

	if err = dev.Configure(); err != nil {
		t.Fatal("Configure failed:", err)
	}

	sids := Uint32Array{}
	err = dev.GetStorageIDs(&sids)
	if err != nil {
		t.Fatalf("GetStorageIDs failed: %v", err)
	}

	if len(sids.Values) == 0 {
		t.Fatalf("No storages")
	}

	id := sids.Values[0]

	// 500 + 512 triggers the null read case on both sides.
	const testSize = 500 + 512
	name := fmt.Sprintf("mtp-doodle-test%x", rand.Int31())

	send := ObjectInfo{
		StorageID:        id,
		ObjectFormat:     OFC_Undefined,
		ParentObject:     0xFFFFFFFF,
		Filename:         name,
		CompressedSize:   uint32(testSize),
		ModificationDate: time.Now(),
		Keywords:         "bla",
	}
	data := make([]byte, testSize)
	for i := range data {
		data[i] = byte('0' + i%10)
	}

	_, _, handle, err := dev.SendObjectInfo(id, 0xFFFFFFFF, &send)
	if err != nil {
		t.Fatal("SendObjectInfo failed:", err)
	} else {
		buf := bytes.NewBuffer(data)
		t.Logf("Sent objectinfo handle: 0x%x\n", handle)
		err = dev.SendObject(buf, int64(len(data)))
		if err != nil {
			t.Log("SendObject failed:", err)
		}
	}

	magicStr := "life universe"
	magicOff := 21
	magicSize := 42

	err = dev.AndroidBeginEditObject(handle)
	if err != nil {
		t.Errorf("AndroidBeginEditObject: %v", err)
		return
	} else {
		err = dev.AndroidTruncate(handle, int64(magicSize))
		if err != nil {
			t.Errorf("AndroidTruncate: %v", err)
		}
		buf := bytes.NewBufferString(magicStr)
		err = dev.AndroidSendPartialObject(handle, int64(magicOff), uint32(buf.Len()), buf)
		if err != nil {
			t.Errorf("AndroidSendPartialObject: %v", err)
		}
		if buf.Len() > 0 {
			t.Errorf("buffer not consumed")
		}
		err = dev.AndroidEndEditObject(handle)
		if err != nil {
			t.Errorf("AndroidEndEditObject: %v", err)
		}
	}

	buf := &bytes.Buffer{}
	err = dev.GetObject(handle, buf)
	if err != nil {
		t.Errorf("GetObject: %v", err)
	}

	if buf.Len() != magicSize {
		t.Errorf("truncate failed:: %v", err)
	}
	for i := 0; i < len(magicStr); i++ {
		data[i+magicOff] = magicStr[i]
	}
	want := string(data[:magicSize])
	if buf.String() != want {
		t.Errorf("read result was %q, want %q", buf.String(), want)
	}

	buf = &bytes.Buffer{}
	err = dev.AndroidGetPartialObject64(handle, buf, int64(magicOff), 5)
	if err != nil {
		t.Errorf("AndroidGetPartialObject64: %v", err)
	}
	want = magicStr[:5]
	got := buf.String()
	if got != want {
		t.Errorf("AndroidGetPartialObject64: got %q want %q", got, want)
	}

	// Try write at end of file.
	err = dev.AndroidBeginEditObject(handle)
	if err != nil {
		t.Errorf("AndroidBeginEditObject: %v", err)
		return
	} else {
		buf := bytes.NewBufferString(magicStr)
		err = dev.AndroidSendPartialObject(handle, int64(magicSize), uint32(buf.Len()), buf)
		if err != nil {
			t.Errorf("AndroidSendPartialObject: %v", err)
		}
		if buf.Len() > 0 {
			t.Errorf("buffer not consumed")
		}
		err = dev.AndroidEndEditObject(handle)
		if err != nil {
			t.Errorf("AndroidEndEditObject: %v", err)
		}
	}
	buf = &bytes.Buffer{}
	err = dev.GetObject(handle, buf)
	if err != nil {
		t.Errorf("GetObject: %v", err)
	}
	want = string(data[:magicSize]) + magicStr
	got = buf.String()
	if got != want {
		t.Errorf("GetObject: got %q want %q", got, want)
	}

	err = dev.DeleteObject(handle)
	if err != nil {
		t.Fatalf("DeleteObject failed: %v", err)
	}
}

func TestDeviceProperties(t *testing.T) {
	dev, err := SelectDevice("")
	if err != nil {
		t.Fatal(err)
	}
	defer dev.Close()
	
	dev.DataPrint = VerboseTest()
	dev.DebugPrint = VerboseTest()
	dev.USBPrint = VerboseTest()
	err = dev.Configure()
	if err != nil {
		t.Log("Configure failed:", err)
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
		t.Fatal("GetDevicePropDesc FriendlyName failed:", err)
	} else {
		t.Logf("%s: %#v\n", DPC_names[DPC_MTP_DeviceFriendlyName], friendly)
	}
	before := friendly.CurrentValue.(string)

	newVal := fmt.Sprintf("gomtp device_test %x", rand.Int31())
	str := StringValue{newVal}
	err = dev.SetDevicePropValue(DPC_MTP_DeviceFriendlyName, &str)
	if err != nil {
		t.Error("SetDevicePropValue failed:", err)
	}

	str.Value = ""
	err = dev.GetDevicePropValue(DPC_MTP_DeviceFriendlyName, &str)
	if err != nil {
		t.Error("GetDevicePropValue failed:", err)
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
		t.Error("SetDevicePropValue failed:", err)
	}

	// Test object properties.
	props := Uint16Array{}
	err = dev.GetObjectPropsSupported(OFC_Undefined, &props)
	if err != nil {
		t.Errorf("GetObjectPropsSupported failed: %v\n", err)
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
			t.Errorf("GetObjectPropDesc(%s) failed: %v\n", name, err)
		} else {
			t.Logf("GetObjectPropDesc(%s) value: %#v %T\n", name, objPropDesc,
				InstantiateType(objPropDesc.DataType).Interface())
		}
	}

}

func TestDeviceInfo(t *testing.T) {
	dev, err := SelectDevice("")
	if err != nil {
		t.Fatal(err)
	}
	defer dev.Close()

	i, _ := dev.ID()
	t.Log("device:", i)
	info := DeviceInfo{}
	err = dev.GetDeviceInfo(&info)
	if err != nil {
		t.Error("GetDeviceInfo failed:", err)
	} else {
		t.Logf("%v\n", &info)
	}
}

func TestDeviceStorage(t *testing.T) {
	dev, err := SelectDevice("")
	if err != nil {
		t.Fatal(err)
	}
	defer dev.Close()

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
		t.Log("Configure failed:", err)
	}

	sids := Uint32Array{}
	err = dev.GetStorageIDs(&sids)
	if err != nil {
		t.Fatalf("GetStorageIDs failed: %v", err)
	} else {
		t.Logf("%#v\n", sids)
	}

	if len(sids.Values) == 0 {
		t.Fatalf("No storages")
	}

	id := sids.Values[0]
	var storageInfo StorageInfo
	dev.GetStorageInfo(id, &storageInfo)
	if err != nil {
		t.Fatalf("GetStorageInfo failed:", err)
	} else {
		t.Logf("%#v\n", storageInfo)
	}

	resp, err := dev.GetNumObjects(id, 0x0, 0x0)
	if err != nil {
		t.Fatalf("GenericRPC failed:", err)
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
		t.Fatalf("SendObjectInfo failed:", err)
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
		t.Fatalf("GetObjectHandles failed: %v", err)
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
		t.Fatalf("GetObjectInfo failed: %v", err)
	} else {
		t.Logf("info %#v\n", backInfo)
	}

	var objSize Uint64Value
	err = dev.GetObjectPropValue(handle, OPC_ObjectSize, &objSize)
	if err != nil {
		t.Fatalf("GetObjectPropValue failed: %v", err)
	} else {
		t.Logf("info %#v\n", objSize)
	}

	if objSize.Value != testSize {
		t.Errorf("object size error: got %v, want %v", objSize.Value, testSize)
	}

	backBuf := &bytes.Buffer{}
	err = dev.GetObject(handle, backBuf)
	if err != nil {
		t.Fatalf("GetObject failed: %v", err)
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
		t.Fatalf("DeleteObject failed: %v", err)
	}
}
