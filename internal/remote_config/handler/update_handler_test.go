package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"configuration-management-service/internal/remote_config/model"
	"configuration-management-service/internal/remote_config/service"
	srvMock "configuration-management-service/internal/remote_config/service/mocks"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestUpdate(t *testing.T) {
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
			name:     "unsupported media type",
			in:       input{ct: "text/plain", name: "qris", body: `{}`},
			mockFunc: func(m *srvMock.MockIService) {},
			ex: expected{
				code: http.StatusUnsupportedMediaType,
				json: `{"error":{"code":"Unsupported Media Type","message":"content-type must be application/json","details":null}}`,
			},
		},
		{
			name:     "missing name param",
			in:       input{ct: echo.MIMEApplicationJSON, name: "   ", body: `{"data":{}}`},
			mockFunc: func(m *srvMock.MockIService) {},
			ex: expected{
				code: http.StatusBadRequest,
				json: `{"error":{"code":"Bad Request","message":"name is required","details":null}}`,
			},
		},
		{
			name:     "invalid json",
			in:       input{ct: echo.MIMEApplicationJSON, name: "qris", body: `{`},
			mockFunc: func(m *srvMock.MockIService) {},
			ex: expected{
				code: http.StatusBadRequest,
				// Echo's Bind error includes details like "code=400, message=unexpected EOF, internal=unexpected EOF"
				json: `{"error":{"code":"Bad Request","message":"invalid JSON","details":"code=400, message=unexpected EOF, internal=unexpected EOF"}}`,
			},
		},
		{
			name: "service not found → 404",
			in:   input{ct: echo.MIMEApplicationJSON, name: "qris", body: `{"data":{"enabled":true}}`},
			mockFunc: func(m *srvMock.MockIService) {
				m.EXPECT().Update(gomock.Any(), "qris", json.RawMessage(`{"enabled":true}`)).
					Return(model.RemoteConfig{}, service.ErrNotFound)
			},
			ex: expected{
				code: http.StatusNotFound,
				json: `{"error":{"code":"Not Found","message":"not found","details":null}}`,
			},
		},
		{
			name: "success",
			in:   input{ct: echo.MIMEApplicationJSON, name: "qris", body: `{"data":{"enabled":true}}`},
			mockFunc: func(m *srvMock.MockIService) {
				m.EXPECT().Update(gomock.Any(), "qris", json.RawMessage(`{"enabled":true}`)).
					Return(model.RemoteConfig{Name: "qris", Type: "feature_toggle", Version: 2, Data: json.RawMessage(`{"enabled":true}`)}, nil)
			},
			ex: expected{
				code: http.StatusOK,
				// Include created_at because handler includes it in JSON
				json: `{"name":"qris","type":"feature_toggle","version":2,"data":{"enabled":true},"created_at":""}`,
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

			// Use placeholder segment to avoid raw-space panics; set path param separately
			req := httptest.NewRequest(http.MethodPut, "/configs/_placeholder", bytes.NewBufferString(tc.in.body))
			req.Header.Set(echo.HeaderContentType, tc.in.ct)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("name")
			c.SetParamValues(tc.in.name)

			_ = h.Update(c)

			res := rec.Result()
			defer res.Body.Close()
			b, _ := io.ReadAll(res.Body)

			assert.Equal(t, tc.ex.code, res.StatusCode, string(b))
			assert.JSONEq(t, tc.ex.json, string(b))
		})
	}
}
