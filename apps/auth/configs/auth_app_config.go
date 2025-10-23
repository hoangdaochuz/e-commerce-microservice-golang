package auth_config

import (
	redis_config "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/configs/redis"
	custom_nats "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/custom-nats"
)

type AuthAppConfig struct {
	NatsConfigs   custom_nats.NatsConfig
	ZitadelConfig *ZitadelConfig
	RedisConfig   *redis_config.RedisConfig
}
