package handler

import (
	"configuration-management-service/internal/remote_config/model"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func (h *handler) Create(c echo.Context) error {
	if !isJSON(c) {
		return writeErr(c, http.StatusUnsupportedMediaType, "content-type must be application/json", nil)
	}

	var req model.RemoteConfigCreateRequest
	if err := c.Bind(&req); err != nil {
		return writeErr(c, http.StatusBadRequest, "invalid JSON", err.Error())
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Type = strings.TrimSpace(req.Type)
	if req.Type == "" || req.Name == "" {
		return writeErr(c, http.StatusBadRequest, "type and name are required", nil)
	}

	cfg, err := h.srv.Create(c.Request().Context(), req.Type, req.Name, req.Data)
	if err != nil {
		return h.writeServiceError(c, err)
	}

	return c.JSON(http.StatusCreated, cfg)
}
