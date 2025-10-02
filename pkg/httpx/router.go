package httpx

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Config struct {
	EnableCORS   bool          // allow wide-open CORS for local/dev
	MaxBodyBytes int64         // default body limit applied globally (0 = disabled)
	Timeout      time.Duration // per-request server timeout (0 = disabled)
}

func NewEcho(cfg *Config) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.Recover())
	e.Use(middleware.Secure())
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())

	if cfg != nil {
		if cfg.Timeout > 0 {
			e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
				Timeout: cfg.Timeout,
			}))
		}
		if cfg.MaxBodyBytes > 0 {
			e.Use(middleware.BodyLimitWithConfig(middleware.BodyLimitConfig{
				Limit: strconv.FormatInt(cfg.MaxBodyBytes, 10),
			}))
		}
		if cfg.EnableCORS {
			e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
				AllowOrigins: []string{"*"},
				AllowMethods: []string{
					http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions,
				},
				AllowHeaders: []string{"Content-Type", "Authorization"},
			}))
		}
	}

	e.GET("/healthz", func(c echo.Context) error { return c.String(http.StatusOK, "ok") })

	return e
}

func WriteBodyLimiter(maxBytes int64) echo.MiddlewareFunc {
	if maxBytes <= 0 {
		return func(next echo.HandlerFunc) echo.HandlerFunc { return next }
	}
	return middleware.BodyLimitWithConfig(middleware.BodyLimitConfig{Limit: strconv.FormatInt(maxBytes, 10)})
}
