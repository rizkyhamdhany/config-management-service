package handler

import (
	"configuration-management-service/internal/remote_config/model"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func (h *handler) Rollback(c echo.Context) error {
	if !isJSON(c) {
		return writeErr(c, http.StatusUnsupportedMediaType, "content-type must be application/json", nil)
	}

	name := strings.TrimSpace(c.Param("name"))
	if name == "" {
		return writeErr(c, http.StatusBadRequest, "name is required", nil)
	}

	var req model.RemoteConfigRollbackRequest
	if err := c.Bind(&req); err != nil {
		return writeErr(c, http.StatusBadRequest, "invalid JSON", err.Error())
	}
	if req.Version <= 0 {
		return writeErr(c, http.StatusBadRequest, "invalid version", "version must be a positive integer")
	}

	cfg, err := h.srv.Rollback(c.Request().Context(), name, req.Version)
	if err != nil {
		return h.writeServiceError(c, err)
	}
	return c.JSON(http.StatusOK, cfg)
}
