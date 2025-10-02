package handler

import (
	"configuration-management-service/internal/remote_config/model"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func (h *handler) Update(c echo.Context) error {
	if !isJSON(c) {
		return writeErr(c, http.StatusUnsupportedMediaType, "content-type must be application/json", nil)
	}

	name := strings.TrimSpace(c.Param("name"))
	if name == "" {
		return writeErr(c, http.StatusBadRequest, "name is required", nil)
	}

	var req model.RemoteConfigUpdateRequest
	if err := c.Bind(&req); err != nil {
		return writeErr(c, http.StatusBadRequest, "invalid JSON", err.Error())
	}

	cfg, err := h.srv.Update(c.Request().Context(), name, req.Data)
	if err != nil {
		return h.writeServiceError(c, err)
	}
	return c.JSON(http.StatusOK, cfg)
}
