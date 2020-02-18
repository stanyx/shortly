package cache

import (
	"sync"
	"time"
)

type MemoryCache struct {
	sync.WaitGroup
	c      sync.Map
	stopCh chan struct{}
}

type CacheItem struct {
	ExpireTime time.Time
	Value      interface{}
}

func NewMemoryCache() *MemoryCache {

	stopCh := make(chan struct{}, 1)
	ch := &MemoryCache{c: sync.Map{}, stopCh: stopCh}
	ch.Add(1)
	// expired records clean loop
	go func() {
	loop:
		for {
			select {
			case <-stopCh:
				break loop
			default:
				ch.Range(func(key, value interface{}) bool {
					if v, ok := value.(CacheItem); ok {
						if v.ExpireTime.Unix() > 0 && v.ExpireTime.Before(time.Now()) {
							ch.Delete(key)
						}
					}
					return true
				})
			}
			time.Sleep(time.Second)
		}
		ch.Done()
	}()

	return ch
}

func (ch *MemoryCache) Close() error {
	close(ch.stopCh)
	ch.Wait()
	return nil
}

func (ch *MemoryCache) Load(key interface{}) (interface{}, bool) {
	i, ok := ch.c.Load(key)
	if !ok {
		return nil, false
	}
	return i.(CacheItem), true
}

func (ch *MemoryCache) Store(key interface{}, value interface{}) {
	ch.c.Store(key, CacheItem{Value: value})
}

func (ch *MemoryCache) StoreExp(key interface{}, value interface{}, ttl int) {
	ch.c.Store(key, CacheItem{
		ExpireTime: time.Now().Add(time.Second * time.Duration(ttl)),
		Value:      value,
	})
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
