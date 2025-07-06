package configs

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type ServiceRegistryConfig struct {
	RequestTimeout time.Duration `mapstructure:"request_timeout"`
	NATSUrl        string        `mapstructure:"nats_url"`
	NATSUser       string        `mapstructure:"nats_user"`
	NATSPassword   string        `mapstructure:"nats_password"`
}

type Config struct {
	ServiceRegistry ServiceRegistryConfig `mapstructure:"service_registry"`
	Apigateway      ApigatewayConfig      `mapstructure:"apigateway"`
	// Database --> Later
	// Log --> Later
}

type ApigatewayConfig struct {
	Port string `mapstructure:"port"`
}

func setDefaults() {
	viper.SetDefault("service_registry.request_timeout", 30*time.Second)
	viper.SetDefault("service_registry.nats_url", "nats://localhost:4222")
	viper.SetDefault("service_registry.nats_user", "nats_user")
	viper.SetDefault("service_registry.nats_password", "nats_pass")
	viper.SetDefault("apigateway.port", "8080")

	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Debug: Show current working directory
	pwd, _ := os.Getwd()
	fmt.Printf("Current working directory: %s\n", pwd)

	// Add multiple possible config paths
	viper.AddConfigPath(".")             // Current directory
	viper.AddConfigPath("./configs")     // configs subdirectory
	viper.AddConfigPath("../configs")    // configs in parent directory
	viper.AddConfigPath("../../configs") // configs in grandparent directory
	viper.AddConfigPath("/configs")      // absolute path (if needed)
	viper.AddConfigPath("../../")        // root
	setDefaults()

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		fmt.Println("❌ No config file found, using defaults and environment variables")
		fmt.Printf("   Searched in: %s\n", pwd)
	} else {
		fmt.Printf("✅ Using config file: %s\n", viper.ConfigFileUsed())
	}
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return &config, nil
}
