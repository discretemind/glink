package stream

import (
	"errors"
	"reflect"
)

func (s *DataStream) FilterByField(fieldName string, fieldValue interface{}) *DataStream {
	return Stream(s, func(event *Event) (*Event, error) {
		v := reflect.ValueOf(event.Payload)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		if v.Kind() == reflect.Ptr {
			return nil, errors.New("Invalid type")
		}
		field := v.FieldByName(fieldName)
		if field.Interface() == fieldValue {
			return event, nil
		} else {
			return nil, nil
		}
	})
}

func (s *DataStream) Filter(f func(value interface{}) bool) *DataStream {
	return Stream(s, func(event *Event) (*Event, error) {
		if f(event.Payload) {
			return event, nil
		}
		return nil, nil
	})
}
