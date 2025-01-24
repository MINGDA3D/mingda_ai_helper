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

func (s *DBService) SaveMachineInfo(info *models.MachineInfo) error {
	result := s.db.Create(info)
	return result.Error
}

func (s *DBService) SaveUserSettings(settings *models.UserSettings) error {
	result := s.db.Create(settings)
	return result.Error
}

func (s *DBService) SavePredictionResult(result *models.PredictionResult) error {
	dbResult := s.db.Create(result)
	return dbResult.Error
} 