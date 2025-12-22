// Package fixtures provides auth test data fixtures
package fixtures

import (
	"net/http"
	"net/http/httptest"
)

// Auth fixtures for testing

// ValidUsername returns a valid test username
func ValidUsername() string {
	return "testuser@example.com"
}

// ValidPassword returns a valid test password
func ValidPassword() string {
	return "securePassword123!"
}

// ValidAccessToken returns a valid test access token
func ValidAccessToken() string {
	return "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.test"
}

// ValidSessionID returns a valid session ID for testing
func ValidSessionID() string {
	return "session_abc123def456"
}

// AuthCookie returns a test authentication cookie
func AuthCookie(cookieName, value string) *http.Cookie {
	return &http.Cookie{
		Name:     cookieName,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
	}
}

// RequestWithAuthCookie creates a request with an auth cookie attached
func RequestWithAuthCookie(method, path, cookieName, cookieValue string) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	req.AddCookie(AuthCookie(cookieName, cookieValue))
	return req
}

// RequestWithoutAuthCookie creates a request without any auth cookie
func RequestWithoutAuthCookie(method, path string) *http.Request {
	return httptest.NewRequest(method, path, nil)
}

// SampleUserProfile returns sample user profile data
type SampleUserProfileData struct {
	Username  string
	Email     string
	FirstName string
	LastName  string
	Gender    string
}

// SampleUserProfile returns a sample user profile for testing
func SampleUserProfile() SampleUserProfileData {
	return SampleUserProfileData{
		Username:  "testuser",
		Email:     "testuser@example.com",
		FirstName: "Test",
		LastName:  "User",
		Gender:    "other",
	}
}

// LoginCallbackData represents callback data for login flow testing
type LoginCallbackData struct {
	Code        string
	State       string
	Error       string
	Description string
}

// ValidLoginCallback returns valid login callback data
func ValidLoginCallback() LoginCallbackData {
	return LoginCallbackData{
		Code:  "auth_code_123",
		State: "state_abc",
	}
}

// ErrorLoginCallback returns error login callback data
func ErrorLoginCallback() LoginCallbackData {
	return LoginCallbackData{
		Error:       "access_denied",
		Description: "User denied access",
	}
}
