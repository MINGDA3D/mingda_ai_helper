package models

import (
	"gorm.io/gorm"
)

// MachineInfo 设备信息模型
type MachineInfo struct {
	gorm.Model
	MachineSN    string `gorm:"column:machine_sn;type:varchar(64);uniqueIndex;not null"`
	MachineModel string `gorm:"column:machine_model;type:varchar(64);not null"`
	AuthToken    string `gorm:"column:auth_token;type:varchar(255);not null"`
}

// TableName 指定表名
func (MachineInfo) TableName() string {
	return "machine_info"
}

// GetMachineInfo 获取设备信息
func GetMachineInfo(db *gorm.DB) (*MachineInfo, error) {
	var info MachineInfo
	result := db.First(&info)
	if result.Error != nil {
		return nil, result.Error
	}
	return &info, nil
}

// SaveMachineInfo 保存设备信息
func SaveMachineInfo(db *gorm.DB, info *MachineInfo) error {
	var count int64
	db.Model(&MachineInfo{}).Count(&count)
	if count > 0 {
		return db.Model(&MachineInfo{}).Updates(info).Error
	}
	return db.Create(info).Error
} 