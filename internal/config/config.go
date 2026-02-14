package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 全局配置结构
type Config struct {
	Server ServerConfig `yaml:"server"`
	AI     AIConfig     `yaml:"ai"`
	Redis  RedisConfig  `yaml:"redis"`
}

// ServerConfig 服务配置
type ServerConfig struct {
	Port string `yaml:"port"`
	Mode string `yaml:"mode"`
}

// AIConfig AI配置
type AIConfig struct {
	BaseURL   string  `yaml:"base_url"`
	Model     string  `yaml:"model"`
	Temperature float64 `yaml:"temperature"`
	MaxTokens int     `yaml:"max_tokens"`
	Timeout   int     `yaml:"timeout"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// GlobalConfig 全局配置实例
var GlobalConfig *Config

// Load 加载配置文件
func Load(configPath string) (*Config, error) {
	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析YAML
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 环境变量覆盖
	if apiKey := os.Getenv("AI_API_KEY"); apiKey != "" {
		// MVP阶段API_KEY通过环境变量传递
	}

	GlobalConfig = &cfg
	return &cfg, nil
}
