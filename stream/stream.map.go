package stream

func (s *DataStream) Map(f func(value interface{}) (interface{}, error)) *DataStream {
	return Stream(s, func(event *Event) (*Event, error) {
		mapped, err := f(event.Payload)
		if err != nil {
			return nil, err
		}
		return &Event{
			Timestamp: event.Timestamp,
			Payload:   mapped,
		}, nil
	})
}
