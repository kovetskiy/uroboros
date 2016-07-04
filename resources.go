package main

import (
	"io/ioutil"
	"log"
	"net/url"

	"github.com/kovetskiy/ko"
	"github.com/kovetskiy/stash"
	"github.com/seletskiy/hierr"
)

type config struct {
	Web struct {
		Listen string `required:"true"`
		BasicURL string `toml:"basic_url" required:"true"`
	} `toml:"web" required:"true"`

	Tasks struct {
		Threads int `required:"true"`
	} `required:"true"`

	Resources struct {
		Stash struct {
			Address  string `required:"true"`
			Username string `required:"true"`
			Password string `required:"true"`
		} `required:"true"`
		Linters map[string]string `required:"true"`
	} `required:"true"`
}

type resources struct {
	config  *config
	stash   stash.Stash
	queue   *Queue
	linters map[string]string
}

func GetResources(path string) (*resources, error) {
	var config config
	if err := ko.Load(path, &config); err != nil {
		return nil, err
	}

	stash.Log = log.New(ioutil.Discard, "", 0)

	stashURL, err := url.Parse(config.Resources.Stash.Address)
	if err != nil {
		return nil, hierr.Errorf(
			err,
			"can't parse Stash address",
		)
	}

	return &resources{
		stash: stash.NewClient(
			config.Resources.Stash.Username,
			config.Resources.Stash.Password,
			stashURL,
		),
		queue:   NewQueue(getLogger("queue")),
		linters: config.Resources.Linters,
		config:  &config,
	}, nil
}
