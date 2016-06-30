package main

import (
	"os/exec"

	"github.com/kovetskiy/executil"
	"github.com/kovetskiy/lorg"
)

type Processor interface {
	SetResources(*Resources)
	SetLogger(*lorg.Log)
	Process()
}

type processor struct {
	resources *Resources
	logger    *lorg.Log
}

func NewProcessor(task Task) Processor {
	switch target := task.(type) {
	case *TaskStashPullRequest:
		return NewProcessorStashPullRequest(target)
	}

	panic("unexpected task")
}

func (processor *processor) SetResources(resources *Resources) {
	processor.resources = resources
}

func (processor *processor) SetLogger(logger *lorg.Log) {
	processor.logger = logger
}

func (processor *processor) execute(
	command *exec.Cmd,
) ([]byte, []byte, error) {
	dir := "./"
	if command.Dir != "" {
		dir = command.Dir
	}

	processor.logger.Debugf("exec %q at %s", command.Args, dir)

	return  executil.Run(command)
}

