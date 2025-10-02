package service

import (
	"context"
	"encoding/json"
	"testing"

	"configuration-management-service/internal/remote_config/model"
	"configuration-management-service/internal/remote_config/repository"
	repoMock "configuration-management-service/internal/remote_config/repository/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_service_Create(t *testing.T) {
	type exRes struct {
		res model.RemoteConfig
		err error
	}

	cases := []struct {
		name     string
		schema   string
		cfgName  string
		data     json.RawMessage
		valErr   error
		mockFunc func(m *repoMock.MockIRepo)
		ex       exRes
	}{
		{
			name:     "when invalid input - empty schema or name should return ErrInvalidInput",
			schema:   " ",
			cfgName:  "",
			data:     json.RawMessage(`{}`),
			valErr:   nil,
			mockFunc: func(m *repoMock.MockIRepo) {},
			ex:       exRes{res: model.RemoteConfig{}, err: ErrInvalidInput},
		},
		{
			name:     "when validator error should return ErrInvalidInput",
			schema:   "feature_toggle",
			cfgName:  "x",
			data:     json.RawMessage(`{"enabled":true}`),
			valErr:   assert.AnError, // any non-nil error triggers ErrInvalidInput mapping
			mockFunc: func(m *repoMock.MockIRepo) {},
			ex:       exRes{res: model.RemoteConfig{}, err: ErrInvalidInput},
		},
		{
			name:    "when already exists maps should return ErrAlreadyExists",
			schema:  "feature_toggle",
			cfgName: "x",
			data:    json.RawMessage(`{"enabled":true}`),
			valErr:  nil,
			mockFunc: func(m *repoMock.MockIRepo) {
				m.EXPECT().
					Create(gomock.Any(), "feature_toggle", "x", json.RawMessage(`{"enabled":true}`)).
					Return(model.RemoteConfig{}, repository.ErrAlreadyExists)
			},
			ex: exRes{res: model.RemoteConfig{}, err: ErrAlreadyExists},
		},
		{
			name:    "when repo not found maps should return ErrNotFound",
			schema:  "feature_toggle",
			cfgName: "y",
			data:    json.RawMessage(`{}`),
			valErr:  nil,
			mockFunc: func(m *repoMock.MockIRepo) {
				m.EXPECT().
					Create(gomock.Any(), "feature_toggle", "y", json.RawMessage(`{}`)).
					Return(model.RemoteConfig{}, repository.ErrNotFound)
			},
			ex: exRes{res: model.RemoteConfig{}, err: ErrNotFound},
		},
		{
			name:    "when success",
			schema:  "feature_toggle",
			cfgName: "qris",
			data:    json.RawMessage(`{"enabled":true}`),
			valErr:  nil,
			mockFunc: func(m *repoMock.MockIRepo) {
				m.EXPECT().
					Create(gomock.Any(), "feature_toggle", "qris", json.RawMessage(`{"enabled":true}`)).
					Return(model.RemoteConfig{
						Name:    "qris",
						Type:    "feature_toggle",
						Version: 1,
						Data:    json.RawMessage(`{"enabled":true}`),
					}, nil)
			},
			ex: exRes{
				res: model.RemoteConfig{
					Name:    "qris",
					Type:    "feature_toggle",
					Version: 1,
					Data:    json.RawMessage(`{"enabled":true}`),
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := repoMock.NewMockIRepo(ctrl)
			tc.mockFunc(repo)

			svc := service{
				repo:      repo,
				validator: stubValidator{err: tc.valErr},
			}

			got, err := svc.Create(context.Background(), tc.schema, tc.cfgName, tc.data)

			if tc.valErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidInput)
			} else {
				assert.Equal(t, tc.ex.err, err)
			}
			assert.Equal(t, tc.ex.res, got)
		})
	}
}
