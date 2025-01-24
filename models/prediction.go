package models

type PredictionResult struct {
	TaskID           string  `json:"task_id" db:"task_id"`
	PredictionStatus int     `json:"prediction_status" db:"prediction_status"`
	PredictionModel  string  `json:"prediction_model" db:"prediction_model"`
	HasDefect        bool    `json:"has_defect" db:"has_defect"`
	DefectType       string  `json:"defect_type" db:"defect_type"`
	Confidence       float64 `json:"confidence" db:"confidence"`
}

func (p *PredictionResult) TableName() string {
	return "prediction_results"
} 