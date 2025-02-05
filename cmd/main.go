package main

import (
	"fmt"
	"log"
	"mingda_ai_helper/config"
	"mingda_ai_helper/handlers"
	"mingda_ai_helper/models"
	"mingda_ai_helper/services"
	"os"
	"path/filepath"
)

// 确保数据库目录存在
func ensureDBDirectory(dbPath string) error {
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("创建数据库目录失败: %v", err)
	}
	return nil
}

// 测试机器信息
func testMachineInfo(dbService *services.DBService) error {
	fmt.Println("\n=== 测试机器信息 ===")
	machineInfo := &models.MachineInfo{
		MachineSN:    "M1P2004A1154004",
		MachineModel: "MD-400D",
		AuthToken:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkZXZpY2VfaWQiOjExNDcsImRldmljZV9zbiI6Ik0xUDIwMDRBMTE1NDAwNCIsImV4cCI6MTczNzg3NzIwNSwiaWF0IjoxNzM3NzkwODA1LCJpc3MiOiJtaW5nZGEtY2xvdWQifQ.Gf6DjPR0w1boT0TtWEyMuUTXCoJOpMyUDj0-nWw3mDM",

	}
	
	if err := dbService.SaveMachineInfo(machineInfo); err != nil {
		return fmt.Errorf("保存机器信息失败: %v", err)
	}
	fmt.Println("保存机器信息成功")

	info, err := dbService.GetMachineInfo()
	if err != nil {
		return fmt.Errorf("获取机器信息失败: %v", err)
	}
	fmt.Printf("获取机器信息成功: SN=%s, Model=%s\n", info.MachineSN, info.MachineModel)
	return nil
}

// 测试用户设置
func testUserSettings(dbService *services.DBService) error {
	fmt.Println("\n=== 测试用户设置 ===")
	settings := &models.UserSettings{
		EnableAI:            true,
		EnableCloudAI:       false,
		ConfidenceThreshold: 80,
		PauseOnThreshold:    true,
	}

	if err := dbService.SaveUserSettings(settings); err != nil {
		return fmt.Errorf("保存用户设置失败: %v", err)
	}
	fmt.Println("保存用户设置成功")

	savedSettings, err := dbService.GetUserSettings()
	if err != nil {
		return fmt.Errorf("获取用户设置失败: %v", err)
	}
	fmt.Printf("获取用户设置成功: EnableAI=%v, Threshold=%d\n", 
		savedSettings.EnableAI, savedSettings.ConfidenceThreshold)
	return nil
}

// 测试预测结果
func testPredictionResults(dbService *services.DBService) error {
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
		return fmt.Errorf("保存预测结果失败: %v", err)
	}
	fmt.Println("保存预测结果成功")

	result, err := dbService.GetPredictionResult("TASK001")
	if err != nil {
		return fmt.Errorf("获取预测结果失败: %v", err)
	}
	fmt.Printf("获取预测结果成功: TaskID=%s, HasDefect=%v, Confidence=%.1f\n",
		result.TaskID, result.HasDefect, result.Confidence)

	results, err := dbService.ListPredictionResults(5)
	if err != nil {
		return fmt.Errorf("获取预测结果列表失败: %v", err)
	}
	fmt.Printf("最近的预测结果数量: %d\n", len(results))
	return nil
}

func main() {
	fmt.Println("开始加载配置文件...")
	// 加载配置文件
	cfg, err := config.LoadConfig("config/config.yaml")
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

	// 确保数据库目录存在
	fmt.Println("检查数据库目录...")
	if err := ensureDBDirectory(cfg.Database.Path); err != nil {
		log.Fatalf("创建数据库目录失败: %v", err)
	}
	fmt.Println("数据库目录检查完成")

	// 初始化日志服务
	fmt.Println("初始化日志服务...")
	logService, err := services.NewLogService(cfg.Logging.Level, cfg.Logging.File)
	if err != nil {
		log.Fatalf("初始化日志服务失败: %v", err)
	}
	fmt.Println("日志服务初始化成功")

	// 初始化数据库服务
	fmt.Println("开始初始化数据库服务...")
	dbService, err := services.NewDBService(cfg.Database.Path)
	if err != nil {
		log.Fatalf("初始化数据库服务失败: %v", err)
	}

	// 确保在程序退出时关闭数据库连接
	sqlDB, err := dbService.DB().DB()
	if err != nil {
		log.Fatalf("获取数据库实例失败: %v", err)
	}
	defer sqlDB.Close()
	fmt.Println("数据库服务初始化成功")

	// 初始化Moonraker客户端
	fmt.Println("初始化Moonraker客户端...")
	moonrakerClient := services.NewMoonrakerClient(cfg.Moonraker, logService)
	fmt.Println("Moonraker客户端初始化成功")

	// 初始化本地AI服务
	fmt.Println("初始化本地AI服务...")
	callbackURL := fmt.Sprintf("http://%s:%d/api/v1/ai/callback", cfg.Moonraker.Host, 8081)
	aiService := services.NewLocalAIService(cfg.AI.LocalURL, callbackURL, dbService)
	fmt.Println("本地AI服务初始化成功")

	// 初始化云端AI服务
	fmt.Println("初始化云端AI服务...")
	cloudAIService := services.NewCloudAIService(cfg.AI.CloudURL, dbService)
	fmt.Println("云端AI服务初始化成功")

	// 初始化监控服务
	fmt.Println("初始化监控服务...")
	monitorService := services.NewMonitorService(moonrakerClient, aiService, cloudAIService, dbService, logService)
	if err := monitorService.Start(); err != nil {
		log.Fatalf("启动监控服务失败: %v", err)
	}
	fmt.Println("监控服务启动成功")

	// 设置HTTP路由
	fmt.Println("设置HTTP路由...")
	router := handlers.SetupRouter(aiService, dbService, logService)
	
	fmt.Println("HTTP路由设置完成")

	// 启动HTTP服务器
	fmt.Println("启动HTTP服务器...")
	if err := router.Run(":8081"); err != nil {
		log.Fatalf("启动HTTP服务器失败: %v", err)
	}
} 