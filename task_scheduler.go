package main

import (
	"fmt"
	"sync/atomic"

	"github.com/kovetskiy/lorg"
)

type TaskScheduler struct {
	queue     TaskQueue
	threads   int
	scheduled int64
	logger    *lorg.Log
	resources Resources
}

func NewTaskScheduler(
	queue TaskQueue,
	threads int,
) *TaskScheduler {
	return &TaskScheduler{queue: queue}
}

func (scheduler *TaskScheduler) HandleAndServe() {
	for i := 0; i < scheduler.threads; i++ {
		go scheduler.handle()
	}
}

func (scheduler *TaskScheduler) handle() {
	task := scheduler.queue.Pop()
	scheduler.serve(task)
}

func (scheduler *TaskScheduler) serve(task Task) {
	task.SetID(
		atomic.AddInt64(&scheduler.scheduled, 1),
	)
	task.SetLogger(
		scheduler.logger.NewChildWithPrefix(
			fmt.Sprintf("[task-%d]", task.GetID()),
		),
	)

	scheduler.logger.Infof("launching task-%d", task.GetID())

	processor := NewTaskProcessor(task)
	processor.SetResources(scheduler.resources)
	processor.Process()
}
