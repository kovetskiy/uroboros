package main

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kovetskiy/executil"
	"github.com/kovetskiy/stash"
	"github.com/seletskiy/hierr"
)

type ProcessorStashPullRequest struct {
	processor
	task        *TaskStashPullRequest
	pullRequest stash.PullRequest
	gopath      string
	sources     string
	makeBuild   bool
	makeTest    bool
}

func NewProcessorStashPullRequest(
	task *TaskStashPullRequest,
) *ProcessorStashPullRequest {
	return &ProcessorStashPullRequest{task: task}
}

func (processor *ProcessorStashPullRequest) Process() {
	processor.task.SetState(TaskStateProcessing)

	err := processor.process()
	if err != nil {
		processor.logger.Error(err)
		processor.task.SetState(TaskStateError)
		return
	}

	processor.task.SetState(TaskStateSuccess)
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

	err := processor.process()
	if err != nil {
		return err
	}

	err := processor.build()
	if err != nil {
		return err
	}

	err = processor.examineMakefile()
	if err != nil {
		return err
	}

	return nil
}


func (processor *ProcessorStashPullRequest)	 build() error {
	return nil
}


func (processor *ProcessorStashPullRequest) examineMakefile() error {
	return nil
}


func (processor *ProcessorStashPullRequest) fetch() error {
	processor.logger.Infof("retrieving information about pull-request")

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
		"retrieving information about repository",
	)

	cloneURL, err := processor.getCloneURL()
	if err != nil {
		return hierr.Errorf(
			err,
			"can't obtain repository clone URL",
		)
	}

	processor.logger.Infof(
		"cloning repository %s", cloneURL,
	)

	gopath, sources, err := processor.clone(cloneURL)
	if err != nil {
		return hierr.Errorf(
			err,
			"can't clone repository %s", cloneURL,
		)
	}

	processor.logger.Infof(
		"switching to branch %s", branch,
	)

	err = processor.checkout(sources, branch)
	if err != nil {
		return hierr.Errorf(
			err,
			"can't checkout repository branch to %s",
			branch,
		)
	}

	processor.logger.Infof(
		"fetching project and its dependencies",
	)

	err = processor.goget(gopath, sources)
	if err != nil {
		if executil.IsExitError(err) {
			if runErr, ok := err.(*executil.Error); ok {
				output := strings.Split(string(runErr.Output), "\n")
				for _, line := range output {
					processor.logger.Errorf("[go get] %s", line)
				}

				return errors.New("can't fetch project and its dependencies")
			}
		}

		return hierr.Errorf(
			err,
			"can't fetch project and its dependencies",
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

	_, _, err = processor.execute(
		exec.Command("git", "clone", url, sources),
	)

	return gopath, sources, err
}

func (processor *ProcessorStashPullRequest) checkout(
	sources, branch string,
) error {
	cmd := exec.Command("git", "checkout", branch)
	cmd.Dir = sources

	_, _, err := processor.execute(cmd)
	return err
}

func (processor *ProcessorStashPullRequest) goget(
	gopath, sources string,
) error {
	cmd := exec.Command("go", "get", "-v", "-d")
	cmd.Dir = sources
	cmd.Env = append(
		[]string{
			"GOPATH=" + gopath,
		},
		os.Environ()...,
	)

	_, _, err := processor.execute(cmd)
	return err
}

func (processor *ProcessorStashPullRequest) getMakefileTarget() {

}
