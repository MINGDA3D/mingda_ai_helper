package services

import (
	"context"
	"fmt"
	"io"
	"mingda_ai_helper/models"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
)

// MonitorService 监控服务
type MonitorService struct {
	moonrakerClient *MoonrakerClient
	aiService       AIService
	cloudAIService  AIService  // 添加云端AI服务
	dbService       *DBService
	logService      *LogService
	
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup

	// 监控间隔
	statusCheckInterval time.Duration
	snapshotInterval   time.Duration

	// AI服务计数器
	aiCounter int
}

// NewMonitorService 创建新的监控服务
func NewMonitorService(
	moonrakerClient *MoonrakerClient,
	localAIService AIService,
	cloudAIService AIService,
	dbService *DBService,
	logService *LogService,
) *MonitorService {
	ctx, cancel := context.WithCancel(context.Background())
	return &MonitorService{
		moonrakerClient:     moonrakerClient,
		aiService:           localAIService,
		cloudAIService:      cloudAIService,
		dbService:           dbService,
		logService:          logService,
		ctx:                 ctx,
		cancel:             cancel,
		statusCheckInterval: time.Minute,      // 1分钟检查一次状态
		snapshotInterval:   time.Minute * 3,   // 3分钟拍照一次
		aiCounter:          0,                 // 初始化AI计数器
	}
}

// Start 启动监控服务
func (s *MonitorService) Start() error {
	s.logService.Info("监控服务启动")
	
	// 连接到 Moonraker
	if err := s.moonrakerClient.Connect(); err != nil {
		s.logService.Error("连接Moonraker失败", zap.Error(err))
		return err
	}
	
	// 启动监控协程
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.monitor()
	}()

	return nil
}

// Stop 停止监控服务
func (s *MonitorService) Stop() {
	s.logService.Info("监控服务停止")
	s.cancel()
	s.moonrakerClient.Close()
	s.wg.Wait()
}

// getLocalIP 获取本地IP地址
func (s *MonitorService) getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", fmt.Errorf("无法获取局域网IP地址")
}

// getSnapshot 获取摄像头快照
func (s *MonitorService) getSnapshot(url string) (string, error) {
	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 发送GET请求
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("获取快照失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("获取快照失败，状态码: %d", resp.StatusCode)
	}

	// 生成保存路径
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("获取用户主目录失败: %v", err)
	}
	
	timestamp := time.Now().Format("20060102_150405")
	savePath := filepath.Join(homeDir, "printer_data", "ai_snapshots", fmt.Sprintf("snapshot_%s.jpg", timestamp))

	// 创建保存目录
	if err := os.MkdirAll(filepath.Dir(savePath), 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %v", err)
	}

	// 创建文件
	file, err := os.Create(savePath)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	// 将响应内容写入文件
	if _, err := io.Copy(file, resp.Body); err != nil {
		return "", fmt.Errorf("保存图片失败: %v", err)
	}

	return savePath, nil
}

// monitor 监控打印状态
func (s *MonitorService) monitor() {
	statusTicker := time.NewTicker(s.statusCheckInterval)
	snapshotTicker := time.NewTicker(s.snapshotInterval)
	defer statusTicker.Stop()
	defer snapshotTicker.Stop()

	var lastSnapshotTime time.Time

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-statusTicker.C:
			// 获取用户设置
			settings, err := s.dbService.GetUserSettings()
			if err != nil {
				s.logService.Error("获取用户设置失败", zap.Error(err))
				continue
			}

			// 如果AI功能未启用，跳过检查
			if !settings.EnableAI {
				continue
			}

			// 获取打印状态
			status, err := s.moonrakerClient.GetPrinterStatus()
			if err != nil {
				s.logService.Error("获取打印机状态失败", zap.Error(err))
				continue
			}

			// 如果不在打印状态，跳过后续操作
			if !status.IsPrinting {
				continue
			}

			s.logService.Info("打印机正在打印中，AI监控已启用")

		case <-snapshotTicker.C:
			// 检查是否需要拍照
			if time.Since(lastSnapshotTime) < s.snapshotInterval {
				continue
			}

			// 获取用户设置
			settings, err := s.dbService.GetUserSettings()
			if err != nil {
				s.logService.Error("获取用户设置失败", zap.Error(err))
				continue
			}

			// 如果AI功能未启用，跳过拍照
			if !settings.EnableAI {
				continue
			}

			// 获取打印状态
			status, err := s.moonrakerClient.GetPrinterStatus()
			if err != nil {
				s.logService.Error("获取打印机状态失败", zap.Error(err))
				continue
			}

			// 如果不在打印状态，跳过拍照
			if !status.IsPrinting {
				continue
			}

			// 获取本地IP
			localIP, err := s.getLocalIP()
			if err != nil {
				s.logService.Error("获取本地IP失败", zap.Error(err))
				continue
			}

			// 构造摄像头URL
			cameraURL := fmt.Sprintf("http://%s/webcam/?action=snapshot", localIP)
			
			// 获取快照
			s.logService.Info("开始获取摄像头快照", zap.String("url", cameraURL))
			savePath, err := s.getSnapshot(cameraURL)
			if err != nil {
				s.logService.Error("获取快照失败", zap.Error(err))
				continue
			}

			// 更新最后拍照时间
			lastSnapshotTime = time.Now()

			// 选择AI服务（每4次循环使用1次云端服务）
			var currentAIService AIService
			useCloudAI := s.aiCounter%4 == 3 && settings.EnableCloudAI
			if useCloudAI {
				currentAIService = s.cloudAIService
				s.logService.Info("使用云端AI服务")
			} else {
				currentAIService = s.aiService
				s.logService.Info("使用本地AI服务")
			}
			s.aiCounter++

			// 调用AI服务进行预测
			s.logService.Info("开始AI预测", 
				zap.String("image_path", savePath),
				zap.Bool("use_cloud", useCloudAI))

			// 创建初始预测结果
			result := &models.PredictionResult{
				TaskID:           fmt.Sprintf("PT%s", time.Now().Format("20060102150405")),
				PredictionStatus: models.StatusProcessing,
				PredictionModel:  "local_ai",
			}

			// 保存初始预测结果
			if err := s.dbService.SavePredictionResult(result); err != nil {
				s.logService.Error("保存初始预测结果失败", zap.Error(err))
				continue
			}

			// 发送预测请求（结果会通过回调处理）
			if _, err := currentAIService.PredictWithFile(s.ctx, savePath); err != nil {
				s.logService.Error("AI预测失败", zap.Error(err))
				continue
			}

			s.logService.Info("预测请求已发送，等待回调处理",
				zap.String("task_id", result.TaskID))
		}
	}
} 