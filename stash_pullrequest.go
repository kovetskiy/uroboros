package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/kovetskiy/executil"
	"github.com/kovetskiy/stash"
	"github.com/seletskiy/hierr"
	"github.com/seletskiy/tplutil"
)

type ProcessorStashPullRequest struct {
	processor

	task        *TaskStashPullRequest
	pullRequest stash.PullRequest
	gopath      string
	sources     string
	makefile    struct {
		build bool
		test  bool
	}
}

func NewProcessorStashPullRequest(
	task *TaskStashPullRequest,
) *ProcessorStashPullRequest {
	return &ProcessorStashPullRequest{task: task}
}

func (processor *ProcessorStashPullRequest) Process() {
	processor.task.SetState(TaskStateProcessing)

	processor.logger.Infof(
		":: retrieving information about pull request",
	)

	var err error
	processor.pullRequest, err = processor.resources.stash.GetPullRequest(
		processor.task.Project,
		processor.task.Repository,
		processor.task.Identifier,
	)
	if err != nil {
		processor.logger.Error(hierr.Errorf(
			err,
			"can't obtain information about specified pull request",
		))
		processor.task.SetState(TaskStateError)
		return
	}

	err = processor.process()
	if err != nil {
		processor.logger.Error(err)
		processor.task.SetState(TaskStateError)
		processor.comment(TemplateCommentBuildFailure)
		return
	}

	processor.logger.Infof(":: build passing")
	processor.task.SetState(TaskStateSuccess)
	processor.comment(TemplateCommentBuildPassing)
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

	err := processor.ensureBadge()
	if err != nil {
		return err
	}

	err = processor.fetch()
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

func (processor *ProcessorStashPullRequest) ensureBadge() error {
	badge, err := tplutil.ExecuteToString(
		TemplateBadge,
		map[string]interface{}{
			"basic_url": processor.resources.config.Web.BasicURL,
			"slug":      processor.task.GetIdentifier(),
		},
	)
	if err != nil {
		return err
	}

	if strings.Contains(processor.pullRequest.Description, badge) {
		processor.logger.Debugf(
			"no need to edit pull request, badge already added",
		)
		return nil
	}

	processor.logger.Debugf(
		"updating pull request, adding badge to description",
	)

	description := badge + "\n" + processor.pullRequest.Description

	reviewers := []string{}
	for _, reviewer := range processor.pullRequest.Reviewers {
		reviewers = append(reviewers, reviewer.User.Name)
	}

	_, err = processor.resources.stash.UpdatePullRequest(
		processor.task.Project,
		processor.task.Repository,
		processor.task.Identifier,
		processor.pullRequest.Version,
		processor.pullRequest.Title, description, "", reviewers,
	)

	if err != nil {
		return hierr.Errorf(
			err,
			"can't add badge to pull request description",
		)
	}

	return nil
}

func (processor *ProcessorStashPullRequest) comment(
	template *template.Template,
) {
	text, err := tplutil.ExecuteToString(template, map[string]interface{}{
		"id":        processor.task.GetUniqueID(),
		"logs":      processor.task.GetBuffer().String(),
		"errors":    processor.task.GetErrorBuffer().String(),
		"basic_url": processor.resources.config.Web.BasicURL,
	})
	if err != nil {
		processor.logger.Error(err)
		return
	}

	processor.logger.Debugf("creating comment to pull request")

	comment, err := processor.resources.stash.CreateComment(
		processor.task.Project,
		processor.task.Repository,
		processor.task.Identifier,
		text,
	)
	if err != nil {
		processor.logger.Error(
			hierr.Errorf(
				err,
				"can't create comment in pull request",
			),
		)
		return
	}

	processor.logger.Debugf("comment #%v created", comment.ID)
}

func (processor *ProcessorStashPullRequest) lint() error {
	for linter, cmd := range processor.resources.linters {
		processor.logger.Infof(
			":: lintering source code using %s",
			linter,
		)

		_, err := processor.spawn("sh", "-c", cmd)
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
					"linter %s exited with non-zero exit code",
					linter,
				)
			}

			return hierr.Errorf(
				err,
				"an error occurred while lintering source code",
			)
		}
	}

	return nil
}

func (processor *ProcessorStashPullRequest) build() error {
	var stderr string
	var err error
	if processor.makefile.build {
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

			if processor.makefile.build {
				return errors.New("make build exited with non-zero exit code")
			} else {
				return errors.New("go build exited with non-zero exit code")
			}
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
	if processor.makefile.test {
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

			if processor.makefile.test {
				return errors.New("make test exited with non-zero exit code")
			} else {
				return errors.New("go test exited with non-zero exit code")
			}
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
			processor.makefile.build = true
		}

		if strings.HasPrefix(line, "test:") {
			processor.makefile.test = true
		}

		if processor.makefile.build && processor.makefile.test {
			break
		}
	}

	return nil
}

func (processor *ProcessorStashPullRequest) fetch() error {
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

	err = processor.prepareSources(cloneURL, branch)
	if err != nil {
		return hierr.Errorf(
			err,
			"can't clone repository %s", cloneURL,
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

			return errors.New("go get exited with non-zero exit code")
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

func (processor *ProcessorStashPullRequest) prepareSources(
	url, branch string,
) error {
	gopath, err := ioutil.TempDir(os.TempDir(), "uroboros_")
	if err != nil {
		return hierr.Errorf(
			err, "can't create temporary directory",
		)
	}

	sources := filepath.Join(
		gopath, "src",
		processor.task.Host, processor.task.Project, processor.task.Repository,
	)

	_, err = processor.spawn("git", "clone", url, sources)
	if err != nil {
		return err
	}

	processor.gopath = gopath
	processor.sources = sources

	processor.logger.Infof(
		":: switching to branch %s", branch,
	)

	_, err = processor.spawn("git", "checkout", branch)
	if err != nil {
		return err
	}

	return nil
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
