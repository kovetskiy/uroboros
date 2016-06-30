package main

import (
	"sync"

	"github.com/zazab/zhash"
)

type cacheHash struct {
	*sync.Mutex
	hash zhash.Hash
}

var (
	cache = &cacheHash{
		&sync.Mutex{},
		zhash.NewHash(),
	}
)

func (cache *cacheHash) Get(path ...string) string {
	value, _ := cache.hash.GetString(path...)
	return value
}

func (cache *cacheHash) Set(value string, path ...string) {
	cache.hash.Set(value, path...)
}
