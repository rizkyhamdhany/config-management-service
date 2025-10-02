package service

import (
	"configuration-management-service/internal/remote_config/repository"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"configuration-management-service/internal/remote_config/model"
	repoMock "configuration-management-service/internal/remote_config/repository/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_service_Update(t *testing.T) {
	type exRes struct {
		res model.RemoteConfig
		err error
	}

	cases := []struct {
		name     string
		cfgName  string
		data     json.RawMessage
		valErr   error
		mockFunc func(m *repoMock.MockIRepo)
		ex       exRes
	}{
		{
			name:     "when invalid input - empty name should return ErrInvalidInput",
			cfgName:  "   ",
			data:     json.RawMessage(`{}`),
			valErr:   nil,
			mockFunc: func(m *repoMock.MockIRepo) {},
			ex:       exRes{res: model.RemoteConfig{}, err: ErrInvalidInput},
		},
		{
			name:    "when latest not found maps should return ErrNotFound",
			cfgName: "key",
			data:    json.RawMessage(`{}`),
			valErr:  nil,
			mockFunc: func(m *repoMock.MockIRepo) {
				m.EXPECT().Latest(gomock.Any(), "key").Return(model.RemoteConfig{}, repository.ErrNotFound)
			},
			ex: exRes{res: model.RemoteConfig{}, err: ErrNotFound},
		},
		{
			name:    "when validator returns error should return ErrInvalidInput",
			cfgName: "key",
			data:    json.RawMessage(`{"bad":true}`),
			valErr:  errors.New("invalid schema"),
			mockFunc: func(m *repoMock.MockIRepo) {
				m.EXPECT().Latest(gomock.Any(), "key").Return(model.RemoteConfig{Name: "key", Type: "feature_toggle", Version: 2}, nil)
			},
			ex: exRes{res: model.RemoteConfig{}, err: ErrInvalidInput},
		},
		{
			name:    "when append not found maps should return ErrNotFound",
			cfgName: "key",
			data:    json.RawMessage(`{"ok":true}`),
			valErr:  nil,
			mockFunc: func(m *repoMock.MockIRepo) {
				m.EXPECT().Latest(gomock.Any(), "key").Return(model.RemoteConfig{Name: "key", Type: "feature_toggle", Version: 2}, nil)
				m.EXPECT().Append(gomock.Any(), "key", json.RawMessage(`{"ok":true}`)).Return(model.RemoteConfig{}, repository.ErrNotFound)
			},
			ex: exRes{res: model.RemoteConfig{}, err: ErrNotFound},
		},
		{
			name:    "when success",
			cfgName: "key",
			data:    json.RawMessage(`{"ok":true}`),
			valErr:  nil,
			mockFunc: func(m *repoMock.MockIRepo) {
				m.EXPECT().Latest(gomock.Any(), "key").Return(model.RemoteConfig{Name: "key", Type: "feature_toggle", Version: 2}, nil)
				m.EXPECT().Append(gomock.Any(), "key", json.RawMessage(`{"ok":true}`)).Return(model.RemoteConfig{Name: "key", Type: "feature_toggle", Version: 3, Data: json.RawMessage(`{"ok":true}`)}, nil)
			},
			ex: exRes{res: model.RemoteConfig{Name: "key", Type: "feature_toggle", Version: 3, Data: json.RawMessage(`{"ok":true}`)}, err: nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := repoMock.NewMockIRepo(ctrl)
			tc.mockFunc(repo)

			svc := service{repo: repo, validator: stubValidator{err: tc.valErr}}

			got, err := svc.Update(context.Background(), tc.cfgName, tc.data)
			if tc.valErr != nil && errors.Is(err, ErrInvalidInput) {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tc.ex.err, err)
			}
			assert.Equal(t, tc.ex.res, got)
		})
	}
}
