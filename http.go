package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/kovetskiy/lorg"
	"github.com/seletskiy/hierr"
)

type HTTPHandler struct {
	address string
	logger  *lorg.Log
	queue   *TaskQueue
}

func NewHTTPHandler(
	logger *lorg.Log,
	queue *TaskQueue,
) (*HTTPHandler, error) {
	handler := &HTTPHandler{
		logger: logger,
		queue:  queue,
	}

	return handler, nil
}

func (handler *HTTPHandler) Handle(address string) error {
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return hierr.Errorf(
			err, "can't resolve %s", address,
		)
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return hierr.Errorf(
			err,
			"can't listen %s", addr,
		)
	}

	handler.logger.Infof("listening on %s", address)
	return http.Serve(listener, handler)
}

func (handler *HTTPHandler) ServeHTTP(
	response http.ResponseWriter, request *http.Request,
) {
	//requestURL := strings.TrimPrefix(request.URL.Path, "/")

	//cloneURL, branch, err := handler.getCloneURLAndBranch(requestURL)
	//if err != nil {
	//logger.Warning(err)
	//http.Error(response, err.Error(), http.StatusBadRequest)
	//return
	//}

	//logger.Debugf("cloning repository '%s'", cloneURL)

	//gopathDirectory, repositoryDirectory, err := cloneRepository(cloneURL)
	//if err != nil {
	//logger.Infof("can't clone repository '%s': %s", cloneURL, err)
	//http.Error(response, err.Error(), http.StatusInternalHTTPHandlerError)
	//return
	//}

	//logger.Debugf(
	//"checkout repository '%s' to '%s'",
	//repositoryDirectory,
	//branch,
	//)

	//err = checkoutBranch(repositoryDirectory, branch)
	//if err != nil {
	//logger.Infof(
	//"can't checkout %s (%s) to '%s': %s",
	//repositoryDirectory, cloneURL, branch, err,
	//)
	//http.Error(response, err.Error(), http.StatusInternalHTTPHandlerError)
	//return
	//}

	//logger.Debugf("running go get for %s", repositoryDirectory)

	//err = goget(gopathDirectory, repositoryDirectory)
	//if err != nil {
	//logger.Warningf(
	//"go get for %s with gopath=%s (%s) for branch %s failed: %s ",
	//gopathDirectory, repositoryDirectory, cloneURL, branch, err,
	//)
	//http.Error(response, err.Error(), http.StatusInternalHTTPHandlerError)
	//return
	//}

	////logger.Infof(
	////"linting %s with args %s",
	////repositoryDirectory, handler.lintArgs,
	////)

	////output := lint(gopathDirectory, repositoryDirectory, handler.lintArgs)

	////_, err = response.Write([]byte(output))
	////if err != nil {
	////logger.Println(err)
	////}

	////logger.Debugf("removing temporary directory %s", gopathDirectory)

	////err = os.RemoveAll(gopathDirectory)
	////if err != nil {
	////logger.Println(err)
	////}
}

//func (handler *HTTPHandler) getCloneURLAndBranch(
//url string,
//) (cloneURL string, branch string, err error) {
//matches := reStashURL.FindStringSubmatch(url)
//if len(matches) > 0 {
//var (
//project     = matches[4]
//repository  = matches[5]
//pullRequest = matches[6]
//)

//return handler.api.GetPullRequestInfo(
//project, repository, pullRequest,
//)
//}

//return url, "origin/master", nil
//}

func cloneRepository(url string) (string, string, error) {
	gopathDirectory, err := ioutil.TempDir(os.TempDir(), "uroboros_")
	if err != nil {
		return "", "", fmt.Errorf("can't create temp directory: %s", err)
	}

	repositoryDirectory := filepath.Join(
		gopathDirectory, "src", "linterd-target",
	)

	_, err = Execute(
		"git", "clone", url, repositoryDirectory,
	)

	return gopathDirectory, repositoryDirectory, err
}

func checkoutBranch(repository string, branch string) error {
	_, err := ExecuteWithDir(repository, "git", "checkout", branch)

	return err
}

func goget(gopath, repository string) error {
	_, err := ExecuteWithGo(
		repository, gopath,
		"go", "get", "-v",
	)

	return err
}
