package stream

import (
	"fmt"
	"golang.org/x/crypto/blake2b"
	"reflect"
)

type KeyedEvent struct {
	Key   interface{}
	Event Event
}

type KeyedStream struct {
	Sync chan chan KeyedEvent
}

//
//
//func (s *DataStream) KeyBy(f func (interface{}) bool) *DataStream {
//	return Stream(s, func(event *Event) (*Event, error) {
//
//		if f(event.Payload){
//			return e
//		}
//
//		v := reflect.ValueOf(event.Payload)
//
//
//		field := v.FieldByName(fieldName)
//		if field.Interface() == fieldValue {
//			return event, nil
//		} else {
//			return nil, nil
//		}
//	})
//}

func (s *DataStream) KeyByFieldSync(fields ...string) *DataStream {
	return Stream(s, func(event *Event) (*Event, error) {
		v := reflect.ValueOf(event.Payload)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		key := ""
		for _, f := range fields {
			key += fmt.Sprintf("%v;", v.FieldByName(f).Interface())
		}
		h, _ := blake2b.New256([]byte(key))
		e := KeyedEvent{
			Key: h,
		}
		fmt.Println(e)

		//field := v.FieldByName(fieldName)
		//if field.Interface() == fieldValue {
		//	return event, nil
		//} else {
		//	return nil, nil
		//}
		return nil, nil
	}).Name("Key By")
}
