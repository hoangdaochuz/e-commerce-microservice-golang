package redis_pkg

import (
	redis_config "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/configs/redis"
	di "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/dependency-injection"
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

var _ = di.Make(NewRedis)

func NewRedis(redisConfig *redis_config.RedisConfig) *Redis {
	return &Redis{
		client: redis.NewClient(&redis.Options{
			Addr:         redisConfig.Address + ":" + redisConfig.Port,
			MaxIdleConns: redisConfig.MaxIdle,
		}),
	}
}

func (r *Redis) GetClient() *redis.Client {
	return r.client
}
