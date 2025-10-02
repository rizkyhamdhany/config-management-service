package auth

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

const HeaderAPIKey = "X-Api-Key"

func StaticKeyMiddleware(expectedKey string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key := c.Request().Header.Get(HeaderAPIKey)
			if key == "" || key != expectedKey {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or missing API key")
			}
			return next(c)
		}
	}
}
