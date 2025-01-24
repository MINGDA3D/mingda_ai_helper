package handlers

import (
	"github.com/gin-gonic/gin"
	"mingda_ai_helper/services"
)

func SetupRouter(
	aiService services.AIService,
	dbService *services.DBService,
	logService *services.LogService,
) *gin.Engine {
	router := gin.Default()

	// 健康检查
	router.GET("/api/v1/ai/health", HealthCheck)

	// 设备管理
	router.POST("/api/v1/machine/register", MachineRegister(dbService, logService))
	router.POST("/api/v1/token/refresh", TokenRefresh(dbService, logService))

	// 用户设置
	router.POST("/api/v1/settings/sync", SettingsSync(dbService, logService))

	// AI预测
	router.POST("/api/v1/predict", Predict(aiService, dbService, logService))
	router.POST("/api/v1/ai/callback", AICallback(dbService, logService))

	// 打印机控制
	router.POST("/api/v1/printer/pause", PrinterPause(logService))

	return router
} 