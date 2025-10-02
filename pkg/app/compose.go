package app

import (
	"configuration-management-service/db"
	"configuration-management-service/internal/remote_config"
	"configuration-management-service/pkg/auth"
	"configuration-management-service/pkg/config"
	"configuration-management-service/pkg/httpx"
	"context"
	"time"

	"github.com/labstack/echo/v4"
)

func BuildServer() (*echo.Echo, func(ctx context.Context) error, error) {
	cfg := config.Load()

	sqlDB, err := db.Open(db.Config{DSN: cfg.DSN})
	if err != nil {
		return nil, nil, err
	}
	if err := db.MigrateSQLFiles(sqlDB, "db/migrations"); err != nil {
		return nil, nil, err
	}

	e := httpx.NewEcho(&httpx.Config{
		EnableCORS:   false,
		MaxBodyBytes: 2 << 20, // 2 MiB global
		Timeout:      15 * time.Second,
	})
	writeLimit := httpx.WriteBodyLimiter(1 << 20)

	e.GET("/healthz", httpx.HealthHandler(cfg.Service, cfg.Version, sqlDB))
	api := e.Group("/api", auth.StaticKeyMiddleware(cfg.StaticKey))

	remoteConfigModule := remote_config.InitModule(sqlDB)
	remoteConfigModule.RegisterRoute(api, writeLimit)

	return e, e.Shutdown, nil
}
