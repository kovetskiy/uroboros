package main

import (
	"fmt"
	"regexp"
	"strings"
)

var reStashURL = regexp.MustCompile(
	`(https?://(.*)/)` +
		`((users|projects)/([^/]+))` +
		`/repos/([^/]+)` +
		`/pull-requests/(\d+)`,
)

type TaskStashPullRequest struct {
	task
	URL        string
	BasicURL   string
	Host       string
	Project    string
	Repository string
	Identifier string
}

func NewTaskStashPullRequest(url string) (*TaskStashPullRequest, error) {
	matches := reStashURL.FindStringSubmatch(url)
	if len(matches) == 0 {
		return nil, fmt.Errorf("URL doesn't seem like Stash Pull Request")
	}

	task := &TaskStashPullRequest{
		URL:        url,
		BasicURL:   matches[1],
		Host:       matches[2],
		Project:    strings.ToLower(matches[5]),
		Repository: matches[6],
		Identifier: matches[7],
	}

	task.identifier = fmt.Sprintf(
		"%s/%s/%s/%s",
		task.Host,
		task.Project,
		task.Repository,
		task.Identifier,
	)

	return task, nil
}

func (request *TaskStashPullRequest) GetTitle() string {
	return fmt.Sprintf(
		"[stash pull-request] %s/%s/%s #%s",
		request.Host, request.Project, request.Repository,
		request.Identifier,
	)
}
