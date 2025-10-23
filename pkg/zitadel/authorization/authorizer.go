package zitadel_authorization

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	cache_pkg "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/cache"
	zitadel_pkg "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/zitadel"
	"github.com/zitadel/oidc/v3/pkg/client"
	"github.com/zitadel/zitadel-go/v3/pkg/authorization"
	"github.com/zitadel/zitadel-go/v3/pkg/authorization/oauth"
	"github.com/zitadel/zitadel-go/v3/pkg/zitadel"
)

type ZitadelAuthorizer struct {
	authorizer *authorization.Authorizer[*oauth.IntrospectionContext]
	cache      cache_pkg.Cache
}

type Authorizer interface {
	Introspect(ctx context.Context, token string) (*zitadel_pkg.ZitadelClaim, error)
}

func NewZitadelAuthorizer(ctx context.Context, authDomain string, zitadelApiKeyBase64 string, cache cache_pkg.Cache) (*ZitadelAuthorizer, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if zitadelApiKeyBase64 == "" {
		return nil, fmt.Errorf("zitadel api key is required")
	}
	introspectDomain := authDomain
	if strings.HasPrefix(introspectDomain, "https://") {
		introspectDomain = strings.ReplaceAll(introspectDomain, "https://", "")
	}

	decodeApiKey, err := base64.StdEncoding.DecodeString(zitadelApiKeyBase64)
	if err != nil {
		return nil, err
	}
	keyFile, err := client.ConfigFromKeyFileData(decodeApiKey)
	if err != nil {
		return nil, err
	}
	zitadelDomainConf := zitadel.New(introspectDomain)
	authZitadel, err := authorization.New(ctx, zitadelDomainConf, oauth.WithIntrospection[*oauth.IntrospectionContext](oauth.JWTProfileIntrospectionAuthentication(keyFile)))
	if err != nil {
		return nil, err
	}
	return &ZitadelAuthorizer{
		authorizer: authZitadel,
		cache:      cache,
	}, nil
}

func (a *ZitadelAuthorizer) Introspect(ctx context.Context, token string) (*zitadel_pkg.ZitadelClaim, error) {
	value, err := a.cache.GetOrSetWithEx(ctx, token, func() (any, error) {
		return a.introspect(ctx, token)
	}, 30)
	if err != nil {
		return nil, err
	}

	return value.(*zitadel_pkg.ZitadelClaim), nil
}

func (a *ZitadelAuthorizer) introspect(ctx context.Context, token string) (*zitadel_pkg.ZitadelClaim, error) {
	if token == "" {
		return nil, fmt.Errorf("token must not be empty")
	}
	accessToken := token
	if !strings.HasPrefix(accessToken, "Bearer ") {
		accessToken = "Bearer " + accessToken
	}

	introspectCtx, err := a.authorizer.CheckAuthorization(ctx, accessToken)
	if err != nil || introspectCtx == nil {
		if errors.Is(err, &authorization.UnauthorizedErr{}) {
			return nil, fmt.Errorf("unauthorize")
		}
		return nil, fmt.Errorf("fail to check authorization: %w", err)
	}
	zitadelClaimByte, err := introspectCtx.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("fail to marshal introspect context: %w", err)
	}
	var zitadelClaim zitadel_pkg.ZitadelClaim
	err = json.Unmarshal(zitadelClaimByte, &zitadelClaim)
	if err != nil {
		return nil, fmt.Errorf("fail to unmarshal zitadel claim byte: %w", err)
	}
	zitadelClaim.Token = token
	return &zitadelClaim, nil
}
