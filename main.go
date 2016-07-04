package main

import (
	"github.com/kovetskiy/godocs"
	"github.com/kovetskiy/lorg"
	"github.com/reconquest/colorgful"
	"github.com/seletskiy/hierr"
)

var (
	version = "1.0"
	usage   = `uroboros ` + version +
		` - the continious integration snake which will gobble u and ur code

@TODO

Usage:
    uroboros [options]

Options:
    -c --config <path>  Specify configuration file.
                         [default: /etc/uroboros/uroboros.conf]
    --debug             Debug mode.
    --trace             Trace mode.
    -h --help           Show this help.
`
)

var (
	globalLogger = lorg.NewLog()
	debugMode    = false
	traceMode    = false
)

func main() {
	args, err := godocs.Parse(usage, version, godocs.UsePager)
	if err != nil {
		fatalln(err)
	}

	globalLogger.SetFormat(
		colorgful.MustApplyDefaultTheme(
			"${time} ${level:[%s]:right:short} ${prefix}%s",
			colorgful.Dark,
		),
	)

	debugMode = args["--debug"].(bool)
	if debugMode {
		globalLogger.SetLevel(lorg.LevelDebug)
	}

	traceMode = args["--trace"].(bool)
	if traceMode {
		globalLogger.SetLevel(lorg.LevelTrace)
	}

	resources, err := GetResources(args["--config"].(string))
	if err != nil {
		hierr.Fatalf(
			err,
			"can't configure uroboros resources",
		)
	}

	var (
		scheduler = NewScheduler(getLogger("scheduler"), resources)
		server    = NewWebServer(getLogger("server"), resources)
	)

	scheduler.Schedule(resources.config.Tasks.Threads)

	if err = server.ListenAndServe(resources.config.Web.Listen); err != nil {
		hierr.Fatalf(
			err,
			"can't serve http connections",
		)
	}
}
