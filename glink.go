package glink

import (
	"fmt"
	"github.com/discretemind/glink/stream"
	"sync"
)

type IManager interface {
	Error(err error)
}

type ITaskSetup interface {
	Task(name string, input func(input stream.IInputStream), out func(*stream.DataStream))
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

func (j *job) Task(name string, input func(input stream.IInputStream), out func(*stream.DataStream)) {
	j.Lock()
	defer j.Unlock()

	_, ok := j.tasks[name]
	if !ok {
		fmt.Println("Task name ", name)
		j.tasks[name] = func() {
			inStream := stream.InputStream()
			input(inStream)
			out(inStream.DataStream)
		}
	}
}

func (j *job) Run() {
	for _, t := range j.tasks {
		fmt.Println("RUn  task ")
		t()
	}
}

//func (j *job) Input()
