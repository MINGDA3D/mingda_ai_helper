package main

import (
	"context"
	"fmt"
	"log"
	"mingda_ai_helper/config"
	"mingda_ai_helper/services"
	"net"
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

func main() {
	// 获取局域网IP
	localIP, err := getLocalIP()
	if err != nil {
		log.Fatalf("获取局域网IP失败: %v", err)
	}
	fmt.Printf("局域网IP地址: %s\n", localIP)

	// 加载配置文件
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// 初始化数据库服务
	dbService, err := services.NewDBService(cfg.Database.Path)
	if err != nil {
		log.Fatalf("初始化数据库服务失败: %v", err)
	}

	// 初始化本地AI服务
	callbackURL := fmt.Sprintf("http://%s:8080/api/v1/ai/callback", localIP)  // 使用实际IP
	aiService := services.NewLocalAIService(cfg.AI.LocalURL, callbackURL, dbService)

	// 生成任务ID
	taskID := fmt.Sprintf("PT%s", time.Now().Format("20060102150405"))

	// 构造图片URL（使用实际IP）
	imageURL := fmt.Sprintf("http://%s/webcam/?action=snapshot", localIP)

	fmt.Printf("\n开始发送预测请求:\n")
	fmt.Printf("TaskID: %s\n", taskID)
	fmt.Printf("ImageURL: %s\n", imageURL)
	fmt.Printf("CallbackURL: %s\n", callbackURL)

	// 发送预测请求
	result, err := aiService.Predict(context.Background(), imageURL, taskID)
	if err != nil {
		log.Fatalf("发送预测请求失败: %v", err)
	}

	fmt.Printf("预测请求已发送，初始结果: %+v\n", result)
	fmt.Println("等待AI服务回调...")

	// 等待30秒，让回调有时间处理
	time.Sleep(30 * time.Second)

	// 查询最终结果
	finalResult, err := dbService.GetPredictionResult(taskID)
	if err != nil {
		log.Fatalf("获取预测结果失败: %v", err)
	}

	fmt.Printf("\n最终预测结果:\n")
	fmt.Printf("TaskID: %s\n", finalResult.TaskID)
	fmt.Printf("状态: %v\n", finalResult.PredictionStatus)
	fmt.Printf("模型: %s\n", finalResult.PredictionModel)
	fmt.Printf("是否有缺陷: %v\n", finalResult.HasDefect)
	fmt.Printf("缺陷类型: %s\n", finalResult.DefectType)
	fmt.Printf("置信度: %.2f%%\n", finalResult.Confidence)
} 