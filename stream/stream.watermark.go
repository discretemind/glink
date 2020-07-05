package stream

import "time"

func (s *DataStream) Watermark(f func(value interface{}) time.Time) *DataStream {
	return Stream(s, func(event *Event) (*Event, error) {
		event.Timestamp = f(event.Payload)
		return event, nil
	})
}
