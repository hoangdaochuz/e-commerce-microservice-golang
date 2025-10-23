package cache_pkg

import (
	"context"
	"fmt"
	"time"

	di "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/dependency-injection"
	redis_pkg "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/redis"
	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	redis *redis.Client
}

func NewRedisCache(redisPkg *redis_pkg.Redis) *RedisCache {
	client := redisPkg.GetClient()
	return &RedisCache{
		redis: client,
	}
}

var _ = di.Make(NewRedisCache)

func (c *RedisCache) Get(ctx context.Context, key string) (any, error) {
	resultCmd := c.redis.Get(ctx, key)
	if resultCmd.Err() != nil {
		return nil, resultCmd.Err()
	}
	value, err := resultCmd.Result()
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (c *RedisCache) Set(ctx context.Context, key string, value any) error {
	statusCmd := c.redis.Set(ctx, key, value, 0)
	return statusCmd.Err()
}

func (c *RedisCache) SetEx(ctx context.Context, key string, value any, seconds int) error {
	statusCmd := c.redis.Set(ctx, key, value, time.Second*time.Duration(seconds))
	return statusCmd.Err()
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	resultCmd := c.redis.Del(ctx, key)
	return resultCmd.Err()
}

func (c *RedisCache) IsExist(ctx context.Context, key string) (bool, error) {
	resultCmd := c.redis.Get(ctx, key)
	_, err := resultCmd.Result()
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *RedisCache) GetOrSet(ctx context.Context, key string, callback func() (any, error)) (any, error) {
	resultCmd := c.redis.Get(ctx, key)
	value, err := resultCmd.Result()
	if err != nil || value == "" {
		value, err := callback()
		if err != nil {
			return nil, err
		}
		statusCmd := c.redis.Set(ctx, key, value, 0)
		if statusCmd.Err() != nil {
			return value, fmt.Errorf("fail to set value to redis")
		}
		return value, nil

	}
	return value, nil
}

func (c *RedisCache) GetOrSetWithEx(ctx context.Context, key string, callback func() (any, error), seconds int) (any, error) {
	resultCmd := c.redis.Get(ctx, key)
	value, err := resultCmd.Result()
	if err != nil || value == "" {
		value, err := callback()
		if err != nil {
			return nil, err
		}
		statusCmd := c.redis.Set(ctx, key, value, time.Second*time.Duration(seconds))
		if statusCmd.Err() != nil {
			return value, fmt.Errorf("fail to set value to redis")
		}
		return value, nil

	}
	return value, nil
}
