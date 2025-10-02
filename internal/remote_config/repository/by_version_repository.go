package repository

import (
	"configuration-management-service/internal/remote_config/model"
	"context"
)

func (r *repo) ByVersion(ctx context.Context, name string, version int) (model.RemoteConfig, error) {
	const q = `
		SELECT name, type, version, data, created_at
		FROM configs
		WHERE name = ? AND version = ?
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, q, name, version)
	return scanConfig(row)
}
