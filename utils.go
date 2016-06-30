package main

//func ExecuteWithDir(
//dir, name string, args ...string,
//) (string, error) {
//cmd := exec.Command(name, args...)
//cmd.Dir = dir

//coreLogger.Debugf("exec %q at %s", cmd.Args, dir)

//stdout, _, err := executil.Run(cmd)
//return string(stdout), err
//}

//func ExecuteWithGo(
//dir, gopath, name string, args ...string,
//) (string, error) {
//cmd := exec.Command(name, args...)
//cmd.Dir = dir
//cmd.Env = append(
//[]string{
//"GOPATH=" + gopath,
//"GO15VENDOREXPERIMENT=1",
//},
//os.Environ()...,
//)

//coreLogger.Debugf("exec %q at %s with GOPATH=%s", cmd.Args, dir, gopath)

//stdout, _, err := executil.Run(cmd)
//return string(stdout), err
//}

//func Execute(name string, args ...string) (string, error) {
//cmd := exec.Command(name, args...)

//coreLogger.Debugf("exec %q", cmd.Args)

//stdout, _, err := executil.Run(cmd)
//return string(stdout), err
//}
