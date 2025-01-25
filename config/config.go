package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config 总配置结构
type Config struct {
	Moonraker MoonrakerConfig `mapstructure:"moonraker"`
	AI        AIConfig        `mapstructure:"ai"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Logging   LoggingConfig   `mapstructure:"logging"`
}

// MoonrakerConfig Moonraker连接配置
type MoonrakerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// AIConfig AI服务配置
type AIConfig struct {
	LocalURL  string `mapstructure:"local_url"`
	CloudURL  string `mapstructure:"cloud_url"`
	Timeout   int    `mapstructure:"timeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	File       string `mapstructure:"file"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
}

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	fmt.Println("开始设置配置文件路径...")
	
	// 设置配置文件的查找路径
	viper.AddConfigPath("config")         // 相对于工作目录的config目录
	viper.AddConfigPath(".")              // 当前工作目录
	viper.SetConfigName("config")         // 配置文件名（不带扩展名）
	viper.SetConfigType("yaml")           // 配置文件类型

	fmt.Println("尝试读取配置文件...")
	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}
	fmt.Printf("成功读取配置文件: %s\n", viper.ConfigFileUsed())

	fmt.Println("开始解析配置文件...")
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	fmt.Println("开始验证配置项...")
	// 验证必要的配置项
	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	fmt.Println("开始创建必要目录...")
	// 创建必要的目录
	if err := createRequiredDirectories(&config); err != nil {
		return nil, err
	}

	fmt.Println("配置加载完成")
	return &config, nil
}

// validateConfig 验证配置项
func validateConfig(config *Config) error {
	// 验证Moonraker配置
	if config.Moonraker.Port <= 0 || config.Moonraker.Port > 65535 {
		return fmt.Errorf("无效的Moonraker端口号: %d", config.Moonraker.Port)
	}

	// 验证AI配置
	if config.AI.Timeout <= 0 {
		return fmt.Errorf("无效的AI超时时间: %d", config.AI.Timeout)
	}

	return nil
}

// createRequiredDirectories 创建必要的目录
func createRequiredDirectories(config *Config) error {
	// 创建数据库目录
	if err := os.MkdirAll(filepath.Dir(config.Database.Path), 0755); err != nil {
		return fmt.Errorf("创建数据库目录失败: %v", err)
	}

	// 创建日志目录
	if err := os.MkdirAll(filepath.Dir(config.Logging.File), 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %v", err)
	}

	return nil
} 