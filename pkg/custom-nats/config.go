package custom_nats

import "github.com/spf13/viper"

const (
	NatsURLKey                = "nats_auth.nats_url"
	NatsAppAccUserNameKey     = "nats_auth.nats_apps.0.username"
	NatsAppAccountPasswordKey = "nats_auth.nats_apps.0.password"
)

func init() {
	viper.SetDefault(NatsURLKey, "nats://localhost:4222")
	viper.SetDefault(NatsAppAccUserNameKey, "app")
	viper.SetDefault(NatsAppAccountPasswordKey, "app")
}

type NatsConfig struct {
	NatsUrl                string
	NatsAppAccUserName     string
	NatsAppAccountPassword string
}

func GetNatsConfig() NatsConfig {
	return NatsConfig{
		NatsUrl:                viper.GetString(NatsURLKey),
		NatsAppAccUserName:     viper.GetString(NatsAppAccUserNameKey),
		NatsAppAccountPassword: viper.GetString(NatsAppAccountPasswordKey),
	}
}
