package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v5"
	firebaseadapter "hackz-mosa-back/internal/adapter/firebase"
	httpadapter "hackz-mosa-back/internal/adapter/http"
	postgresadapter "hackz-mosa-back/internal/adapter/postgres"
	"hackz-mosa-back/internal/config"
	"hackz-mosa-back/internal/server"
	"hackz-mosa-back/internal/usecase"
)

const shutdownTimeout = 10 * time.Second

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil {
		slog.Error("application stopped", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	users, err := postgresadapter.NewUserRepository(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer users.Close()

	if err := users.Migrate(ctx); err != nil {
		return err
	}

	verifier, err := firebaseadapter.NewTokenVerifier(ctx, cfg.FirebaseProjectID)
	if err != nil {
		return err
	}
	signIn := usecase.NewSignIn(verifier, users)
	authHandler := httpadapter.NewAuthHandler(signIn)
	app := server.New(server.Dependencies{
		AuthHandler:    authHandler,
		AllowedOrigins: cfg.CORSAllowedOrigins,
	})
	address := ":" + cfg.Port

	slog.Info("starting server", "address", address)
	startConfig := echo.StartConfig{
		Address:         address,
		GracefulTimeout: shutdownTimeout,
	}
	if err := startConfig.Start(ctx, app); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve HTTP: %w", err)
	}
	return nil
}
