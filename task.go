package main

import (
	"bytes"
)

type TaskState int

var (
	TaskStateUnknown    TaskState = 0
	TaskStateQueued     TaskState = 10
	TaskStateProcessing TaskState = 20
	TaskStateError      TaskState = 30
	TaskStateSuccess    TaskState = 40
)


func (state TaskState) String() string {
	switch state {
	case TaskStateQueued:
		return "queued"
	case TaskStateProcessing:
		return "processing"
	case TaskStateError:
		return "error"
	case TaskStateSuccess:
		return "success"
	default:
		return "unknown"
	}
}



type Task interface {
	GetID() int64
	SetID(int64)
	GetState() TaskState
	SetState(TaskState)
	GetBuffer() *bytes.Buffer
	GetErrorBuffer() *bytes.Buffer
	GetTitle() string
}

type task struct {
	id     int64
	state  TaskState
	buffer *bytes.Buffer
	errorBuffer *bytes.Buffer
}

func (task *task) GetID() int64 {
	return task.id
}

func (task *task) SetID(id int64) {
	task.id = id
}

func (task *task) GetState() TaskState {
	return task.state
}

func (task *task) SetState(state TaskState) {
	task.state = state
}

func (task *task) GetBuffer() *bytes.Buffer {
	if task.buffer == nil {
		task.buffer = &bytes.Buffer{}
	}

	return task.buffer
}


func (task *task) GetErrorBuffer() *bytes.Buffer {
	if task.errorBuffer == nil {
		task.errorBuffer = &bytes.Buffer{}
	}

	return task.errorBuffer
}
