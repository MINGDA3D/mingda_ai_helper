package models

import (
	"time"
	"gorm.io/gorm"
)

// Machine 设备信息模型
type Machine struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	SN        string    `gorm:"column:sn;type:varchar(64);uniqueIndex;not null"`
	Model     string    `gorm:"column:model;type:varchar(64);not null"`
	Token     string    `gorm:"column:token;type:varchar(255);not null"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP"`
}

// TableName 指定表名
func (Machine) TableName() string {
	return "machine_info"
}

// GetMachine 获取设备信息
func GetMachine(db *gorm.DB) (*Machine, error) {
	var machine Machine
	result := db.First(&machine)
	if result.Error != nil {
		return nil, result.Error
	}
	return &machine, nil
}

// SaveMachine 保存设备信息
func SaveMachine(db *gorm.DB, machine *Machine) error {
	// 如果已存在记录，则更新
	var count int64
	db.Model(&Machine{}).Count(&count)
	if count > 0 {
		return db.Model(&Machine{}).Updates(machine).Error
	}
	// 否则创建新记录
	return db.Create(machine).Error
} 