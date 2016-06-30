package main

import (
	"net/http"

	"github.com/kovetskiy/godocs"
)

var (
	version = "1.0"
	usage   = `uroboros ` + version +
		` - the continious integration snake which will gobble ur code

@TODO

Usage:
	uroboros [options]

Options:
    -h --help  Show this help.
	-c <path>  Specify configuration file.
				[default: /etc/uroboros/uroboros.conf]
`
)

var (
	logger = getLogger()
)

func main() {
	args, err := godocs.Parse(usage, version, godocs.UsePager)

	var (
		configPath = args["-c"].(string)
	)

	config, err := getConfig(configPath)
	if err != nil {
		logger.Fatalf("can't load configuration file %s: %s", configPath, err)
	}

	stashAPI, err := getStashAPI(
		config.Stash.Host, config.Stash.User, config.Stash.Password,
	)
	if err != nil {
		logger.Fatalf("can't create stash api resource: %s", err)
	}

	logger.Infof("listening on %s", config.HTTP.ListenAddress)

	err = http.ListenAndServe(config.HTTP.ListenAddress, server)
	if err != nil {
		logger.Fatal(err)
	}
}
