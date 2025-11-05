package zitadel_authentication

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	zitadel_pkg "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/zitadel"
	"github.com/zitadel/oidc/v3/pkg/crypto"
	zitadel_http "github.com/zitadel/oidc/v3/pkg/http"
	zitadel_oidc "github.com/zitadel/oidc/v3/pkg/oidc"
	"github.com/zitadel/zitadel-go/v3/pkg/authentication"
	"golang.org/x/oauth2"
)

type Auth[T any] struct {
	ctx               context.Context
	Session           SessionHandler[T]
	CodeVerifierStore SessionHandler[string]
	Config            Config
	OIDCConfig        OIDCConfig
	SessionCookieName string
	OIDCDiscovery     OIDCDiscovery
}

func NewAuth[T any](ctx context.Context, session SessionHandler[T], config Config, sessionCookieName string, oidcDiscovery OIDCDiscovery, codeVerifierStore SessionHandler[string]) (*Auth[T], error) {
	oidcConfig, err := oidcDiscovery.Discovery(ctx, config.AuthDomain)
	if err != nil {
		return nil, err
	}
	return &Auth[T]{
		ctx:               ctx,
		Session:           session,
		CodeVerifierStore: codeVerifierStore,
		Config:            config,
		SessionCookieName: sessionCookieName,
		OIDCDiscovery:     oidcDiscovery,
		OIDCConfig:        *oidcConfig,
	}, nil
}

func (a *Auth[T]) GetAuthEncryptKey() string {
	return a.Config.EncryptKey
}

func getDecodeEncryptKey(key string) (string, error) {
	bytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (a *Auth[T]) SetSessionCookie(cookieHandler CookieHandler, sessionId, cookiePath string) error {
	decodeEncryptKey, err := getDecodeEncryptKey(a.Config.EncryptKey)
	if err != nil {
		return err
	}

	encryptSessionId, err := crypto.EncryptAES(sessionId, decodeEncryptKey)
	if err != nil {
		return fmt.Errorf("fail to encrypt session id: %w", err)
	}
	return cookieHandler.SetCookie(a.SessionCookieName, encryptSessionId, cookiePath, a.Config.ExpiredInSeconds)
}

func (a *Auth[T]) DeleteSessionCookie(cookieHandler CookieHandler, cookiePath string) error {
	return cookieHandler.DelCookie(a.SessionCookieName, cookiePath)
}

type ClaimGetter[T any] interface {
	Get() (*T, error)
}

func IsAuthenticated[T any](claimGetter ClaimGetter[T]) (*T, error) {
	return claimGetter.Get()
}

func (a *Auth[T]) logoutUrl(idToken, postLogoutUrl string) (string, error) {
	state := authentication.State{
		RequestedURI: "",
	}
	decodeEncryptKey, err := getDecodeEncryptKey(a.Config.EncryptKey)
	if err != nil {
		return "", err
	}

	encryptedState, err := state.Encrypt(decodeEncryptKey)
	if err != nil {
		return "", fmt.Errorf("fail to encrypt state: %w", err)
	}
	req := zitadel_oidc.EndSessionRequest{
		IdTokenHint:           idToken,
		ClientID:              a.Config.ZitadelClientID,
		PostLogoutRedirectURI: postLogoutUrl,
		State:                 encryptedState,
	}
	endSessionUrl, err := url.Parse(a.Config.EndSessionEndpoint)
	if err != nil {
		return "", fmt.Errorf("fail to parse EndSessionEndpoint: %w", err)
	}
	reqParams, err := zitadel_http.URLEncodeParams(req, zitadel_http.Encoder(zitadel_oidc.NewEncoder()))
	if err != nil {
		return "", fmt.Errorf("fail to encoded request params: %w", err)
	}
	endSessionUrl.RawQuery = reqParams.Encode()
	return endSessionUrl.String(), nil
}

func (a *Auth[T]) LogoutUrl(cookieHanlder CookieHandler, cookiesPath, idToken, postLogoutURL string) (string, error) {
	cookieValue, err := cookieHanlder.GetCookie(a.SessionCookieName)
	if err != nil {
		return "", err
	}

	decodeEncryptKey, err := getDecodeEncryptKey(a.Config.EncryptKey)
	if err != nil {
		return "", err
	}
	sessionId, err := crypto.DecryptAES(cookieValue, decodeEncryptKey)
	if err != nil {
		return "", err
	}

	// delete cookie of user-agent (browser)
	err = cookieHanlder.DelCookie(a.SessionCookieName, cookiesPath)
	if err != nil {
		return "", fmt.Errorf("fail to delete cookie")
	}

	// delete session info in redis
	err = a.Session.Del(a.ctx, sessionId)
	if err != nil {
		return "", fmt.Errorf("fail to delete session redis: %w", err)
	}

	endSessionUrl, err := a.logoutUrl(idToken, postLogoutURL)
	if err != nil {
		return "", err
	}
	return endSessionUrl, nil
}

type (
	ScopeGetter func() []ScopeOps
	ScopeOps    func() string
)

func WithScope(scope string) ScopeOps {
	return func() string {
		return scope
	}
}

func WithOpenIdScope() ScopeOps {
	return WithScope("openid")
}

func WithEmailScope() ScopeOps {
	return WithScope("email")
}

func WithProfileScope() ScopeOps {
	return WithScope("profile")
}

func WithAddressScope() ScopeOps {
	return WithScope("address")
}

func WithPhoneScope() ScopeOps {
	return WithScope("phone")
}

func WithOfflineScope() ScopeOps {
	return WithScope("offline_access")
}

func WithProjectRoleScope(role string) ScopeOps {
	return WithScope(fmt.Sprintf("urn:zitadel:iam:org:project:role:%s", role))
}

type LoginOps func() oauth2.AuthCodeOption

func WithLoginHint(hint string) LoginOps {
	return func() oauth2.AuthCodeOption {
		return oauth2.SetAuthURLParam("login_hint", hint)
	}
}

func (a *Auth[T]) storeCodeVerifier(codeVerifier string, encryptedState string) error {
	return a.CodeVerifierStore.Set(a.ctx, encryptedState, codeVerifier, a.Config.ExpiredInSeconds)
}

// AuthCodeUrl generate a URL redirect a agent (web browser) to login UI page
func (a *Auth[T]) AuthCodeUrl(r *http.Request, w http.ResponseWriter, postLoginSuccessURI string, scopeGetter ScopeGetter, loginOpts []LoginOps) (string, error) {
	var scopes []string
	state := &authentication.State{
		RequestedURI: postLoginSuccessURI,
	}

	decodeEnctyptKey, err := getDecodeEncryptKey(a.Config.EncryptKey)
	if err != nil {
		return "", err
	}

	encryptedState, err := state.Encrypt(decodeEnctyptKey)
	if err != nil {
		return "", fmt.Errorf("fail to encrypt state: %w", err)
	}

	// err = storeStateCookie(cookieHandler, encryptedState, postLoginSuccessURI)
	// if err != nil {
	// 	return "", fmt.Errorf("fail to store state to cookie")
	// }

	codeVerifier := oauth2.GenerateVerifier()
	err = a.storeCodeVerifier(codeVerifier, encryptedState)
	if err != nil {
		return "", fmt.Errorf("fail to store code verifier")
	}

	// store code_verifier into cookie to use on exchange code
	codeChallenge := oauth2.S256ChallengeFromVerifier(codeVerifier)

	loginOptions := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	}

	for _, loginOpt := range loginOpts {
		loginOptions = append(loginOptions, loginOpt())
	}

	scopesOpts := scopeGetter()
	for _, scopeOpt := range scopesOpts {
		scopes = append(scopes, scopeOpt())
	}

	oauth2Config := oauth2.Config{
		ClientID:    a.Config.ZitadelClientID,
		Endpoint:    oauth2.Endpoint{AuthURL: a.Config.AuthorizationEndpoint},
		RedirectURL: a.Config.RedirectURI,
		Scopes:      scopes,
	}

	authUrl := oauth2Config.AuthCodeURL(encryptedState, loginOptions...)
	return authUrl, nil
}

type Token struct {
	Token   *oauth2.Token
	IdToken string
	State   string
}

func (a *Auth[T]) getToken(r *http.Request, redirectURI string) (*Token, error) {

	queryParams := r.URL.Query()
	code := queryParams.Get("code")
	state := queryParams.Get("state")
	errors := queryParams.Get("error")
	errDesc := queryParams.Get("error_description")

	if code == "" {
		return nil, fmt.Errorf("code is empty")
	}
	if errors != "" || errDesc != "" {
		return nil, fmt.Errorf("error when make authorization request. error: %s, error descrition: %s, state: %s", errors, errDesc, state)
	}

	codeVerifier, err := a.CodeVerifierStore.Get(a.ctx, state)
	if err != nil {
		return nil, fmt.Errorf("fail to get code verifier from session store: %w", err)
	}
	if codeVerifier == nil || *codeVerifier == "" {
		return nil, fmt.Errorf("cannot get code verifier from session store")
	}
	err = a.CodeVerifierStore.Del(a.ctx, state)
	if err != nil {
		return nil, fmt.Errorf("fail to delete code verifier: %w", err)
	}

	authCodeOption := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("code_verifier", *codeVerifier),
		oauth2.SetAuthURLParam("client_id", a.Config.ZitadelClientID),
	}

	config := oauth2.Config{
		// ClientID:    a.Config.ZitadelClientID,
		Endpoint:    oauth2.Endpoint{TokenURL: a.Config.TokenEndpoint},
		RedirectURL: redirectURI,
	}

	token, err := config.Exchange(r.Context(), code, authCodeOption...)
	if err != nil {
		return nil, fmt.Errorf("fail while exchange code: %w", err)
	}

	idToken := token.Extra("id_token").(string)
	return &Token{
		Token:   token,
		IdToken: idToken,
		State:   state,
	}, nil
}

func (a *Auth[T]) userInfo(ctx context.Context, token *Token) (*zitadel_pkg.ZitadelClaim, error) {
	accessToken := token.Token.AccessToken
	if accessToken == "" {
		return nil, fmt.Errorf("access token is empty")
	}
	httpClient := http.Client{
		Transport: http.DefaultTransport,
	}
	req, err := http.NewRequestWithContext(ctx, "GET", a.Config.UserInfoEndpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var claims zitadel_pkg.ZitadelClaim
	err = json.Unmarshal(bodyBytes, &claims)
	if err != nil {
		return nil, err
	}
	claims.IdToken = token.IdToken
	claims.Token = accessToken
	return &claims, nil
}

func (a *Auth[T]) Callback(r *http.Request, w http.ResponseWriter, appPath string, convertToInternalClaims func(zitadelClaim zitadel_pkg.ZitadelClaim, token *Token, sessionId string) (*T, error)) error {
	// app path is a entry point: example http://localhost:8080/pro/next-shop --> path is /pro/next-shop will set in cookie.
	// So when other request with enpoint match with /pro/next-shop  will have cookie and backend can read this cookie.
	// Because all request will start with endpoint something like that, so the cookie will be set and can be read by backend.
	httpCookieHandler := NewHttpCookie(r, w)
	token, err := a.getToken(r, a.Config.RedirectURI)
	if err != nil {
		return fmt.Errorf("fail to get token: %w", err)
	}

	sessionId := uuid.New()
	err = a.SetSessionCookie(httpCookieHandler, sessionId.String(), appPath)
	if err != nil {
		return fmt.Errorf("fail to set cookie: %w", err)
	}
	zitadelClaims, err := a.userInfo(r.Context(), token)
	if err != nil {
		return fmt.Errorf("fail to get user info from token: %w", err)
	}
	if zitadelClaims == nil {
		return fmt.Errorf("user claims not found")
	}
	// convert claims to internal user claim
	internalClaims, err := convertToInternalClaims(*zitadelClaims, token, sessionId.String())
	if err != nil {
		return fmt.Errorf("fail to convert zitadel claim to internal claim: %w", err)
	}
	if internalClaims == nil {
		return fmt.Errorf("internal claim is empty")
	}
	//  store user info to session by session id
	err = a.Session.Set(a.ctx, sessionId.String(), *internalClaims, a.Config.ExpiredInSeconds)
	if err != nil {
		return fmt.Errorf("fail to store user info to session store")
	}
	// redirect to entry point of application
	decodeEncryptKey, err := getDecodeEncryptKey(a.Config.EncryptKey)
	if err != nil {
		return err
	}
	decryptedState, err := authentication.DecryptState(token.State, decodeEncryptKey)
	if err != nil {
		return fmt.Errorf("fail to decrypt a state: %w", err)
	}
	http.Redirect(w, r, decryptedState.RequestedURI, http.StatusFound)
	return nil
}

func (a *Auth[T]) GetCurrentUser(ctx context.Context, cookieHandler CookieHandler) (*T, error) {
	cookies, err := cookieHandler.GetCookie(a.SessionCookieName)
	if err != nil {
		return nil, err
	}
	decodeEncryptKey, err := getDecodeEncryptKey(a.Config.EncryptKey)
	if err != nil {
		return nil, err
	}
	sessionId, err := crypto.DecryptAES(cookies, decodeEncryptKey)
	if err != nil {
		return nil, err
	}
	claims, err := a.Session.Get(ctx, sessionId)
	if err != nil {
		return nil, err
	}
	if claims == nil {
		return nil, fmt.Errorf("session info expired")
	}
	return claims, nil
}
