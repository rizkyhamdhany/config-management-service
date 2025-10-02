package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"configuration-management-service/internal/remote_config/model"
	srvMock "configuration-management-service/internal/remote_config/repository/mocks"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	type input struct {
		name    string // path param
		version string // query param, empty means nil
		ifNone  string // If-None-Match header
	}
	type expected struct {
		code int
		json string
		etag string
	}

	cases := []struct {
		name     string
		mockFunc func(m *srvMock.MockIService)
		in       input
		ex       expected
	}{
		{
			name: "when missing config name should status code 400",
			// IMPORTANT: use a placeholder in URL; inject spaces via SetParamValues
			in: input{name: "   "},
			mockFunc: func(m *srvMock.MockIService) {
				// no calls expected; handler should reject before calling service
			},
			ex: expected{
				code: http.StatusBadRequest,
				json: `{"error":{"code":"Bad Request","message":"name is required","details":null}}`,
			},
		},
		{
			name: "when latest success, no If-None-Match",
			in:   input{name: "qris"},
			mockFunc: func(m *srvMock.MockIService) {
				m.EXPECT().Get(gomock.Any(), "qris", (*int)(nil)).
					Return(model.RemoteConfig{
						Name:    "qris",
						Type:    "feature_toggle",
						Version: 5,
						Data:    []byte(`{"enabled":true}`),
					}, nil)
			},
			ex: expected{
				code: http.StatusOK,
				json: `{"name":"qris","type":"feature_toggle","version":5,"data":{"enabled":true},"created_at":""}`,
				etag: weakETag("qris", 5),
			},
		},
		{
			name: "when If-None-Match matches should 304",
			in:   input{name: "qris", ifNone: weakETag("qris", 5)},
			mockFunc: func(m *srvMock.MockIService) {
				m.EXPECT().Get(gomock.Any(), "qris", (*int)(nil)).
					Return(model.RemoteConfig{
						Name:    "qris",
						Type:    "feature_toggle",
						Version: 5,
						Data:    []byte(`{"enabled":true}`),
					}, nil)
			},
			ex: expected{
				code: http.StatusNotModified,
				json: ``,
				etag: weakETag("qris", 5),
			},
		},
		{
			name: "when by version success",
			in:   input{name: "qris", version: "2"},
			mockFunc: func(m *srvMock.MockIService) {
				v := 2
				m.EXPECT().Get(gomock.Any(), "qris", &v).
					Return(model.RemoteConfig{
						Name:    "qris",
						Type:    "feature_toggle",
						Version: 2,
						Data:    []byte(`{"enabled":true}`),
					}, nil)
			},
			ex: expected{
				code: http.StatusOK,
				json: `{"name":"qris","type":"feature_toggle","version":2,"data":{"enabled":true},"created_at":""}`,
				etag: weakETag("qris", 2),
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

			// Always use a placeholder path segment; inject param separately.
			url := "/configs/_placeholder"
			if tc.in.version != "" {
				url += "?version=" + tc.in.version
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)
			if tc.in.ifNone != "" {
				req.Header.Set("If-None-Match", tc.in.ifNone)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("name")
			c.SetParamValues(tc.in.name) // may be "   " to trigger trim → empty

			_ = h.Get(c)

			res := rec.Result()
			defer res.Body.Close()
			b, _ := io.ReadAll(res.Body)

			assert.Equal(t, tc.ex.code, res.StatusCode, string(b))
			if tc.ex.json != "" {
				assert.JSONEq(t, tc.ex.json, string(b))
			} else {
				assert.Equal(t, "", string(b))
			}
			if tc.ex.etag != "" {
				assert.Equal(t, tc.ex.etag, res.Header.Get("ETag"))
			}
		})
	}
}
