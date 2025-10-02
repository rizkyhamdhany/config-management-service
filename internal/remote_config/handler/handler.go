package handler

import (
	"configuration-management-service/internal/remote_config/service"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

type IHandler interface {
	Create(c echo.Context) error
	Update(c echo.Context) error
	Get(c echo.Context) error
	List(c echo.Context) error
	Rollback(c echo.Context) error
}

type handler struct {
	srv service.IService
}

func NewHandler(srv service.IService) IHandler {
	return &handler{srv: srv}
}

// --- helpers ---
func (h *handler) writeServiceError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, service.ErrNotFound):
		return writeErr(c, http.StatusNotFound, err.Error(), nil)
	case errors.Is(err, service.ErrAlreadyExists):
		return writeErr(c, http.StatusConflict, err.Error(), nil)
	case errors.Is(err, service.ErrInvalidInput):
		return writeErr(c, http.StatusBadRequest, "invalid input", err.Error())
	default:
		return writeErr(c, http.StatusInternalServerError, "internal error", nil)
	}
}

func writeErr(c echo.Context, code int, msg string, details any) error {
	return c.JSON(code, map[string]any{
		"error": map[string]any{
			"code":    http.StatusText(code),
			"message": msg,
			"details": details,
		},
	})
}

func isJSON(c echo.Context) bool {
	ct := c.Request().Header.Get(echo.HeaderContentType)
	return strings.HasPrefix(ct, echo.MIMEApplicationJSON)
}

func weakETag(name string, version int) string {
	h := sha1.Sum([]byte(name + ":" + strconv.Itoa(version)))
	return `W/"` + hex.EncodeToString(h[:8]) + `"`
}
