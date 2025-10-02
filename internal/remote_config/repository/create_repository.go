package repository

import (
	"configuration-management-service/internal/remote_config/model"
	"context"
	"encoding/json"
	"fmt"
)

func (r *repo) Create(ctx context.Context, schemaType, name string, data json.RawMessage) (model.RemoteConfig, error) {
	const q = `
		INSERT INTO configs(name, type, version, data)
		VALUES(?, ?, 1, ?)
	`
	_, err := r.db.ExecContext(ctx, q, name, schemaType, string(data))
	if err != nil {
		if isUniqueViolation(err) {
			return model.RemoteConfig{}, ErrAlreadyExists
		}
		return model.RemoteConfig{}, fmt.Errorf("create: %w", err)
	}
	return r.ByVersion(ctx, name, 1)
}
