package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"configuration-management-service/internal/remote_config/model"
	srvMock "configuration-management-service/internal/remote_config/service/mocks"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	type input struct {
		name string
	}
	type expected struct {
		code int
		json string
	}

	cases := []struct {
		name     string
		mockFunc func(m *srvMock.MockIService)
		in       input
		ex       expected
	}{
		{
			name: "when missing config name should status code 400",
			in:   input{name: " "},
			mockFunc: func(m *srvMock.MockIService) {
				// no service calls expected
			},
			ex: expected{
				code: http.StatusBadRequest,
				json: `{"error":{"code":"Bad Request","message":"name is required","details":null}}`,
			},
		},
		{
			name: "when success",
			in:   input{name: "qris"},
			mockFunc: func(m *srvMock.MockIService) {
				m.EXPECT().ListVersions(gomock.Any(), "qris").
					Return([]model.RemoteConfig{{Name: "qris", Type: "feature_toggle", Version: 1}}, nil)
			},
			ex: expected{
				code: http.StatusOK,
				json: `{"versions":[{"name":"qris","type":"feature_toggle","version":1,"data":null,"created_at":""}]}`,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			srv := srvMock.NewMockIService(ctrl)
			tc.mockFunc(srv)
			h := NewHandler(srv)

			req := httptest.NewRequest(http.MethodGet, "/configs/_placeholder/versions", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("name")
			c.SetParamValues(tc.in.name)

			_ = h.List(c)

			res := rec.Result()
			defer res.Body.Close()
			b, _ := io.ReadAll(res.Body)

			assert.Equal(t, tc.ex.code, res.StatusCode, string(b))
			assert.JSONEq(t, tc.ex.json, string(b))
		})
	}
}
