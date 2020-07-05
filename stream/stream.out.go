package stream

import "fmt"

func (s *DataStream) Print() {
	s.BindOut(func(event *Event) {
		fmt.Printf("Print %s, %+v %v\n", s.name, event.Payload, event.Timestamp)
	})
}
