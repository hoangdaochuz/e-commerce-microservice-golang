package redis_config

import (
	di "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/dependency-injection"
	"github.com/spf13/viper"
)

const (
	AddressKey = "redis.address"
	PortKey    = "redis.port"
	MaxIdleKey = "redis.max_idle"
)

type RedisConfig struct {
	Address  string
	Port     string
	Password string
	MaxIdle  int
	Db       *int
}

var _ = di.Make[*RedisConfig](GetRedisConfig)

// func init() {
// 	viper.SetDefault("redis.address", "localhost")
// 	viper.SetDefault("redis.port", "6379")
// 	viper.SetDefault("redis.max_idle", 9)
// }

func GetRedisConfig() *RedisConfig {
	return &RedisConfig{
		Address: viper.GetString(AddressKey),
		Port:    viper.GetString(PortKey),
		MaxIdle: viper.GetInt(MaxIdleKey),
	}
}
