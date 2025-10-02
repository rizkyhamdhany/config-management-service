package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

func (h *handler) Get(c echo.Context) error {
	name := strings.TrimSpace(c.Param("name"))
	if name == "" {
		return writeErr(c, http.StatusBadRequest, "name is required", nil)
	}

	var v *int
	if q := strings.TrimSpace(c.QueryParam("version")); q != "" {
		iv, err := strconv.Atoi(q)
		if err != nil || iv <= 0 {
			return writeErr(c, http.StatusBadRequest, "invalid version", "version must be a positive integer")
		}
		v = &iv
	}

	cfg, err := h.srv.Get(c.Request().Context(), name, v)
	if err != nil {
		return h.writeServiceError(c, err)
	}

	etag := weakETag(cfg.Name, cfg.Version)
	c.Response().Header().Set("ETag", etag)
	c.Response().Header().Set("Cache-Control", "no-cache")

	if inm := c.Request().Header.Get("If-None-Match"); inm != "" && inm == etag {
		return c.NoContent(http.StatusNotModified)
	}

	return c.JSON(http.StatusOK, cfg)
}
