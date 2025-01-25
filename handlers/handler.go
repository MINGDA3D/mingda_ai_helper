package handlers

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"mingda_ai_helper/models"
	"mingda_ai_helper/pkg/response"
	"mingda_ai_helper/services"
)

// HealthCheck 健康检查
func HealthCheck(c *gin.Context) {
	response.Success(c, gin.H{"status": "ok"})
}

// MachineRegister 设备注册
func MachineRegister(db services.DBInterface, log services.LogInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			MachineModel string `json:"machine_model" binding:"required"`
			MachineSN    string `json:"machine_sn" binding:"required"`
			AuthToken    string `json:"auth_token" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			response.ValidationError(c, "无效的请求参数")
			return
		}

		// 保存机器信息
		machine := &models.MachineInfo{
			MachineSN:    req.MachineSN,
			MachineModel: req.MachineModel,
			AuthToken:    req.AuthToken,
		}

		if err := db.SaveMachineInfo(machine); err != nil {
			log.Error("保存机器信息失败", zap.Error(err))
			response.ServerError(c, "保存机器信息失败")
			return
		}

		response.Success(c, gin.H{"status": "ok"})
	}
}

// TokenRefresh Token刷新
func TokenRefresh(db services.DBInterface, log services.LogInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			MachineSN string `json:"machine_sn" binding:"required"`
			NewToken  string `json:"new_token" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			response.ValidationError(c, "无效的请求参数")
			return
		}

		// 更新数据库中的token
		if err := db.UpdateMachineToken(req.MachineSN, req.NewToken); err != nil {
			log.Error("更新token失败", zap.Error(err))
			response.ServerError(c, "更新token失败")
			return
		}

		response.Success(c, gin.H{"status": "ok"})
	}
}

// SettingsSync 同步用户设置
func SettingsSync(db services.DBInterface, log services.LogInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		var settings models.UserSettings
		if err := c.ShouldBindJSON(&settings); err != nil {
			response.ValidationError(c, "无效的设置参数")
			return
		}

		// 验证参数
		if settings.ConfidenceThreshold < 0 || settings.ConfidenceThreshold > 100 {
			response.ValidationError(c, "置信度阈值必须在0-100之间")
			return
		}

		if err := db.SaveUserSettings(&settings); err != nil {
			log.Error("保存用户设置失败", zap.Error(err))
			response.ServerError(c, "保存用户设置失败")
			return
		}

		response.Success(c, gin.H{"status": "ok"})
	}
}

// Predict AI预测请求
func Predict(ai services.AIService, db services.DBInterface, log services.LogInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ImageURL    string `json:"image_url" binding:"required,url"`
			TaskID      string `json:"task_id" binding:"required"`
			CallbackURL string `json:"callback_url" binding:"required,url"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			response.ValidationError(c, "无效的请求参数")
			return
		}

		// 创建预测任务
		result := &models.PredictionResult{
			TaskID:           req.TaskID,
			PredictionStatus: models.StatusPending,
		}

		// 保存初始状态
		if err := db.SavePredictionResult(result); err != nil {
			log.Error("保存预测任务失败", zap.Error(err))
			response.ServerError(c, "保存预测任务失败")
			return
		}

		// 异步执行预测
		go func() {
			result, err := ai.Predict(c.Request.Context(), req.ImageURL, req.TaskID)
			if err != nil {
				log.Error("预测失败", zap.Error(err))
				return
			}

			if err := db.SavePredictionResult(result); err != nil {
				log.Error("保存预测结果失败", zap.Error(err))
			}
		}()

		response.Success(c, gin.H{"task_id": req.TaskID})
	}
}

// AICallback AI回调处理
func AICallback(db services.DBInterface, log services.LogInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		var result models.PredictionResult
		if err := c.ShouldBindJSON(&result); err != nil {
			response.ValidationError(c, "无效的回调参数")
			return
		}

		if result.TaskID == "" {
			response.ValidationError(c, "任务ID不能为空")
			return
		}

		if err := db.SavePredictionResult(&result); err != nil {
			log.Error("保存预测结果失败", zap.Error(err))
			response.ServerError(c, "保存预测结果失败")
			return
		}

		// 检查是否需要暂停打印
		settings, err := db.GetUserSettings()
		if err != nil {
			log.Error("获取用户设置失败", zap.Error(err))
			response.ServerError(c, "获取用户设置失败")
			return
		}

		if settings.PauseOnThreshold && result.Confidence >= float64(settings.ConfidenceThreshold) {
			// TODO: 调用打印机暂停接口
			log.Info("触发打印暂停", 
				zap.String("task_id", result.TaskID), 
				zap.Float64("confidence", result.Confidence))
		}

		response.Success(c, gin.H{"status": "ok"})
	}
}

// PrinterPause 打印机暂停
func PrinterPause(log services.LogInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			MachineSN string `json:"machine_sn" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			response.ValidationError(c, "无效的请求参数")
			return
		}

		// TODO: 实现打印机暂停逻辑，需要调用Moonraker客户端
		log.Info("请求打印机暂停", zap.String("machine_sn", req.MachineSN))

		response.Success(c, gin.H{"status": "ok"})
	}
} 