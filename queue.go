package main

import (
	"sync"
	"sync/atomic"

	"github.com/kovetskiy/lorg"
)

type Queue struct {
	channel chan Task
	queued  int64
	poped   int64
	logger  *lorg.Log
	tasks   []Task
	mutex   *sync.Mutex
}

func NewQueue(logger *lorg.Log) *Queue {
	queue := &Queue{
		channel: make(chan Task),
		logger:  logger,
		mutex:   &sync.Mutex{},
	}

	return queue
}

func (queue *Queue) Push(task Task) int64 {
	uniqueID := atomic.AddInt64(&queue.queued, 1)

	task.SetUniqueID(uniqueID)
	task.SetState(TaskStateQueued)

	queue.mutex.Lock()
	queue.tasks = append(queue.tasks, task)
	queue.mutex.Unlock()

	go func() {
		queue.channel <- task
	}()

	queue.logger.Debugf(
		"[%d/%d] push #%d",
		queue.poped, queue.queued, task.GetUniqueID(),
	)

	return uniqueID
}

func (queue *Queue) Pop() Task {
	task := <-queue.channel
	atomic.AddInt64(&queue.poped, 1)

	queue.logger.Debugf(
		"[%d/%d] pop #%d",
		queue.poped, queue.queued, task.GetUniqueID(),
	)

	return task
}

func (queue *Queue) GetTaskByIdentifier(identifier string) Task {
	for i := len(queue.tasks) - 1; i >= 0; i-- {
		if queue.tasks[i].GetIdentifier() == identifier {
			return queue.tasks[i]
		}
	}

	return nil
}

func (queue *Queue) GetTaskByUniqueID(id int) Task {
	if id <= len(queue.tasks) && id >= 1 {
		return queue.tasks[id-1]
	}

	return nil
}
