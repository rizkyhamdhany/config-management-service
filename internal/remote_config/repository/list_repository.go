package repository

import (
	"configuration-management-service/internal/remote_config/model"
	"context"
	"encoding/json"
)

func (r *repo) List(ctx context.Context, name string) ([]model.RemoteConfig, error) {
	const q = `
		SELECT name, type, version, data, created_at
		FROM configs
		WHERE name = ?
		ORDER BY version ASC
	`
	rows, err := r.db.QueryContext(ctx, q, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.RemoteConfig
	for rows.Next() {
		var cfg model.RemoteConfig
		var dataStr string
		if err := rows.Scan(&cfg.Name, &cfg.Type, &cfg.Version, &dataStr, &cfg.CreatedAt); err != nil {
			return nil, err
		}
		cfg.Data = json.RawMessage(dataStr)
		out = append(out, cfg)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
