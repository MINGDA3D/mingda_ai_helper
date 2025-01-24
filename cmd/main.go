package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"mingda_ai_helper/config"
	"mingda_ai_helper/models"
	"mingda_ai_helper/services"
)

func main() {
	fmt.Println("开始加载配置文件...")
	// 加载配置文件
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}
	fmt.Println("配置文件加载成功")

	fmt.Println("开始初始化日志服务...")
	// 初始化日志服务
	logService, err := services.NewLogService(cfg.Logging.Level, cfg.Logging.File)
	if err != nil {
		log.Fatalf("初始化日志服务失败: %v", err)
	}
	defer logService.Sync()
	fmt.Println("日志服务初始化成功")

	fmt.Println("开始初始化数据库服务...")
	// 初始化数据库服务
	dbService, err := services.NewDBService(cfg.Database.Path)
	if err != nil {
		logService.Error("初始化数据库服务失败", "error", err)
		os.Exit(1)
	}
	defer dbService.Close()
	fmt.Println("数据库服务初始化成功")

	fmt.Println("开始获取设备信息...")
	// 获取设备信息
	machineInfo, err := models.GetMachineInfo(dbService.DB())
	if err != nil {
		logService.Error("获取设备信息失败", "error", err)
		os.Exit(1)
	}
	fmt.Println("设备信息获取成功")

	// 打印配置信息
	logService.Info("设备信息",
		"型号", machineInfo.MachineModel,
		"SN码", machineInfo.MachineSN)
	
	logService.Info("Moonraker配置",
		"地址", fmt.Sprintf("%s:%d", cfg.Moonraker.Host, cfg.Moonraker.Port))
	
	logService.Info("AI服务配置",
		"本地AI", cfg.AI.Local.Enabled,
		"云端AI", cfg.AI.Cloud.Enabled)

	fmt.Println("开始初始化AI服务...")
	// 初始化AI服务
	aiService, err := services.NewAIService(cfg.AI, machineInfo.AuthToken)
	if err != nil {
		logService.Error("初始化AI服务失败", "error", err)
		os.Exit(1)
	}
	fmt.Println("AI服务初始化成功")

	fmt.Println("开始初始化Moonraker客户端...")
	// 初始化Moonraker客户端
	moonrakerClient := services.NewMoonrakerClient(cfg.Moonraker, logService)
	fmt.Println("Moonraker客户端初始化成功")

	fmt.Println("开始初始化监控服务...")
	// 初始化监控服务
	monitorService := services.NewMonitorService(moonrakerClient, aiService, dbService, logService)
	fmt.Println("监控服务初始化成功")

	fmt.Println("开始启动监控服务...")
	// 启动监控
	if err := monitorService.Start(); err != nil {
		logService.Error("启动监控服务失败", "error", err)
		os.Exit(1)
	}
	fmt.Println("监控服务启动成功")

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println("程序已启动，等待中断信号...")
	<-sigChan

	// 优雅关闭
	logService.Info("收到退出信号，正在关闭服务...")
	monitorService.Stop()
} 