package model

import "encoding/json"

type RemoteConfig struct {
	Name      string          `json:"name"`
	Type      string          `json:"type"`
	Version   int             `json:"version"`
	Data      json.RawMessage `json:"data"`
	CreatedAt string          `json:"created_at"`
}

type RemoteConfigCreateRequest struct {
	Type string          `json:"type"`
	Name string          `json:"name"`
	Data json.RawMessage `json:"data"`
}

type RemoteConfigUpdateRequest struct {
	Data json.RawMessage `json:"data"`
}

type RemoteConfigRollbackRequest struct {
	Version int `json:"version"`
}
