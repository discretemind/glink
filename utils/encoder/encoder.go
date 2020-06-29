package encoder

import (
	"reflect"
	"unsafe"
)

var (
	IntSize     = int(unsafe.Sizeof(int(0)))
	UintSize    = int(unsafe.Sizeof(uint(0)))
	UintptrSize = int(unsafe.Sizeof(uintptr(0)))
)

func DecodeRaw(data []byte, obj interface{}) (err error) {
	r := newBitReader()
	switch val := obj.(type) {
	default:
		return r.decode(data, reflect.TypeOf(obj), reflect.ValueOf(obj))
	case reflect.Value:
		return r.decode(data, val.Type(), val)
	}
}

func EncodeRaw(obj interface{}) (result []byte) {
	var root = Encoder()
	st := reflect.TypeOf(obj)
	sv := reflect.ValueOf(obj)

	if st.Kind() == reflect.Ptr {
		sv = sv.Elem()
		st = st.Elem()
	}

	root.writeValue(st, sv)
	return root.data
}
