package handler

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"configuration-management-service/internal/remote_config/model"
	srvMock "configuration-management-service/internal/remote_config/repository/mocks"
	"configuration-management-service/internal/remote_config/service"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestRollback(t *testing.T) {
	type input struct {
		ct   string
		name string
		body string
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
			in:   input{ct: echo.MIMEApplicationJSON, name: " ", body: `{"version":2}`},
			mockFunc: func(m *srvMock.MockIService) {
				// no calls expected
			},
			ex: expected{
				code: http.StatusBadRequest,
				json: `{"error":{"code":"Bad Request","message":"name is required","details":null}}`,
			},
		},
		{
			name:     "when invalid json should status code 400 and error message",
			in:       input{ct: echo.MIMEApplicationJSON, name: "qris", body: `{`},
			mockFunc: func(m *srvMock.MockIService) {},
			ex: expected{
				code: http.StatusBadRequest,
				// Echo's Bind includes a verbose details string:
				json: `{"error":{"code":"Bad Request","message":"invalid JSON","details":"code=400, message=unexpected EOF, internal=unexpected EOF"}}`,
			},
		},
		{
			name:     "when invalid version should status code 400 and error message",
			in:       input{ct: echo.MIMEApplicationJSON, name: "qris", body: `{"version":0}`},
			mockFunc: func(m *srvMock.MockIService) {},
			ex: expected{
				code: http.StatusBadRequest,
				json: `{"error":{"code":"Bad Request","message":"invalid version","details":"version must be a positive integer"}}`,
			},
		},
		{
			name: "when service not found should status code 404 and error message",
			in:   input{ct: echo.MIMEApplicationJSON, name: "qris", body: `{"version":2}`},
			mockFunc: func(m *srvMock.MockIService) {
				m.EXPECT().Rollback(gomock.Any(), "qris", 2).
					Return(model.RemoteConfig{}, service.ErrNotFound)
			},
			ex: expected{
				code: http.StatusNotFound,
				json: `{"error":{"code":"Not Found","message":"not found","details":null}}`,
			},
		},
		{
			name: "when success",
			in:   input{ct: echo.MIMEApplicationJSON, name: "qris", body: `{"version":2}`},
			mockFunc: func(m *srvMock.MockIService) {
				m.EXPECT().Rollback(gomock.Any(), "qris", 2).
					Return(model.RemoteConfig{Name: "qris", Type: "feature_toggle", Version: 3, Data: []byte(`{"enabled":true}`)}, nil)
			},
			ex: expected{
				code: http.StatusOK,
				// Include created_at because your handler includes it in JSON
				json: `{"name":"qris","type":"feature_toggle","version":3,"data":{"enabled":true},"created_at":""}`,
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

			// Use placeholder path and inject :name via params (avoids raw space panic)
			req := httptest.NewRequest(http.MethodPost, "/configs/_placeholder/rollback", bytes.NewBufferString(tc.in.body))
			req.Header.Set(echo.HeaderContentType, tc.in.ct)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("name")
			c.SetParamValues(tc.in.name)

			_ = h.Rollback(c)

			res := rec.Result()
			defer res.Body.Close()
			b, _ := io.ReadAll(res.Body)

			assert.Equal(t, tc.ex.code, res.StatusCode, string(b))
			assert.JSONEq(t, tc.ex.json, string(b))
		})
	}
}
