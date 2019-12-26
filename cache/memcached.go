package cache

import (
	"log"
	"github.com/bradfitz/gomemcache/memcache"
)

type MemcachedCache struct {
	c *memcache.Client
	logger *log.Logger
}

func NewMemcachedCache(serverList []string, logger *log.Logger) *MemcachedCache {
	return &MemcachedCache{
		c:      memcache.New(serverList...),
		logger: logger,
	}
}

func (ch *MemcachedCache) Load(key interface{}) (interface{}, bool) {
	url, err := ch.c.Get(key.(string))
	if err != nil {
		ch.logger.Printf("cache get error, cause: %+v\n", err)
		return nil, false
	}
	return string(url.Value), true
}

func (ch *MemcachedCache) Store(key interface{}, value interface{}) {
	err := ch.c.Set(&memcache.Item{Key: key.(string), Value: []byte(value.(string))})
	if err != nil {
		ch.logger.Printf("cache set(key=%s) error, cause: %+v\n", key, err)
	}
}

func (ch *MemcachedCache) Delete(key interface{}) {
	ch.c.Delete(key.(string))
}

func (ch *MemcachedCache) Range(f func(key interface{}, value interface{}) bool) {
	// TODO - not implemented
	panic("memcached - not implemented range method")
}