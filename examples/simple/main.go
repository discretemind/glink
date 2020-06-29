package main

import (
	"fmt"
	"github.com/discretemind/glink"
	"github.com/discretemind/glink/stream"
	"time"
)

type packet struct {
	Type string
	Num  int
}

func main() {
	fmt.Println("Start")
	done := make(chan bool, 1)
	job := glink.Standalone()
	job.Task("simple", func(input stream.IInputStream) {

		go func() {
			i := 0
			for i < 10 {
				fmt.Println("Task ", i)
				if i%2 == 0 {
					input.Push(packet{
						Type: "type X",
						Num:  i,
					})
				} else {
					input.Push(packet{
						Type: "type Y",
						Num:  i,
					})
				}
				time.Sleep(1 * time.Second)
				i++
			}
			close(done)
		}()
	}, func(flow *stream.DataStream) {
		flow.Map(func(value interface{}) (interface{}, error) {
			p := value.(packet)
			p.Num *= 2
			return p, nil
		}).Name("All items").Print()

		flow.FilterByField("Type", "type X").Map(func(value interface{}) (interface{}, error) {
			p := value.(packet)
			p.Num *= 2
			return p, nil
		}).Name("Filtered X").Print()

		flow.Filter(func(value interface{}) bool {
			return value.(packet).Type == "type Y"
		}).Map(func(value interface{}) (interface{}, error) {
			p := value.(packet)
			p.Num *= 10
			return p, nil
		}).Name("Filtered Y").Print()

	})

	go job.Run()

	fmt.Println("Wait")
	<-done
}
