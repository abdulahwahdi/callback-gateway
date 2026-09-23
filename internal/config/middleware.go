package config

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"webhook-middleware/internal/pkg/response"
)

// DashboardAuth guards /dashboard/* routes with a static API key, when one
// is configured. In local/dev, leaving DASHBOARD_API_KEY unset keeps the
// dashboard open. In any shared environment, set it and have your
// dashboard frontend send it as the X-API-Key header.
func DashboardAuth(apiKey string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if apiKey == "" {
				return next(c)
			}
			if c.Request().Header.Get("X-API-Key") != apiKey {
				return response.Err(c, http.StatusUnauthorized, "invalid or missing X-API-Key")
			}
			return next(c)
		}
	}
}
