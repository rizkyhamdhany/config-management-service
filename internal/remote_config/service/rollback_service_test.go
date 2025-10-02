package service

import (
	"configuration-management-service/internal/remote_config/repository"
	"context"
	"testing"

	"configuration-management-service/internal/remote_config/model"
	repoMock "configuration-management-service/internal/remote_config/repository/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_service_Rollback(t *testing.T) {
	type exRes struct {
		res model.RemoteConfig
		err error
	}

	cases := []struct {
		name     string
		cfgName  string
		version  int
		mockFunc func(m *repoMock.MockIRepo)
		ex       exRes
	}{
		{
			name:     "when invalid input empty name or bad version should return ErrInvalidInput",
			cfgName:  " ",
			version:  0,
			mockFunc: func(m *repoMock.MockIRepo) {},
			ex:       exRes{res: model.RemoteConfig{}, err: ErrInvalidInput},
		},
		{
			name:    "when target not found should return ErrNotFound",
			cfgName: "key",
			version: 2,
			mockFunc: func(m *repoMock.MockIRepo) {
				m.EXPECT().ByVersion(gomock.Any(), "key", 2).Return(model.RemoteConfig{}, repository.ErrNotFound)
			},
			ex: exRes{res: model.RemoteConfig{}, err: ErrNotFound},
		},
		{
			name:    "when append not found ErrNotFound should return",
			cfgName: "key",
			version: 2,
			mockFunc: func(m *repoMock.MockIRepo) {
				m.EXPECT().ByVersion(gomock.Any(), "key", 2).Return(model.RemoteConfig{Name: "key", Type: "feature_toggle", Version: 2, Data: []byte(`{"a":1}`)}, nil)
				m.EXPECT().Append(gomock.Any(), "key", []byte(`{"a":1}`)).Return(model.RemoteConfig{}, repository.ErrNotFound)
			},
			ex: exRes{res: model.RemoteConfig{}, err: ErrNotFound},
		},
		{
			name:    "when success",
			cfgName: "key",
			version: 2,
			mockFunc: func(m *repoMock.MockIRepo) {
				m.EXPECT().ByVersion(gomock.Any(), "key", 2).Return(model.RemoteConfig{Name: "key", Type: "feature_toggle", Version: 2, Data: []byte(`{"a":1}`)}, nil)
				m.EXPECT().Append(gomock.Any(), "key", []byte(`{"a":1}`)).Return(model.RemoteConfig{Name: "key", Type: "feature_toggle", Version: 3, Data: []byte(`{"a":1}`)}, nil)
			},
			ex: exRes{res: model.RemoteConfig{Name: "key", Type: "feature_toggle", Version: 3, Data: []byte(`{"a":1}`)}, err: nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := repoMock.NewMockIRepo(ctrl)
			tc.mockFunc(repo)
			svc := service{repo: repo}

			got, err := svc.Rollback(context.Background(), tc.cfgName, tc.version)
			assert.Equal(t, tc.ex.err, err)
			assert.Equal(t, tc.ex.res, got)
		})
	}
}
