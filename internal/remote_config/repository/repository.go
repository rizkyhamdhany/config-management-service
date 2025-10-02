package repository

import (
	"configuration-management-service/internal/remote_config/model"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

type IRepo interface {
	Create(ctx context.Context, schemaType, name string, data json.RawMessage) (model.RemoteConfig, error)
	Append(ctx context.Context, name string, data json.RawMessage) (model.RemoteConfig, error)
	Latest(ctx context.Context, name string) (model.RemoteConfig, error)
	ByVersion(ctx context.Context, name string, version int) (model.RemoteConfig, error)
	List(ctx context.Context, name string) ([]model.RemoteConfig, error)
}

type repo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) IRepo {
	return &repo{db: db}
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanConfig(row rowScanner) (model.RemoteConfig, error) {
	var cfg model.RemoteConfig
	var dataStr string
	if err := row.Scan(&cfg.Name, &cfg.Type, &cfg.Version, &dataStr, &cfg.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.RemoteConfig{}, ErrNotFound
		}
		return model.RemoteConfig{}, err
	}
	cfg.Data = json.RawMessage(dataStr)
	return cfg, nil
}

func byVersionTx(ctx context.Context, tx *sql.Tx, name string, version int) (model.RemoteConfig, error) {
	const q = `
		SELECT name, type, version, data, created_at
		FROM configs
		WHERE name = ? AND version = ?
		LIMIT 1
	`
	return scanConfig(tx.QueryRowContext(ctx, q, name, version))
}

func isUniqueViolation(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique constraint") ||
		strings.Contains(msg, "unique") && strings.Contains(msg, "constraint") ||
		strings.Contains(msg, "constraint failed")
}
