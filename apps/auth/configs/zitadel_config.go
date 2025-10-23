package auth_config

import (
	di "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/dependency-injection"
	"github.com/spf13/viper"
)

const (
	ClientIdKey              = "zitadel_configs.client_id"
	RedirectUriKey           = "zitadel_configs.redirect_uri"
	AuthorizationEndpointKey = "zitadel_configs.authorization_endpoint"
	EndsessionEndpointKey    = "zitadel_configs.endsession_endpoint"
	IntrospectionEndpointKey = "zitadel_configs.introspection_endpoint"
	TokenEndpointKey         = "zitadel_configs.token_endpoint"
	UserInfoEndpointKey      = "zitadel_configs.userinfo_endpoint"
	ApiKeyBase64Key          = "zitadel_configs.api_key_base64"
	AuthDomainKey            = "zitadel_configs.auth_domain"
	ZitadelEncryptKey        = "zitadel_configs.encrypt_key"
)

var _ = di.Make(NewZitadelConfig)

type ZitadelConfig struct {
	ClientId              string
	RedirectUri           string
	AuthorizationEndpoint string
	EndsessionEndpoint    string
	IntrospectionEndpoint string
	TokenEndpoint         string
	UserInfoEndpoint      string
	ApiKeyBase64          string
	AuthDomain            string
	EncryptKey            string
}

func NewZitadelConfig() *ZitadelConfig {
	return &ZitadelConfig{
		ClientId:              viper.GetString(ClientIdKey),
		RedirectUri:           viper.GetString(RedirectUriKey),
		AuthorizationEndpoint: viper.GetString(AuthorizationEndpointKey),
		EndsessionEndpoint:    viper.GetString(EndsessionEndpointKey),
		IntrospectionEndpoint: viper.GetString(IntrospectionEndpointKey),
		TokenEndpoint:         viper.GetString(TokenEndpointKey),
		UserInfoEndpoint:      viper.GetString(UserInfoEndpointKey),
		ApiKeyBase64:          viper.GetString(ApiKeyBase64Key),
		AuthDomain:            viper.GetString(AuthDomainKey),
		EncryptKey:            viper.GetString(ZitadelEncryptKey),
	}
}
