package handlers

import (
	"github.com/gin-gonic/gin"
	"mingda_ai_helper/handlers/middleware"
	"mingda_ai_helper/services"
)

func SetupRouter(
	aiService services.AIService,
	dbService *services.DBService,
	logService *services.LogService,
) *gin.Engine {
	router := gin.New() // 使用gin.New()而不是Default()以自定义中间件

	// 添加全局中间件
	router.Use(gin.Recovery())
	router.Use(middleware.ErrorHandler(logService))
	router.Use(middleware.RequestLogger(logService))

	// API路由组
	v1 := router.Group("/api/v1")
	{
		// 健康检查
		v1.GET("/ai/health", HealthCheck)

		// 设备管理
		v1.POST("/machine/register", MachineRegister(dbService, logService))
		v1.POST("/token/refresh", TokenRefresh(dbService, logService))

		// 用户设置
		v1.POST("/settings/sync", SettingsSync(dbService, logService))

		// AI预测
		v1.POST("/predict", Predict(aiService, dbService, logService))
		v1.POST("/ai/callback", AICallback(dbService, logService))

		// 打印机控制
		v1.POST("/printer/pause", PrinterPause(logService))
	}

	return router
} 