package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"mingda_ai_helper/config"
	"mingda_ai_helper/services"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

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

func getSnapshot(url string, savePath string) error {
	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 发送GET请求
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("获取快照失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("获取快照失败，状态码: %d", resp.StatusCode)
	}

	// 创建保存目录
	if err := os.MkdirAll(filepath.Dir(savePath), 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// 创建文件
	file, err := os.Create(savePath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	// 将响应内容写入文件
	if _, err := io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("保存图片失败: %v", err)
	}

	return nil
}

func startCallbackServer(dbService *services.DBService, logService *services.LogService) {
	// 设置路由
	router := handlers.SetupRouter(nil, dbService, logService)

	// 在新的goroutine中启动服务器
	go func() {
		fmt.Println("启动回调服务器在 :8081 端口")
		if err := router.Run(":8081"); err != nil {
			log.Fatalf("启动回调服务器失败: %v", err)
		}
	}()

	// 等待服务器启动
	time.Sleep(1 * time.Second)
}
func main() {
	// 加载配置
	cfg, err := config.LoadConfig("../config/config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化数据库服务
	dbService, err := services.NewDBService(cfg.Database.Path)
	if err != nil {
		log.Fatalf("初始化数据库服务失败: %v", err)
	}

	// 启动回调服务器
	startCallbackServer(dbService, logService)

	// 创建云端AI服务实例
	cloudAI := services.NewCloudAIService(cfg.AI.CloudURL, dbService)

	// 获取本地IP地址
	localIP, err := getLocalIP()
	if err != nil {
		log.Fatalf("\033[31m获取本地IP失败: %v\033[0m\n", err)
	}

	// 构造摄像头URL和保存路径
	cameraURL := fmt.Sprintf("http://%s/webcam/?action=snapshot", localIP)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("\033[31m获取用户主目录失败: %v\033[0m\n", err)
	}
	
	// 生成保存路径
	timestamp := time.Now().Format("20060102_150405")
	savePath := filepath.Join(homeDir, "printer_data", "ai_snapshots", fmt.Sprintf("snapshot_%s.jpg", timestamp))

	fmt.Printf("\033[33m正在从摄像头获取图片: %s\033[0m\n", cameraURL)
	
	// 获取并保存快照
	if err := getSnapshot(cameraURL, savePath); err != nil {
		log.Fatalf("\033[31m获取快照失败: %v\033[0m\n", err)
	}

	fmt.Printf("\033[32m快照已保存到: %s\033[0m\n", savePath)

	// 调用预测接口
	fmt.Printf("\033[33m开始处理图片...\033[0m\n")
	result, err := cloudAI.PredictWithFile(context.Background(), savePath)
	if err != nil {
		log.Fatalf("\033[31m预测失败: %v\033[0m\n", err)
	}

	fmt.Printf("\033[32m预测成功!\033[0m\n")
	fmt.Printf("任务ID: %s\n", result.TaskID)
	fmt.Printf("预测状态: %d\n", result.PredictionStatus)
	fmt.Printf("预测模型: %s\n", result.PredictionModel)

		// 等待30秒，让回调有时间处理
		time.Sleep(30 * time.Second)

		// 查询最终结果
		finalResult, err := dbService.GetPredictionResult(taskID)
		if err != nil {
			log.Fatalf("获取预测结果失败: %v", err)
		}
	
		fmt.Printf("\n=== 最终预测结果 ===\n")
		printJSON("结果", finalResult)
	
		// 打印状态说明
		fmt.Printf("\n状态说明：\n")
		fmt.Printf("0: 等待处理\n")
		fmt.Printf("1: 处理中\n")
		fmt.Printf("2: 处理完成\n")
} 