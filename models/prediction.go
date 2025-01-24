package models

import (
	"gorm.io/gorm"
)

// PredictionStatus 预测状态
type PredictionStatus int

const (
	StatusPending PredictionStatus = iota
	StatusProcessing
	StatusCompleted
)

// PredictionResult 预测结果模型
type PredictionResult struct {
	gorm.Model
	TaskID           string          `gorm:"column:task_id;type:varchar(64);uniqueIndex;not null"`
	PredictionStatus PredictionStatus `gorm:"column:prediction_status;not null;check:prediction_status IN (0, 1, 2)"`
	PredictionModel  string          `gorm:"column:prediction_model;type:varchar(64);not null"`
	HasDefect        bool            `gorm:"column:has_defect;not null"`
	DefectType       string          `gorm:"column:defect_type;type:varchar(64)"`
	Confidence       float64         `gorm:"column:confidence;check:confidence BETWEEN 0 AND 100"`
}

// TableName 指定表名
func (PredictionResult) TableName() string {
	return "prediction_results"
}

// GetPredictionResult 获取预测结果
func GetPredictionResult(db *gorm.DB, taskID string) (*PredictionResult, error) {
	var result PredictionResult
	if err := db.Where("task_id = ?", taskID).First(&result).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

// SavePredictionResult 保存预测结果
func SavePredictionResult(db *gorm.DB, result *PredictionResult) error {
	return db.Create(result).Error
}

// UpdatePredictionStatus 更新预测状态
func UpdatePredictionStatus(db *gorm.DB, taskID string, status PredictionStatus) error {
	return db.Model(&PredictionResult{}).Where("task_id = ?", taskID).Update("prediction_status", status).Error
} 