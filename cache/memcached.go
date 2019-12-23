package cache

import (
	"github.com/bradfitz/gomemcache/memcache"
)

type MemcachedCache struct {
	c *memcache.Client
}

func NewMemcachedCache() *MemcachedCache {
	return &MemcachedCache{c: memcache.New("10.0.0.1:11211", "10.0.0.2:11211", "10.0.0.3:11212")}
}

func (ch *MemcachedCache) Load(key interface{}) (interface{}, bool) {
	url, err := ch.c.Get(key.(string))
	if err != nil {
		return nil, false
	}
	return string(url.Value), true
}

func (ch *MemcachedCache) Store(key interface{}, value interface{}) {
	ch.c.Set(&memcache.Item{Key: key.(string), Value: []byte(value.(string))})
}

func (ch *MemcachedCache) Delete(key interface{}) {
	ch.c.Delete(key.(string))
}

func (ch *MemcachedCache) Range(f func(key interface{}, value interface{}) bool) {
	// TODO - not implemented
	panic("memcached - not implemented range method")
}