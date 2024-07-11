package internal

type Cache interface {
	Set(key string, value interface{}) error
	Get(key string) (interface{}, error)
}

var defaultImpl *Cache
