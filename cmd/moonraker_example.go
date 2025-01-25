package main

import (
	"fmt"
	"time"
	"path/filepath"
	"os"

	"go.uber.org/zap"
	"mingda_ai_helper/config"
	"mingda_ai_helper/services"
)

func main() {
	// 创建日志目录
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Printf("创建日志目录失败: %v\n", err)
		return
	}

	// 创建日志服务
	logService, err := services.NewLogService("debug", filepath.Join(logDir, "moonraker.log"))
	if err != nil {
		fmt.Printf("初始化日志服务失败: %v\n", err)
		return
	}

	// 创建Moonraker配置
	cfg := config.MoonrakerConfig{
		Host: "localhost", // 替换为你的Moonraker服务器地址
		Port: 7125,       // 替换为你的Moonraker服务器端口
	}

	// 创建Moonraker客户端
	client := services.NewMoonrakerClient(cfg, logService)

	// 连接到Moonraker服务器
	if err := client.Connect(); err != nil {
		fmt.Printf("连接Moonraker失败: %v\n", err)
		return
	}
	defer client.Close()

	// 获取打印机状态
	status, err := client.GetPrinterStatus()
	if err != nil {
		fmt.Printf("获取打印机状态失败: %v\n", err)
		return
	}

	// 打印状态信息
	fmt.Printf("打印机连接状态: %s\n", status.Webhooks.State)
	fmt.Printf("打印机消息: %s\n", status.Webhooks.Message)
	
	// 打印进度信息
	fmt.Printf("打印进度: %.1f%%\n", status.VirtualSdcard.Progress*100)
	fmt.Printf("是否正在打印: %v\n", status.VirtualSdcard.IsActive)
	fmt.Printf("文件位置: %d\n", status.VirtualSdcard.FilePosition)

	// 打印作业信息
	fmt.Printf("当前文件: %s\n", status.PrintStats.Filename)
	fmt.Printf("打印状态: %s\n", status.PrintStats.State)
	fmt.Printf("打印时长: %.1f秒\n", status.PrintStats.PrintDuration)
	fmt.Printf("总时长: %.1f秒\n", status.PrintStats.TotalDuration)
	if status.PrintStats.Message != "" {
		fmt.Printf("状态消息: %s\n", status.PrintStats.Message)
	}

	// 如果正在打印，尝试暂停打印
	if status.PrintStats.State == "printing" {
		fmt.Println("\n尝试暂停打印...")
		if err := client.PausePrint(); err != nil {
			fmt.Printf("暂停打印失败: %v\n", err)
			return
		}
		fmt.Println("打印已暂停")

		// 等待2秒后再次获取状态
		time.Sleep(2 * time.Second)
		
		status, err = client.GetPrinterStatus()
		if err != nil {
			fmt.Printf("获取打印机状态失败: %v\n", err)
			return
		}
		fmt.Printf("当前状态: %s\n", status.PrintStats.State)
	}
} 