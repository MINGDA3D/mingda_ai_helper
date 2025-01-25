package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"mingda_ai_helper/models"
	"mingda_ai_helper/services"
)

func main() {
	// 初始化数据库服务
	dbService, err := services.NewDBService("printer_data.db")
	if err != nil {
		log.Fatalf("初始化数据库服务失败: %v", err)
	}

	// 创建HTTP处理函数
	http.HandleFunc("/api/v1/settings/sync", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "只支持POST方法", http.StatusMethodNotAllowed)
			return
		}

		// 解析请求体
		var settings models.UserSettings
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&settings); err != nil {
			http.Error(w, "无效的JSON格式", http.StatusBadRequest)
			return
		}

		// 打印收到的请求
		fmt.Printf("收到设置请求:\n")
		fmt.Printf("启用AI: %v\n", settings.EnableAI)
		fmt.Printf("启用云端AI: %v\n", settings.EnableCloudAI)
		fmt.Printf("置信度阈值: %d\n", settings.ConfidenceThreshold)
		fmt.Printf("超过阈值暂停: %v\n", settings.PauseOnThreshold)

		// 验证置信度阈值
		if settings.ConfidenceThreshold < 0 || settings.ConfidenceThreshold > 100 {
			http.Error(w, "置信度阈值必须在0-100之间", http.StatusBadRequest)
			return
		}

		// 保存设置到数据库
		if err := dbService.SaveUserSettings(&settings); err != nil {
			log.Printf("保存设置失败: %v", err)
			http.Error(w, "保存设置失败", http.StatusInternalServerError)
			return
		}

		// 返回成功响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    0,
			"message": "success",
			"data": map[string]interface{}{
				"status": "ok",
			},
		})
	})

	// 启动服务器
	fmt.Println("服务器正在监听 :8081 端口...")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
} 