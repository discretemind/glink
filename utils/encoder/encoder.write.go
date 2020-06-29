package encoder

import (
	"encoding/binary"
	"log"
	"math"
	"reflect"
)

type bitEncoder struct {
	data []byte
}

func Encoder() (res *bitEncoder) {
	res = &bitEncoder{}
	return
}

func (b *bitEncoder) writeValue(t reflect.Type, v reflect.Value) {
	switch t.Kind() {
	case reflect.Struct:
		b.writeStruct(v)
	case reflect.Array:
		b.writeArray(v, t.Len())
	case reflect.Slice:
		b.writeSlice(v, v.Len())
	case reflect.String:
		b.writeString(v.String())
	case reflect.Bool:
		b.writeBool(v.Bool())
	case reflect.Int8:
		b.writeByte(byte(v.Int()))
	case reflect.Int16:
		b.uintWriteAsBytesBinary(uint64(v.Int()), 2)
	case reflect.Int32:
		b.uintWriteAsBytesBinary(uint64(v.Int()), 4)
	case reflect.Int64:
		b.uintWriteAsBytesBinary(uint64(v.Int()), 8)
	case reflect.Int:
		b.uintWriteAsBytesBinary(uint64(v.Int()), IntSize)
	case reflect.Uint8:
		b.writeByte(byte(v.Uint()))
	case reflect.Uint16:
		b.uintWriteAsBytesBinary(v.Uint(), 2)
	case reflect.Uint32:
		b.uintWriteAsBytesBinary(v.Uint(), 4)
	case reflect.Uint64:
		b.uintWriteAsBytesBinary(v.Uint(), 8)
	case reflect.Uint:
		b.uintWriteAsBytesBinary(v.Uint(), UintSize)
	case reflect.Uintptr:
		b.uintWriteAsBytesBinary(v.Uint(), UintptrSize)
	case reflect.Float32:
		b.uintWriteAsBytesBinary(uint64(math.Float32bits(float32(v.Float()))), 4)
	case reflect.Float64:
		b.uintWriteAsBytesBinary(math.Float64bits(v.Float()), 8)
	case reflect.Map:
		b.writeMap(v)
	default:
		log.Panic("Encoding unhandled Kind " + v.Kind().String())
	}
}

func (b *bitEncoder) writeStruct(v reflect.Value) {
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		v := v.Field(i)
		f := t.Field(i)
		if f.Tag.Get("enc") != "-" {
			if v.CanSet() || f.Name != "_" {
				b.writeValue(f.Type, v)
			}
		}
	}
}

func (b *bitEncoder) writeMap(v reflect.Value) {
	keys := v.MapKeys()
	b.uintWriteAsBytesBinary(uint64(len(keys)), 4)
	var t reflect.Type
	if len(keys) > 0 {
		t = keys[0].Type()
	}
	for _, value := range keys {
		b.writeValue(t, value)
		val := v.MapIndex(value)
		b.writeValue(val.Type(), v.MapIndex(value))
	}
}

func (b *bitEncoder) uintWriteAsBytesBinary(v uint64, size int) {
	switch size {
	case 2:
		uintBytes := [2]byte{}
		binary.BigEndian.PutUint16(uintBytes[:], uint16(v))
		b.write(uintBytes[:])
		return
	case 4:
		uintBytes := [4]byte{}
		binary.BigEndian.PutUint32(uintBytes[:], uint32(v))
		b.write(uintBytes[:])
		return
	case 8:
		uintBytes := [8]byte{}
		binary.BigEndian.PutUint64(uintBytes[:], v)
		b.write(uintBytes[:])
		return
	}

}

func (b *bitEncoder) writeArray(v reflect.Value, size int) {
	if size > 0 {
		switch v.Index(0).Kind() {
		case reflect.Struct:
			for i := 0; i < size; i++ {
				b.writeStruct(v.Index(i))
			}
		case reflect.Array:
			for i := 0; i < size; i++ {
				b.writeArray(v.Index(i), v.Index(i).Len())
			}
		case reflect.Slice:
			for i := 0; i < size; i++ {
				b.writeSlice(v, v.Index(i).Len())
			}
		case reflect.String:
			for i := 0; i < size; i++ {
				b.writeString(v.Index(i).String())
			}
		case reflect.Bool:
			for i := 0; i < size; i++ {
				b.writeBool(v.Index(i).Bool())
			}
		case reflect.Int8:
			bf := make([]byte, size)
			for i := 0; i < size; i++ {
				bf[i] = byte(v.Index(i).Int())
			}
			b.write(bf)
		case reflect.Int16:
			for i := 0; i < size; i++ {
				b.uintWriteAsBytesBinary(uint64(v.Index(i).Int()), 2)
			}
		case reflect.Int32:
			for i := 0; i < size; i++ {
				b.uintWriteAsBytesBinary(uint64(v.Index(i).Int()), 4)
			}
		case reflect.Int64:
			for i := 0; i < size; i++ {
				b.uintWriteAsBytesBinary(uint64(v.Index(i).Int()), 8)
			}
		case reflect.Int:
			for i := 0; i < size; i++ {
				b.uintWriteAsBytesBinary(uint64(v.Index(i).Int()), IntSize)
			}
		case reflect.Uint8:
			bf := make([]byte, size)
			for i := 0; i < size; i++ {
				bf[i] = byte(v.Index(i).Uint())
			}
			b.write(bf)
		case reflect.Uint16:
			for i := 0; i < size; i++ {
				b.uintWriteAsBytesBinary(v.Index(i).Uint(), 2)
			}
		case reflect.Uint32:
			for i := 0; i < size; i++ {
				b.uintWriteAsBytesBinary(v.Index(i).Uint(), 4)
			}
		case reflect.Uint64:
			for i := 0; i < size; i++ {
				b.uintWriteAsBytesBinary(v.Index(i).Uint(), 8)
			}
		case reflect.Uint:
			for i := 0; i < size; i++ {
				b.uintWriteAsBytesBinary(v.Index(i).Uint(), UintSize)
			}
		case reflect.Uintptr:
			for i := 0; i < size; i++ {
				b.uintWriteAsBytesBinary(v.Index(i).Uint(), UintptrSize)
			}
		case reflect.Float32:
			for i := 0; i < size; i++ {
				uintBytes := [4]byte{}
				binary.BigEndian.PutUint32(uintBytes[:], uint32(math.Float32bits(float32(v.Index(i).Float()))))
				b.write(uintBytes[:])
			}
		case reflect.Float64:
			for i := 0; i < size; i++ {
				uintBytes := [8]byte{}
				binary.BigEndian.PutUint64(uintBytes[:], math.Float64bits(v.Index(i).Float()))
				b.write(uintBytes[:])
			}
		case reflect.Map:
			for i := 0; i < size; i++ {
				b.writeMap(v.Index(i))
			}
		default:
			log.Panic("Encoding unhandled Kind " + v.Kind().String())
		}
	}
}

func (b *bitEncoder) writeBool(v bool) {
	if v {
		b.writeByte(byte(1))
	} else {
		b.writeByte(byte(0))
	}
}

func (b *bitEncoder) write(data []byte) {
	b.data = append(b.data, data...)
}

func (b *bitEncoder) writeByte(data byte) {
	b.data = append(b.data, data)
}

func (b *bitEncoder) writeSlice(v reflect.Value, size int) {
	b.uintWriteAsBytesBinary(uint64(size), 4)
	b.writeArray(v, size)
}

func (b *bitEncoder) writeString(line string) {
	b.uintWriteAsBytesBinary(uint64(len(line)), 4)
	b.write([]byte(line))
}
