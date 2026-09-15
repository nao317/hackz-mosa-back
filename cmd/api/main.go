package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v5"
	"hackz-mosa-back/internal/server"
)

const shutdownTimeout = 10 * time.Second

func main() {
	app := server.New()
	address := ":" + envOrDefault("PORT", "8080")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	slog.Info("starting server", "address", address)
	config := echo.StartConfig{
		Address:         address,
		GracefulTimeout: shutdownTimeout,
	}
	if err := config.Start(ctx, app); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server stopped unexpectedly", "error", err)
		os.Exit(1)
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
