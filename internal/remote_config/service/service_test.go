package service

import (
	"configuration-management-service/internal/remote_config/validator"
	"encoding/json"
)

type stubValidator struct{ err error }

func (s stubValidator) Validate(schemaType string, data json.RawMessage) error { return s.err }

var _ validator.ISchemaValidator = (*stubValidator)(nil)
