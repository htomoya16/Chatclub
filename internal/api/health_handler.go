package api

import (
	"backend/internal/service"
	"net/http"

	"github.com/labstack/echo/v4"
)

type HealthHandler struct {
	healthSvc service.HealthService
}

func NewHealthHandler(healthSvc service.HealthService) *HealthHandler {
	return &HealthHandler{healthSvc: healthSvc}
}

// プロセスが動いているか
func (h *HealthHandler) Livez(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

// 依存到達確認
func (h *HealthHandler) Readyz(c echo.Context) error {
	// Not Readyのとき
	if !h.healthSvc.Ready(c.Request().Context()) {
		// 503 Service Unavailable
		return c.String(http.StatusServiceUnavailable, "not ready")
	}
	// Readyのとき
	// 200 OK
	return c.NoContent(http.StatusOK)
}

// 人間/監視向けの総合診断
func (h *HealthHandler) Healthz(c echo.Context) error {
	rep := h.healthSvc.Report(c.Request().Context())
	// 200 OK
	code := http.StatusOK
	if !rep.Ready {
		// 503 Service Unavailable
		code = http.StatusServiceUnavailable
	}
	return c.JSON(code, rep)
}
