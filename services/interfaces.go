package services

import (
	"go.uber.org/zap"
	"mingda_ai_helper/models"
)

// DBInterface 数据库服务接口
type DBInterface interface {
	SaveMachineInfo(info *models.MachineInfo) error
	GetMachineInfo() (*models.MachineInfo, error)
	UpdateMachineToken(machineSN, newToken string) error
	SaveUserSettings(settings *models.UserSettings) error
	GetUserSettings() (*models.UserSettings, error)
	SavePredictionResult(result *models.PredictionResult) error
}

// LogInterface 日志服务接口
type LogInterface interface {
	Info(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
} 