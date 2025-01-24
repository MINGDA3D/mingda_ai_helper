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
	query := `INSERT INTO machine_info (machine_sn, machine_model, auth_token) 
			  VALUES (?, ?, ?)`
	_, err := s.db.Exec(query, info.MachineSN, info.MachineModel, info.AuthToken)
	return err
}

func (s *DBService) SaveUserSettings(settings *models.UserSettings) error {
	query := `INSERT INTO user_settings (enable_ai, enable_cloud_ai, confidence_threshold, pause_on_threshold) 
			  VALUES (?, ?, ?, ?)`
	_, err := s.db.Exec(query, settings.EnableAI, settings.EnableCloudAI, 
						settings.ConfidenceThreshold, settings.PauseOnThreshold)
	return err
}

func (s *DBService) SavePredictionResult(result *models.PredictionResult) error {
	query := `INSERT INTO prediction_results (task_id, prediction_status, prediction_model, 
			  has_defect, defect_type, confidence) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := s.db.Exec(query, result.TaskID, result.PredictionStatus, result.PredictionModel,
						result.HasDefect, result.DefectType, result.Confidence)
	return err
} 