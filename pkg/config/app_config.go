package config

import (
	"os"
)

type App struct {
	DSN       string
	Service   string
	Version   string
	StaticKey string
}

func Load() App {
	dsn := os.Getenv("DSN")
	if dsn == "" {
		dsn = "file:./data/configs.db?_pragma=busy_timeout=5000&_pragma=journal_mode=WAL"
	}
	staticKey := os.Getenv("S2S_STATIC_KEY")
	if staticKey == "" {
		staticKey = "super-secret-123"
	}
	return App{
		DSN:       dsn,
		Service:   os.Getenv("SERVICE_NAME"),
		Version:   os.Getenv("SERVICE_VERSION"),
		StaticKey: staticKey,
	}
}
