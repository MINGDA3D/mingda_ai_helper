package main

import (
	"fmt"
	"log"
	"mingda_ai_helper/config"
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
	fmt.Printf("=== 配置信息结束 ===\n")
} 