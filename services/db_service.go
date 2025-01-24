package services

import (
	"database/sql"
	"mingda_ai_helper/models"
	_ "github.com/mattn/go-sqlite3"
)

type DBService struct {
	db *sql.DB
}

func NewDBService(dbPath string) (*DBService, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	service := &DBService{db: db}
	if err := service.initTables(); err != nil {
		db.Close()
		return nil, err
	}

	return service, nil
}

func (s *DBService) initTables() error {
	// 创建机器信息表
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS machine_info (
			machine_sn TEXT PRIMARY KEY,
			machine_model TEXT NOT NULL,
			auth_token TEXT NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	// 创建用户设置表
	_, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS user_settings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			enable_ai BOOLEAN NOT NULL,
			enable_cloud_ai BOOLEAN NOT NULL,
			confidence_threshold INTEGER NOT NULL CHECK(confidence_threshold BETWEEN 0 AND 100),
			pause_on_threshold BOOLEAN NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	// 创建预测结果表
	_, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS prediction_results (
			task_id TEXT PRIMARY KEY,
			prediction_status INTEGER NOT NULL CHECK(prediction_status IN (0, 1, 2)),
			prediction_model TEXT NOT NULL,
			has_defect BOOLEAN NOT NULL,
			defect_type TEXT,
			confidence REAL CHECK(confidence BETWEEN 0 AND 100)
		)
	`)
	if err != nil {
		return err
	}

	return nil
}

func (s *DBService) Close() error {
	return s.db.Close()
}

func (s *DBService) DB() *sql.DB {
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