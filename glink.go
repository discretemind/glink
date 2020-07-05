package glink

import (
	"fmt"
	"github.com/discretemind/glink/stream"
	"sync"
	"time"
)

type IManager interface {

	Error(err error)
}

type ITaskSetup interface {
	Task(name string, input func(input stream.IInputStream), watermark ...func(interface{}) time.Time) *stream.DataStream
	Run()
}

type job struct {
	sync.Mutex
	tasks   map[string]func()
	manager IManager
}

func New(manager IManager) ITaskSetup {
	res := &job{
		manager: manager,
		tasks:   make(map[string]func()),
	}
	return res
}

func Cluster(url string, token string) ITaskSetup {
	return New(ClusterManager(url, token))
}

func Standalone() ITaskSetup {
	return New(StandaloneManager())
}

func (j *job) Task(name string, input func(input stream.IInputStream), watermark ...func(interface{}) time.Time) *stream.DataStream {
	j.Lock()
	defer j.Unlock()

	_, ok := j.tasks[name]
	if !ok {
		fmt.Println("Task name ", name)
		inStream := stream.InputStream()

		j.tasks[name] = func() {
			input(inStream)
		}

		var out *stream.DataStream
		if len(watermark) != 0 {
			out = inStream.Watermark(watermark[0])
		} else {
			out = inStream.Watermark(func(meg interface{}) time.Time {
				return time.Now()
			})
		}

		return out
	}
	return nil
}

func (j *job) Run() {
	for _, t := range j.tasks {
		fmt.Println("RUn  task ")
		t()
	}
}
