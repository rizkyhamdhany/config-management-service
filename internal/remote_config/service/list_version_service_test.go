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

func Test_service_ListVersions(t *testing.T) {
	type exRes struct {
		res []model.RemoteConfig
		err error
	}

	cases := []struct {
		name     string
		cfgName  string
		mockFunc func(m *repoMock.MockIRepo)
		ex       exRes
	}{
		{
			name:     "when invalid input - empty name should return ErrInvalidInput",
			cfgName:  " ",
			mockFunc: func(m *repoMock.MockIRepo) {},
			ex:       exRes{res: nil, err: ErrInvalidInput},
		},
		{
			name:    "when repo not found maps to ErrNotFound should return ErrNotFound",
			cfgName: "key",
			mockFunc: func(m *repoMock.MockIRepo) {
				m.EXPECT().List(gomock.Any(), "key").Return(nil, repository.ErrNotFound)
			},
			ex: exRes{res: nil, err: ErrNotFound},
		},
		{
			name:    "when success",
			cfgName: "key",
			mockFunc: func(m *repoMock.MockIRepo) {
				m.EXPECT().List(gomock.Any(), "key").Return([]model.RemoteConfig{{Name: "key", Type: "feature_toggle", Version: 1}}, nil)
			},
			ex: exRes{res: []model.RemoteConfig{{Name: "key", Type: "feature_toggle", Version: 1}}, err: nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := repoMock.NewMockIRepo(ctrl)
			tc.mockFunc(repo)
			svc := service{repo: repo}

			got, err := svc.ListVersions(context.Background(), tc.cfgName)
			assert.Equal(t, tc.ex.err, err)
			assert.Equal(t, tc.ex.res, got)
		})
	}
}
