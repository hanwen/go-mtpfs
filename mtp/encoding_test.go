package mtp

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

const deviceInfoStr = `6400 0600
0000 6400 266d 0069 0063 0072 006f 0073
006f 0066 0074 002e 0063 006f 006d 003a
0020 0031 002e 0030 003b 0020 0061 006e
0064 0072 006f 0069 0064 002e 0063 006f
006d 003a 0020 0031 002e 0030 003b 0000
0000 001e 0000 0001 1002 1003 1004 1005
1006 1007 1008 1009 100a 100b 100c 100d
1014 1015 1016 1017 101b 1001 9802 9803
9804 9805 9810 9811 98c1 95c2 95c3 95c4
95c5 9504 0000 0002 4003 4004 4005 4003
0000 0001 d402 d403 5000 0000 001a 0000
0000 3001 3004 3005 3008 3009 300b 3001
3802 3804 3807 3808 380b 380d 3801 b902
b903 b982 b983 b984 b905 ba10 ba11 ba14
ba82 ba06 b905 6100 7300 7500 7300 0000
084e 0065 0078 0075 0073 0020 0037 0000
0004 3100 2e00 3000 0000 1130 0031 0035
0064 0032 0035 0036 0038 0035 0038 0034
0038 0030 0032 0031 0062 0000 00`

const objInfoStr = `0100 0100
0130 0000 0010 0000 0000 0000 0000 0000
0000 0000 0000 0000 0000 0000 0000 0000
0000 0000 0000 0000 0000 0000 0000 0000
064d 0075 0073 0069 0063 0000 0000 1032
0030 0030 0030 0030 0031 0030 0031 0054
0031 0039 0031 0031 0033 0030 0000 0000`

func parseHex(s string) []byte {
	hex := strings.Replace(s, " ", "", -1)
	hex = strings.Replace(hex, "\n", "", -1)
	buf := bytes.NewBufferString(hex)
	bin := make([]byte, len(hex)/2)

	_, err := fmt.Fscanf(buf, "%x", &bin)
	if err != nil {
		panic(err)
	}
	if buf.Len() > 0 {
		panic("consume")
	}
	return bin
}

func diffIndex(a, b []byte) error {
	l := len(b)
	if len(a) < len(b) {
		l = len(a)
	}

	for i := 0; i < l; i++ {
		if a[i] != b[i] {
			return fmt.Errorf("data idx 0x%x got %x want %x",
				i, a[i], b[i])
		}
	}

	if len(a) != len(b) {
		return fmt.Errorf("length mismatch got %d want %d",
			len(a), len(b))
	}
	return nil
}

func TestDecode(t *testing.T) {
	bin := parseHex(deviceInfoStr)
	var info DeviceInfo
	buf := bytes.NewBuffer(bin)
	err := Decode(buf, &info)
	if err != nil {
		t.Fatalf("unexpected decode error %v", err)
	}

	buf = &bytes.Buffer{}
	err = Encode(buf, &info)
	if err != nil {
		t.Fatalf("unexpected encode error %v", err)
	}

	err = diffIndex(buf.Bytes(), bin)
	if err != nil {
		t.Error(err)

		fmt.Println("got")
		hexDump(buf.Bytes())
		fmt.Println("want")
		hexDump(bin)
	}
}

func TestDecodeObjInfo(t *testing.T) {
	bin := parseHex(objInfoStr)
	var info ObjectInfo
	buf := bytes.NewBuffer(bin)
	err := Decode(buf, &info)
	if err != nil {
		t.Fatalf("unexpected decode error %v", err)
	}

	buf = &bytes.Buffer{}
	err = Encode(buf, &info)
	if err != nil {
		t.Fatalf("unexpected encode error %v", err)
	}

	err = diffIndex(buf.Bytes(), bin)
	if err != nil {
		t.Error(err)

		fmt.Println("got")
		hexDump(buf.Bytes())
		fmt.Println("want")
		hexDump(bin)
	}
}

type TestStr struct {
	S string
}

func TestEncodeStrEmpty(t *testing.T) {
	b := &bytes.Buffer{}
	err := Encode(b, &TestStr{})
	if err != nil {
		t.Fatalf("unexpected encode error %v", err)
	}
	if string(b.Bytes()) != "\000" {
		t.Fatalf("string encode mismatch %q ", b.Bytes())
	}
}

type TimeValue struct {
	Value time.Time
}

func TestDecodeTime(t *testing.T) {
	ts := &TestStr{"20120101T010022."}
	samsung := &bytes.Buffer{}
	if err := Encode(samsung, ts); err != nil {
		t.Fatalf("str encode failed: %v", err)
	}

	tv := &TimeValue{}
	if err := Decode(samsung, tv); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	buf := bytes.Buffer{}
	if err := Encode(&buf, tv); err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	if err := Decode(&buf, ts); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	want := "20120101T010022"
	got := ts.S
	if got != want {
		t.Errorf("time encode/decode: got %q want %q", got, want)
	}
}

func TestVariantDPD(t *testing.T) {
	uint16range := PropDescRangeForm{
		MinimumValue: uint16(1),
		MaximumValue: uint16(11),
		StepSize:     uint16(2),
	}

	fixed := DevicePropDescFixed{
		DevicePropertyCode:  DPC_BatteryLevel,
		DataType:            DTC_UINT16,
		GetSet:              DPGS_GetSet,
		FactoryDefaultValue: uint16(3),
		CurrentValue:        uint16(5),
		FormFlag:            DPFF_Range,
	}

	dp := DevicePropDesc{fixed, &uint16range}

	buf := &bytes.Buffer{}
	err := Encode(buf, &dp)
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}

	back := DevicePropDesc{}
	if err := Decode(buf, &back); err != nil {
		t.Fatalf("encode error: %v", err)
	}

	if !reflect.DeepEqual(back, dp) {
		t.Fatalf("reflect.DeepEqual failed: got %#v, want %#v",
			dp, back)
	}
}

func DisabledTestVariantOPD(t *testing.T) {
	uint16enum := PropDescEnumForm{
		Values: []DataDependentType{uint16(1), uint16(11), uint16(2)},
	}

	fixed := ObjectPropDescFixed{
		ObjectPropertyCode:  OPC_WirelessConfigurationFile,
		DataType:            DTC_UINT16,
		GetSet:              DPGS_GetSet,
		FactoryDefaultValue: uint16(3),
		GroupCode:           0x21,
		FormFlag:            DPFF_Enumeration,
	}

	dp := ObjectPropDesc{fixed, &uint16enum}

	buf := &bytes.Buffer{}
	err := Encode(buf, &dp)
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}

	back := ObjectPropDesc{}
	if err := Decode(buf, &back); err != nil {
		t.Fatalf("encode error: %v", err)
	}

	if !reflect.DeepEqual(back, dp) {
		t.Fatalf("reflect.DeepEqual failed: got %#v, want %#v",
			dp, back)
	}
}
