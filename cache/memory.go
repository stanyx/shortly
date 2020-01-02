package cache

import (
	"sync"
)

type MemoryCache struct {
	c sync.Map
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{c: sync.Map{}}
}

func (ch *MemoryCache) Load(key interface{}) (interface{}, bool) {
	return ch.c.Load(key) 
}

func (ch *MemoryCache) Store(key interface{}, value interface{}) {
	ch.c.Store(key, value)
}

func (ch *MemoryCache) Delete(key interface{}) {
	ch.c.Delete(key)
}

func (ch *MemoryCache) Range(f func(key interface{}, value interface{}) bool) {
	ch.c.Range(f)
}

func (ch *MemoryCache) Ping() error {
	return nil
}