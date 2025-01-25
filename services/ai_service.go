package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mingda_ai_helper/models"
	"net/http"
	"time"
)

type AIService interface {
	Predict(ctx context.Context, imageURL string, taskID string) (*models.PredictionResult, error)
}

// PredictRequest AI预测请求结构体
type PredictRequest struct {
	ImageURL    string `json:"image_url"`
	TaskID      string `json:"task_id"`
	CallbackURL string `json:"callback_url"`
}

type LocalAIService struct {
	localURL    string
	callbackURL string
	httpClient  *http.Client
	dbService   *DBService
}

type CloudAIService struct {
	endpoint string
	apiKey   string
}

func NewLocalAIService(localURL, callbackURL string, dbService *DBService) *LocalAIService {
	return &LocalAIService{
		localURL:    localURL,
		callbackURL: callbackURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		dbService: dbService,
	}
}

func NewCloudAIService(endpoint, apiKey string) *CloudAIService {
	return &CloudAIService{
		endpoint: endpoint,
		apiKey:   apiKey,
	}
}

func (s *LocalAIService) Predict(ctx context.Context, imageURL string, taskID string) (*models.PredictionResult, error) {
	// 创建预测请求
	reqBody := PredictRequest{
		ImageURL:    imageURL,
		TaskID:      taskID,
		CallbackURL: s.callbackURL,
	}

	// 将请求体转换为JSON
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %v", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", s.localURL+"/api/v1/predict", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request failed: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("predict request failed with status: %d", resp.StatusCode)
	}

	// 创建初始预测结果
	result := &models.PredictionResult{
		TaskID:           taskID,
		PredictionStatus: models.StatusProcessing,
	}

	// 保存初始预测结果到数据库
	if err := s.dbService.SavePredictionResult(result); err != nil {
		return nil, fmt.Errorf("save initial prediction result failed: %v", err)
	}

	return result, nil
}

func (s *CloudAIService) Predict(ctx context.Context, imageURL string, taskID string) (*models.PredictionResult, error) {
	// TODO: 实现云端AI预测逻辑
	return nil, fmt.Errorf("not implemented")
} 