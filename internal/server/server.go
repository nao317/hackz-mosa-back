package server

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

type response struct {
	Message string `json:"message"`
}

// New builds the HTTP server and registers its middleware and routes.
func New() *echo.Echo {
	app := echo.New()
	app.Use(middleware.RequestLogger())
	app.Use(middleware.Recover())

	app.GET("/", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, response{Message: "Hello, Echo!"})
	})
	app.GET("/health", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, response{Message: "ok"})
	})

	return app
}
