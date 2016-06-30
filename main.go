package main

import (
	"github.com/kovetskiy/godocs"
	"github.com/kovetskiy/lorg"
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
    -h --help           Show this help.
    -c --config <path>  Specify configuration file.
                         [default: /etc/uroboros/uroboros.conf]
    --debug             Debug mode.
    --trace             Trace mode.
`
)

var (
	coreLogger = getLogger()
	debugMode  = false
	traceMode  = false
)

func main() {
	args, err := godocs.Parse(usage, version, godocs.UsePager)
	if err != nil {
		fatalln(err)
	}

	debugMode = args["--debug"].(bool)
	if debugMode {
		coreLogger.SetLevel(lorg.LevelDebug)
	}

	traceMode = args["--trace"].(bool)
	if traceMode {
		coreLogger.SetLevel(lorg.LevelTrace)
	}

	config, err := getConfig(args["--config"].(string))
	if err != nil {
		hierr.Fatalf(
			err,
			"can't configure uroboros server",
		)
	}

	resources, err := GetResources(coreLogger, config)
	if err != nil {
		hierr.Fatalf(
			err,
			"can't configure uroboros resources",
		)
	}

	var (
		scheduler = NewScheduler(
			coreLogger.NewChildWithPrefix("[scheduler]"), resources,
		)

		handler = NewHTTPHandler(
			coreLogger.NewChildWithPrefix("[handler]"),
			resources,
		)
	)

	scheduler.Schedule(config.Tasks.Threads)

	if err = handler.Listen(config.HTTP.Address); err != nil {
		hierr.Fatalf(
			err,
			"can't serve http connections",
		)
	}
}
