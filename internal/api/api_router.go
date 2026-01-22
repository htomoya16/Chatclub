package api

import (
	"github.com/labstack/echo/v4"
)

func SetupRoutes(
	// 引数
	e *echo.Echo,
	healthHandler *HealthHandler,
	discord *WhitelistHandler) {

	api := e.Group("/api")

	// ヘルスチェック
	api.GET("/livez", healthHandler.Livez)
	api.GET("/readyz", healthHandler.Readyz)
	api.GET("/healthz", healthHandler.Healthz)

	// Discord管理用
	g := e.Group("/discord")
	g.POST("/whitelist/add", discord.Add)
	g.POST("/whitelist/remove", discord.Remove)
}
