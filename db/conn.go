package db

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

type Config struct{ DSN string }

func Open(cfg Config) (*sql.DB, error) {
	db, err := sql.Open("sqlite", cfg.DSN)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(10)
	return db, db.Ping()
}
