package stream

import "fmt"

type FaultStream struct {
	DataStream
}

func (s *DataStream) Fault() *FaultStream {
	result := &FaultStream{
		DataStream: DataStream{
			ctx: s.Context(),
		},
	}
	s.BindOut(func(event *Event) {
		fmt.Println("FaultOut ", event)
		for _, out := range result.outs {
			out(event)
		}
	})
	s.Name("Fault")
	return result
}
