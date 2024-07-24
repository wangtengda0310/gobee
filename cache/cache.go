package cache

import "github.com/wangtengda0310/gobee/cache/internal"

type Cache internal.Cache

var impl = internal.BigCacheImpl{}

func Set(key string, value interface{}) {
	impl.Set(key, value)
}

func Get(key string) interface{} {
	get, err := impl.Get(key)
	if err != nil {
		return nil
	}
	return get
}
