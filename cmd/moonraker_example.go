package main

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"mingda_ai_helper/config"
	"mingda_ai_helper/services"
)

func main() {
	// 创建logger
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// 创建日志服务
	logService := services.NewLogService(logger)

	// 创建Moonraker配置
	cfg := config.MoonrakerConfig{
		Host: "localhost", // 替换为你的Moonraker服务器地址
		Port: 7125,       // 替换为你的Moonraker服务器端口
	}

	// 创建Moonraker客户端
	client := services.NewMoonrakerClient(cfg, logService)

	// 连接到Moonraker服务器
	if err := client.Connect(); err != nil {
		logService.Error("连接Moonraker失败", zap.Error(err))
		return
	}
	defer client.Close()

	// 获取打印机状态
	status, err := client.GetPrinterStatus()
	if err != nil {
		logService.Error("获取打印机状态失败", zap.Error(err))
		return
	}

	// 打印状态信息
	fmt.Printf("打印机状态: %s\n", status.State)
	fmt.Printf("打印机消息: %s\n", status.Message)
	fmt.Printf("喷头温度: %.1f°C (目标: %.1f°C)\n", 
		status.Temperature.Tool0.Actual, 
		status.Temperature.Tool0.Target)
	fmt.Printf("热床温度: %.1f°C (目标: %.1f°C)\n",
		status.Temperature.Bed.Actual,
		status.Temperature.Bed.Target)

	// 获取打印进度
	progress, err := client.GetPrintProgress()
	if err != nil {
		logService.Error("获取打印进度失败", zap.Error(err))
		return
	}
	fmt.Printf("打印进度: %.1f%%\n", progress*100)

	// 如果正在打印，尝试暂停打印
	if status.State == "printing" {
		fmt.Println("\n尝试暂停打印...")
		if err := client.PausePrint(); err != nil {
			logService.Error("暂停打印失败", zap.Error(err))
			return
		}
		fmt.Println("打印已暂停")

		// 等待2秒后再次获取状态
		time.Sleep(2 * time.Second)
		
		status, err = client.GetPrinterStatus()
		if err != nil {
			logService.Error("获取打印机状态失败", zap.Error(err))
			return
		}
		fmt.Printf("当前状态: %s\n", status.State)
	}
} 