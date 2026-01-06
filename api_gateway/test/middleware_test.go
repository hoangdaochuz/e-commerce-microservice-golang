package api_gateway_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apigateway "github.com/hoangdaochuz/ecommerce-microservice-golang/api_gateway"
	"github.com/stretchr/testify/require"
)

func mockHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{status: "success"}`))
	})
}

func Test_ContentTypeMiddleware(t *testing.T) {
	t.Run("Test set content-type header successfully", func(t *testing.T) {
		handler := apigateway.ContentTypeMiddleware(mockHandler())
		request := httptest.NewRequest("GET", "/api/v1/test", nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		require.Equal(t, http.StatusOK, response.Code)
		require.Equal(t, "application/json", response.Header().Get("Content-Type"))
	})
}

func Test_CorsMiddleware(t *testing.T) {
	t.Run("Test set CORS headers successfully", func(t *testing.T) {
		handler := apigateway.CorsMiddleware(mockHandler())
		request := httptest.NewRequest("GET", "/api/v1/test", nil)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, request)

		require.Equal(t, http.StatusOK, response.Code)
		require.Equal(t, "GET, POST, PUT, PATCH, DELETE, OPTIONS", response.Header().Get("Access-Control-Allow-Methods"))
		require.Equal(t, "Content-Type, Authorization", response.Header().Get("Access-Control-Allow-Headers"))
		require.Equal(t, "true", response.Header().Get("Access-Control-Allow-Credentials"))
	})
}

func Test_AuthMiddleware(t *testing.T) {
	t.Run("Skip check auth when path in skip check path", func(t *testing.T) {
		handler := apigateway.AuthMiddleware(mockHandler())
		request := httptest.NewRequest("GET", "/api/v1/auth/Login", nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		require.Equal(t, http.StatusOK, response.Code)
		require.Equal(t, `{status: "success"}`, response.Body.String())
	})

	t.Run("Unauthorize when access resource with no credentials", func(t *testing.T) {
		handler := apigateway.AuthMiddleware(mockHandler())
		req := httptest.NewRequest("POST", "/api/v1/order/GetOrderbyId", nil)
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)

		require.Equal(t, http.StatusUnauthorized, res.Code)
	})
}

func Test_MiddlewareChain(t *testing.T) {
	t.Run("Test middleware chain", func(t *testing.T) {
		order := []string{}
		middleware1 := func(handler http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, "before-m1")
				handler.ServeHTTP(w, r)
				order = append(order, "after-m1")
			})
		}
		middleware2 := func(handler http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, "before-m2")
				handler.ServeHTTP(w, r)
				order = append(order, "after-m2")
			})
		}

		primeHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "handle-prime-handler")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{status:"OK"}`))
		})

		handler := apigateway.MiddlewareChain(primeHandler, middleware1, middleware2)
		req := httptest.NewRequest("POST", "/api/v1/test", nil)
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		require.Equal(t, http.StatusOK, res.Code)
		require.Equal(t, 5, len(order))
		require.Equal(t, strings.Join(order, "."), "before-m1.before-m2.handle-prime-handler.after-m2.after-m1")
	})
}
