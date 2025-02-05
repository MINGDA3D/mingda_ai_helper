package main

import (
	"context"
	"fmt"
	"log"
	"mingda_ai_helper/config"
	"mingda_ai_helper/models"
	"mingda_ai_helper/services"
	"time"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig("../config/config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化日志服务
	logService, err := services.NewLogService(cfg.Logging.Level, cfg.Logging.File)
	if err != nil {
		log.Fatalf("初始化日志服务失败: %v", err)
	}

	// 初始化数据库服务
	dbService, err := services.NewDBService(cfg.Database.File)
	if err != nil {
		log.Fatalf("初始化数据库服务失败: %v", err)
	}

	// 初始化AI服务
	aiService := services.NewAIService(cfg.AI, dbService, logService)

	// 生成设备SN
	timestamp := time.Now().Format("150405")
	deviceSN := fmt.Sprintf("M1P2004A1%s", timestamp)
	deviceModel := "MD-400D"

	fmt.Printf("\n=== 开始设备认证流程 ===\n")
	fmt.Printf("设备SN: %s\n", deviceSN)
	fmt.Printf("设备型号: %s\n", deviceModel)

	// 1. 注册设备
	fmt.Printf("\n1. 注册设备...\n")
	secret, err := aiService.RegisterDevice(context.Background(), deviceSN, deviceModel)
	if err != nil {
		log.Fatalf("注册设备失败: %v", err)
	}
	fmt.Printf("设备密钥: %s\n", secret)

	// 2. 设备认证
	fmt.Printf("\n2. 设备认证...\n")
	token, err := aiService.AuthDevice(context.Background(), deviceSN, secret)
	if err != nil {
		log.Fatalf("设备认证失败: %v", err)
	}
	fmt.Printf("认证Token: %s\n", token)

	// 保存设备信息到数据库
	err = dbService.SaveMachineInfo(&models.MachineInfo{
		MachineSN:    deviceSN,
		MachineModel: deviceModel,
		AuthToken:    token,
	})
	if err != nil {
		log.Fatalf("保存设备信息失败: %v", err)
	}

	// 3. 刷新Token
	fmt.Printf("\n3. 刷新Token...\n")
	newToken, err := aiService.RefreshToken(context.Background(), token)
	if err != nil {
		log.Fatalf("刷新Token失败: %v", err)
	}
	fmt.Printf("新Token: %s\n", newToken)

	// 更新数据库中的Token
	err = dbService.UpdateMachineToken(deviceSN, newToken)
	if err != nil {
		log.Fatalf("更新Token失败: %v", err)
	}

	fmt.Printf("\n=== 设备认证流程完成 ===\n")
} 
