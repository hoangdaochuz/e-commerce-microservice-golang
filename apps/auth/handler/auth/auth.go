package auth

import (
	"context"

	"github.com/hoangdaochuz/ecommerce-microservice-golang/apps/auth/api/auth"
	auth_service "github.com/hoangdaochuz/ecommerce-microservice-golang/apps/auth/services/auth"
	di "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/dependency-injection"
)

type AuthServiceApp struct {
	auth.UnimplementedAuthenticateServiceServer
	authService *auth_service.AuthService
	// other field
}

func NewAuthServiceApp(authService *auth_service.AuthService) *AuthServiceApp {
	return &AuthServiceApp{
		authService: authService,
	}
}

var _ = di.Make(NewAuthServiceApp)

func (a *AuthServiceApp) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	// implement later
	err := a.authService.Login(ctx)
	return nil, err
}

func (a *AuthServiceApp) Callback(ctx context.Context, req *auth.CallbackRequest) (*auth.CallbackResponse, error) {
	// implement later
	return nil, nil
}

func (a *AuthServiceApp) ValidateToken(ctx context.Context, req *auth.ValidateTokenRequest) (*auth.ValidateTokenResponse, error) {
	// implement later
	return nil, nil
}
