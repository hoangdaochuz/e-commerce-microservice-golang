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

type Redis struct {
	Address string `mapstructure:"address"`
	Port    string `mapstructure:"port"`
	MaxIdle int    `mapstructure:"max_idle"`
}

type ZitadelConfigs struct {
	ClientId              string `mapstructure:"client_id"`
	RedirectURI           string `mapstructure:"redirect_uri"`
	AuthorizationEndpoint string `mapstructure:"authorization_endpoint"`
	EndSessionEndpoint    string `mapstructure:"endsession_endpoint"`
	IntrospectionEndpoint string `mapstructure:"introspection_endpoint"`
	TokenEndpoint         string `mapstructure:"token_endpoint"`
	UserInfoEndpoint      string `mapstructure:"userinfo_endpoint"`
	ApiKeyBase64          string `mapstructure:"api_key_base64"`
	AuthDomain            string `mapstructure:"auth_domain"`
	EncryptKey            string `mapstructure:"encrypt_key"`
	SessionExpiredSeconds string `mapstructure:"session_expired_seconds"`
}

type AuthToken struct {
	RsaKeyPairFilePath   string `mapstructure:"rsa_key_pair_file_path"`
	RsaPublicKeyFilePath string `mapstructure:"rsa_public_key_file_path"`
}

type GeneralConfig struct {
	BackendEndpoint       string `mapstructure:"backend_endpoint"`
	FrontendUserEndpoint  string `mapstructure:"frontend_user_endpoint"`
	FrontendAdminEndpoint string `mapstructure:"frontent_admin_endpoint"`
}

type CircuitBreakerCommon struct {
	MaxRequest           int     `mapstructure:"max_requests"`
	Interval             int     `mapstructure:"interval"`
	Timeout              int     `mapstructure:"timeout"`
	FailureThreshold     int     `mapstructure:"failure_threshold"`
	FailureRateThreshold float64 `mapstructure:"failure_rate_threshold"`
	MinRequests          int     `mapstructure:"min_requests"`
}

type CircuitBreakerNats struct {
	Services map[string]CircuitBreakerCommon `mapstructure:"services"`
}

type CircuitBreakerDatabases struct {
	CircuitBreakerCommon
	SeparateReadWrite bool `mapstructure:"separate_read_write"`
	FallbackToMemory  bool `mapstructure:"fallback_to_memory"`
}

type CircuitBreakerExternalAPIs struct {
	Zitadel CircuitBreakerCommon `mapstructure:"zitadel"`
}

type CircuitBreaker struct {
	Defaults     CircuitBreakerCommon       `mapstructure:"defaults"`
	Databases    CircuitBreakerDatabases    `mapstructure:"databases"`
	Nats         CircuitBreakerNats         `mapstructure:"nats"`
	ExternalAPIs CircuitBreakerExternalAPIs `mapstructure:"external_apis"`
}

type Config struct {
	ServiceRegistry ServiceRegistryConfig `mapstructure:"service_registry"`
	Apigateway      ApigatewayConfig      `mapstructure:"apigateway"`
	OrderDatabase   OrderDatabase         `mapstructure:"order_database"`
	NatsAuth        NATSAuth              `mapstructure:"nats_auth"`
	Redis           Redis                 `mapstructure:"redis"`
	ZitadelConfigs  ZitadelConfigs        `mapstructure:"zitadel_configs"`
	AuthToken       AuthToken             `mapstructure:"auth_token"`
	GeneralConfig   GeneralConfig         `mapstructure:"general_config"`
	CircuitBreaker  CircuitBreaker        `mapstructure:"circuit_breaker"`
	// Database --> Later
	// Log --> Later
}

type ApigatewayConfig struct {
	Port string `mapstructure:"port"`
}

func setDefaults() {
	viper.SetDefault("service_registry.request_timeout", 30*time.Second)
	viper.SetDefault("nats_auth.nats_url", "nats://localhost:4222")
	viper.SetDefault("nats_auth.xkey_private", "x_private_key")
	viper.SetDefault("nats_auth.issuer_private", "issuer_private")

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

	viper.SetDefault("zitadel_configs.client_id", "XXXXXXXXXXXX")
	viper.SetDefault("zitadel_configs.redirect_uri", "XXXXXXXXXXXX")
	viper.SetDefault("zitadel_configs.authorization_endpoint", "https://e-commerce-golang-project-icglms.us1.zitadel.cloud/oauth/v2/authorize")
	viper.SetDefault("zitadel_configs.endsession_endpoint", "https://e-commerce-golang-project-icglms.us1.zitadel.cloud/oidc/v1/end_session")
	viper.SetDefault("zitadel_configs.introspection_endpoint", "https://e-commerce-golang-project-icglms.us1.zitadel.cloud/oauth/v2/introspect")
	viper.SetDefault("zitadel_configs.token_endpoint", "https://e-commerce-golang-project-icglms.us1.zitadel.cloud/oauth/v2/token")
	viper.SetDefault("zitadel_configs.userinfo_endpoint", "https://e-commerce-golang-project-icglms.us1.zitadel.cloud/oidc/v1/userinfo")
	viper.SetDefault("zitadel_configs.auth_domain", "https://e-commerce-golang-project-icglms.us1.zitadel.cloud")
	viper.SetDefault("zitadel_configs.api_key_base64", "YOUR_API_KEY_BASE64")
	viper.SetDefault("zitadel_configs.encrypt_key", "YOUR_ENCRYPT_KEY")
	viper.SetDefault("zitadel_configs.cookie_name", "ecommerce-cookie")
	viper.SetDefault("zitadel_configs.session_expired_seconds", 604800)

	viper.SetDefault("redis.address", "localhost")
	viper.SetDefault("redis.port", "6379")
	viper.SetDefault("redis.max_idle", 9)

	viper.SetDefault("auth_token.rsa_key_pair_file_path", "apps/auth/resources/rsa-key-pair.pem")
	viper.SetDefault("auth_token.rsa_public_key_file_path", "apps/auth/resources/rsa-public.pem")
	viper.SetDefault("general_config.backend_endpoint", "http://localhost:8080")
	viper.SetDefault("general_config.frontend_user_endpoint", "http://localhost:3000")
	viper.SetDefault("general_config.frontend_admin_endpoint", "http://localhost:3001")

	// Circuit Breaker
	viper.SetDefault("circuit_breaker.defaults.max_requests", 3)
	viper.SetDefault("circuit_breaker.defaults.interval", 60)
	viper.SetDefault("circuit_breaker.defaults.timeout", 120)
	viper.SetDefault("circuit_breaker.defaults.failure_threshold", 5)
	viper.SetDefault("circuit_breaker.defaults.failure_rate_threshold", 0.6)
	viper.SetDefault("circuit_breaker.defaults.min_requests", 10)

	viper.SetDefault("circuit_breaker.nats.services.auth.max_requests", 3)
	viper.SetDefault("circuit_breaker.nats.services.auth.interval", 30)
	viper.SetDefault("circuit_breaker.nats.services.auth.timeout", 20)
	viper.SetDefault("circuit_breaker.nats.services.auth.failure_threshold", 3)
	viper.SetDefault("circuit_breaker.nats.services.auth.failure_rate_threshold", 0.4)
	viper.SetDefault("circuit_breaker.nats.services.auth.min_requests", 10)

	viper.SetDefault("circuit_breaker.nats.services.order.max_requests", 3)
	viper.SetDefault("circuit_breaker.nats.services.order.interval", 30)
	viper.SetDefault("circuit_breaker.nats.services.order.timeout", 20)
	viper.SetDefault("circuit_breaker.nats.services.order.failure_threshold", 3)
	viper.SetDefault("circuit_breaker.nats.services.order.failure_rate_threshold", 0.4)
	viper.SetDefault("circuit_breaker.nats.services.order.min_requests", 10)

	viper.SetDefault("circuit_breaker.databases.max_requests", 3)
	viper.SetDefault("circuit_breaker.databases.interval", 30)
	viper.SetDefault("circuit_breaker.databases.timeout", 20)
	viper.SetDefault("circuit_breaker.databases.failure_threshold", 5)
	viper.SetDefault("circuit_breaker.databases.failure_rate_threshold", 0.4)
	viper.SetDefault("circuit_breaker.databases.min_requests", 10)
	viper.SetDefault("circuit_breaker.databases.separate_read_write", true)
	viper.SetDefault("circuit_breaker.databases.fallback_to_memory", true)

	viper.SetDefault("circuit_breaker.external_apis.zitadel.max_requests", 3)
	viper.SetDefault("circuit_breaker.external_apis.zitadel.interval", 30)
	viper.SetDefault("circuit_breaker.external_apis.zitadel.timeout", 20)
	viper.SetDefault("circuit_breaker.external_apis.zitadel.failure_threshold", 5)
	viper.SetDefault("circuit_breaker.external_apis.zitadel.failure_rate_threshold", 0.4)
	viper.SetDefault("circuit_breaker.external_apis.zitadel.min_requests", 10)

}

func init() {
	setDefaults()
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

func LoadNatsCircuitBreakerConfigByServiceName(serviceName string) *CircuitBreakerCommon {
	result := &CircuitBreakerCommon{
		MaxRequest:           viper.GetInt(fmt.Sprintf("circuit_breaker.nats.services.%s.max_requests", serviceName)),
		Interval:             viper.GetInt(fmt.Sprintf("circuit_breaker.nats.services.%s.interval", serviceName)),
		Timeout:              viper.GetInt(fmt.Sprintf("circuit_breaker.nats.services.%s.timeout", serviceName)),
		FailureThreshold:     viper.GetInt(fmt.Sprintf("circuit_breaker.nats.services.%s.failure_threshold", serviceName)),
		FailureRateThreshold: viper.GetFloat64(fmt.Sprintf("circuit_breaker.nats.services.%s.failure_rate_threshold", serviceName)),
		MinRequests:          viper.GetInt(fmt.Sprintf("circuit_breaker.nats.services.%s.min_requests", serviceName)),
	}
	return result
}
