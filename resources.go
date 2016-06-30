package main

import (
	"io/ioutil"
	"log"
	"net/url"

	"github.com/kovetskiy/lorg"
	"github.com/kovetskiy/stash"
	"github.com/seletskiy/hierr"
)

type Resources struct {
	stash stash.Stash
	queue *Queue
}

func GetResources(logger *lorg.Log, config *config) (*Resources, error) {
	stash.Log = log.New(ioutil.Discard, "", 0)

	stashURL, err := url.Parse(config.Resources.Stash.Address)
	if err != nil {
		return nil, hierr.Errorf(
			err,
			"can't parse stash address",
		)
	}

	stashClient := stash.NewClient(
		config.Resources.Stash.Username,
		config.Resources.Stash.Password,
		stashURL,
	)

	queue := NewQueue(logger.NewChildWithPrefix("[queue]"))

	resources := &Resources{
		stash: stashClient,
		queue: queue,
	}

	return resources, nil
}
