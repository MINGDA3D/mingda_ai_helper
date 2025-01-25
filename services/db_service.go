package services

import (
	"mingda_ai_helper/models"
	"gorm.io/gorm"
	"gorm.io/driver/sqlite"
)

type DBService struct {
	db *gorm.DB
}

func NewDBService(dbPath string) (*DBService, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	service := &DBService{db: db}
	if err := service.initTables(); err != nil {
		return nil, err
	}

	return service, nil
}

func (s *DBService) initTables() error {
	// 自动迁移表结构
	return s.db.AutoMigrate(
		&models.MachineInfo{},
		&models.UserSettings{},
		&models.PredictionResult{},
	)
}

func (s *DBService) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (s *DBService) DB() *gorm.DB {
	return s.db
}

// 机器信息相关操作
func (s *DBService) GetMachineInfo() (*models.MachineInfo, error) {
	var info models.MachineInfo
	result := s.db.First(&info)
	if result.Error != nil {
		return nil, result.Error
	}
	return &info, nil
}

func (s *DBService) SaveMachineInfo(info *models.MachineInfo) error {
	var count int64
	s.db.Model(&models.MachineInfo{}).Count(&count)
	if count > 0 {
		return s.db.Model(&models.MachineInfo{}).Updates(info).Error
	}
	return s.db.Create(info).Error
}

func (s *DBService) UpdateMachineToken(machineSN string, newToken string) error {
	return s.db.Model(&models.MachineInfo{}).
		Where("machine_sn = ?", machineSN).
		Update("auth_token", newToken).Error
}

// 用户设置相关操作
func (s *DBService) GetUserSettings() (*models.UserSettings, error) {
	var settings models.UserSettings
	result := s.db.First(&settings)
	if result.Error != nil {
		return nil, result.Error
	}
	return &settings, nil
}

func (s *DBService) SaveUserSettings(settings *models.UserSettings) error {
	var count int64
	s.db.Model(&models.UserSettings{}).Count(&count)
	if count > 0 {
		return s.db.Model(&models.UserSettings{}).Updates(settings).Error
	}
	return s.db.Create(settings).Error
}

// 预测结果相关操作
func (s *DBService) GetPredictionResult(taskID string) (*models.PredictionResult, error) {
	var result models.PredictionResult
	if err := s.db.Where("task_id = ?", taskID).First(&result).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *DBService) SavePredictionResult(result *models.PredictionResult) error {
	return s.db.Create(result).Error
}

func (s *DBService) UpdatePredictionStatus(taskID string, status models.PredictionStatus) error {
	return s.db.Model(&models.PredictionResult{}).
		Where("task_id = ?", taskID).
		Update("prediction_status", status).Error
}

func (s *DBService) ListPredictionResults(limit int) ([]models.PredictionResult, error) {
	var results []models.PredictionResult
	err := s.db.Order("created_at desc").Limit(limit).Find(&results).Error
	return results, err
}

func (s *DBService) DeletePredictionResult(taskID string) error {
	return s.db.Where("task_id = ?", taskID).Delete(&models.PredictionResult{}).Error
} 