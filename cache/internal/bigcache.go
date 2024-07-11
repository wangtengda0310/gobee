package internal

// https://github.com/allegro/bigcache

type BigCacheImpl struct {
}

func (c *BigCacheImpl) Set(key string, value interface{}) error {
	//cache := bigcache.BigCache{}
	println(key, value)
	//cache.Set(key, []byte(fmt.Sprintf("%v", value)))
	return nil
}

func (c *BigCacheImpl) Get(key string) (interface{}, error) {
	return "nil", nil
}
