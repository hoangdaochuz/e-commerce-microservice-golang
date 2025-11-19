package zitadel_authentication

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hoangdaochuz/ecommerce-microservice-golang/configs"
	cache_pkg "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/cache"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/circuitbreaker"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/httpclient"
	zitadel_pkg "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/zitadel"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/shared"
)

type AuthBreakerConfig struct {
	HttpClientConfig *httpclient.Config
	// BreakerConfig    *circuitbreaker.Config
	Breaker *circuitbreaker.Breaker[*httpclient.HTTPResponse]
	// wheter allow using cached value when error
	IsEnableCache bool

	// cache tts
	CacheTTL time.Duration
}

func DefaultAuthBreakerConfig() *AuthBreakerConfig {
	return &AuthBreakerConfig{
		HttpClientConfig: httpclient.DefaultConfig(),
		Breaker:          circuitbreaker.NewBreaker[*httpclient.HTTPResponse](circuitbreaker.ToCircuitBreakerConfig(shared.ZITADEL_CIRCUIT_BREAKER, configs.LoadDefaultCircuitBreakerConfig())),
		IsEnableCache:    true,
		CacheTTL:         120 * time.Second,
	}
}

type AuthBreaker[T any] struct {
	Auth[T]
	httpClientBreaker *httpclient.BreakerHTTPClient
	isEnableCached    bool
	cacheTTL          time.Duration
	cacheStore        cache_pkg.Cache
}

func NewAuthBreaker[T any](authBreakerConfig *AuthBreakerConfig, auth Auth[T], cache cache_pkg.Cache) *AuthBreaker[T] {
	if authBreakerConfig == nil {
		authBreakerConfig = DefaultAuthBreakerConfig()
	}
	return &AuthBreaker[T]{
		httpClientBreaker: httpclient.NewBreakerHTTPClient(authBreakerConfig.HttpClientConfig, authBreakerConfig.Breaker),
		isEnableCached:    authBreakerConfig.IsEnableCache,
		cacheTTL:          authBreakerConfig.CacheTTL,
		cacheStore:        cache,
		Auth:              auth,
	}
}

func (ab *AuthBreaker[T]) UserInfo(ctx context.Context, token *Token, userInfoConfig string) (*zitadel_pkg.ZitadelClaim, error) {
	accessToken := token.Token.AccessToken
	if accessToken == "" {
		return nil, fmt.Errorf("access token is empty")
	}
	req, err := http.NewRequestWithContext(ctx, "GET", userInfoConfig, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	var res *httpclient.HTTPResponse
	if ab.isEnableCached {
		res, err = ab.httpClientBreaker.DoWithFallback(ctx, req, func() (*httpclient.HTTPResponse, error) {
			// call to get from database / cache

			return nil, nil
		})
	} else {
		res, err = ab.httpClientBreaker.Do(ctx, req)
	}
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received response with status different 200")
	}

	var claims zitadel_pkg.ZitadelClaim
	err = json.Unmarshal(res.Body, &claims)
	if err != nil {
		return nil, err
	}
	claims.IdToken = token.IdToken
	claims.Token = accessToken

	// store cache if enable cache
	// if ab.isEnableCached {
	// 	err := ab.cacheStore.SetEx(ctx, token.IdToken, claims, int(ab.cacheTTL))
	// 	if err != nil {
	// 		return nil, fmt.Errorf("fail to store user claims into cache")
	// 	}
	// }

	return &claims, nil
}
