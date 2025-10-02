package httpx

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func HealthHandler(service, version string, db *sql.DB) echo.HandlerFunc {
	started := time.Now()
	return func(c echo.Context) error {
		uptime := time.Since(started)

		resp := echo.Map{
			"status":   http.StatusOK,
			"service":  service,
			"version":  version,
			"uptime_s": uptime.Seconds(),
		}

		if db != nil {
			ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
			defer cancel()

			if err := db.PingContext(ctx); err != nil {
				resp["status"] = http.StatusServiceUnavailable
				resp["db"] = echo.Map{
					"ok":    false,
					"error": err.Error(),
				}
				return c.JSON(http.StatusServiceUnavailable, resp)
			}
			resp["db"] = echo.Map{"ok": true}
		}

		return c.JSON(http.StatusOK, resp)
	}
}
