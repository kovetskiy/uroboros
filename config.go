package main

import (
	"os"

	"github.com/jinzhu/configor"
	"github.com/seletskiy/hierr"
)

type config struct {
	HTTP struct {
		Address string `toml:"address" required:"true"`
	} `toml:"http"`

	Tasks struct {
		Threads int `toml:"threads" required:"true"`
	} `toml:"tasks" required:"true"`

	Resources struct {
		Stash struct {
			Address  string `toml:"address" required:"true"`
			Username string `toml:"username" required:"true"`
			Password string `toml:"password" required:"true"`
		} `toml:"stash" required:"true"`
		Linters map[string]string `toml:"linters"`
	} `toml:"resources" required:"true"`
}

func getConfig(path string) (*config, error) {
	config := &config{}

	_, err := os.Stat(path)
	if err != nil {
		return nil, hierr.Errorf(
			err, "can't stat %s", path,
		)
	}

	err = configor.Load(config, path)
	if err != nil {
		return nil, hierr.Errorf(
			err, "problem with configuration data",
		)
	}

	return config, err
}
