package stream

import (
	"time"
)

type IInputStream interface {
	Push(event interface{})
	//Run()
}

type inputStream struct {
	*DataStream
	watermark func(meg interface{}) time.Time
}

func InputStream() (result *inputStream) {
	result = &inputStream{
		DataStream: &DataStream{
			ctx: &Context{},
		},
	}
	return
}

func (s *inputStream) Watermark(f func(meg interface{}) time.Time) *DataStream {
	s.watermark = f
	return s.DataStream
}

func (s *inputStream) Push(msg interface{}) {
	evt := &Event{
		Payload: msg,
	}
	if s.watermark != nil {
		evt.Timestamp = s.watermark(msg)
	} else {
		evt.Timestamp = time.Now()
	}
	//fmt.Println("Push ", evt.Payload)
	for _, out := range s.outs {
		out(evt)
	}
}
