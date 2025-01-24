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
	Local  LocalAIConfig  `mapstructure:"local"`
	Cloud  CloudAIConfig  `mapstructure:"cloud"`
}

// LocalAIConfig 本地AI配置
type LocalAIConfig struct {
	Enabled   bool   `mapstructure:"enabled"`
	ModelPath string `mapstructure:"model_path"`
}

// CloudAIConfig 云端AI配置
type CloudAIConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Endpoint string `mapstructure:"endpoint"`
	Timeout  int    `mapstructure:"timeout"`
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
func LoadConfig() (*Config, error) {
	// 获取可执行文件所在目录
	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("获取可执行文件路径失败: %v", err)
	}
	exeDir := filepath.Dir(exePath)

	// 设置配置文件的查找路径
	viper.AddConfigPath(exeDir)           // 可执行文件目录
	viper.AddConfigPath("config")         // 相对于工作目录的config目录
	viper.AddConfigPath("../config")      // 上级目录的config目录
	viper.AddConfigPath(".")              // 当前工作目录

	viper.SetConfigName("config")         // 配置文件名（不带扩展名）
	viper.SetConfigType("yaml")           // 配置文件类型

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 验证必要的配置项
	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	// 创建必要的目录
	if err := createRequiredDirectories(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// validateConfig 验证配置项
func validateConfig(config *Config) error {
	// 验证Moonraker配置
	if config.Moonraker.Port <= 0 || config.Moonraker.Port > 65535 {
		return fmt.Errorf("无效的Moonraker端口号: %d", config.Moonraker.Port)
	}

	// 验证AI配置
	if !config.AI.Local.Enabled && !config.AI.Cloud.Enabled {
		return fmt.Errorf("本地AI和云端AI至少需要启用一个")
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

	// 如果启用了本地AI，创建模型目录
	if config.AI.Local.Enabled {
		if err := os.MkdirAll(config.AI.Local.ModelPath, 0755); err != nil {
			return fmt.Errorf("创建AI模型目录失败: %v", err)
		}
	}

	return nil
} 