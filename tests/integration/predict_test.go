package integration

import (
	"context"
	"fmt"
	"log"
	"mingda_ai_helper/config"
	"mingda_ai_helper/services"
	"time"
)

func TestPredict() {
	// 加载配置
	cfg, err := config.LoadConfig("../../config/config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化数据库服务
	dbService, err := services.NewDBService(cfg.Database.Path)
	if err != nil {
		log.Fatalf("初始化数据库服务失败: %v", err)
	}

	// 初始化日志服务
	logService, err := services.NewLogService(cfg.Logging.Level, cfg.Logging.File)
	if err != nil {
		log.Fatalf("初始化日志服务失败: %v", err)
	}

	// 创建本地AI服务实例
	callbackURL := fmt.Sprintf("http://%s:8081/api/v1/ai/callback", cfg.Moonraker.Host)
	localAI := services.NewLocalAIService(cfg.AI.LocalURL, callbackURL, dbService)

	// 启动回调服务器
	startCallbackServer(dbService, logService)

	// 测试本地预测
	fmt.Println("\n=== 测试本地预测 ===")
	localResult, err := localAI.Predict(
		context.Background(),
		"http://example.com/test.jpg",
		"TEST001",
	)
	if err != nil {
		log.Printf("\033[31m本地预测失败: %v\033[0m\n", err)
	} else {
		fmt.Printf("\033[32m本地预测成功，任务ID: %s\033[0m\n", localResult.TaskID)
	}

	// 创建云端AI服务实例
	cloudAI := services.NewCloudAIService(cfg.AI.CloudURL, dbService)

	// 测试云端预测
	fmt.Println("\n=== 测试云端预测 ===")
	cloudResult, err := cloudAI.PredictWithFile(
		context.Background(),
		"test_data/test_image.jpg",
	)
	if err != nil {
		log.Printf("\033[31m云端预测失败: %v\033[0m\n", err)
	} else {
		fmt.Printf("\033[32m云端预测成功，任务ID: %s\033[0m\n", cloudResult.TaskID)
	}

	// 等待回调处理
	fmt.Println("\n等待回调处理...")
	time.Sleep(10 * time.Second)
} 