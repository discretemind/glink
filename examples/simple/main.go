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

func input(done chan bool) func(input stream.IInputStream) {
	return func(input stream.IInputStream) {
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
	}

}

func main() {
	fmt.Println("Start")
	done := make(chan bool, 1)
	job := glink.Standalone()
	inputSet := job.Task("simple", input(done), func(i interface{}) time.Time {
		return time.Now()
	})
	set2 := inputSet.Map(func(value interface{}) (interface{}, error) {
		p := value.(packet)
		p.Num *= 2
		return p, nil
	}).Name("All items")

	set2.Print()
	set3 := set2.FilterByField("Type", "type X").Map(func(value interface{}) (interface{}, error) {
		p := value.(packet)
		p.Num *= 2
		return p, nil
	}).Name("Filtered X")

	set3.Print()
	set3.Filter(func(value interface{}) bool {
		return value.(packet).Type == "type Y"
	}).Map(func(value interface{}) (interface{}, error) {
		p := value.(packet)
		p.Num *= 10
		return p, nil
	}).Name("Filtered Y").Print()


	go job.Run()

	fmt.Println("Wait")
	<-done
}
