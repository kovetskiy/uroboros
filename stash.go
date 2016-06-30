package main

import (
	"fmt"

	"github.com/bndr/gopencils"
)

type StashAPI struct {
	*gopencils.Resource
}

type ResponseStashPullRequest struct {
	FromRef struct {
		// I dunno why stash developers names branch field as `displayId`
		Branch     string `json:"displayId"`
		Repository struct {
			Links struct {
				// there is can be http and ssh clone urls
				Clone []struct {
					Href string
					Name string
				} `json:"clone"`
			} `json:"links"`
		} `json:"repository"`
	} `json:"fromRef"`
}

func getStashAPI(host, user, pass string) (*StashAPI, error) {
	return &StashAPI{
		gopencils.Api(
			"http://"+host+"/rest/api/1.0",
			&gopencils.BasicAuth{user, pass},
		),
	}, nil
}

// GetPullRequestInfo calls stash api and returns info about pull request:
// ssh clone url and full branch name (like refs/heads/dev)
func (api *StashAPI) GetPullRequestInfo(
	project, repository, pullRequest string,
) (string, string, error) {
	request, err := api.Res("projects").Res(project).
		Res("repos").Res(repository).
		Res("pull-requests").Res(pullRequest, &ResponseStashPullRequest{}).
		Get()
	if err != nil {
		return "", "", err
	}

	info := *request.Response.(*ResponseStashPullRequest)

	sshCloneURL := ""
	for _, clone := range info.FromRef.Repository.Links.Clone {
		if clone.Name != "ssh" {
			continue
		}

		sshCloneURL = clone.Href
	}

	if sshCloneURL == "" {
		return "", "", fmt.Errorf(
			"can't get ssh clone url of specified pull request",
		)
	}

	branch := "origin/" + info.FromRef.Branch

	return sshCloneURL, branch, nil
}
