package service

import (
	"configuration-management-service/internal/remote_config/model"
	"configuration-management-service/internal/remote_config/repository"
	"context"
	"errors"
	"strings"
)

func (s service) Rollback(ctx context.Context, name string, version int) (model.RemoteConfig, error) {
	name = strings.TrimSpace(name)
	if name == "" || version <= 0 {
		return model.RemoteConfig{}, ErrInvalidInput
	}

	// Fetch the target historical version
	target, err := s.repo.ByVersion(ctx, name, version)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.RemoteConfig{}, ErrNotFound
		}
		return model.RemoteConfig{}, err
	}

	cfg, err := s.repo.Append(ctx, name, target.Data)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.RemoteConfig{}, ErrNotFound
		}
		return model.RemoteConfig{}, err
	}
	return cfg, nil
}
