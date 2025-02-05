package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"time"
)

type AppConfig struct {
	TargetURL         string        `yaml:"target_url"`
	FetchInterval     time.Duration `yaml:"fetch_interval"`
	MaxDisplay        int           `yaml:"max_display"`
	DatabasePath      string        `yaml:"database_path"`
	LogPath           string        `yaml:"log_path"`
	NotificationMethods []string    `yaml:"notification_methods"`
}

func LoadConfig(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var cfg AppConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}
	return &cfg, nil
}
