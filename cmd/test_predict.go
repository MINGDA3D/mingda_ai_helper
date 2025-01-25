package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mingda_ai_helper/config"
	"mingda_ai_helper/handlers"
	"mingda_ai_helper/services"
	"net"
	"net/http"
	"time"
)

// getLocalIP 获取局域网IP地址
func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", fmt.Errorf("无法获取局域网IP地址")
}

func printJSON(label string, v interface{}) {
	data, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		fmt.Printf("%s: 错误 - %v\n", label, err)
		return
	}
	fmt.Printf("%s:\n%s\n", label, string(data))
}

func startCallbackServer(dbService *services.DBService, logService *services.LogService) {
	// 设置路由
	router := handlers.SetupRouter(nil, dbService, logService)

	// 在新的goroutine中启动服务器
	go func() {
		fmt.Println("启动回调服务器在 :8081 端口")
		if err := router.Run(":8081"); err != nil {
			log.Fatalf("启动回调服务器失败: %v", err)
		}
	}()

	// 等待服务器启动
	time.Sleep(1 * time.Second)
}

func main() {
	// 获取局域网IP
	localIP, err := getLocalIP()
	if err != nil {
		log.Fatalf("获取局域网IP失败: %v", err)
	}
	fmt.Printf("局域网IP地址: %s\n\n", localIP)

	// 加载配置文件
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// 初始化日志服务
	logService, err := services.NewLogService(cfg.Logging.Level, cfg.Logging.File)
	if err != nil {
		log.Fatalf("初始化日志服务失败: %v", err)
	}

	// 初始化数据库服务
	dbService, err := services.NewDBService(cfg.Database.Path)
	if err != nil {
		log.Fatalf("初始化数据库服务失败: %v", err)
	}

	// 启动回调服务器
	startCallbackServer(dbService, logService)

	// 初始化本地AI服务
	aiServerURL := "http://localhost:5000"  // AI服务地址
	callbackURL := "http://localhost:8081/api/v1/ai/callback"  // 回调地址使用8081端口
	aiService := services.NewLocalAIService(aiServerURL, callbackURL, dbService)

	// 生成任务ID
	taskID := fmt.Sprintf("PT%s", time.Now().Format("20060102150405"))

	// 构造图片URL（使用实际IP）
	imageURL := fmt.Sprintf("http://%s/webcam/?action=snapshot", localIP)

	// 构造预测请求
	predictRequest := services.PredictRequest{
		ImageURL:    imageURL,
		TaskID:      taskID,
		CallbackURL: callbackURL,
	}

	fmt.Printf("=== 发送预测请求 ===\n")
	fmt.Printf("AI服务地址: %s\n", aiServerURL)
	printJSON("请求内容", predictRequest)

	// 发送预测请求
	result, err := aiService.Predict(context.Background(), imageURL, taskID)
	if err != nil {
		log.Fatalf("发送预测请求失败: %v", err)
	}

	fmt.Printf("\n=== 初始预测结果 ===\n")
	printJSON("结果", result)
	fmt.Println("\n等待AI服务回调...")

	// 等待30秒，让回调有时间处理
	time.Sleep(30 * time.Second)

	// 查询最终结果
	finalResult, err := dbService.GetPredictionResult(taskID)
	if err != nil {
		log.Fatalf("获取预测结果失败: %v", err)
	}

	fmt.Printf("\n=== 最终预测结果 ===\n")
	printJSON("结果", finalResult)

	// 打印状态说明
	fmt.Printf("\n状态说明：\n")
	fmt.Printf("0: 等待处理\n")
	fmt.Printf("1: 处理中\n")
	fmt.Printf("2: 处理完成\n")
} 