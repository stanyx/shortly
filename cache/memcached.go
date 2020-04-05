package cache

import (
	"fmt"
	"log"

	"github.com/memcachier/mc"
)

type MemcacheCredentials struct {
	Username string
	Password string
}

type MemcachedCache struct {
	c      *mc.Client
	logger *log.Logger
}

func NewMemcachedCache(serverList string, logger *log.Logger, credentials MemcacheCredentials) (*MemcachedCache, error) {

	mc := mc.NewMC(serverList, credentials.Username, credentials.Password)

	return &MemcachedCache{
		c:      mc,
		logger: logger,
	}, nil
}

func (ch *MemcachedCache) Close() error {
	ch.c.Quit()
	return nil
}

func (ch *MemcachedCache) Load(key interface{}) (interface{}, bool) {
	val, _, _, err := ch.c.Get(key.(string))
	if err != nil {
		ch.logger.Printf("cache get(key=%v) error, cause: %+v\n", key, err)
		return nil, false
	}
	return val, true
}

func (ch *MemcachedCache) Store(key interface{}, value interface{}) {
	_, err := ch.c.Set(key.(string), value.(string), 0, 0, 0)
	if err != nil {
		ch.logger.Printf("cache set(key=%v) error, cause: %+v\n", key, err)
	}
}

func (ch *MemcachedCache) StoreExp(key interface{}, value interface{}, ttl int) {
	_, err := ch.c.Set(key.(string), value.(string), 0, uint32(ttl), 0)
	if err != nil {
		ch.logger.Printf("cache set(key=%v) error, cause: %+v\n", key, err)
	}
}

func (ch *MemcachedCache) Delete(key interface{}) {
	err := ch.c.Del(key.(string))
	if err != nil {
		ch.logger.Printf("cache delete(key=%s) error, cause: %+v\n", key, err)
	}
}

func (ch *MemcachedCache) Range(f func(key interface{}, value interface{}) bool) {
	// TODO - not implemented
	panic("memcached - not implemented range method")
}

func (ch *MemcachedCache) Ping() error {
	stats, err := ch.c.Stats()
	fmt.Printf("cache stats: %+v, err: %v\n", stats, err)
	return err
}
