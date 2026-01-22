package api

import (
	"net/http"

	"backend/internal/service"

	"github.com/labstack/echo/v4"
)

type WhitelistHandler struct {
	svc service.WhitelistService
}

func NewWhitelistHandler(s service.WhitelistService) *WhitelistHandler {
	return &WhitelistHandler{svc: s}
}

func (h *WhitelistHandler) Add(c echo.Context) error {
	type req struct {
		UserID string `json:"user_id"`
		Note   string `json:"note"`
	}
	var r req

	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if r.UserID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "userID required")
	}

	// Discord用のホワイトリストなので platform を "discord" で固定
	if err := h.svc.Add(c.Request().Context(), "discord", r.UserID, r.Note); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusCreated)
}

func (h *WhitelistHandler) Remove(c echo.Context) error {
	type req struct {
		UserID string `json:"user_id"`
	}
	var r req

	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if r.UserID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "userID required")
	}

	if err := h.svc.Remove(c.Request().Context(), "discord", r.UserID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}
