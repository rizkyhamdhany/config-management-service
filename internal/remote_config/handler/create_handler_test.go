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

func TestCreate(t *testing.T) {
	type input struct {
		ct   string
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
			name:     "when unsupported media type should status code 415",
			in:       input{ct: "text/plain", body: `{}`},
			mockFunc: func(m *srvMock.MockIService) {},
			ex: expected{
				code: http.StatusUnsupportedMediaType,
				json: `{"error":{"code":"Unsupported Media Type","message":"content-type must be application/json","details":null}}`,
			},
		},
		{
			name:     "when invalid json should status code 400 and error message",
			in:       input{ct: echo.MIMEApplicationJSON, body: `{`},
			mockFunc: func(m *srvMock.MockIService) {},
			ex: expected{
				code: http.StatusBadRequest,
				json: `{"error":{"code":"Bad Request","message":"invalid JSON","details":"code=400, message=unexpected EOF, internal=unexpected EOF"}}`,
			},
		},
		{
			name:     "when missing type/name should status code 400 and error message",
			in:       input{ct: echo.MIMEApplicationJSON, body: `{"type":"","name":"   ","data":{}}`},
			mockFunc: func(m *srvMock.MockIService) {},
			ex: expected{
				code: http.StatusBadRequest,
				json: `{"error":{"code":"Bad Request","message":"type and name are required","details":null}}`,
			},
		},
		{
			name: "when config is already exists should status code 409 and error message",
			in:   input{ct: echo.MIMEApplicationJSON, body: `{"type":"feature_toggle","name":"qris","data":{"enabled":true}}`},
			mockFunc: func(m *srvMock.MockIService) {
				m.EXPECT().Create(gomock.Any(), "feature_toggle", "qris", json.RawMessage(`{"enabled":true}`)).
					Return(model.RemoteConfig{}, service.ErrAlreadyExists)
			},
			ex: expected{
				code: http.StatusConflict,
				json: `{"error":{"code":"Conflict","message":"already exists","details":null}}`,
			},
		},
		{
			name: "when success",
			in:   input{ct: echo.MIMEApplicationJSON, body: `{"type":"feature_toggle","name":"qris","data":{"enabled":true}}`},
			mockFunc: func(m *srvMock.MockIService) {
				m.EXPECT().Create(gomock.Any(), "feature_toggle", "qris", json.RawMessage(`{"enabled":true}`)).
					Return(model.RemoteConfig{Name: "qris", Type: "feature_toggle", Version: 1, Data: json.RawMessage(`{"enabled":true}`)}, nil)
			},
			ex: expected{
				code: http.StatusCreated,
				json: `{"name":"qris","type":"feature_toggle","version":1,"data":{"enabled":true},"created_at":""}`,
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

			req := httptest.NewRequest(http.MethodPost, "/configs", bytes.NewBufferString(tc.in.body))
			req.Header.Set(echo.HeaderContentType, tc.in.ct)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// call handler directly
			err := h.Create(c)
			if err != nil {
				_ = err // echo returns error to be handled by server; direct call returns after JSON sent
			}

			res := rec.Result()
			defer res.Body.Close()
			b, _ := io.ReadAll(res.Body)

			assert.Equal(t, tc.ex.code, res.StatusCode, string(b))
			assert.JSONEq(t, tc.ex.json, string(b))
		})
	}
}
