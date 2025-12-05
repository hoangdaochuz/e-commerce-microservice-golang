package auth_service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/hoangdaochuz/ecommerce-microservice-golang/apps/auth/api/auth"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/apps/auth/handler/claims"
	auth_session "github.com/hoangdaochuz/ecommerce-microservice-golang/apps/auth/handler/session"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/configs"
	cache_pkg "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/cache"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/circuitbreaker"
	custom_nats "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/custom-nats"
	di "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/dependency-injection"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/httpclient"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/logging"
	zitadel_pkg "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/zitadel"
	zitadel_authentication "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/zitadel/authentication"
	zitadel_authorization "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/zitadel/authorization"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/shared"
	"github.com/spf13/viper"
)

type AuthService struct {
	// ctx context.Context
	// zitadelAuth       *zitadel_authentication.Auth[claims.Claim]
	zitadelAuthBreaker *zitadel_authentication.AuthBreaker[claims.Claim]
	zitadelAuthorizer  zitadel_authorization.Authorizer
	// cookieHandler     zitadel_authentication.CookieHandler
}

var _ = di.Make[*AuthService](NewAuthService)

const (
	AUTH_DOMAIN_KEY             = "zitadel_configs.auth_domain"
	API_KEY_BASE64_KEY          = "zitadel_configs.api_key_base64"
	ENCRYPT_KEY                 = "zitadel_configs.encrypt_key"
	USER_INFO_ENDPOINT_KEY      = "zitadel_configs.userinfo_endpoint"
	TOKEN_ENDPOINT_KEY          = "zitadel_configs.token_endpoint"
	INTROSPECTION_ENDPOINT_KEY  = "zitadel_configs.introspection_endpoint"
	AUTHORIZE_ENDPOINT_KEY      = "zitadel_configs.authorization_endpoint"
	REDIRECT_URI_KEY            = "zitadel_configs.redirect_uri"
	CLIENT_ID                   = "zitadel_configs.client_id"
	END_SESSION_ENDPOINT_KEY    = "zitadel_configs.endsession_endpoint"
	COOKIE_NAME_KEY             = "zitadel_configs.cookie_name"
	FRONTEND_USER_ENDPOINT_KEY  = "general_config.frontend_user_endpoint"
	FRONTEND_ADMIN_ENDPOINT_KEY = "general_config.frontend_admin_endpoint"
	SESSION_EXPIRED_SECONDS     = "zitadel_configs.session_expired_seconds"
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
	sessionExpiredSeconds, err := strconv.Atoi(viper.GetString(SESSION_EXPIRED_SECONDS))
	if err != nil {
		return nil, err
	}
	authConfig := zitadel_authentication.Config{
		AuthDomain:            authDomain,
		EncryptKey:            viper.GetString(ENCRYPT_KEY),
		ZitadelClientID:       viper.GetString(CLIENT_ID),
		ExpiredInSeconds:      sessionExpiredSeconds,
		EndSessionEndpoint:    viper.GetString(END_SESSION_ENDPOINT_KEY),
		TokenEndpoint:         viper.GetString(TOKEN_ENDPOINT_KEY),
		AuthorizationEndpoint: viper.GetString(AUTHORIZE_ENDPOINT_KEY),
		RedirectURI:           viper.GetString(REDIRECT_URI_KEY),
		PostLoginSuccessURI:   viper.GetString(FRONTEND_USER_ENDPOINT_KEY),
		UserInfoEndpoint:      viper.GetString(USER_INFO_ENDPOINT_KEY),
	}
	// sessionHandler := auth_session.NewRedisSession[claims.Claim]()

	err = auth_session.MakeRedisSession[claims.Claim]()
	if err != nil {
		return nil, err
	}

	err = auth_session.MakeRedisSession[string]()
	if err != nil {
		return nil, err
	}

	var sessionHandler *auth_session.RedisSession[claims.Claim]
	err = di.Resolve(func(redisSession *auth_session.RedisSession[claims.Claim]) {
		sessionHandler = redisSession
	})
	if err != nil {
		return nil, err
	}

	var codeVerifierStore *auth_session.RedisSession[string]
	err = di.Resolve(func(verifierStore *auth_session.RedisSession[string]) {
		codeVerifierStore = verifierStore
	})
	if err != nil {
		return nil, err
	}

	oidcDiscovery := zitadel_authentication.NewOIDCDiscoveryImpl()

	zitadelAuth, err := zitadel_authentication.NewAuth(context.TODO(), sessionHandler, authConfig, viper.GetString(COOKIE_NAME_KEY), oidcDiscovery, codeVerifierStore)
	if err != nil {
		logging.GetSugaredLogger().Errorf("fail to create auth service: %v", err)
		return nil, fmt.Errorf("fail to create auth service: %w", err)
	}

	var redisCache *cache_pkg.RedisCache
	di.Resolve(func(cache *cache_pkg.RedisCache) {
		redisCache = cache
	})
	authorizer, err := zitadel_authorization.NewZitadelAuthorizer(context.TODO(), authDomain, zitadelKeyBase64, redisCache)
	if err != nil {
		logging.GetSugaredLogger().Errorf("fail to create authorizer: %v", err)
		return nil, fmt.Errorf("fail to create authorizer: %w", err)
	}

	zitadelCircuitBreakerConfig := configs.LoadExternalApiCircuitBreakerConfigByApiProviderName("zitadel")
	zitadelAuthBreakerRegistry := circuitbreaker.GetRegistry[*httpclient.HTTPResponse]()
	zitadelHTTPBreaker, err := zitadelAuthBreakerRegistry.GetOrCreateBreaker(shared.ZITADEL_CIRCUIT_BREAKER, circuitbreaker.ToCircuitBreakerConfig(shared.ZITADEL_CIRCUIT_BREAKER, zitadelCircuitBreakerConfig))
	if err != nil {
		logging.GetSugaredLogger().Errorf("fail to get or create zitadel auth breaker: %v", err)
		return nil, fmt.Errorf("fail to get or create zitadel auth breaker: %w", err)
	}
	authBreakerConfig := &zitadel_authentication.AuthBreakerConfig{
		HttpClientConfig: httpclient.DefaultConfig(),
		Breaker:          zitadelHTTPBreaker,
		IsEnableCache:    false,
		// CacheTTL: ,
	}

	return &AuthService{
		// ctx:               ctx,
		zitadelAuthBreaker: zitadel_authentication.NewAuthBreaker(authBreakerConfig, *zitadelAuth, nil),
		zitadelAuthorizer:  authorizer,
	}, nil
}

func (srv *AuthService) Login(ctx context.Context, req *auth.LoginRequest) (*auth.RedirectResponse, error) {
	reqCtx := ctx.Value(shared.HTTPRequest_ContextKey)
	r := reqCtx.(*http.Request)
	w := custom_nats.NewResponseBuilderWithHeader(nil)

	loginURL, err := srv.zitadelAuthBreaker.AuthCodeUrl(r, w, viper.GetString(FRONTEND_USER_ENDPOINT_KEY), func() []zitadel_authentication.ScopeOps {
		var scopes []zitadel_authentication.ScopeOps
		scopes = append(scopes, getDefaultScopes()...)
		scopes = append(scopes, getProjectRoleScopes("admin")...)
		return scopes
	}, []zitadel_authentication.LoginOps{
		zitadel_authentication.WithLoginHint(req.Username),
	})
	if err != nil {
		logging.GetSugaredLogger().Errorf("fail to get login url: %v", err)
		return nil, fmt.Errorf("fail to get login url: %w", err)
	}
	return &auth.RedirectResponse{
		IsSuccess:   true,
		RedirectURL: loginURL,
	}, nil
}

type BodyCallback struct {
	Code        string
	State       string
	Error       string
	Description string
}

func (srv *AuthService) Callback(ctx context.Context, req *auth.CallbackRequest) (*custom_nats.Response, error) {
	reqCtx := ctx.Value(shared.HTTPRequest_ContextKey)
	postReq := reqCtx.(*http.Request)
	// Hack code
	// Convert to get request for oauth flow
	body, err := postReq.GetBody()
	if err != nil {
		return nil, err
	}
	urlObject := postReq.URL
	path := urlObject.Path
	host := urlObject.Host
	headers := postReq.Header

	var bodyObj BodyCallback
	err = json.NewDecoder(body).Decode(&bodyObj)
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequest("GET", host+path, nil)
	if err != nil {
		return nil, err
	}

	q := r.URL.Query()
	q.Add("code", bodyObj.Code)
	q.Add("state", bodyObj.State)
	r.URL.RawQuery = q.Encode()
	r.Header = headers

	w := custom_nats.NewResponseBuilderWithHeader(nil)
	err = srv.zitadelAuthBreaker.Callback(r, w, "/", func(zitadelClaim zitadel_pkg.ZitadelClaim, token *zitadel_authentication.Token, sessionId string) (*claims.Claim, error) {
		converter := claims.NewClaimConverter()
		claims, err := converter.ConvertToInternalClaims(&zitadelClaim, claims.ClaimConverterRequest{
			SessionId: sessionId,
			TokenId:   token.IdToken,
			Token:     token.Token.AccessToken,
		})
		if err != nil {
			return nil, err
		}
		return claims, nil
	}, srv.zitadelAuthBreaker)
	if err != nil {
		return nil, err
	}

	return w.Build(), nil
}

func (srv *AuthService) ValidateToken() (bool, error) {
	return false, nil
}

func (srv *AuthService) GetMyProfile(ctx context.Context, req *auth.EmptyRequest) (*auth.GetMyProfileResponse, error) {
	rCtx := ctx.Value(shared.HTTPRequest_ContextKey)
	r := rCtx.(*http.Request)
	httpCookieHandler := zitadel_authentication.NewHttpCookie(r, nil)

	claims, err := srv.zitadelAuthBreaker.GetCurrentUser(ctx, httpCookieHandler)
	if err != nil {
		return nil, fmt.Errorf("fail to get current user: %w", err)
	}
	return &auth.GetMyProfileResponse{
		Username:  claims.Username,
		Email:     claims.Email,
		FirstName: claims.GivenName,
		LastName:  claims.FamilyName,
		Gender:    claims.Gender,
	}, nil
}

func (srv *AuthService) Logout(ctx context.Context, req *auth.EmptyRequest) (*auth.RedirectResponse, error) {
	w := custom_nats.NewResponseBuilderWithHeader(nil)
	rCtx := ctx.Value(shared.HTTPRequest_ContextKey)
	r := rCtx.(*http.Request)

	httpCookieHandler := zitadel_authentication.NewHttpCookie(r, w)

	claims, err := srv.zitadelAuthBreaker.GetCurrentUser(ctx, httpCookieHandler)
	if err != nil {
		return nil, err
	}
	if claims == nil {
		return nil, fmt.Errorf("claims is empty")
	}
	idToken := claims.IdToken
	endSessionUrl, err := srv.zitadelAuthBreaker.LogoutUrl(httpCookieHandler, "/", idToken, srv.zitadelAuthBreaker.Config.PostLoginSuccessURI)
	if err != nil {
		return nil, err
	}
	if endSessionUrl == "" {
		return nil, fmt.Errorf("end session url is empty")
	}

	return &auth.RedirectResponse{
		IsSuccess:   true,
		RedirectURL: endSessionUrl,
	}, nil
}
