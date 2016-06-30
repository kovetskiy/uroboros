package main

import (
	"github.com/jinzhu/configor"
)

type config struct {
	HTTP struct {
		ListenAddress string `toml:"listen_address" required:"true" default:":80"`
	} `toml:"http"`

	Stash struct {
		Host     string `toml:"host" required:"true"`
		User     string `toml:"user" required:"true"`
		Password string `toml:"password" required:"true"`
	} `toml:"stash" required:"true"`
}

func getConfig(path string) (*config, error) {
	config := new(config)
	err := configor.Load(&config, path)
	return config, err
}
