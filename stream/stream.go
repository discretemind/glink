package stream

import (
	"time"
)

type Event struct {
	Timestamp time.Time `json:"tm"`
	Payload   interface{}
}

type FilterHandler func(event *Event) (*Event, error)
type PushHandler func(event *Event)

type Context struct {

}

type IStreamSource interface {
	Out(f PushHandler)
	FaultOut(f PushHandler)
}

type DataStream struct {
	ctx    *Context
	name   string
	id     string
	outs   []PushHandler
	faults []PushHandler
}

func Stream(from *DataStream, handler FilterHandler) (result *DataStream) {
	result = &DataStream{
		ctx: from.Context(),
	}
	from.BindOut(func(event *Event) {
		outEvent, err := handler(event)
		if err != nil {
			for _, out := range result.faults {
				out(outEvent)
			}
			return
		}
		if outEvent != nil{
			for _, out := range result.outs {
				out(outEvent)
			}
		}
	})
	return
}

func (s *DataStream) Context() *Context {
	return s.ctx
}

func (s *DataStream) Name(name string) *DataStream {
	s.name = name
	return s
}

func (s *DataStream) ID(id string) *DataStream {
	s.id = id
	return s
}

func (s *DataStream) BindOut(f PushHandler) {
	s.outs = append(s.outs, f)
}

func (s *DataStream) BindFault(f PushHandler) {
	s.faults = append(s.faults, f)
}
