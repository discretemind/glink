package stream

import "fmt"

func (s *DataStream) Print() {
	s.BindOut(func(event *Event) {
		fmt.Println("Print ", s.name, event.Payload)
	})
}
