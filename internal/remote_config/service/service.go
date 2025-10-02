package service

import (
	"configuration-management-service/internal/remote_config/model"
	"configuration-management-service/internal/remote_config/repository"
	"configuration-management-service/internal/remote_config/validator"
	"context"
	"encoding/json"
	"errors"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrInvalidInput  = errors.New("invalid input")
)

type IService interface {
	Create(ctx context.Context, schemaType, name string, data json.RawMessage) (model.RemoteConfig, error)
	Update(ctx context.Context, name string, data json.RawMessage) (model.RemoteConfig, error)
	Get(ctx context.Context, name string, version *int) (model.RemoteConfig, error)
	ListVersions(ctx context.Context, name string) ([]model.RemoteConfig, error)
	Rollback(ctx context.Context, name string, version int) (model.RemoteConfig, error)
}

type service struct {
	repo      repository.IRepo
	validator validator.ISchemaValidator
}

func NewService(repo repository.IRepo, schemaValidator validator.ISchemaValidator) IService {
	return service{
		repo:      repo,
		validator: schemaValidator,
	}
}
