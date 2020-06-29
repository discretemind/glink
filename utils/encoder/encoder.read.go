package encoder

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"reflect"
)

type bitReader struct {
	Error error
}

func newBitReader() (result *bitReader) {
	result = &bitReader{}
	return
}

func (r *bitReader) decode(data []byte, t reflect.Type, v reflect.Value) error {
	switch v.Kind() {
	case reflect.Ptr:
		t = t.Elem()
		v = v.Elem()
	default:
		return fmt.Errorf("invalid type %s", reflect.TypeOf(v).String())
	}

	d := Decoder(data)
	r.read(d, t, v)
	return nil
}

func (r *bitReader) read(d *dec, t reflect.Type, v reflect.Value) {
	switch t.Kind() {
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			fieldValue := v.Field(i)
			f := t.Field(i)
			if f.Tag.Get("enc") != "-" {
				r.read(d, fieldValue.Type(), fieldValue)
			}
		}

	case reflect.Array:
		if v.Len() > 0 {
			tp := v.Index(0).Type()
			switch tp.Kind() {
			case reflect.Uint8:
				bytes := make([]byte, v.Len())
				for i := 0; i < v.Len(); i++ {
					bytes[i] = d.readByte()
				}
				reflect.Copy(v, reflect.ValueOf(bytes))
			default:
				for i := 0; i < v.Len(); i++ {
					r.read(d, tp, v.Index(i))
				}
			}
		}

	case reflect.Slice:
		length := int(d.readUintBinary(4))
		reflectSlice := reflect.MakeSlice(v.Type(), length, length)
		v.Set(reflectSlice)
		if length == 0 {
			return
		}
		t := v.Index(0).Type()
		switch t.Kind() {
		case reflect.Float64:
			for i := 0; i < length; i++ {
				reflectSlice.Index(i).SetFloat(math.Float64frombits(d.readUintBinary(8)))
			}
			return
		case reflect.Uint8:
			bytes := make([]byte, length)
			for i := 0; i < length; i++ {
				bytes[i] = d.readByte()
			}
			reflect.Copy(reflectSlice, reflect.ValueOf(bytes))
		default:
			tp := reflectSlice.Index(0).Type()
			for i := 0; i < length; i++ {
				r.read(d, tp, reflectSlice.Index(i))
			}
		}
	case reflect.String:
		str := d.readString()
		v.SetString(str)
	case reflect.Bool:
		v.SetBool(d.readByte() == 1)
	case reflect.Int8:
		v.SetInt(int64(d.readByte()))
	case reflect.Int16:
		v.SetInt(d.readIntBinary(2))
	case reflect.Int32:
		v.SetInt(d.readIntBinary(4))
	case reflect.Int64:
		v.SetInt(d.readIntBinary(8))
	case reflect.Int:
		v.SetInt(d.readIntBinary(IntSize))
	case reflect.Uint8:
		v.SetUint(uint64(d.readByte()))
	case reflect.Uint16:
		v.SetUint(d.readUintBinary(2))
	case reflect.Uint32:
		v.SetUint(d.readUintBinary(4))
	case reflect.Uint64:
		v.SetUint(d.readUintBinary(8))
	case reflect.Uint:
		v.SetUint(d.readUintBinary(UintSize))
	case reflect.Uintptr:
		v.SetUint(d.readUintBinary(UintptrSize))
	case reflect.Float32:
		fv := math.Float32frombits(uint32(d.readUintBinary(4)))
		v.SetFloat(float64(fv))
	case reflect.Float64:
		fv := math.Float64frombits(d.readUintBinary(8))
		v.SetFloat(fv)
	case reflect.Map:
		length := int(d.readUintBinary(4))
		reflectMap := reflect.MakeMapWithSize(v.Type(), length)
		for i := 0; i < length; i++ {
			key := reflect.New(v.Type().Key()).Elem()
			value := reflect.New(v.Type().Elem()).Elem()
			r.read(d, key.Type(), key)
			r.read(d, value.Type(), value)
			reflectMap.SetMapIndex(key, value)
		}

		v.Set(reflectMap)
	default:
		log.Panic("Decoding unhandled Kind " + v.Kind().String())
	}
}

type dec struct {
	buf          []byte
	indexCurrent int
}

func Decoder(data []byte) *dec {
	return &dec{buf: data, indexCurrent: 0}
}

func (d *dec) readByte() byte {
	return d.take(1)[0]
}

func (d *dec) readString() string {
	l := int(d.readUintBinary(4))
	return string(d.take(l))
}

func (d *dec) take(step int) (res []byte) {
	res = d.buf[d.indexCurrent:(d.indexCurrent + step)]
	d.indexCurrent += step
	return
}

func (d *dec) readIntBinary(s int) int64 {
	b := d.take(s)
	return bytesIntBinary(b, s)
}

func (d *dec) readUintBinary(s int) uint64 {
	data := d.take(s)
	return BytesUintBinary(data, s)
}

func BytesUintBinary(b []byte, size int) (res uint64) {
	if size == 2 {
		res = uint64(binary.BigEndian.Uint16(b))
	} else if size == 4 {
		res = uint64(binary.BigEndian.Uint32(b))
	} else if size == 8 {
		res = binary.BigEndian.Uint64(b)
	}
	return
}

func bytesIntBinary(b []byte, size int) (res int64) {
	if size == 2 {
		res = int64(int16(binary.BigEndian.Uint16(b)))
	} else if size == 4 {
		res = int64(int32(binary.BigEndian.Uint32(b)))
	} else if size == 8 {
		res = int64(binary.BigEndian.Uint64(b))
	}
	return
}
