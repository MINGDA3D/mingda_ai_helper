package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
    "go.uber.org/zap"
	"github.com/gorilla/websocket"
	"mingda_ai_helper/config"
)

// PrinterStatus 打印机状态
type PrinterStatus struct {
	Webhooks struct {
		State   string `json:"state"`
		Message string `json:"message"`
	} `json:"webhooks"`
	VirtualSdcard struct {
		Progress          float64 `json:"progress"`
		IsActive         bool    `json:"is_active"`
		FilePosition     int     `json:"file_position"`
	} `json:"virtual_sdcard"`
	PrintStats struct {
		Filename     string  `json:"filename"`
		TotalDuration float64 `json:"total_duration"`
		PrintDuration float64 `json:"print_duration"`
		State        string  `json:"state"`
		Message      string  `json:"message"`
	} `json:"print_stats"`
}

// IsPrinting 判断打印机是否正在打印
func (s *PrinterStatus) IsPrinting() bool {
	return s.PrintStats.State == "printing" && s.VirtualSdcard.IsActive
}

// MoonrakerClient Moonraker客户端
type MoonrakerClient struct {
	config     config.MoonrakerConfig
	wsConn     *websocket.Conn
	httpClient *http.Client
	logService *LogService
	baseURL    string
	
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	mu         sync.Mutex
}

// NewMoonrakerClient 创建新的Moonraker客户端
func NewMoonrakerClient(cfg config.MoonrakerConfig, logService *LogService) *MoonrakerClient {
	ctx, cancel := context.WithCancel(context.Background())
	baseURL := fmt.Sprintf("http://%s:%d", cfg.Host, cfg.Port)
	return &MoonrakerClient{
		config:     cfg,
		logService: logService,
		ctx:        ctx,
		cancel:     cancel,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: baseURL,
	}
}

// Connect 连接到Moonraker WebSocket服务
func (c *MoonrakerClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.wsConn != nil {
		return nil
	}

	// 构建WebSocket URL
	u := url.URL{
		Scheme: "ws",
		Host:   fmt.Sprintf("%s:%d", c.config.Host, c.config.Port),
		Path:   "/websocket",
	}

	c.logService.Info("正在连接到Moonraker", zap.String("url", u.String()))

	// 建立连接
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("连接Moonraker失败: %v", err)
	}

	c.wsConn = conn
	c.logService.Info("成功连接到Moonraker")

	// 启动接收消息的协程
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.readPump()
	}()

	// 启动心跳协程
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.heartbeat()
	}()

	return nil
}

// Close 关闭连接
func (c *MoonrakerClient) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.wsConn != nil {
		c.cancel()
		c.wsConn.Close()
		c.wsConn = nil
		c.wg.Wait()
	}
}

// readPump 持续读取WebSocket消息
func (c *MoonrakerClient) readPump() {
	defer func() {
		c.mu.Lock()
		if c.wsConn != nil {
			c.wsConn.Close()
			c.wsConn = nil
		}
		c.mu.Unlock()
	}()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			_, message, err := c.wsConn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					c.logService.Error("读取WebSocket消息失败", zap.Error(err))
				}
				return
			}
			// TODO: 处理接收到的消息
			c.logService.Debug("收到消息", zap.ByteString("message", message))
		}
	}
}

// heartbeat 定期发送心跳
func (c *MoonrakerClient) heartbeat() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.mu.Lock()
			if c.wsConn != nil {
				if err := c.wsConn.WriteMessage(websocket.PingMessage, nil); err != nil {
					c.logService.Error("发送心跳失败", zap.Error(err))
				}
			}
			c.mu.Unlock()
		}
	}
}

// GetPrinterStatus 获取打印机状态
func (c *MoonrakerClient) GetPrinterStatus() (*PrinterStatus, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/printer/objects/query?webhooks&virtual_sdcard&print_stats")
	if err != nil {
		return nil, fmt.Errorf("获取打印机状态失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取打印机状态失败，状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var result struct {
		Result struct {
			Status struct {
				Webhooks      struct {
					State   string `json:"state"`
					Message string `json:"message"`
				} `json:"webhooks"`
				VirtualSdcard struct {
					Progress      float64 `json:"progress"`
					IsActive     bool    `json:"is_active"`
					FilePosition int     `json:"file_position"`
				} `json:"virtual_sdcard"`
				PrintStats   struct {
					Filename     string  `json:"filename"`
					TotalDuration float64 `json:"total_duration"`
					PrintDuration float64 `json:"print_duration"`
					State        string  `json:"state"`
					Message      string  `json:"message"`
				} `json:"print_stats"`
			} `json:"status"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	status := &PrinterStatus{
		Webhooks:      result.Result.Status.Webhooks,
		VirtualSdcard: result.Result.Status.VirtualSdcard,
		PrintStats:    result.Result.Status.PrintStats,
	}

	return status, nil
}

// PausePrint 暂停打印
func (c *MoonrakerClient) PausePrint() error {
	req, err := http.NewRequest("POST", c.baseURL+"/printer/print/pause", nil)
	if err != nil {
		return fmt.Errorf("创建暂停打印请求失败: %v", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("发送暂停打印请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("暂停打印失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

// GetPrintProgress 获取打印进度
func (c *MoonrakerClient) GetPrintProgress() (float64, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/printer/objects/query?print_stats")
	if err != nil {
		return 0, fmt.Errorf("获取打印进度失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("获取打印进度失败，状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("读取响应失败: %v", err)
	}

	var result struct {
		Result struct {
			Status struct {
				Progress float64 `json:"progress"`
			} `json:"status"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("解析响应失败: %v", err)
	}

	return result.Result.Status.Progress, nil
} 