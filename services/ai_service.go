package services

import (
	"context"
	"errors"
	"mingda_ai_helper/models"
)

type AIService interface {
	Predict(ctx context.Context, imageURL string, taskID string) (*models.PredictionResult, error)
}

type LocalAIService struct {
	modelPath string
}

type CloudAIService struct {
	endpoint string
	apiKey   string
}

func NewLocalAIService(modelPath string) *LocalAIService {
	return &LocalAIService{
		modelPath: modelPath,
	}
}

func NewCloudAIService(endpoint, apiKey string) *CloudAIService {
	return &CloudAIService{
		endpoint: endpoint,
		apiKey:   apiKey,
	}
}

func (s *LocalAIService) Predict(ctx context.Context, imageURL string, taskID string) (*models.PredictionResult, error) {
	// TODO: 实现本地AI预测逻辑
	return nil, errors.New("not implemented")
}

func (s *CloudAIService) Predict(ctx context.Context, imageURL string, taskID string) (*models.PredictionResult, error) {
	// TODO: 实现云端AI预测逻辑
	return nil, errors.New("not implemented")
} 