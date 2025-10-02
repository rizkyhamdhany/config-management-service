package service

import (
	"configuration-management-service/internal/remote_config/model"
	"configuration-management-service/internal/remote_config/repository"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

func (s service) Update(ctx context.Context, name string, data json.RawMessage) (model.RemoteConfig, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return model.RemoteConfig{}, ErrInvalidInput
	}
	if len(data) == 0 {
		return model.RemoteConfig{}, fmt.Errorf("%w: empty data", ErrInvalidInput)
	}

	latest, err := s.repo.Latest(ctx, name)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.RemoteConfig{}, ErrNotFound
		}
		return model.RemoteConfig{}, err
	}

	if err := s.validator.Validate(latest.Type, data); err != nil {
		return model.RemoteConfig{}, fmt.Errorf("%w: %s", ErrInvalidInput, err.Error())
	}

	cfg, err := s.repo.Append(ctx, name, data)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.RemoteConfig{}, ErrNotFound
		}
		return model.RemoteConfig{}, err
	}
	return cfg, nil
}
