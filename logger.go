package main

import (
	"fmt"
	"os"

	"github.com/kovetskiy/lorg"
	"github.com/kovetskiy/spinner-go"
	"github.com/reconquest/colorgful"
)

func getLogger() *lorg.Log {
	logger := lorg.NewLog()

	logger.SetFormat(
		colorgful.MustApplyDefaultTheme(
			"${time} ${level:[%s]:right:short} ${prefix}%s",
			colorgful.Dark,
		),
	)

	return logger
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
	coreLogger.Debugf(format, values...)
}

func tracef(format string, values ...interface{}) {
	coreLogger.Tracef(format, values...)
}

func debugln(value interface{}) {
	debugf("%s", value)
}

func traceln(value interface{}) {
	tracef("%s", value)
}
