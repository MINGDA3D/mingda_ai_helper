package handlers

import (
	"github.com/gin-gonic/gin"
	"mingda_ai_helper/models"
	"mingda_ai_helper/services"
	"net/http"
)

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func MachineRegister(db *services.DBService, log *services.LogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var machine models.MachineInfo
		if err := c.ShouldBindJSON(&machine); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// TODO: 生成认证token
		// TODO: 保存机器信息

		c.JSON(http.StatusOK, gin.H{"auth_token": machine.AuthToken})
	}
}

func TokenRefresh(db *services.DBService, log *services.LogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现token刷新逻辑
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func SettingsSync(db *services.DBService, log *services.LogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var settings models.UserSettings
		if err := c.ShouldBindJSON(&settings); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := db.SaveUserSettings(&settings); err != nil {
			log.Error("Failed to save user settings")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func Predict(ai services.AIService, db *services.DBService, log *services.LogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ImageURL    string `json:"image_url"`
			TaskID      string `json:"task_id"`
			CallbackURL string `json:"callback_url"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := ai.Predict(c.Request.Context(), req.ImageURL, req.TaskID)
		if err != nil {
			log.Error("Failed to make prediction")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err := db.SavePredictionResult(result); err != nil {
			log.Error("Failed to save prediction result")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func AICallback(db *services.DBService, log *services.LogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var result models.PredictionResult
		if err := c.ShouldBindJSON(&result); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := db.SavePredictionResult(&result); err != nil {
			log.Error("Failed to save prediction result from callback")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func PrinterPause(log *services.LogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现打印机暂停逻辑
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
} 