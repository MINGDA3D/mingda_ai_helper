package main

import (
	"fmt"
	"log"
	"mingda_ai_helper/config"
	"mingda_ai_helper/models"
	"mingda_ai_helper/services"
	"time"
)

func main() {
	fmt.Println("开始加载配置文件...")
	// 加载配置文件
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}
	fmt.Println("配置文件加载成功")

	// 打印配置信息
	fmt.Printf("\n=== 配置信息 ===\n")
	fmt.Printf("Moonraker配置:\n")
	fmt.Printf("  - 地址: %s:%d\n", cfg.Moonraker.Host, cfg.Moonraker.Port)
	
	fmt.Printf("\nAI服务配置:\n")
	fmt.Printf("  - 本地服务地址: %s\n", cfg.AI.LocalURL)
	fmt.Printf("  - 云端服务地址: %s\n", cfg.AI.CloudURL)
	fmt.Printf("  - 超时时间: %d秒\n", cfg.AI.Timeout)

	fmt.Printf("\n数据库配置:\n")
	fmt.Printf("  - 数据库路径: %s\n", cfg.Database.Path)

	fmt.Printf("\n日志配置:\n")
	fmt.Printf("  - 日志级别: %s\n", cfg.Logging.Level)
	fmt.Printf("  - 日志文件: %s\n", cfg.Logging.File)
	fmt.Printf("  - 单文件大小: %dMB\n", cfg.Logging.MaxSize)
	fmt.Printf("  - 备份数量: %d\n", cfg.Logging.MaxBackups)
	fmt.Printf("  - 保留天数: %d\n", cfg.Logging.MaxAge)
	fmt.Printf("=== 配置信息结束 ===\n\n")

	// 初始化数据库服务
	fmt.Println("开始初始化数据库服务...")
	dbService, err := services.NewDBService(cfg.Database.Path)
	if err != nil {
		log.Fatalf("初始化数据库服务失败: %v", err)
	}
	defer dbService.Close()
	fmt.Println("数据库服务初始化成功")

	// 测试机器信息
	fmt.Println("\n=== 测试机器信息 ===")
	machineInfo := &models.MachineInfo{
		MachineSN:    "TEST001",
		MachineModel: "MingDa-D2",
		AuthToken:    "test-token-123",
	}
	
	if err := dbService.SaveMachineInfo(machineInfo); err != nil {
		log.Printf("保存机器信息失败: %v", err)
	} else {
		fmt.Println("保存机器信息成功")
	}

	if info, err := dbService.GetMachineInfo(); err != nil {
		log.Printf("获取机器信息失败: %v", err)
	} else {
		fmt.Printf("获取机器信息成功: SN=%s, Model=%s\n", info.MachineSN, info.MachineModel)
	}

	// 测试用户设置
	fmt.Println("\n=== 测试用户设置 ===")
	settings := &models.UserSettings{
		EnableAI:            true,
		EnableCloudAI:       false,
		ConfidenceThreshold: 80,
		PauseOnThreshold:    true,
	}

	if err := dbService.SaveUserSettings(settings); err != nil {
		log.Printf("保存用户设置失败: %v", err)
	} else {
		fmt.Println("保存用户设置成功")
	}

	if savedSettings, err := dbService.GetUserSettings(); err != nil {
		log.Printf("获取用户设置失败: %v", err)
	} else {
		fmt.Printf("获取用户设置成功: EnableAI=%v, Threshold=%d\n", 
			savedSettings.EnableAI, savedSettings.ConfidenceThreshold)
	}

	// 测试预测结果
	fmt.Println("\n=== 测试预测结果 ===")
	predictionResult := &models.PredictionResult{
		TaskID:           "TASK001",
		PredictionStatus: models.StatusCompleted,
		PredictionModel:  "local-model-v1",
		HasDefect:        true,
		DefectType:       "stringing",
		Confidence:       95.5,
	}

	if err := dbService.SavePredictionResult(predictionResult); err != nil {
		log.Printf("保存预测结果失败: %v", err)
	} else {
		fmt.Println("保存预测结果成功")
	}

	if result, err := dbService.GetPredictionResult("TASK001"); err != nil {
		log.Printf("获取预测结果失败: %v", err)
	} else {
		fmt.Printf("获取预测结果成功: TaskID=%s, HasDefect=%v, Confidence=%.1f\n",
			result.TaskID, result.HasDefect, result.Confidence)
	}

	// 测试列表查询
	if results, err := dbService.ListPredictionResults(5); err != nil {
		log.Printf("获取预测结果列表失败: %v", err)
	} else {
		fmt.Printf("最近的预测结果数量: %d\n", len(results))
	}

	fmt.Println("\n数据库测试完成")
} 