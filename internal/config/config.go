package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port               string
	DatabaseURL        string
	FirebaseProjectID  string
	CORSAllowedOrigins []string
}

func Load() (Config, error) {
	config := Config{
		Port:              envOrDefault("PORT", "8080"),
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		FirebaseProjectID: os.Getenv("FIREBASE_PROJECT_ID"),
		CORSAllowedOrigins: csvEnvOrDefault(
			"CORS_ALLOWED_ORIGINS",
			"http://localhost:5173,http://localhost:3000",
		),
	}

	if config.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if config.FirebaseProjectID == "" {
		return Config{}, fmt.Errorf("FIREBASE_PROJECT_ID is required")
	}

	return config, nil
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func csvEnvOrDefault(key, fallback string) []string {
	values := strings.Split(envOrDefault(key, fallback), ",")
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			result = append(result, value)
		}
	}
	return result
}
