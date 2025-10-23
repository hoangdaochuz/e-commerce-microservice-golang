package cache_pkg

import (
	"context"
	"time"

	"github.com/patrickmn/go-cache"
	"golang.org/x/sync/singleflight"
)

type LocalCache struct {
	cache *cache.Cache
	group singleflight.Group
}

func NewDefaultLocalCache() *LocalCache {
	return &LocalCache{
		cache: cache.New(cache.DefaultExpiration, cache.DefaultExpiration),
	}
}

func NewLocalCacheNoExpiration() *LocalCache {
	return &LocalCache{
		cache: cache.New(cache.NoExpiration, cache.NoExpiration),
	}
}

func NewLocalCacheWithExpiration(defaultExpiration, cleanupInterval time.Duration) *LocalCache {
	return &LocalCache{
		cache: cache.New(defaultExpiration, cleanupInterval),
	}
}

func (c *LocalCache) Get(ctx context.Context, key string) (any, error) {
	value, found := c.cache.Get(key)
	if found {
		return value, nil
	}
	return nil, nil
}

func (c *LocalCache) Set(ctx context.Context, key string, value any) error {
	c.cache.Set(key, value, -1)
	return nil
}

func (c *LocalCache) SetEx(ctx context.Context, key string, value any, seconds int) error {
	c.cache.Set(key, value, time.Second*time.Duration(seconds))
	return nil
}

func (c *LocalCache) Delete(ctx context.Context, key string) error {
	c.cache.Delete(key)
	return nil
}

func (c *LocalCache) IsExist(ctx context.Context, key string) (bool, error) {
	_, found := c.cache.Get(key)
	return found, nil
}

func (c *LocalCache) GetOrSet(ctx context.Context, key string, callback func() (any, error)) (any, error) {
	value, found := c.cache.Get(key)
	if found {
		return value, nil
	}
	var err error
	// var shareBool bool
	value, err, _ = c.group.Do(key, func() (interface{}, error) {
		return callback()
	})
	if err != nil {
		return nil, err
	}
	c.cache.Set(key, value, -1)
	return value, nil
}

func (c *LocalCache) GetOrSetWithEx(ctx context.Context, key string, callback func() (any, error), seconds int) (any, error) {
	value, found := c.cache.Get(key)
	if found {
		return value, nil
	}
	var err error
	value, err, _ = c.group.Do(key, func() (interface{}, error) {
		return callback()
	})
	if err != nil {
		return nil, err
	}
	c.cache.Set(key, value, time.Second*time.Duration(seconds))
	return value, nil
}
