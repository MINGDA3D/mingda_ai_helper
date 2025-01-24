package services

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"mingda_ai_helper/config"
)

// MoonrakerClient Moonraker客户端
type MoonrakerClient struct {
	config     config.MoonrakerConfig
	wsConn     *websocket.Conn
	logService *LogService
	
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	mu         sync.Mutex
}

// NewMoonrakerClient 创建新的Moonraker客户端
func NewMoonrakerClient(cfg config.MoonrakerConfig, logService *LogService) *MoonrakerClient {
	ctx, cancel := context.WithCancel(context.Background())
	return &MoonrakerClient{
		config:     cfg,
		logService: logService,
		ctx:        ctx,
		cancel:     cancel,
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
func (c *MoonrakerClient) GetPrinterStatus() error {
	// TODO: 实现获取打印机状态的逻辑
	return nil
}

// PausePrint 暂停打印
func (c *MoonrakerClient) PausePrint() error {
	// TODO: 实现暂停打印的逻辑
	return nil
} 