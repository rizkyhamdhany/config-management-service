package handler

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func (h *handler) List(c echo.Context) error {
	name := strings.TrimSpace(c.Param("name"))
	if name == "" {
		return writeErr(c, http.StatusBadRequest, "name is required", nil)
	}

	res, err := h.srv.ListVersions(c.Request().Context(), name)
	if err != nil {
		return h.writeServiceError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"versions": res})
}
