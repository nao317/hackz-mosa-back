package config

import "testing"

func TestLoad(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/app")
	t.Setenv("FIREBASE_PROJECT_ID", "firebase-project")
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:3000, https://app.example.com ")

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(config.CORSAllowedOrigins) != 2 {
		t.Fatalf("origins = %#v", config.CORSAllowedOrigins)
	}
}

func TestLoadRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("FIREBASE_PROJECT_ID", "firebase-project")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want an error")
	}
}
