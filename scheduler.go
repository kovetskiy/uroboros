package main

import (
	"fmt"
	"os"
	"sync/atomic"

	"github.com/kovetskiy/lorg"
)

type Scheduler struct {
	logger    *lorg.Log
	resources *resources
	scheduled int64
}

func NewScheduler(
	logger *lorg.Log,
	resources *resources,
) *Scheduler {
	return &Scheduler{
		logger:    logger,
		resources: resources,
	}
}

func (scheduler *Scheduler) Schedule(threads int) {
	for i := 1; i <= threads; i++ {
		scheduler.logger.Infof("[%d/%d] spawning thread", i, threads)
		go scheduler.schedule()
	}
}

func (scheduler *Scheduler) schedule() {
	for {
		task := scheduler.resources.queue.Pop()
		scheduler.serve(task)
	}
}

func (scheduler *Scheduler) serve(task Task) {
	atomic.AddInt64(&scheduler.scheduled, 1)

	scheduler.logger.Infof("serving task#%d", task.GetID())
	scheduler.logger.Tracef("%#v", task)

	logger := scheduler.logger.NewChildWithPrefix(
		fmt.Sprintf("[task#%d]", task.GetID()),
	)

	logger.SetOutput(
		lorg.NewOutput(
			os.Stderr,
		).SetLevelWriterCondition(
			lorg.LevelError,
			os.Stderr,
			uncolored{unprefixed{task.GetBuffer()}},
			uncolored{unprefixed{task.GetErrorBuffer()}},
		).SetLevelWriterCondition(
			lorg.LevelWarning,
			os.Stderr,
			uncolored{unprefixed{task.GetBuffer()}},
		).SetLevelWriterCondition(
			lorg.LevelInfo,
			os.Stderr,
			uncolored{unprefixed{task.GetBuffer()}},
		),
	)

	processor := NewProcessor(task)
	processor.SetResources(scheduler.resources)
	processor.SetLogger(logger)
	processor.Process()
}
