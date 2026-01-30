package api

import (
	"github.com/labstack/echo/v4"
)

func SetupRoutes(
	// 引数
	e *echo.Echo,
	healthHandler *HealthHandler,
	sf6AssetHandler *SF6AssetHandler) {

	api := e.Group("/api")

	// ヘルスチェック
	api.GET("/livez", healthHandler.Livez)
	api.GET("/readyz", healthHandler.Readyz)
	api.GET("/healthz", healthHandler.Healthz)
	api.GET("/sf6/character/:tool", sf6AssetHandler.CharacterImage)
}
