//go:build integration

package auth

import (
	"context"
	"net"
	"testing"

	"github.com/hoangdaochuz/ecommerce-microservice-golang/apps/auth/api/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

// Note: Auth service has complex external dependencies (Zitadel, Redis, etc.)
// These integration tests focus on the gRPC transport layer with mocked services

// MockAuthServiceForIntegration provides a mock for auth service testing
type MockAuthServiceForIntegration struct {
	auth.UnimplementedAuthenticateServiceServer
	mock.Mock
}

func (m *MockAuthServiceForIntegration) Login(ctx context.Context, req *auth.LoginRequest) (*auth.RedirectResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.RedirectResponse), args.Error(1)
}

func (m *MockAuthServiceForIntegration) ValidateToken(ctx context.Context, req *auth.ValidateTokenRequest) (*auth.ValidateTokenResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.ValidateTokenResponse), args.Error(1)
}

func (m *MockAuthServiceForIntegration) GetMyProfile(ctx context.Context, req *auth.EmptyRequest) (*auth.GetMyProfileResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.GetMyProfileResponse), args.Error(1)
}

func (m *MockAuthServiceForIntegration) Logout(ctx context.Context, req *auth.EmptyRequest) (*auth.RedirectResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.RedirectResponse), args.Error(1)
}

// setupAuthGRPCServer sets up a gRPC server with bufconn for testing
func setupAuthGRPCServer(t *testing.T, mockService *MockAuthServiceForIntegration) (*grpc.Server, *bufconn.Listener) {
	lis := bufconn.Listen(bufSize)
	server := grpc.NewServer()

	auth.RegisterAuthenticateServiceServer(server, mockService)

	go func() {
		if err := server.Serve(lis); err != nil {
			// Server stopped, expected during test cleanup
		}
	}()

	t.Cleanup(func() {
		server.Stop()
		lis.Close()
	})

	return server, lis
}

// dialBufconn creates a gRPC client connection to the bufconn listener
func dialBufconn(ctx context.Context, lis *bufconn.Listener) (*grpc.ClientConn, error) {
	return grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}

func TestAuthService_Integration_Login(t *testing.T) {
	mockService := new(MockAuthServiceForIntegration)
	_, lis := setupAuthGRPCServer(t, mockService)

	ctx := context.Background()
	conn, err := dialBufconn(ctx, lis)
	require.NoError(t, err)
	defer conn.Close()

	client := auth.NewAuthenticateServiceClient(conn)

	t.Run("login returns redirect URL via gRPC", func(t *testing.T) {
		expectedResponse := &auth.RedirectResponse{
			IsSuccess:   true,
			RedirectURL: "https://auth.example.com/authorize?client_id=test",
		}

		mockService.On("Login", mock.Anything, mock.MatchedBy(func(req *auth.LoginRequest) bool {
			return req.Username == "testuser@example.com"
		})).Return(expectedResponse, nil).Once()

		resp, err := client.Login(ctx, &auth.LoginRequest{
			Username: "testuser@example.com",
		})

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.True(t, resp.IsSuccess)
		assert.Contains(t, resp.RedirectURL, "https://auth.example.com")
		mockService.AssertExpectations(t)
	})

	t.Run("login handles errors via gRPC", func(t *testing.T) {
		mockService.On("Login", mock.Anything, mock.Anything).Return(nil, assert.AnError).Once()

		resp, err := client.Login(ctx, &auth.LoginRequest{
			Username: "error@example.com",
		})

		require.Error(t, err)
		assert.Nil(t, resp)
		mockService.AssertExpectations(t)
	})
}

func TestAuthService_Integration_GetMyProfile(t *testing.T) {
	mockService := new(MockAuthServiceForIntegration)
	_, lis := setupAuthGRPCServer(t, mockService)

	ctx := context.Background()
	conn, err := dialBufconn(ctx, lis)
	require.NoError(t, err)
	defer conn.Close()

	client := auth.NewAuthenticateServiceClient(conn)

	t.Run("get profile returns user data via gRPC", func(t *testing.T) {
		expectedResponse := &auth.GetMyProfileResponse{
			Username:  "testuser",
			Email:     "testuser@example.com",
			FirstName: "Test",
			LastName:  "User",
			Gender:    "other",
		}

		mockService.On("GetMyProfile", mock.Anything, mock.Anything).Return(expectedResponse, nil).Once()

		resp, err := client.GetMyProfile(ctx, &auth.EmptyRequest{})

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "testuser", resp.Username)
		assert.Equal(t, "testuser@example.com", resp.Email)
		assert.Equal(t, "Test", resp.FirstName)
		assert.Equal(t, "User", resp.LastName)
		mockService.AssertExpectations(t)
	})

	t.Run("get profile handles unauthenticated error via gRPC", func(t *testing.T) {
		mockService.On("GetMyProfile", mock.Anything, mock.Anything).Return(nil, assert.AnError).Once()

		resp, err := client.GetMyProfile(ctx, &auth.EmptyRequest{})

		require.Error(t, err)
		assert.Nil(t, resp)
		mockService.AssertExpectations(t)
	})
}

func TestAuthService_Integration_Logout(t *testing.T) {
	mockService := new(MockAuthServiceForIntegration)
	_, lis := setupAuthGRPCServer(t, mockService)

	ctx := context.Background()
	conn, err := dialBufconn(ctx, lis)
	require.NoError(t, err)
	defer conn.Close()

	client := auth.NewAuthenticateServiceClient(conn)

	t.Run("logout returns end session URL via gRPC", func(t *testing.T) {
		expectedResponse := &auth.RedirectResponse{
			IsSuccess:   true,
			RedirectURL: "https://auth.example.com/v2/endsession?post_logout_redirect_uri=https://app.example.com",
		}

		mockService.On("Logout", mock.Anything, mock.Anything).Return(expectedResponse, nil).Once()

		resp, err := client.Logout(ctx, &auth.EmptyRequest{})

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.True(t, resp.IsSuccess)
		assert.Contains(t, resp.RedirectURL, "endsession")
		mockService.AssertExpectations(t)
	})
}

func TestAuthService_Integration_ValidateToken(t *testing.T) {
	mockService := new(MockAuthServiceForIntegration)
	_, lis := setupAuthGRPCServer(t, mockService)

	ctx := context.Background()
	conn, err := dialBufconn(ctx, lis)
	require.NoError(t, err)
	defer conn.Close()

	client := auth.NewAuthenticateServiceClient(conn)

	t.Run("validates token successfully via gRPC", func(t *testing.T) {
		expectedResponse := &auth.ValidateTokenResponse{
			IsValid: true,
		}

		mockService.On("ValidateToken", mock.Anything, mock.MatchedBy(func(req *auth.ValidateTokenRequest) bool {
			return req.Token == "valid_token_123"
		})).Return(expectedResponse, nil).Once()

		resp, err := client.ValidateToken(ctx, &auth.ValidateTokenRequest{
			Token: "valid_token_123",
		})

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.True(t, resp.IsValid)
		mockService.AssertExpectations(t)
	})

	t.Run("returns invalid for expired token via gRPC", func(t *testing.T) {
		expectedResponse := &auth.ValidateTokenResponse{
			IsValid: false,
		}

		mockService.On("ValidateToken", mock.Anything, mock.Anything).Return(expectedResponse, nil).Once()

		resp, err := client.ValidateToken(ctx, &auth.ValidateTokenRequest{
			Token: "expired_token",
		})

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.False(t, resp.IsValid)
		mockService.AssertExpectations(t)
	})
}

