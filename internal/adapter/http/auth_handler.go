package http

import (
	"context"
	"errors"
	"log/slog"
	stdhttp "net/http"
	"strings"

	"github.com/labstack/echo/v5"
	"hackz-mosa-back/internal/domain"
)

type signInUseCase interface {
	Execute(ctx context.Context, idToken string) (domain.User, error)
}

type AuthHandler struct {
	signIn signInUseCase
}

func NewAuthHandler(signIn signInUseCase) *AuthHandler {
	return &AuthHandler{signIn: signIn}
}

func (h *AuthHandler) Register(group *echo.Group) {
	group.POST("/auth/login", h.authenticate)
	group.GET("/me", h.authenticate)
}

func (h *AuthHandler) authenticate(c *echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store")
	c.Response().Header().Set("Pragma", "no-cache")

	idToken, ok := bearerToken(c.Request().Header.Get("Authorization"))
	if !ok {
		return errorResponse(c, stdhttp.StatusUnauthorized, "unauthorized", "valid Firebase ID token is required")
	}

	user, err := h.signIn.Execute(c.Request().Context(), idToken)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidToken):
			return errorResponse(c, stdhttp.StatusUnauthorized, "unauthorized", "valid Firebase ID token is required")
		case errors.Is(err, domain.ErrUnsupportedProvider):
			return errorResponse(c, stdhttp.StatusForbidden, "unsupported_provider", "authentication provider is not supported")
		default:
			slog.Error("authentication request failed", "error", err)
			return errorResponse(c, stdhttp.StatusInternalServerError, "internal_error", "internal server error")
		}
	}

	return c.JSON(stdhttp.StatusOK, map[string]domain.User{"user": user})
}

func bearerToken(header string) (string, bool) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}

func errorResponse(c *echo.Context, status int, code, message string) error {
	return c.JSON(status, map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
