package main

type Task interface {
	GetID() int64
	SetID(int64)
}

type task struct {
	id int64
}

func (task *task) GetID() int64 {
	return task.id
}

func (task *task) SetID(id int64) {
	task.id = id
}
