package service

import (
	"configuration-management-service/internal/remote_config/model"
	"configuration-management-service/internal/remote_config/repository"
	"context"
	"errors"
	"strings"
)

func (s service) Get(ctx context.Context, name string, version *int) (model.RemoteConfig, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return model.RemoteConfig{}, ErrInvalidInput
	}

	if version == nil {
		cfg, err := s.repo.Latest(ctx, name)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return model.RemoteConfig{}, ErrNotFound
			}
			return model.RemoteConfig{}, err
		}
		return cfg, nil
	}

	cfg, err := s.repo.ByVersion(ctx, name, *version)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.RemoteConfig{}, ErrNotFound
		}
		return model.RemoteConfig{}, err
	}
	return cfg, nil
}
