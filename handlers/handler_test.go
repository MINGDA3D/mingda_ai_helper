package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"mingda_ai_helper/models"
	"mingda_ai_helper/pkg/response"
)

// MockDBService 模拟数据库服务
type MockDBService struct {
	mock.Mock
}

func (m *MockDBService) SaveMachineInfo(info *models.MachineInfo) error {
	args := m.Called(info)
	return args.Error(0)
}

func (m *MockDBService) GetMachineInfo() (*models.MachineInfo, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.MachineInfo), args.Error(1)
}

func (m *MockDBService) UpdateMachineToken(machineSN, newToken string) error {
	args := m.Called(machineSN, newToken)
	return args.Error(0)
}

func (m *MockDBService) SaveUserSettings(settings *models.UserSettings) error {
	args := m.Called(settings)
	return args.Error(0)
}

func (m *MockDBService) GetUserSettings() (*models.UserSettings, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserSettings), args.Error(1)
}

func (m *MockDBService) SavePredictionResult(result *models.PredictionResult) error {
	args := m.Called(result)
	return args.Error(0)
}

// MockAIService 模拟AI服务
type MockAIService struct {
	mock.Mock
}

func (m *MockAIService) Predict(ctx context.Context, imageURL string, taskID string) (*models.PredictionResult, error) {
	args := m.Called(ctx, imageURL, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PredictionResult), args.Error(1)
}

// MockLogService 模拟日志服务
type MockLogService struct {
	mock.Mock
}

func (m *MockLogService) Info(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *MockLogService) Error(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

// 测试辅助函数
func setupTestRouter(db *MockDBService, ai *MockAIService, log *MockLogService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	return SetupRouter(ai, db, log)
}

// 测试健康检查接口
func TestHealthCheck(t *testing.T) {
	db := new(MockDBService)
	ai := new(MockAIService)
	log := new(MockLogService)
	router := setupTestRouter(db, ai, log)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/ai/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "success", resp.Message)
}

// 测试设备注册接口
func TestMachineRegister(t *testing.T) {
	db := new(MockDBService)
	ai := new(MockAIService)
	log := new(MockLogService)
	router := setupTestRouter(db, ai, log)

	// 准备测试数据
	reqBody := map[string]string{
		"machine_model": "TestModel",
		"machine_sn":    "TEST001",
		"auth_token":    "test-token",
	}
	jsonBody, _ := json.Marshal(reqBody)

	// 设置Mock期望
	db.On("SaveMachineInfo", mock.AnythingOfType("*models.MachineInfo")).Return(nil)
	log.On("Error", mock.Anything, mock.Anything).Return()

	// 发送请求
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/machine/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// 验证结果
	assert.Equal(t, http.StatusOK, w.Code)
	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "success", resp.Message)
}

// 测试设置同步接口
func TestSettingsSync(t *testing.T) {
	db := new(MockDBService)
	ai := new(MockAIService)
	log := new(MockLogService)
	router := setupTestRouter(db, ai, log)

	// 准备测试数据
	settings := models.UserSettings{
		EnableAI:            true,
		EnableCloudAI:       false,
		ConfidenceThreshold: 80,
		PauseOnThreshold:    true,
	}
	jsonBody, _ := json.Marshal(settings)

	// 设置Mock期望
	db.On("SaveUserSettings", mock.AnythingOfType("*models.UserSettings")).Return(nil)
	log.On("Error", mock.Anything, mock.Anything).Return()

	// 发送请求
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/settings/sync", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// 验证结果
	assert.Equal(t, http.StatusOK, w.Code)
	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)
}

// 测试预测请求接口
func TestPredict(t *testing.T) {
	db := new(MockDBService)
	ai := new(MockAIService)
	log := new(MockLogService)
	router := setupTestRouter(db, ai, log)

	// 准备测试数据
	reqBody := map[string]string{
		"image_url":    "http://example.com/test.jpg",
		"task_id":      "TASK001",
		"callback_url": "http://callback.example.com",
	}
	jsonBody, _ := json.Marshal(reqBody)

	// 设置Mock期望
	db.On("SavePredictionResult", mock.AnythingOfType("*models.PredictionResult")).Return(nil)
	log.On("Error", mock.Anything, mock.Anything).Return()

	// 发送请求
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/predict", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// 验证结果
	assert.Equal(t, http.StatusOK, w.Code)
	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	// 验证返回的task_id
	data, ok := resp.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "TASK001", data["task_id"])
}

// 测试AI回调接口
func TestAICallback(t *testing.T) {
	db := new(MockDBService)
	ai := new(MockAIService)
	log := new(MockLogService)
	router := setupTestRouter(db, ai, log)

	// 准备测试数据
	result := models.PredictionResult{
		TaskID:           "TASK001",
		PredictionStatus: models.StatusCompleted,
		PredictionModel:  "test-model",
		HasDefect:        true,
		DefectType:       "stringing",
		Confidence:       95.5,
	}
	jsonBody, _ := json.Marshal(result)

	// 设置Mock期望
	db.On("SavePredictionResult", mock.AnythingOfType("*models.PredictionResult")).Return(nil)
	db.On("GetUserSettings").Return(&models.UserSettings{
		ConfidenceThreshold: 90,
		PauseOnThreshold:   true,
	}, nil)
	log.On("Error", mock.Anything, mock.Anything).Return()
	log.On("Info", mock.Anything, mock.Anything).Return()

	// 发送请求
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/ai/callback", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// 验证结果
	assert.Equal(t, http.StatusOK, w.Code)
	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)
}

// 测试参数验证错误
func TestValidationErrors(t *testing.T) {
	db := new(MockDBService)
	ai := new(MockAIService)
	log := new(MockLogService)
	router := setupTestRouter(db, ai, log)

	testCases := []struct {
		name     string
		path     string
		method   string
		body     interface{}
		expected int
	}{
		{
			name:   "注册缺少必填参数",
			path:   "/api/v1/machine/register",
			method: "POST",
			body: map[string]string{
				"machine_model": "TestModel",
			},
			expected: http.StatusBadRequest,
		},
		{
			name:   "设置阈值超出范围",
			path:   "/api/v1/settings/sync",
			method: "POST",
			body: map[string]interface{}{
				"enable_ai":             true,
				"confidence_threshold":   150,
				"pause_on_threshold":    true,
			},
			expected: http.StatusBadRequest,
		},
		{
			name:   "预测请求URL格式错误",
			path:   "/api/v1/predict",
			method: "POST",
			body: map[string]string{
				"image_url":    "invalid-url",
				"task_id":      "TASK001",
				"callback_url": "http://callback.example.com",
			},
			expected: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tc.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tc.method, tc.path, bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expected, w.Code)
		})
	}
} 