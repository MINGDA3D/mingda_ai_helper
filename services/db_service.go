package services

import (
	"errors"
	"mingda_ai_helper/models"
	"gorm.io/gorm"
	"gorm.io/driver/sqlite"
	"time"
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
	var existingSettings models.UserSettings
	result := s.db.First(&existingSettings)
	if result.Error == nil {
		// 如果记录存在，更新它
		return s.db.Model(&existingSettings).Updates(settings).Error
	} else if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// 如果记录不存在，创建新记录
		return s.db.Create(settings).Error
	}
	// 其他错误
	return result.Error
}

// 预测结果相关操作
func (s *DBService) GetPredictionResult(taskID string) (*models.PredictionResult, error) {
	var result models.PredictionResult
	
	err := s.db.Where("task_id = ?", taskID).First(&result).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 记录不存在时返回nil, nil
		}
		return nil, err // 其他错误正常返回
	}
	
	return &result, nil
}

func (s *DBService) SavePredictionResult(result *models.PredictionResult) error {
	// 检查是否已存在相同taskID的记录
	var existingResult models.PredictionResult
	err := s.db.Where("task_id = ?", result.TaskID).First(&existingResult).Error
	if err == nil {
		// 如果记录已存在，且新记录的状态比旧记录更新，则更新记录
		if result.PredictionStatus > existingResult.PredictionStatus {
			return s.db.Model(&models.PredictionResult{}).
				Where("task_id = ?", result.TaskID).
				Updates(map[string]interface{}{
					"prediction_status": result.PredictionStatus,
					"prediction_model": result.PredictionModel,
					"has_defect":      result.HasDefect,
					"defect_type":     result.DefectType,
					"confidence":      result.Confidence,
					"updated_at":      time.Now(),
				}).Error
		}
		// 如果新记录的状态不比旧记录更新，则忽略
		return nil
	} else if err == gorm.ErrRecordNotFound {
		// 如果记录不存在，则创建新记录
		return s.db.Create(result).Error
	}
	// 其他错误直接返回
	return err
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