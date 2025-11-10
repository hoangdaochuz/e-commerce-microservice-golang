package auth

import (
	"context"

	"github.com/hoangdaochuz/ecommerce-microservice-golang/apps/auth/api/auth"
	auth_service "github.com/hoangdaochuz/ecommerce-microservice-golang/apps/auth/services/auth"
	custom_nats "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/custom-nats"
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

var _ = di.Make[*AuthServiceApp](NewAuthServiceApp)

func (a *AuthServiceApp) Login(ctx context.Context, req *auth.LoginRequest) (*auth.RedirectResponse, error) {
	return a.authService.Login(ctx, req)
}

func (a *AuthServiceApp) Callback(ctx context.Context, req *auth.CallbackRequest) (*custom_nats.Response, error) {
	return a.authService.Callback(ctx, req)
}

func (a *AuthServiceApp) ValidateToken(ctx context.Context, req *auth.ValidateTokenRequest) (*auth.ValidateTokenResponse, error) {
	// implement later
	return nil, nil
}

func (a *AuthServiceApp) GetMyProfile(ctx context.Context, req *auth.EmptyRequest) (*auth.GetMyProfileResponse, error) {
	return a.authService.GetMyProfile(ctx, req)
}

func (a *AuthServiceApp) Logout(ctx context.Context, req *auth.EmptyRequest) (*auth.RedirectResponse, error) {
	return a.authService.Logout(ctx, req)
}
