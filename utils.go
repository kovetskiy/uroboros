package main

import (
	"os"
	"os/exec"

	"github.com/kovetskiy/executil"
)

func ExecuteWithDir(
	dir, name string, args ...string,
) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir

	logger.Debugf("exec %q at %s", cmd.Args, dir)

	stdout, _, err := executil.Run(cmd)
	return string(stdout), err
}

func ExecuteWithGo(
	dir, gopath, name string, args ...string,
) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = append(
		[]string{
			"GOPATH=" + gopath,
			"GO15VENDOREXPERIMENT=1",
		},
		os.Environ()...,
	)

	logger.Debugf("exec %q at %s with GOPATH=%s", cmd.Args, dir, gopath)

	stdout, _, err := executil.Run(cmd)
	return string(stdout), err
}

func Execute(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)

	logger.Debugf("exec %q", cmd.Args)

	stdout, _, err := executil.Run(cmd)
	return string(stdout), err
}
