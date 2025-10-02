package validator

import (
	"encoding/json"
	"errors"

	"github.com/xeipuuv/gojsonschema"
)

type ISchemaValidator interface {
	Validate(schemaType string, data json.RawMessage) error
}

type schemaValidator struct{}

func NewSchemaValidator() ISchemaValidator {
	return schemaValidator{}
}

// Registry of available schemas (hard-coded as required)
var registry = map[string]string{
	"feature_toggle": `
	{
	  "$schema": "http://json-schema.org/draft-07/schema#",
	  "title": "feature_toggle",
	  "type": "object",
	  "properties": {
		"enabled": { "type": "boolean" },
		"rollout_percentage": { "type": "integer", "minimum": 0, "maximum": 100 },
		"tags": { "type": "array", "items": { "type": "string" }, "uniqueItems": true },
		"description": { "type": "string" }
	  },
	  "required": ["enabled"],
	  "additionalProperties": false
	}`,

	"experiment_config": `
	{
	  "$schema": "http://json-schema.org/draft-07/schema#",
	  "title": "experiment_config",
	  "type": "object",
	  "properties": {
		"experiment_key": { "type": "string", "minLength": 1 },
		"active": { "type": "boolean" },
		"variants": {
		  "type": "array",
		  "items": {
			"type": "object",
			"properties": {
			  "name": { "type": "string", "minLength": 1 },
			  "weight": { "type": "number", "minimum": 0 }
			},
			"required": ["name", "weight"],
			"additionalProperties": false
		  },
		  "minItems": 2
		},
		"audience": {
		  "type": "object",
		  "properties": {
			"countries": { "type": "array", "items": { "type": "string" }, "uniqueItems": true },
			"os": { "type": "array", "items": { "type": "string", "enum": ["ios", "android", "web"] }, "uniqueItems": true },
			"min_app_version": { "type": "string" }
		  },
		  "additionalProperties": false
		},
		"description": { "type": "string" }
	  },
	  "required": ["experiment_key", "active", "variants"],
	  "additionalProperties": false
	}`,

	"service_client": `
	{
	  "$schema": "http://json-schema.org/draft-07/schema#",
	  "title": "service_client",
	  "type": "object",
	  "properties": {
		"name": { "type": "string", "minLength": 1 },
		"base_url": { "type": "string", "format": "uri" },
		"timeout_ms": { "type": "integer", "minimum": 100 },
		"retry": {
		  "type": "object",
		  "properties": {
			"max_retries": { "type": "integer", "minimum": 0, "maximum": 10 },
			"backoff_ms": { "type": "integer", "minimum": 0 },
			"jitter": { "type": "boolean" }
		  },
		  "required": ["max_retries"],
		  "additionalProperties": false
		},
		"headers": {
		  "type": "object",
		  "additionalProperties": { "type": "string" }
		},
		"description": { "type": "string" }
	  },
	  "required": ["name", "base_url", "timeout_ms"],
	  "additionalProperties": false
	}`,

	"rate_limit_policy": `
	{
	  "$schema": "http://json-schema.org/draft-07/schema#",
	  "title": "rate_limit_policy",
	  "type": "object",
	  "properties": {
		"identifier_type": { "type": "string", "enum": ["ip", "user", "api_key"] },
		"window_seconds": { "type": "integer", "minimum": 1 },
		"max_requests": { "type": "integer", "minimum": 1 },
		"burst": { "type": "integer", "minimum": 0 },
		"scope": {
		  "type": "array",
		  "items": { "type": "string", "minLength": 1 },
		  "uniqueItems": true
		},
		"description": { "type": "string" }
	  },
	  "required": ["identifier_type", "window_seconds", "max_requests"],
	  "additionalProperties": false
	}`,

	"notification_policy": `
	{
	  "$schema": "http://json-schema.org/draft-07/schema#",
	  "title": "notification_policy",
	  "type": "object",
	  "properties": {
		"channel": { "type": "string", "enum": ["email", "sms", "push"] },
		"enabled": { "type": "boolean" },
		"daily_limit": { "type": "integer", "minimum": 0 },
		"template_id": { "type": "string" },
		"placeholders": { "type": "array", "items": { "type": "string", "minLength": 1 }, "uniqueItems": true },
		"description": { "type": "string" }
	  },
	  "required": ["channel", "enabled"],
	  "additionalProperties": false
	}`,

	"schedule_rule": `
	{
	  "$schema": "http://json-schema.org/draft-07/schema#",
	  "title": "schedule_rule",
	  "type": "object",
	  "properties": {
		"active": { "type": "boolean" },
		"timezone": { "type": "string", "minLength": 1 },
		"cron": { "type": "string", "minLength": 1 },
		"windows": {
		  "type": "array",
		  "items": {
			"type": "object",
			"properties": {
			  "start": { "type": "string", "format": "date-time" },
			  "end": { "type": "string", "format": "date-time" }
			},
			"required": ["start", "end"],
			"additionalProperties": false
		  }
		},
		"description": { "type": "string" }
	  },
	  "required": ["active", "timezone"],
	  "additionalProperties": false
	}`,

	"threshold_policy": `
	{
	  "$schema": "http://json-schema.org/draft-07/schema#",
	  "title": "threshold_policy",
	  "type": "object",
	  "properties": {
		"metric": { "type": "string", "minLength": 1 },
		"unit": { "type": "string", "enum": ["count", "ms", "percent", "amount"] },
		"min": { "type": ["number", "null"] },
		"max": { "type": ["number", "null"] },
		"inclusive": { "type": "boolean", "default": true },
		"enabled": { "type": "boolean" },
		"description": { "type": "string" }
	  },
	  "required": ["metric", "unit", "enabled"],
	  "additionalProperties": false
	}`,
}

func (s schemaValidator) Validate(schemaType string, data json.RawMessage) error {
	schema, ok := registry[schemaType]
	if !ok {
		return errors.New("unknown config type: " + schemaType)
	}
	res, err := gojsonschema.Validate(gojsonschema.NewStringLoader(schema), gojsonschema.NewBytesLoader(data))
	if err != nil {
		return err
	}
	if !res.Valid() {
		if len(res.Errors()) > 0 {
			return errors.New(res.Errors()[0].String())
		}
		return errors.New("validation failed")
	}
	return nil
}
