package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config 全局配置
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Auth     AuthConfig     `yaml:"auth"`
	Upload   UploadConfig   `yaml:"upload"`
	Site     SiteConfig     `yaml:"site"`
}

type ServerConfig struct {
	Port     int    `yaml:"port"`
	HomeHost string `yaml:"home_host"`
	NavHost  string `yaml:"nav_host"`
	Mode     string `yaml:"mode"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type AuthConfig struct {
	SessionSecret     string `yaml:"session_secret"`
	SessionMaxAge     int    `yaml:"session_max_age"`
	MaxLoginAttempts  int    `yaml:"max_login_attempts"`
	LockoutMinutes    int    `yaml:"lockout_minutes"`
}

type UploadConfig struct {
	Dir          string   `yaml:"dir"`
	AllowedTypes []string `yaml:"allowed_types"`
	MaxSizeMB    int      `yaml:"max_size_mb"`
}

type SiteConfig struct {
	Title    string   `yaml:"title"`
	Subtitle string   `yaml:"subtitle"`
	Vision   []string `yaml:"vision"`
	WikiURL  string   `yaml:"wiki_url"`
	NavURL   string   `yaml:"nav_url"`
}

// Load 从文件加载配置
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	// 设置默认值
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.Mode == "" {
		cfg.Server.Mode = "release"
	}
	if cfg.Database.Path == "" {
		cfg.Database.Path = "./data/portal.db"
	}
	if cfg.Auth.SessionMaxAge == 0 {
		cfg.Auth.SessionMaxAge = 168
	}
	if cfg.Auth.MaxLoginAttempts == 0 {
		cfg.Auth.MaxLoginAttempts = 5
	}
	if cfg.Auth.LockoutMinutes == 0 {
		cfg.Auth.LockoutMinutes = 15
	}
	if cfg.Upload.Dir == "" {
		cfg.Upload.Dir = "./web/static/uploads"
	}
	if cfg.Upload.MaxSizeMB == 0 {
		cfg.Upload.MaxSizeMB = 2
	}
	return &cfg, nil
}
