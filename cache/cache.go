package cache

type UrlCache interface {
	Load(key interface{}) (interface{}, bool)
	Store(key interface{}, value interface{})
	StoreExp(key, value interface{}, ttl int)
	Delete(key interface{})
	Range(func(key interface{}, value interface{}) bool)
	Ping() error
	Close() error
}
