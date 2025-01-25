package handlers

import (
	"github.com/gin-gonic/gin"
	"mingda_ai_helper/models"
	"mingda_ai_helper/services"
	"mingda_ai_helper/utils"
)

// HealthCheck 健康检查
func HealthCheck(c *gin.Context) {
	Success(c, gin.H{"status": "ok"})
}

// MachineRegister 设备注册
func MachineRegister(db *services.DBService, log *services.LogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			MachineModel string `json:"machine_model" binding:"required"`
			MachineSN    string `json:"machine_sn" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			ValidationError(c, "无效的请求参数")
			return
		}

		// 生成认证token
		token, err := utils.GenerateToken(req.MachineSN)
		if err != nil {
			ServerError(c, "生成token失败")
			return
		}

		// 保存机器信息
		machine := &models.MachineInfo{
			MachineSN:    req.MachineSN,
			MachineModel: req.MachineModel,
			AuthToken:    token,
		}

		if err := db.SaveMachineInfo(machine); err != nil {
			log.Error("保存机器信息失败", "error", err)
			ServerError(c, "保存机器信息失败")
			return
		}

		Success(c, gin.H{"auth_token": token})
	}
}

// TokenRefresh Token刷新
func TokenRefresh(db *services.DBService, log *services.LogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			MachineSN string `json:"machine_sn" binding:"required"`
			OldToken  string `json:"old_token" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			ValidationError(c, "无效的请求参数")
			return
		}

		// 验证旧token
		if !utils.ValidateToken(req.MachineSN, req.OldToken) {
			UnauthorizedError(c)
			return
		}

		// 生成新token
		newToken, err := utils.GenerateToken(req.MachineSN)
		if err != nil {
			ServerError(c, "生成新token失败")
			return
		}

		// 更新数据库中的token
		if err := db.UpdateMachineToken(req.MachineSN, newToken); err != nil {
			log.Error("更新token失败", "error", err)
			ServerError(c, "更新token失败")
			return
		}

		Success(c, gin.H{"new_token": newToken})
	}
}

// SettingsSync 同步用户设置
func SettingsSync(db *services.DBService, log *services.LogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var settings models.UserSettings
		if err := c.ShouldBindJSON(&settings); err != nil {
			ValidationError(c, "无效的设置参数")
			return
		}

		// 验证参数
		if settings.ConfidenceThreshold < 0 || settings.ConfidenceThreshold > 100 {
			ValidationError(c, "置信度阈值必须在0-100之间")
			return
		}

		if err := db.SaveUserSettings(&settings); err != nil {
			log.Error("保存用户设置失败", "error", err)
			ServerError(c, "保存用户设置失败")
			return
		}

		Success(c, gin.H{"status": "ok"})
	}
}

// Predict AI预测请求
func Predict(ai services.AIService, db *services.DBService, log *services.LogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ImageURL    string `json:"image_url" binding:"required,url"`
			TaskID      string `json:"task_id" binding:"required"`
			CallbackURL string `json:"callback_url" binding:"required,url"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			ValidationError(c, "无效的请求参数")
			return
		}

		// 创建预测任务
		result := &models.PredictionResult{
			TaskID:           req.TaskID,
			PredictionStatus: models.StatusPending,
		}

		// 保存初始状态
		if err := db.SavePredictionResult(result); err != nil {
			log.Error("保存预测任务失败", "error", err)
			ServerError(c, "保存预测任务失败")
			return
		}

		// 异步执行预测
		go func() {
			result, err := ai.Predict(c.Request.Context(), req.ImageURL, req.TaskID)
			if err != nil {
				log.Error("预测失败", "error", err)
				return
			}

			if err := db.SavePredictionResult(result); err != nil {
				log.Error("保存预测结果失败", "error", err)
			}
		}()

		Success(c, gin.H{"task_id": req.TaskID})
	}
}

// AICallback AI回调处理
func AICallback(db *services.DBService, log *services.LogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var result models.PredictionResult
		if err := c.ShouldBindJSON(&result); err != nil {
			ValidationError(c, "无效的回调参数")
			return
		}

		if result.TaskID == "" {
			ValidationError(c, "任务ID不能为空")
			return
		}

		if err := db.SavePredictionResult(&result); err != nil {
			log.Error("保存预测结果失败", "error", err)
			ServerError(c, "保存预测结果失败")
			return
		}

		// 检查是否需要暂停打印
		settings, err := db.GetUserSettings()
		if err != nil {
			log.Error("获取用户设置失败", "error", err)
			ServerError(c, "获取用户设置失败")
			return
		}

		if settings.PauseOnThreshold && result.Confidence >= float64(settings.ConfidenceThreshold) {
			// TODO: 调用打印机暂停接口
			log.Info("触发打印暂停", "task_id", result.TaskID, "confidence", result.Confidence)
		}

		Success(c, gin.H{"status": "ok"})
	}
}

// PrinterPause 打印机暂停
func PrinterPause(log *services.LogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			MachineSN string `json:"machine_sn" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			ValidationError(c, "无效的请求参数")
			return
		}

		// TODO: 实现打印机暂停逻辑，需要调用Moonraker客户端
		log.Info("请求打印机暂停", "machine_sn", req.MachineSN)

		Success(c, gin.H{"status": "ok"})
	}
} 