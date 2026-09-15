package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealth(t *testing.T) {
	app := New(Dependencies{AllowedOrigins: []string{"http://localhost:3000"}})
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()

	app.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if got, want := recorder.Body.String(), "{\"message\":\"ok\"}\n"; got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
}

func TestCORSAllowsFrontendDevelopmentOrigin(t *testing.T) {
	app := New(Dependencies{AllowedOrigins: []string{"http://localhost:5173"}})
	request := httptest.NewRequest(http.MethodOptions, "/api/v1/auth/login", nil)
	request.Header.Set("Origin", "http://localhost:5173")
	request.Header.Set("Access-Control-Request-Method", http.MethodPost)
	recorder := httptest.NewRecorder()

	app.ServeHTTP(recorder, request)

	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("Access-Control-Allow-Origin = %q", got)
	}
}

func TestCORSAllowsPlaylistMutationMethods(t *testing.T) {
	app := New(Dependencies{AllowedOrigins: []string{"http://localhost:5173"}})
	for _, method := range []string{http.MethodPut, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodOptions, "/api/v1/playlists/1", nil)
			request.Header.Set("Origin", "http://localhost:5173")
			request.Header.Set("Access-Control-Request-Method", method)
			recorder := httptest.NewRecorder()

			app.ServeHTTP(recorder, request)

			if got := recorder.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(got, method) {
				t.Fatalf("Access-Control-Allow-Methods = %q, want it to contain %q", got, method)
			}
		})
	}
}
