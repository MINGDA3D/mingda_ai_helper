package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"mingda_ai_helper/models"
	"mingda_ai_helper/services"
)

// AICallbackRequest AI回调请求结构
type AICallbackRequest struct {
	TaskID      string  `json:"task_id"`
	HasDefect   bool    `json:"has_defect"`
	DefectType  string  `json:"defect_type"`
	Confidence  float64 `json:"confidence"`
}

// SettingsHandler 处理设置相关的请求
type SettingsHandler struct {
	dbService      *services.DBService
	moonrakerClient *services.MoonrakerClient
}

// NewSettingsHandler 创建新的设置处理器
func NewSettingsHandler(dbService *services.DBService, moonrakerClient *services.MoonrakerClient) *SettingsHandler {
	return &SettingsHandler{
		dbService:      dbService,
		moonrakerClient: moonrakerClient,
	}
}

// HandleSettingsSync 处理设置同步
func (h *SettingsHandler) HandleSettingsSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只支持POST方法", http.StatusMethodNotAllowed)
		return
	}

	var settings models.UserSettings
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&settings); err != nil {
		http.Error(w, "无效的JSON格式", http.StatusBadRequest)
		return
	}

	fmt.Printf("收到设置请求:\n")
	fmt.Printf("启用AI: %v\n", settings.EnableAI)
	fmt.Printf("启用云端AI: %v\n", settings.EnableCloudAI)
	fmt.Printf("置信度阈值: %d\n", settings.ConfidenceThreshold)
	fmt.Printf("超过阈值暂停: %v\n", settings.PauseOnThreshold)

	if settings.ConfidenceThreshold < 0 || settings.ConfidenceThreshold > 100 {
		http.Error(w, "置信度阈值必须在0-100之间", http.StatusBadRequest)
		return
	}

	if err := h.dbService.SaveUserSettings(&settings); err != nil {
		log.Printf("保存设置失败: %v", err)
		http.Error(w, "保存设置失败", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    0,
		"message": "success",
		"data": map[string]interface{}{
			"status": "ok",
		},
	})
}

// HandleAICallback 处理AI回调
func (h *SettingsHandler) HandleAICallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只支持POST方法", http.StatusMethodNotAllowed)
		return
	}

	var callback AICallbackRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&callback); err != nil {
		log.Printf("解析回调请求失败: %v", err)
		http.Error(w, "无效的JSON格式", http.StatusBadRequest)
		return
	}

	fmt.Printf("\n收到AI预测回调:\n")
	fmt.Printf("任务ID: %s\n", callback.TaskID)
	fmt.Printf("检测到缺陷: %v\n", callback.HasDefect)
	fmt.Printf("缺陷类型: %s\n", callback.DefectType)
	fmt.Printf("置信度: %.2f\n", callback.Confidence)

	settings, err := h.dbService.GetUserSettings()
	if err != nil {
		log.Printf("获取用户设置失败: %v", err)
		http.Error(w, "获取用户设置失败", http.StatusInternalServerError)
		return
	}

	if callback.HasDefect && callback.Confidence >= float64(settings.ConfidenceThreshold) {
		fmt.Printf("检测到打印缺陷，置信度: %.2f，阈值: %d\n", 
			callback.Confidence, settings.ConfidenceThreshold)

		if settings.PauseOnThreshold {
			status, err := h.moonrakerClient.GetPrinterStatus()
			if err != nil {
				log.Printf("获取打印机状态失败: %v", err)
			} else if status.IsPrinting {
				if err := h.moonrakerClient.PausePrint(); err != nil {
					log.Printf("暂停打印失败: %v", err)
				} else {
					fmt.Printf("已暂停打印，任务ID: %s\n", callback.TaskID)
				}
			}
		}
	}

	result := &models.PredictionResult{
		TaskID:           callback.TaskID,
		PredictionStatus: models.StatusCompleted,
		HasDefect:        callback.HasDefect,
		DefectType:       callback.DefectType,
		Confidence:       callback.Confidence,
	}

	if err := h.dbService.SavePredictionResult(result); err != nil {
		log.Printf("保存预测结果失败: %v", err)
		http.Error(w, "保存预测结果失败", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    0,
		"message": "success",
		"data": map[string]interface{}{
			"status": "ok",
		},
	})
} 