package service

import (
	"configuration-management-service/internal/remote_config/model"
	"configuration-management-service/internal/remote_config/repository"
	"context"
	"errors"
	"strings"
)

func (s service) ListVersions(ctx context.Context, name string) ([]model.RemoteConfig, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrInvalidInput
	}
	res, err := s.repo.List(ctx, name)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return res, nil
}
