package examples

import (
	"fmt"
	"log"
	"mingda_ai_helper/config"
	"mingda_ai_helper/services"
	"time"
)

func MoonrakerExample() {
	// 加载配置
	cfg, err := config.LoadConfig("../../config/config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化日志服务
	logService, err := services.NewLogService(cfg.Logging.Level, cfg.Logging.File)
	if err != nil {
		log.Fatalf("初始化日志服务失败: %v", err)
	}

	// 创建Moonraker客户端
	client := services.NewMoonrakerClient(cfg.Moonraker, logService)

	// 连接到Moonraker
	if err := client.Connect(); err != nil {
		log.Fatalf("连接Moonraker失败: %v", err)
	}
	defer client.Close()

	// 每5秒获取一次打印机状态
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	fmt.Println("开始监控打印机状态...")
	for range ticker.C {
		status, err := client.GetPrinterStatus()
		if err != nil {
			log.Printf("获取打印机状态失败: %v", err)
			continue
		}

		fmt.Printf("\n=== 打印机状态 ===\n")
		fmt.Printf("Webhooks状态: %s\n", status.Webhooks.State)
		fmt.Printf("Webhooks消息: %s\n", status.Webhooks.Message)
		
		fmt.Printf("\n虚拟SD卡:\n")
		fmt.Printf("  - 进度: %.1f%%\n", status.VirtualSdcard.Progress * 100)
		fmt.Printf("  - 是否活动: %v\n", status.VirtualSdcard.IsActive)
		fmt.Printf("  - 文件位置: %d\n", status.VirtualSdcard.FilePosition)

		fmt.Printf("\n打印统计:\n")
		fmt.Printf("  - 文件名: %s\n", status.PrintStats.Filename)
		fmt.Printf("  - 总时长: %.1f秒\n", status.PrintStats.TotalDuration)
		fmt.Printf("  - 打印时长: %.1f秒\n", status.PrintStats.PrintDuration)
		fmt.Printf("  - 状态: %s\n", status.PrintStats.State)
		fmt.Printf("  - 消息: %s\n", status.PrintStats.Message)
		fmt.Printf("===================\n")
	}
} 