package cache

type UrlCache interface {
	Load(key interface{}) (interface{}, bool)
	Store(key interface{}, value interface{})
	Delete(key interface{})
	Range(func(key interface{}, value interface{}) bool)
}