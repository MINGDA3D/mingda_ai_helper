package services

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"mingda_ai_helper/models"
	"net/http"
	"os"
	"path/filepath"
	"time"
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

// DeviceRegisterResponse 设备注册响应
type DeviceRegisterResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Secret string `json:"secret"`
	} `json:"data"`
}

// DeviceAuthResponse 设备认证响应
type DeviceAuthResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Token string `json:"token"`
	} `json:"data"`
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
		return nil, fmt.Errorf("failed to get machine info: %v", err)
	}
	if machineInfo == nil {
		return nil, fmt.Errorf("machine info not found")
	}

	// 生成任务ID
	if imagePath == "" {
		return nil, fmt.Errorf("image path is required")
	}

	taskID := fmt.Sprintf("PT%s", time.Now().Format("20060102150405"))

	// 检查文件是否存在
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("image file not found: %s", imagePath)
	}

	// 定义发送请求的函数
	sendRequest := func(token string) (*http.Response, error) {
		// 打开文件
		file, err := os.Open(imagePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open image file: %v", err)
		}
		defer file.Close()

		// 准备multipart表单
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// 添加文件
		part, err := writer.CreateFormFile("file", filepath.Base(imagePath))
		if err != nil {
			return nil, fmt.Errorf("failed to create form file: %v", err)
		}
		if _, err = io.Copy(part, file); err != nil {
			return nil, fmt.Errorf("failed to copy file content: %v", err)
		}

		// 添加task_id
		if err = writer.WriteField("task_id", taskID); err != nil {
			return nil, fmt.Errorf("failed to add task_id field: %v", err)
		}

		if err = writer.Close(); err != nil {
			return nil, fmt.Errorf("failed to close writer: %v", err)
		}

		// 创建上传请求
		req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/device/print/image", body)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %v", err)
		}

		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", "Bearer "+token)

		// 打印请求信息
		fmt.Printf("\n请求URL: %s\n", req.URL.String())
		fmt.Printf("请求方法: %s\n", req.Method)
		fmt.Printf("Content-Type: %s\n", req.Header.Get("Content-Type"))
		fmt.Printf("Authorization: Bearer %s...\n", token[:30])
		fmt.Printf("TaskID: %s\n", taskID)
		fmt.Printf("图片文件: %s\n\n", imagePath)

		return s.httpClient.Do(req)
	}

	// 首次尝试发送请求
	resp, err := sendRequest(machineInfo.AuthToken)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应内容
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// 打印响应信息
	fmt.Printf("响应状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应内容: %s\n\n", string(respBody))

	// 如果是401错误，尝试刷新token并重试
	if resp.StatusCode == http.StatusUnauthorized {
		var errorResp struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal(respBody, &errorResp); err != nil {
			return nil, fmt.Errorf("failed to parse error response: %v", err)
		}

		// 如果是token过期错误
		if errorResp.Code == 1003 {
			fmt.Println("Token已过期，正在刷新...")
			
			// 刷新token
			newToken, err := s.RefreshToken(ctx, machineInfo.AuthToken)
			if err != nil {
				return nil, fmt.Errorf("failed to refresh token: %v", err)
			}

			// 更新数据库中的token
			if err := s.dbService.SaveMachineInfo(&models.MachineInfo{
				MachineSN:    machineInfo.MachineSN,
				MachineModel: machineInfo.MachineModel,
				AuthToken:    newToken,
			}); err != nil {
				return nil, fmt.Errorf("failed to update token in database: %v", err)
			}

			fmt.Println("Token刷新成功，重试请求...")

			// 使用新token重试请求
			resp, err = sendRequest(newToken)
			if err != nil {
				return nil, fmt.Errorf("failed to retry request: %v", err)
			}
			defer resp.Body.Close()

			// 读取重试响应
			respBody, err = io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read retry response body: %v", err)
			}

			// 打印重试响应信息
			fmt.Printf("重试响应状态码: %d\n", resp.StatusCode)
			fmt.Printf("重试响应内容: %s\n\n", string(respBody))
		}
	}

	// 检查最终响应状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned non-200 status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	// 解析响应
	var result struct {
		Code int         `json:"code"`
		Msg  string      `json:"msg"`
		Data interface{} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v, raw response: %s", err, string(respBody))
	}

	if result.Code != 200 {
		return nil, fmt.Errorf("upload failed: %s", result.Msg)
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
		return nil, fmt.Errorf("failed to save prediction result: %v", err)
	}

	return predictionResult, nil
}

// RegisterDevice 注册设备
func (s *CloudAIService) RegisterDevice(ctx context.Context, sn, model string) (string, error) {
	// 准备请求体
	reqBody := map[string]string{
		"sn":    sn,
		"model": model,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request body failed: %v", err)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/devices/register", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("create request failed: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request failed: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response failed: %v", err)
	}

	// 解析响应
	var registerResp DeviceRegisterResponse
	if err := json.Unmarshal(respBody, &registerResp); err != nil {
		return "", fmt.Errorf("unmarshal response failed: %v", err)
	}

	if registerResp.Code != 0 {
		return "", fmt.Errorf("register device failed: %s", registerResp.Message)
	}

	return registerResp.Data.Secret, nil
}

// AuthDevice 设备认证
func (s *CloudAIService) AuthDevice(ctx context.Context, sn, secret string) (string, error) {
	// 生成时间戳
	timestamp := time.Now().Unix()

	// 生成签名: sha256(sn + secret + timestamp)
	signStr := fmt.Sprintf("%s%s%d", sn, secret, timestamp)
	h := sha256.New()
	h.Write([]byte(signStr))
	sign := hex.EncodeToString(h.Sum(nil))

	// 准备请求体
	reqBody := map[string]interface{}{
		"sn":        sn,
		"sign":      sign,
		"timestamp": timestamp,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request body failed: %v", err)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/devices/auth", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("create request failed: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request failed: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response failed: %v", err)
	}

	// 解析响应
	var authResp DeviceAuthResponse
	if err := json.Unmarshal(respBody, &authResp); err != nil {
		return "", fmt.Errorf("unmarshal response failed: %v", err)
	}

	if authResp.Code != 0 {
		return "", fmt.Errorf("auth device failed: %s", authResp.Message)
	}

	return authResp.Data.Token, nil
}

// RefreshToken 刷新token
func (s *CloudAIService) RefreshToken(ctx context.Context, oldToken string) (string, error) {
	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/devices/refresh", nil)
	if err != nil {
		return "", fmt.Errorf("create request failed: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+oldToken)

	// 发送请求
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request failed: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response failed: %v", err)
	}

	// 解析响应
	var refreshResp DeviceAuthResponse
	if err := json.Unmarshal(respBody, &refreshResp); err != nil {
		return "", fmt.Errorf("unmarshal response failed: %v", err)
	}

	if refreshResp.Code != 0 {
		return "", fmt.Errorf("refresh token failed: %s", refreshResp.Message)
	}

	return refreshResp.Data.Token, nil
} 