package zitadel_authentication

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OIDCConfig struct {
	Issuer                           string   `json:"issuer"`
	AuthorizationEndpoint            string   `json:"authorization_endpoint"`
	TokenEndpoint                    string   `json:"token_endpoint"`
	TokenEndpointAuthMethodSupported []string `json:"token_endpoint_auth_method_supported"`
	UserInfoEndpoint                 string   `json:"userinfo_endpoint"`
	EndSessionEndpoint               string   `json:"end_session_endpoint"`
	JWKSURI                          string   `json:"jwks_uri"`
	RegistrationEndpoint             string   `json:"registration_endpoint"`
	ScopesSupported                  []string `json:"scopes_supported"`
	ResponseTypesSupported           []string `json:"response_types_supported"`
	ClaimsSupported                  []string `json:"claims_supported"`
}

type OIDCDiscovery interface {
	Discovery(ctx context.Context, domain string) (*OIDCConfig, error)
}

type OIDCDiscoveryImpl struct{}

func NewOIDCDiscoveryImpl() *OIDCDiscoveryImpl {
	return &OIDCDiscoveryImpl{}
}
func (o *OIDCDiscoveryImpl) Discovery(ctx context.Context, domain string) (*OIDCConfig, error) {
	discoveryEndpoint := fmt.Sprintf("%s/.well-known/openid-configuration", domain)
	httpClient := &http.Client{
		Transport: http.DefaultTransport,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", discoveryEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("fail to create request with context: %w", err)
	}
	response, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fail to get oidc config: %w", err)
	}
	defer response.Body.Close()
	resBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("fail to read response")
	}
	var oidcConfig OIDCConfig
	err = json.Unmarshal(resBytes, &oidcConfig)
	if err != nil {
		return nil, fmt.Errorf("fail to unmarshal response body")
	}
	return &oidcConfig, nil
}

type Config struct {
	AuthDomain            string
	EncryptKey            string
	ZitadelClientID       string
	ExpiredInSeconds      int
	EndSessionEndpoint    string
	TokenEndpoint         string
	AuthorizationEndpoint string
	RedirectURI           string
	PostLoginSuccessURI   string
	UserInfoEndpoint      string
}
