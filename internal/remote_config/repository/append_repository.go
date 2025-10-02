package repository

import (
	"configuration-management-service/internal/remote_config/model"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

func (r *repo) Append(ctx context.Context, name string, data json.RawMessage) (model.RemoteConfig, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return model.RemoteConfig{}, fmt.Errorf("append.begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var schemaType string
	var nextVersion int
	const qSel = `
		SELECT COALESCE(MAX(version), 0) + 1 AS next_version,
		       (SELECT type FROM configs WHERE name = ? ORDER BY version DESC LIMIT 1) AS schemaType
		FROM configs
		WHERE name = ?
	`
	if err := tx.QueryRowContext(ctx, qSel, name, name).Scan(&nextVersion, &schemaType); err != nil {
		return model.RemoteConfig{}, fmt.Errorf("append.select: %w", err)
	}
	if schemaType == "" {
		return model.RemoteConfig{}, ErrNotFound
	}

	const qIns = `
		INSERT INTO configs(name, type, version, data)
		VALUES(?, ?, ?, ?)
	`
	if _, err := tx.ExecContext(ctx, qIns, name, schemaType, nextVersion, string(data)); err != nil {
		if isUniqueViolation(err) {
			return model.RemoteConfig{}, ErrAlreadyExists
		}
		return model.RemoteConfig{}, fmt.Errorf("append.insert: %w", err)
	}

	cfg, err := byVersionTx(ctx, tx, name, nextVersion)
	if err != nil {
		return model.RemoteConfig{}, err
	}

	if err := tx.Commit(); err != nil {
		return model.RemoteConfig{}, fmt.Errorf("append.commit: %w", err)
	}
	return cfg, nil
}
