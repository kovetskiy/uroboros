package main

import "regexp"

var reStashURL = regexp.MustCompile(
	`(https?://.*/)` +
		`((users|projects)/([^/]+))` +
		`/repos/([^/]+)` +
		`/pull-requests/(\d+)`,
)

type TaskBitbucketPullRequest struct {
	task
	URL        string
	Hostname   string
	Project    string
	Repository string
	Identifier int64
}

func NewTaskBitbucketPullRequest(url string) (*TaskBitbucketPullRequest, error) {
	task := &TaskBitbucketPullRequest{}
}
