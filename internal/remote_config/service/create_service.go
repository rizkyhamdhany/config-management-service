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

func (s service) Create(ctx context.Context, schemaType, name string, data json.RawMessage) (model.RemoteConfig, error) {
	schemaType = strings.TrimSpace(schemaType)
	name = strings.TrimSpace(name)
	if schemaType == "" || name == "" {
		return model.RemoteConfig{}, ErrInvalidInput
	}
	if len(data) == 0 {
		return model.RemoteConfig{}, fmt.Errorf("%w: empty data", ErrInvalidInput)
	}

	if err := s.validator.Validate(schemaType, data); err != nil {
		return model.RemoteConfig{}, fmt.Errorf("%w: %s", ErrInvalidInput, err.Error())
	}

	cfg, err := s.repo.Create(ctx, schemaType, name, data)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrAlreadyExists):
			return model.RemoteConfig{}, ErrAlreadyExists
		case errors.Is(err, repository.ErrNotFound):
			return model.RemoteConfig{}, ErrNotFound
		default:
			return model.RemoteConfig{}, err
		}
	}
	return cfg, nil
}
