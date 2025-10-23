package auth_service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hoangdaochuz/ecommerce-microservice-golang/apps/auth/handler/claims"
	auth_session "github.com/hoangdaochuz/ecommerce-microservice-golang/apps/auth/handler/session"
	cache_pkg "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/cache"
	di "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/dependency-injection"
	zitadel_authentication "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/zitadel/authentication"
	zitadel_authorization "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/zitadel/authorization"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/shared"
	"github.com/spf13/viper"
)

type AuthService struct {
	ctx               context.Context
	zitadelAuth       *zitadel_authentication.Auth[claims.Claim]
	zitadelAuthorizer zitadel_authorization.Authorizer
	// cookieHandler     zitadel_authentication.CookieHandler
}

var _ = di.Make(NewAuthService)

const (
	AUTH_DOMAIN_KEY            = "zitadel_configs.auth_domain"
	API_KEY_BASE64_KEY         = "zitadel_configs.api_key_base64"
	ENCRYPT_KEY                = "zitadel_configs.encrypt_key"
	USER_INFO_ENDPOINT_KEY     = "zitadel_configs.userinfo_endpoint"
	TOKEN_ENDPOINT_KEY         = "zitadel_configs.token_endpoint"
	INTROSPECTION_ENDPOINT_KEY = "zitadel_configs.introspection_endpoint"
	AUTHORIZE_ENDPOINT_KEY     = "zitadel_configs.authorization_endpoint"
	REDIRECT_URI_KEY           = "zitadel_configs.redirect_uri"
	CLIENT_ID                  = "zitadel_configs.client_id"
	END_SESSION_ENDPOINT_KEY   = "zitadel_configs.endsession_endpoint"
	COOKIE_NAME                = "ecommerce-cookie"
)

func getDefaultScopes() []zitadel_authentication.ScopeOps {
	return []zitadel_authentication.ScopeOps{
		zitadel_authentication.WithOpenIdScope(),
		zitadel_authentication.WithProfileScope(),
		zitadel_authentication.WithEmailScope(),
		zitadel_authentication.WithAddressScope(),
		zitadel_authentication.WithPhoneScope(),
		zitadel_authentication.WithOfflineScope(),
	}
}

func getProjectRoleScopes(role string) []zitadel_authentication.ScopeOps {
	return []zitadel_authentication.ScopeOps{
		zitadel_authentication.WithProjectRoleScope(role),
	}
}

func NewAuthService() (*AuthService, error) {
	authDomain := viper.GetString(AUTH_DOMAIN_KEY)
	zitadelKeyBase64 := viper.GetString(API_KEY_BASE64_KEY)
	authConfig := zitadel_authentication.Config{
		AuthDomain:            authDomain,
		EncryptKey:            viper.GetString(ENCRYPT_KEY),
		ZitadelClientID:       viper.GetString(CLIENT_ID),
		ExpiredInSeconds:      5 * 60,
		EndSessionEndpoint:    viper.GetString(END_SESSION_ENDPOINT_KEY),
		TokenEndpoint:         viper.GetString(TOKEN_ENDPOINT_KEY),
		AuthorizationEndpoint: viper.GetString(AUTHORIZE_ENDPOINT_KEY),
		RedirectURI:           viper.GetString(REDIRECT_URI_KEY),
		PostLoginSuccessURI:   "/",
		UserInfoEndpoint:      viper.GetString(USER_INFO_ENDPOINT_KEY),
	}
	// sessionHandler := auth_session.NewRedisSession[claims.Claim]()
	var sessionHandler *auth_session.RedisSession[claims.Claim]
	di.Resolve(func(redisSession *auth_session.RedisSession[claims.Claim]) {
		sessionHandler = redisSession
	})

	oidcDiscovery := zitadel_authentication.NewOIDCDiscoveryImpl()

	zitadelAuth, err := zitadel_authentication.NewAuth(context.TODO(), sessionHandler, authConfig, COOKIE_NAME, oidcDiscovery)
	if err != nil {
		return nil, fmt.Errorf("fail to create auth service: %w", err)
	}

	var redisCache *cache_pkg.RedisCache
	di.Resolve(func(cache *cache_pkg.RedisCache) {
		redisCache = cache
	})
	authorizer, err := zitadel_authorization.NewZitadelAuthorizer(context.TODO(), authDomain, zitadelKeyBase64, redisCache)
	if err != nil {
		return nil, fmt.Errorf("fail to create authorizer: %w", err)
	}
	return &AuthService{
		// ctx:               ctx,
		zitadelAuth:       zitadelAuth,
		zitadelAuthorizer: authorizer,
	}, nil
}

func (srv *AuthService) Login(ctx context.Context) error {
	reqCtx := ctx.Value(shared.HTTPRequest_ContextKey)
	resCtx := ctx.Value(shared.HTTPResponse_ContextKey)
	r := reqCtx.(*http.Request)
	fmt.Println("request: ", r)
	w := resCtx.(http.ResponseWriter)
	httpCookieHandler := zitadel_authentication.NewHttpCookie(r, w)

	loginURL, err := srv.zitadelAuth.AuthCodeUrl(httpCookieHandler, "/", func() []zitadel_authentication.ScopeOps {
		var scopes []zitadel_authentication.ScopeOps
		scopes = append(scopes, getDefaultScopes()...)
		scopes = append(scopes, getProjectRoleScopes("admin")...)
		return scopes
	}, []zitadel_authentication.LoginOps{
		zitadel_authentication.WithLoginHint("khai.nguyen"),
	})
	if err != nil {
		return fmt.Errorf("fail to get login url: %w", err)
	}
	fmt.Println("login url: ", loginURL)

	http.Redirect(w, r, loginURL, http.StatusFound)
	return nil
}

func (srv *AuthService) Callback() error {
	return nil
}

func (srv *AuthService) ValidateToken() (bool, error) {
	return false, nil
}
