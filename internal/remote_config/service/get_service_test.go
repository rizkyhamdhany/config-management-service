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

func Test_service_Get(t *testing.T) {
	type exRes struct {
		res model.RemoteConfig
		err error
	}

	ver := 2
	cases := []struct {
		name     string
		cfgName  string
		version  *int
		mockFunc func(m *repoMock.MockIRepo)
		ex       exRes
	}{
		{
			name:     "when invalid input - empty name should return ErrInvalidInput",
			cfgName:  " ",
			version:  nil,
			mockFunc: func(m *repoMock.MockIRepo) {},
			ex:       exRes{res: model.RemoteConfig{}, err: ErrInvalidInput},
		},
		{
			name:    "when version nil should return ErrNotFound",
			cfgName: "key",
			version: nil,
			mockFunc: func(m *repoMock.MockIRepo) {
				m.EXPECT().Latest(gomock.Any(), "key").Return(model.RemoteConfig{}, repository.ErrNotFound)
			},
			ex: exRes{res: model.RemoteConfig{}, err: ErrNotFound},
		},
		{
			name:    "when version nil should return latest",
			cfgName: "key",
			version: nil,
			mockFunc: func(m *repoMock.MockIRepo) {
				m.EXPECT().Latest(gomock.Any(), "key").Return(model.RemoteConfig{Name: "key", Type: "feature_toggle", Version: 5}, nil)
			},
			ex: exRes{res: model.RemoteConfig{Name: "key", Type: "feature_toggle", Version: 5}, err: nil},
		},
		{
			name:    "when version provided should ByVersion not found",
			cfgName: "key",
			version: &ver,
			mockFunc: func(m *repoMock.MockIRepo) {
				m.EXPECT().ByVersion(gomock.Any(), "key", ver).Return(model.RemoteConfig{}, repository.ErrNotFound)
			},
			ex: exRes{res: model.RemoteConfig{}, err: ErrNotFound},
		},
		{
			name:    "when version provided should ByVersion success",
			cfgName: "key",
			version: &ver,
			mockFunc: func(m *repoMock.MockIRepo) {
				m.EXPECT().ByVersion(gomock.Any(), "key", ver).Return(model.RemoteConfig{Name: "key", Type: "feature_toggle", Version: ver}, nil)
			},
			ex: exRes{res: model.RemoteConfig{Name: "key", Type: "feature_toggle", Version: 2}, err: nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := repoMock.NewMockIRepo(ctrl)
			tc.mockFunc(repo)
			svc := service{repo: repo}

			got, err := svc.Get(context.Background(), tc.cfgName, tc.version)
			assert.Equal(t, tc.ex.err, err)
			assert.Equal(t, tc.ex.res, got)
		})
	}
}
