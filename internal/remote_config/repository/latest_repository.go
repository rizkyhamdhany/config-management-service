package repository

import (
	"configuration-management-service/internal/remote_config/model"
	"context"
)

func (r *repo) Latest(ctx context.Context, name string) (model.RemoteConfig, error) {
	const q = `
		SELECT name, type, version, data, created_at
		FROM configs
		WHERE name = ?
		ORDER BY version DESC
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, q, name)
	return scanConfig(row)
}
