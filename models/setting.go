package models

import (
	"gorm.io/gorm"
)

// UserSettings 用户设置模型
type UserSettings struct {
	gorm.Model
	EnableAI             bool `gorm:"column:enable_ai;not null" json:"enable_ai"`
	EnableCloudAI        bool `gorm:"column:enable_cloud_ai;not null" json:"enable_cloud_ai"`
	ConfidenceThreshold  int  `gorm:"column:confidence_threshold;not null;check:confidence_threshold BETWEEN 0 AND 100" json:"confidence_threshold"`
	PauseOnThreshold    bool `gorm:"column:pause_on_threshold;not null" json:"pause_on_threshold"`
}

// TableName 指定表名
func (UserSettings) TableName() string {
	return "user_settings"
}

// GetUserSettings 获取用户设置
func GetUserSettings(db *gorm.DB) (*UserSettings, error) {
	var settings UserSettings
	result := db.First(&settings)
	if result.Error != nil {
		return nil, result.Error
	}
	return &settings, nil
}

// SaveUserSettings 保存用户设置
func SaveUserSettings(db *gorm.DB, settings *UserSettings) error {
	var count int64
	db.Model(&UserSettings{}).Count(&count)
	if count > 0 {
		return db.Model(&UserSettings{}).Updates(settings).Error
	}
	return db.Create(settings).Error
} 