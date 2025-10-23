package claims

import (
	"context"
	"fmt"
	"log"

	"github.com/dgrijalva/jwt-go"
	zitadel_pkg "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/zitadel"
	zitadel_authentication "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/zitadel/authentication"
	zitadel_authorization "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/zitadel/authorization"
	"github.com/zitadel/oidc/v3/pkg/crypto"
)

type Claim struct {
	UserId         string   `json:"user_id"`          // Internal Id
	ExternalUserId string   `json:"external_user_id"` // Zitadel user id
	Username       string   `json:"username"`
	Email          string   `json:"email"`
	Roles          []string `json:"roles"`
	AccessToken    string   `json:"access_token"`
	RefreshToken   string   `json:"refresh_token"`
	AuthSessionId  string   `json:"session_id"`
	jwt.StandardClaims
}

type HttpSessionClaimGetter struct {
	sessionKey     string
	sessionHandler zitadel_authentication.SessionHandler[Claim]
	cookieHandler  zitadel_authentication.CookieHandler
	decryptionKey  string
	authorizer     zitadel_authorization.Authorizer
	ctx            context.Context
}

func NewHttpSessionClaimGetter(sessionKey string, sessionHanlder zitadel_authentication.SessionHandler[Claim], cookieHandler zitadel_authentication.CookieHandler, decryptionKey string, authorizer zitadel_authorization.Authorizer, ctx context.Context) *HttpSessionClaimGetter {
	return &HttpSessionClaimGetter{
		sessionKey:     sessionKey,
		sessionHandler: sessionHanlder,
		cookieHandler:  cookieHandler,
		decryptionKey:  decryptionKey,
		authorizer:     authorizer,
		ctx:            ctx,
	}
}

func (h *HttpSessionClaimGetter) Get() (*Claim, error) {
	cookieValue, err := h.cookieHandler.GetCookie(h.sessionKey)
	if err != nil {
		return nil, fmt.Errorf("fail to get cookie: %w", err)
	}
	if cookieValue == "" {
		return nil, fmt.Errorf("cookie value is empty")
	}
	sessionId, err := crypto.DecryptAES(cookieValue, h.decryptionKey)
	if err != nil {
		return nil, fmt.Errorf("fail to decrypt cookie value: %w", err)
	}

	sessionInfo, err := h.sessionHandler.Get(h.ctx, sessionId)
	if err != nil {
		return nil, fmt.Errorf("failt to get session info: %w", err)
	}
	if sessionInfo == nil {
		return nil, fmt.Errorf("session info is empty")
	}
	accessToken := sessionInfo.AccessToken
	if accessToken != "" {
		zitadelClaims, err := h.authorizer.Introspect(h.ctx, accessToken)
		if err != nil {
			log.Default().Printf("cannot introspect token: %s", err.Error())
			return sessionInfo, fmt.Errorf("session claim getter: unauthorize")
		}
		if zitadelClaims == nil {
			return sessionInfo, fmt.Errorf("session claim getter: unauthorize")
		}
		return sessionInfo, nil
	}
	return sessionInfo, fmt.Errorf("session claim getter: unauthorize: No access token")
}

type ClaimsConverter struct {
}

// func NewClaimsConverter() *ClaimsConverter {
// 	return &ClaimsConverter{}
// }

type ClaimConverter struct {
}

func NewClaimConverter() *ClaimConverter {
	return &ClaimConverter{}
}

type ClaimConverterRequest struct {
	SessionId string
	TokenId   string
	Token     string
}

func (c *ClaimConverter) ConvertToInternalClaims(zitadelClaims *zitadel_pkg.ZitadelClaim, req ClaimConverterRequest) (*Claim, error) {
	externalUserId := zitadelClaims.Sub
	//TODO: Get User in db by external user id
	// userDetail, err :=

	var roles []string
	for roleKey := range zitadelClaims.UrnZitadelIAMOrgProjectRoles {
		roles = append(roles, roleKey)
	}

	token := req.Token
	if token == "" {
		token = zitadelClaims.Token
	}

	return &Claim{
		ExternalUserId: externalUserId,
		Roles:          roles,
		Email:          zitadelClaims.Email,
		// Username: zitadelClaims.PreferredUsername,
		StandardClaims: jwt.StandardClaims{
			Audience:  zitadelClaims.Aud,
			ExpiresAt: zitadelClaims.Exp,
			IssuedAt:  zitadelClaims.Iat,
			Issuer:    zitadelClaims.Iss,
			Id:        zitadelClaims.IdToken,
			Subject:   zitadelClaims.Sub,
		},
		AccessToken:   token,
		AuthSessionId: req.SessionId,
		// RefreshToken: ,
	}, nil
}

type BearTokenGetter struct {
	token      string
	authorizer *zitadel_authorization.ZitadelAuthorizer
	ctx        context.Context
}

func NewBearTokenGetter(token string, authorizer *zitadel_authorization.ZitadelAuthorizer, ctx context.Context) *BearTokenGetter {
	return &BearTokenGetter{
		token:      token,
		authorizer: authorizer,
		ctx:        ctx,
	}
}

func (b *BearTokenGetter) Get() (*Claim, error) {
	res, err := b.authorizer.Introspect(b.ctx, b.token)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, fmt.Errorf("claim not found with this token: unauthorize")
	}
	// convert zitadel claims -> internal claims
	claimConverter := &ClaimConverter{}
	internalClaims, err := claimConverter.ConvertToInternalClaims(res, ClaimConverterRequest{
		Token: b.token,
	})
	if err != nil {
		return nil, fmt.Errorf("fail to convert zitadel claims to internal claims: %w", err)
	}
	return internalClaims, nil
}
