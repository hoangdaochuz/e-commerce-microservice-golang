package configs

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type NATSApp struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Account  string `mapstructure:"account"`
}

type ServiceRegistryConfig struct {
	RequestTimeout time.Duration `mapstructure:"request_timeout"`
}

type OrderDatabase struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBname   string `mapstructure:"dbname"`
}

type NATSAuth struct {
	AuthCallOutSubject string    `mapstructure:"auth_callout_subject"`
	NATSUrl            string    `mapstructure:"nats_url"`
	NATSApps           []NATSApp `mapstructure:"nats_apps"`
	XKeyPrivate        string    `mapstructure:"xkey_private"`
	XKeyPublic         string    `mapstructure:"xkey_public"`
	Issuer             string    `mapstructure:"issuer"`
	IssuerPrivate      string    `mapstructure:"issuer_private"`
}

type Config struct {
	ServiceRegistry ServiceRegistryConfig `mapstructure:"service_registry"`
	Apigateway      ApigatewayConfig      `mapstructure:"apigateway"`
	OrderDatabase   OrderDatabase         `mapstructure:"order_database"`
	NatsAuth        NATSAuth              `mapstructure:"nats_auth"`
	// Database --> Later
	// Log --> Later
}

type ApigatewayConfig struct {
	Port string `mapstructure:"port"`
}

func setDefaults() {
	viper.SetDefault("service_registry.request_timeout", 30*time.Second)
	viper.SetDefault("nats_auth.nats_url", "nats://localhost:4222")

	viper.SetDefault("nats_auth.nats_apps.0.username", "app")
	viper.SetDefault("nats_auth.nats_apps.0.password", "app")
	viper.SetDefault("nats_auth.nats_apps.0.account", "APP")
	viper.SetDefault("nats_auth.nats_apps.1.username", "admin")
	viper.SetDefault("nats_auth.nats_apps.1.password", "admin")
	viper.SetDefault("nats_auth.nats_apps.1.account", "ADMIN")
	viper.SetDefault("nats_auth.nats_apps.2.username", "auth")
	viper.SetDefault("nats_auth.nats_apps.2.password", "auth")
	viper.SetDefault("nats_auth.nats_apps.2.account", "AUTH")

	// viper.SetDefault("service_registry.nats_user", "nats_user")
	// viper.SetDefault("service_registry.nats_password", "nats_pass")
	viper.SetDefault("apigateway.port", "8080")

	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")

	viper.SetDefault("order_database.host", "localhost")
	viper.SetDefault("order_database.port", "5432")
	viper.SetDefault("order_database.user", "postgres")
	viper.SetDefault("order_database.password", "postgres")
	viper.SetDefault("order_database.dbname", "order")
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
	_ = viper.BindEnv("nats_auth.nats_url")
	_ = viper.BindEnv("nats_auth.xkey_private")
	_ = viper.BindEnv("nats_auth.issuer_private")
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
