package auth_session

import (
	"context"
	"fmt"
	"time"

	di "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/dependency-injection"
	redis_pkg "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/redis"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/utils"
)

type RedisSession[T any] struct {
	*redis_pkg.Redis
}

func MakeRedisSession[T any]() error {
	return di.Make[T](NewRedisSession[T])
}

func NewRedisSession[T any](redis *redis_pkg.Redis) *RedisSession[T] {
	return &RedisSession[T]{
		Redis: redis,
	}
}

func (r *RedisSession[T]) Get(ctx context.Context, key string) (*T, error) {
	sessionStr, err := r.Redis.GetClient().Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	return utils.JsonStringToStruct[T](sessionStr)
}

func (r *RedisSession[T]) Set(ctx context.Context, key string, value T, seconds int) error {
	valueStr, err := utils.StructToJsonString(value)
	if err != nil {
		return fmt.Errorf("fail to set redis session: %w", err)
	}
	redisStatus := r.Redis.GetClient().SetEx(ctx, key, valueStr, time.Second*time.Duration(seconds))
	_, err = redisStatus.Result()
	return err
}
func (r *RedisSession[T]) Del(ctx context.Context, key string) error {
	intCmd := r.Redis.GetClient().Del(ctx, key)
	_, err := intCmd.Result()
	return err
}
