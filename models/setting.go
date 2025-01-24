package models

type UserSettings struct {
	ID                 int     `json:"id" db:"id"`
	EnableAI          bool    `json:"enable_ai" db:"enable_ai"`
	EnableCloudAI     bool    `json:"enable_cloud_ai" db:"enable_cloud_ai"`
	ConfidenceThreshold int   `json:"confidence_threshold" db:"confidence_threshold"`
	PauseOnThreshold   bool   `json:"pause_on_threshold" db:"pause_on_threshold"`
}

func (s *UserSettings) TableName() string {
	return "user_settings"
} 