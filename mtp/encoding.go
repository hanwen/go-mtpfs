package mtp

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"
)

var byteOrder = binary.LittleEndian

func decodeStr(r io.Reader) (string, error) {
	var szSlice [1]byte
	_, err := r.Read(szSlice[:])
	if err != nil {
		return "", err
	}
	sz := int(szSlice[0])
	if sz == 0 {
		return "", nil
	}
	utfStr := make([]byte, 4*sz)
	data := make([]byte, 2*sz)
	n, err := r.Read(data)
	if err != nil {
		return "", err
	}
	if n < len(data) {
		return "", fmt.Errorf("underflow")
	}
	w := 0
	for i := 0; i < int(2*sz); i += 2 {
		cp := byteOrder.Uint16(data[i:])
		w += utf8.EncodeRune(utfStr[w:], rune(cp))
	}
	if utfStr[w-1] == 0 {
		w--
	}
	s := string(utfStr[:w])
	return s, nil
}

func encodeStr(buf []byte, s string) ([]byte, error) {
	if s == "" {
		buf[0] = 0
		return buf[:1], nil
	}

	codepoints := 0
	buf = append(buf[:0], 0)

	var rune [2]byte
	for _, r := range s {
		byteOrder.PutUint16(rune[:], uint16(r))
		buf = append(buf, rune[0], rune[1])
		codepoints++
	}
	buf = append(buf, 0, 0)
	codepoints++
	if codepoints > 254 {
		return nil, fmt.Errorf("string too long")
	}

	buf[0] = byte(codepoints)
	return buf, nil
}

func encodeStrField(w io.Writer, f reflect.Value) error {
	out := make([]byte, 2*f.Len()+4)
	enc, err := encodeStr(out, f.Interface().(string))
	if err != nil {
		return err
	}
	_, err = w.Write(enc)
	return err
}

func kindSize(k reflect.Kind) int {
	switch k {
	case reflect.Int8:
		return 1
	case reflect.Int16:
		return 2
	case reflect.Int32:
		return 4
	case reflect.Int64:
		return 8
	case reflect.Uint8:
		return 1
	case reflect.Uint16:
		return 2
	case reflect.Uint32:
		return 4
	default:
		panic(fmt.Sprintf("unknown kind %v", k))
	}
	return 0
}

var nullValue reflect.Value

func decodeArray(r io.Reader, t reflect.Type) (reflect.Value, error) {
	var sz uint32
	err := binary.Read(r, byteOrder, &sz)
	if err != nil {
		return nullValue, err
	}

	ksz := int(kindSize(t.Elem().Kind()))

	data := make([]byte, int(sz)*ksz)
	_, err = r.Read(data)
	if err != nil {
		return nullValue, err
	}
	slice := reflect.MakeSlice(t, int(sz), int(sz))
	for i := 0; i < int(sz); i++ {
		from := data[i*ksz:]
		var val uint64
		switch ksz {
		case 1:
			val = uint64(from[0])
		case 2:
			val = uint64(byteOrder.Uint16(from[0:]))
		case 4:
			val = uint64(byteOrder.Uint32(from[0:]))
		default:
			panic("unimp")
		}

		slice.Index(i).SetUint(val)
	}
	return slice, nil
}

func encodeArray(w io.Writer, val reflect.Value) error {
	sz := uint32(val.Len())
	err := binary.Write(w, byteOrder, &sz)
	if err != nil {
		return err
	}

	kind := val.Type().Elem().Kind()
	ksz := 0
	if kind == reflect.Interface {
		log.Println("OK")
		ksz = int(kindSize(val.Index(0).Elem().Kind()))
	} else {
		ksz = int(kindSize(kind))
	}
	data := make([]byte, int(sz)*ksz)
	for i := 0; i < int(sz); i++ {
		elt := val.Index(i)
		to := data[i*ksz:]

		switch kind {
		case reflect.Uint8:
			to[0] = byte(elt.Uint())
		case reflect.Uint16:
			byteOrder.PutUint16(to, uint16(elt.Uint()))
		case reflect.Uint32:
			byteOrder.PutUint32(to, uint32(elt.Uint()))
		case reflect.Uint64:
			byteOrder.PutUint64(to, uint64(elt.Uint()))

		case reflect.Int8:
			to[0] = byte(elt.Int())
		case reflect.Int16:
			byteOrder.PutUint16(to, uint16(elt.Int()))
		case reflect.Int32:
			byteOrder.PutUint32(to, uint32(elt.Int()))
		case reflect.Int64:
			byteOrder.PutUint64(to, uint64(elt.Int()))
		default:
			panic(fmt.Sprintf("unimplemented: encode for kind %v", kind))
		}
	}
	_, err = w.Write(data)
	return err
}

var timeType = reflect.ValueOf(time.Now()).Type()

const timeFormat = "20060102T150405"
const timeFormatNumTZ = "20060102T150405-0700"

var zeroTime = time.Time{}

func encodeTime(w io.Writer, f reflect.Value) error {
	tptr := f.Addr().Interface().(*time.Time)
	s := ""
	if !tptr.Equal(zeroTime) {
		s = tptr.Format(timeFormat)
	}

	out := make([]byte, 2*len(s)+3)
	out, err := encodeStr(out, s)
	if err != nil {
		return err
	}
	_, err = w.Write(out)
	return err
}

func decodeTime(r io.Reader, f reflect.Value) error {
	s, err := decodeStr(r)
	if err != nil {
		return err
	}
	var t time.Time
	if s != "" {
		// Samsung has trailing dots.
		s = strings.TrimRight(s, ".")
		t, err = time.Parse(timeFormat, s)
		if err != nil {
			// Nokia lumia has numTZ
			t, err = time.Parse(timeFormatNumTZ, s)
			if err != nil {
				return err
			}
		}
	}
	f.Set(reflect.ValueOf(t))
	return nil
}

func decodeField(r io.Reader, f reflect.Value, typeSelector DataTypeSelector) error {
	if !f.CanAddr() {
		return fmt.Errorf("canaddr false")
	}

	if f.Type() == timeType {
		return decodeTime(r, f)
	}

	switch f.Kind() {
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		fallthrough
	case reflect.Int64:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int8:
		return binary.Read(r, byteOrder, f.Addr().Interface())
	case reflect.String:
		s, err := decodeStr(r)
		if err != nil {
			return err
		}
		f.SetString(s)
	case reflect.Slice:
		sl, err := decodeArray(r, f.Type())
		if err != nil {
			return err
		}
		f.Set(sl)
	case reflect.Interface:
		val := InstantiateType(typeSelector)
		decodeField(r, val, typeSelector)
		f.Set(val)
	default:
		panic(fmt.Sprintf("unimplemented kind %v", f))
	}
	return nil
}

func encodeField(w io.Writer, f reflect.Value) error {
	if f.Type() == timeType {
		return encodeTime(w, f)
	}

	switch f.Kind() {
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		fallthrough
	case reflect.Int64:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int8:
		return binary.Write(w, byteOrder, f.Interface())
	case reflect.String:
		return encodeStrField(w, f)
	case reflect.Slice:
		return encodeArray(w, f)
	case reflect.Interface:
		return encodeField(w, f.Elem())
	default:
		panic(fmt.Sprintf("unimplemented kind %v", f))
	}
	return nil
}

// Decode MTP data stream into data structure.
func Decode(r io.Reader, iface interface{}) error {
	decoder, ok := iface.(Decoder)
	if ok {
		return decoder.Decode(r)
	}

	typeSel := DataTypeSelector(0xfe)
	return decodeWithSelector(r, iface, typeSel)
}

func decodeWithSelector(r io.Reader, iface interface{}, typeSel DataTypeSelector) error {
	val := reflect.ValueOf(iface)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("need ptr argument: %T", iface)
	}
	val = val.Elem()
	t := val.Type()

	for i := 0; i < t.NumField(); i++ {
		err := decodeField(r, val.Field(i), typeSel)

		if err != nil {
			return err
		}
		if val.Field(i).Type().Name() == "DataTypeSelector" {
			typeSel = val.Field(i).Interface().(DataTypeSelector)
		}

	}
	return nil
}

// Encode MTP data stream into data structure.
func Encode(w io.Writer, iface interface{}) error {
	encoder, ok := iface.(Encoder)
	if ok {
		return encoder.Encode(w)
	}

	val := reflect.ValueOf(iface)
	if val.Kind() != reflect.Ptr {
		msg := fmt.Sprintf("need ptr argument: %T", iface)
		return fmt.Errorf(msg)
		//panic("need ptr argument: %T")
	}
	val = val.Elem()
	t := val.Type()

	for i := 0; i < t.NumField(); i++ {
		err := encodeField(w, val.Field(i))
		if err != nil {
			return err
		}
	}
	return nil

}

// Instantiates an object of wanted type as addressable value.
func InstantiateType(t DataTypeSelector) reflect.Value {
	var val interface{}
	switch t {
	case DTC_INT8:
		v := int8(0)
		val = &v
	case DTC_UINT8:
		v := int8(0)
		val = &v
	case DTC_INT16:
		v := int16(0)
		val = &v
	case DTC_UINT16:
		v := uint16(0)
		val = &v
	case DTC_INT32:
		v := int32(0)
		val = &v
	case DTC_UINT32:
		v := uint32(0)
		val = &v
	case DTC_INT64:
		v := int64(0)
		val = &v
	case DTC_UINT64:
		v := uint64(0)
		val = &v
	case DTC_INT128:
		v := [16]byte{}
		val = &v
	case DTC_UINT128:
		v := [16]byte{}
		val = &v
	case DTC_STR:
		s := ""
		val = &s
	default:
		panic(fmt.Sprintf("type not known 0x%x", t))
	}

	return reflect.ValueOf(val).Elem()
}

func decodePropDescForm(r io.Reader, selector DataTypeSelector, formFlag uint8) (DataDependentType, error) {
	if formFlag == DPFF_Range {
		f := PropDescRangeForm{}
		err := decodeWithSelector(r, reflect.ValueOf(&f).Interface(),
			selector)
		return &f, err
	} else if formFlag == DPFF_Enumeration {
		f := PropDescEnumForm{}
		err := decodeWithSelector(r, reflect.ValueOf(&f).Interface(),
			selector)
		return &f, err
	}
	return nil, nil
}

func (pd *ObjectPropDesc) Decode(r io.Reader) (err error) {
	err = Decode(r, &pd.ObjectPropDescFixed)
	if err != nil {
		return err
	}
	form, err := decodePropDescForm(r, pd.DataType, pd.FormFlag)
	pd.Form = form
	return err
}

func (pd *DevicePropDesc) Decode(r io.Reader) (err error) {
	err = Decode(r, &pd.DevicePropDescFixed)
	if err != nil {
		return err
	}
	form, err := decodePropDescForm(r, pd.DataType, pd.FormFlag)
	pd.Form = form
	return err
}

func (pd *DevicePropDesc) Encode(w io.Writer) (err error) {
	err = Encode(w, &pd.DevicePropDescFixed)
	if err != nil {
		return err
	}

	err = Encode(w, pd.Form)
	return err
}

func (pd *ObjectPropDesc) Encode(w io.Writer) (err error) {
	err = Encode(w, &pd.ObjectPropDescFixed)
	if err != nil {
		return err
	}
	return Encode(w, pd.Form)
}
