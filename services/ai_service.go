package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"mingda_ai_helper/models"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"encoding/base64"
)

type AIService interface {
	Predict(ctx context.Context, imageURL string, taskID string) (*models.PredictionResult, error)
	PredictWithFile(ctx context.Context, imagePath string) (*models.PredictionResult, error)
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
	baseURL    string
	dbService  *DBService
	httpClient *http.Client
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

func NewCloudAIService(cloudURL string, dbService *DBService) *CloudAIService {
	return &CloudAIService{
		baseURL: cloudURL + "/api/v1",
		dbService: dbService,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
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

	// 打印请求信息（调试用）
	fmt.Printf("\n请求URL: %s\n", req.URL.String())
	fmt.Printf("请求方法: %s\n", req.Method)
	fmt.Printf("Content-Type: %s\n", req.Header.Get("Content-Type"))
	fmt.Printf("请求体: %s\n\n", string(jsonData))

	// 发送请求
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request failed: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应内容
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// 打印响应信息（调试用）
	fmt.Printf("响应状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应内容: %s\n\n", string(respBody))

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned non-200 status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	// 解析响应
	var aiResp struct {
		Detections []struct {
			Bbox       []float64 `json:"bbox"`
			Class      string    `json:"class"`
			Confidence float64   `json:"confidence"`
		} `json:"detections"`
		HasDefect    bool   `json:"has_defect"`
		PredictModel string `json:"predict_model"`
		Status       string `json:"status"`
		TaskID       string `json:"task_id"`
	}

	if err := json.Unmarshal(respBody, &aiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v, raw response: %s", err, string(respBody))
	}

	// 创建预测结果
	result := &models.PredictionResult{
		TaskID:           taskID,
		PredictionStatus: models.StatusProcessing,
		PredictionModel:  aiResp.PredictModel,
		HasDefect:        aiResp.HasDefect,
	}

	// 保存预测结果到数据库
	if err := s.dbService.SavePredictionResult(result); err != nil {
		return nil, fmt.Errorf("failed to save prediction result: %v", err)
	}

	return result, nil
}

func (s *LocalAIService) PredictWithFile(ctx context.Context, imagePath string) (*models.PredictionResult, error) {
	// 检查文件是否存在
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("image file not found: %s", imagePath)
	}

	// 生成任务ID
	taskID := fmt.Sprintf("PT%s", time.Now().Format("20060102150405"))

	// 创建预测请求
	reqBody := PredictRequest{
		ImageURL:    fmt.Sprintf("file://%s", imagePath),
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
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned non-200 status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	// 创建预测结果
	result := &models.PredictionResult{
		TaskID:           taskID,
		PredictionStatus: models.StatusProcessing,
		PredictionModel:  "local_ai",
	}

	// 保存预测结果到数据库
	if err := s.dbService.SavePredictionResult(result); err != nil {
		return nil, fmt.Errorf("failed to save prediction result: %v", err)
	}

	return result, nil
}

func (s *CloudAIService) Predict(ctx context.Context, imageURL string, taskID string) (*models.PredictionResult, error) {
	return nil, fmt.Errorf("not implemented for URL-based prediction")
}

func (s *CloudAIService) PredictWithFile(ctx context.Context, imagePath string) (*models.PredictionResult, error) {
	// 获取机器信息和认证令牌
	machineInfo, err := s.dbService.GetMachineInfo()
	if err != nil {
		return nil, fmt.Errorf("获取机器信息失败: %v", err)
	}

	// 读取图片文件并转换为base64
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, fmt.Errorf("读取图片文件失败: %v", err)
	}
	base64Image := base64.StdEncoding.EncodeToString(imageData)

	// 生成任务ID
	taskID := fmt.Sprintf("PT%s", time.Now().Format("20060102150405"))

	// 构造请求体
	reqBody := struct {
		TaskID    string `json:"task_id"`
		ImageData string `json:"image_data"`
		MachineSN string `json:"machine_sn"`
	}{
		TaskID:    taskID,
		ImageData: base64Image,
		MachineSN: machineInfo.MachineSN,
	}

	// 将请求体转换为JSON
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求体失败: %v", err)
	}

	// 创建上传请求
	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/device/print/image", body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+machineInfo.AuthToken)

	// 打印请求信息
	fmt.Printf("\n请求URL: %s\n", req.URL.String())
	fmt.Printf("请求方法: %s\n", req.Method)
	fmt.Printf("Content-Type: %s\n", req.Header.Get("Content-Type"))
	fmt.Printf("Authorization: Bearer %s...\n", machineInfo.AuthToken[:30])
	fmt.Printf("TaskID: %s\n", taskID)
	fmt.Printf("图片文件: %s\n\n", imagePath)

	// 发送请求
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应内容
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("服务器返回错误状态码: %d, 响应内容: %s", resp.StatusCode, string(respBody))
	}

	// 解析响应
	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			TaskID string `json:"task_id"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v, 原始响应: %s", err, string(respBody))
	}

	if result.Code != 200 {
		return nil, fmt.Errorf("上传失败: %s", result.Msg)
	}

	// 查询预测状态
	statusReq, err := http.NewRequestWithContext(ctx, "GET", 
		fmt.Sprintf("%s/device/print/images?task_id=%s", s.baseURL, taskID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create status request: %v", err)
	}

	statusReq.Header.Set("Authorization", "Bearer "+machineInfo.AuthToken)
	
	statusResp, err := s.httpClient.Do(statusReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %v", err)
	}
	defer statusResp.Body.Close()

	// 创建预测结果
	predictionResult := &models.PredictionResult{
		TaskID:           taskID,
		PredictionStatus: models.StatusProcessing,
		PredictionModel:  "cloud_ai",
	}

	// 保存预测结果到数据库
	if err := s.dbService.SavePredictionResult(predictionResult); err != nil {
		return nil, fmt.Errorf("保存预测结果失败: %v", err)
	}

	return predictionResult, nil
} 