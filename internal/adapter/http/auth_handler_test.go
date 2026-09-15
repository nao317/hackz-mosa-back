package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"hackz-mosa-back/internal/domain"
)

type signInStub struct {
	token string
}

func (s *signInStub) Execute(_ context.Context, token string) (domain.User, error) {
	s.token = token
	return domain.User{FirebaseUID: "firebase-user", Provider: domain.ProviderGoogle}, nil
}

func TestAuthenticate(t *testing.T) {
	service := &signInStub{}
	app := echo.New()
	NewAuthHandler(service).Register(app.Group("/api/v1"))
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	request.Header.Set("Authorization", "Bearer firebase-id-token")
	recorder := httptest.NewRecorder()

	app.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if service.token != "firebase-id-token" {
		t.Fatalf("token = %q", service.token)
	}
	if got := recorder.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", got)
	}
}

func TestAuthenticateRequiresBearerToken(t *testing.T) {
	app := echo.New()
	NewAuthHandler(&signInStub{}).Register(app.Group("/api/v1"))
	request := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	recorder := httptest.NewRecorder()

	app.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}
