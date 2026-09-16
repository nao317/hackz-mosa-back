package server

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	httpadapter "hackz-mosa-back/internal/adapter/http"
)

type response struct {
	Message string `json:"message"`
}

type Dependencies struct {
	AuthHandler       *httpadapter.AuthHandler
	PlaylistHandler   *httpadapter.PlaylistHandler
	MapMappingHandler *httpadapter.MapMappingHandler
	AllowedOrigins    []string
}

// New builds the HTTP server and registers its middleware and routes.
func New(dependencies Dependencies) *echo.Echo {
	app := echo.New()
	app.Use(middleware.RequestLogger())
	app.Use(middleware.Recover())
	app.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: dependencies.AllowedOrigins,
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{"Accept", "Authorization", "Content-Type"},
		MaxAge:       3600,
	}))

	app.GET("/", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, response{Message: "Hello, Echo!"})
	})
	app.GET("/health", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, response{Message: "ok"})
	})
	if dependencies.AuthHandler != nil {
		dependencies.AuthHandler.Register(app.Group("/api/v1"))
	}
	if dependencies.PlaylistHandler != nil {
		dependencies.PlaylistHandler.Register(app.Group("/api/v1"))
	}
	if dependencies.MapMappingHandler != nil {
		dependencies.MapMappingHandler.Register(app.Group("/api/v1"))
	}

	return app
}
