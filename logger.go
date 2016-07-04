package main

import (
	"fmt"
	"os"

	"github.com/kovetskiy/lorg"
	"github.com/kovetskiy/spinner-go"
)

func getLogger(format string, arg ...interface{}) *lorg.Log {
	return globalLogger.NewChildWithPrefix(fmt.Sprintf(format, arg...))
}

func fatalf(format string, values ...interface{}) {
	if spinner.IsActive() {
		spinner.Stop()
	}

	fmt.Fprintf(os.Stderr, format+"\n", values...)
	os.Exit(1)
}

func fatalln(value interface{}) {
	fatalf("%s", value)
}

func debugf(format string, values ...interface{}) {
	globalLogger.Debugf(format, values...)
}

func tracef(format string, values ...interface{}) {
	globalLogger.Tracef(format, values...)
}

func debugln(value interface{}) {
	debugf("%s", value)
}

func traceln(value interface{}) {
	tracef("%s", value)
}
