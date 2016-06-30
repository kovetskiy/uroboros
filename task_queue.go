package main

type TaskQueue struct {
	tasks chan Task
}

func NewTaskQueue() *TaskQueue {
	queue := &TaskQueue{
		tasks: make(chan Task),
	}

	return queue
}

func (queue *TaskQueue) Push(task Task) {
	queue.tasks <- task
}

func (queue *TaskQueue) Pop() Task {
	return <-queue.tasks
}
