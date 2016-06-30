package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kovetskiy/executil"
	"github.com/kovetskiy/stash"
	"github.com/seletskiy/hierr"
)

const (
	StashPullRequestBuildError = `
![build failure](https://img.shields.io/badge/build-failure-red.svg)
$link
` + "```" + `
$buffer
` + "```" + ``

	StashPullRequestBuildSuccess = `
![build passing](https://img.shields.io/badge/build-passing-brightgreen.svg)
$link
` + "```" + `
$buffer
` + "```" + ``
)

type ProcessorStashPullRequest struct {
	processor
	task          *TaskStashPullRequest
	pullRequest   stash.PullRequest
	gopath        string
	sources       string
	buildWithMake bool
	testWithMake  bool
}

func NewProcessorStashPullRequest(
	task *TaskStashPullRequest,
) *ProcessorStashPullRequest {
	return &ProcessorStashPullRequest{task: task}
}

func (processor *ProcessorStashPullRequest) Process() {
	processor.task.SetState(TaskStateProcessing)

	var commentText string

	err := processor.process()
	if err != nil {
		processor.logger.Error(err)
		processor.task.SetState(TaskStateError)
		commentText = StashPullRequestBuildError
	} else {
		processor.logger.Infof(":: build passing")
		processor.task.SetState(TaskStateSuccess)
		commentText = StashPullRequestBuildSuccess
	}

	replacer := strings.NewReplacer(
		`$link`, "http://uroboro.s/task/"+fmt.Sprint(processor.task.GetID()),
		`$buffer`, processor.task.GetBuffer().String(),
	)

	processor.logger.Debugf("creating comment to pull-request")

	_, err = processor.resources.stash.CreateComment(
		processor.task.Project,
		processor.task.Repository,
		processor.task.Identifier,
		replacer.Replace(commentText),
	)
	if err != nil {
		processor.logger.Error(
			hierr.Errorf(
				err,
				"can't create comment in pull-request",
			),
		)
	}
}

func (processor *ProcessorStashPullRequest) process() error {
	defer func() {
		if processor.gopath != "" {
			processor.logger.Debugf("removing directory %s", processor.gopath)

			err := os.RemoveAll(processor.gopath)
			if err != nil {
				processor.logger.Errorf(
					"can't remove directory %s: %s", processor.gopath, err,
				)
			}
		}
	}()

	err := processor.fetch()
	if err != nil {
		return err
	}

	processor.logger.Infof(":: successfully fetched")

	err = processor.lookupMakefileTargets()
	if err != nil {
		return err
	}

	err = processor.build()
	if err != nil {
		return err
	}

	processor.logger.Infof(":: successfully builded")

	err = processor.test()
	if err != nil {
		return err
	}

	processor.logger.Infof(":: successfully tested")

	err = processor.lint()
	if err != nil {
		return err
	}

	return nil
}

func (processor *ProcessorStashPullRequest) lint() error {
	linters := [][]string{
		[]string{
			"govet",
			"go", "tool", "vet", ".",
		},
		[]string{
			"misspell",
			"misspell", ".",
		},
		[]string{
			"ineffassign",
			"ineffassign", ".",
		},
		[]string{
			"gofmt",
			"gofmt", "-s", "-l", ".",
		},
		[]string{
			"gocyclo",
			"gocyclo", "-over", "10", ".",
		},
	}

	for _, linter := range linters {
		processor.logger.Infof(
			":: checking source code using %s",
			linter[0],
		)

		_, err := processor.spawn(linter[1], linter[2:]...)
		if err != nil {
			if executil.IsExitError(err) {
				output := strings.Split(
					string(err.(*executil.Error).Output),
					"\n",
				)
				for _, line := range output {
					processor.logger.Error(line)
				}

				return fmt.Errorf(
					"%s exited with non-zero exit code",
					linter[0],
				)
			}

			return hierr.Errorf(
				err,
				"can't lint project",
			)
		}
	}

	return nil
}

func (processor *ProcessorStashPullRequest) build() error {
	var stderr string
	var err error
	if processor.buildWithMake {
		processor.logger.Infof(":: building project using make build")

		stderr, err = processor.makeBuild()
	} else {
		processor.logger.Infof(":: building project using go build")

		stderr, err = processor.gobuild()
	}

	if err != nil {
		if executil.IsExitError(err) {
			output := strings.Split(stderr, "\n")
			for _, line := range output {
				processor.logger.Error(line)
			}

			return errors.New("build failed")
		}

		return hierr.Errorf(
			err,
			"can't build project",
		)
	}

	return nil
}

func (processor *ProcessorStashPullRequest) test() error {
	var stderr string
	var err error
	if processor.testWithMake {
		processor.logger.Infof(":: testing project using make test")

		stderr, err = processor.makeTest()
	} else {
		processor.logger.Infof(":: testing project using go test")

		stderr, err = processor.gotest()
	}

	if err != nil {
		if executil.IsExitError(err) {
			output := strings.Split(stderr, "\n")
			for _, line := range output {
				processor.logger.Error(line)
			}

			return errors.New("tests failed")
		}

		return hierr.Errorf(
			err,
			"can't test project",
		)
	}

	return nil
}

func (processor *ProcessorStashPullRequest) lookupMakefileTargets() error {
	contents, err := ioutil.ReadFile(
		filepath.Join(processor.sources, "Makefile"),
	)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return hierr.Errorf(
			err,
			"can't read Makefile",
		)
	}

	for _, line := range strings.Split(string(contents), "\n") {
		if strings.HasPrefix(line, "build:") {
			processor.buildWithMake = true
		}

		if strings.HasPrefix(line, "test:") {
			processor.testWithMake = true
		}

		if processor.buildWithMake && processor.testWithMake {
			break
		}
	}

	return nil
}

func (processor *ProcessorStashPullRequest) fetch() error {
	processor.logger.Infof(
		":: retrieving information about pull-request",
	)

	var err error
	processor.pullRequest, err = processor.resources.stash.GetPullRequest(
		processor.task.Project,
		processor.task.Repository,
		processor.task.Identifier,
	)
	if err != nil {
		return hierr.Errorf(
			err,
			"can't obtain information about specified pull-request",
		)
	}

	var branch = processor.pullRequest.FromRef.DisplayID

	processor.logger.Infof(
		":: retrieving information about repository",
	)

	cloneURL, err := processor.getCloneURL()
	if err != nil {
		return hierr.Errorf(
			err,
			"can't obtain repository clone URL",
		)
	}

	processor.logger.Infof(
		":: cloning repository %s", cloneURL,
	)

	gopath, sources, err := processor.clone(cloneURL)
	if err != nil {
		return hierr.Errorf(
			err,
			"can't clone repository %s", cloneURL,
		)
	}

	processor.gopath = gopath
	processor.sources = sources

	processor.logger.Infof(
		":: switching to branch %s", branch,
	)

	err = processor.checkout(branch)
	if err != nil {
		return hierr.Errorf(
			err,
			"can't checkout repository branch to %s",
			branch,
		)
	}

	processor.logger.Infof(
		":: fetching project's dependencies",
	)

	stderr, err := processor.goget()
	if err != nil {
		if executil.IsExitError(err) {
			output := strings.Split(stderr, "\n")
			for _, line := range output {
				processor.logger.Error(line)
			}

			return errors.New("go get failed")
		}

		return hierr.Errorf(
			err,
			"can't fetch project's dependencies",
		)
	}

	return nil
}

func (processor *ProcessorStashPullRequest) getCloneURL() (string, error) {
	url := cache.Get(
		processor.task.Host,
		processor.task.Project,
		processor.task.Repository,
	)
	if url != "" {
		return url, nil
	}

	repository, err := processor.resources.stash.GetRepository(
		processor.task.Project,
		processor.task.Repository,
	)
	if err != nil {
		return "", hierr.Errorf(
			err,
			"can't obtain information about specified repository",
		)
	}

	cache.Set(
		repository.SshUrl(),
		processor.task.Host,
		processor.task.Project,
		processor.task.Repository,
	)

	return repository.SshUrl(), nil
}

func (processor *ProcessorStashPullRequest) clone(
	url string,
) (string, string, error) {
	gopath, err := ioutil.TempDir(os.TempDir(), "uroboros_gopath_")
	if err != nil {
		return "", "", hierr.Errorf(
			err, "can't create temporary directory",
		)
	}

	sources := filepath.Join(
		gopath, "src",
		processor.task.Host, processor.task.Project, processor.task.Repository,
	)

	_, err = processor.spawn("git", "clone", url, sources)

	return gopath, sources, err
}

func (processor *ProcessorStashPullRequest) checkout(branch string) error {
	_, err := processor.spawn("git", "checkout", branch)
	return err
}

func (processor *ProcessorStashPullRequest) goget() (string, error) {
	return processor.spawn("go", "get", "-v", "-d")
}

func (processor *ProcessorStashPullRequest) gobuild() (string, error) {
	return processor.spawn("go", "build", "-gcflags", "-e")
}

func (processor *ProcessorStashPullRequest) gotest() (string, error) {
	return processor.spawn("go", "test", "-gcflags", "-e")
}

func (processor *ProcessorStashPullRequest) makeBuild() (string, error) {
	return processor.spawn("make", "build")
}

func (processor *ProcessorStashPullRequest) makeTest() (string, error) {
	return processor.spawn("make", "test")
}

func (processor *ProcessorStashPullRequest) spawn(
	name string, arg ...string,
) (string, error) {
	cmd := exec.Command(name, arg...)

	if processor.sources != "" {
		cmd.Dir = processor.sources
	}

	if processor.gopath != "" {
		cmd.Env = append(
			[]string{
				"GOPATH=" + processor.gopath,
			},
			os.Environ()...,
		)
	}

	_, stderr, err := processor.processor.execute(cmd)
	return string(stderr), err
}
