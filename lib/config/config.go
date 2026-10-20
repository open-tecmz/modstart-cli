package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Site 一个站点的 CLI 凭证信息
type Site struct {
	URL       string `json:"url"`
	AdminURL  string `json:"adminUrl"`
	App       string `json:"app"`
	AdminPath string `json:"adminPath"`
	Token     string `json:"token"`
	Name      string `json:"name"`
	AddedAt   string `json:"addedAt"`
}

// Config 本地配置，支持登记多个站点
type Config struct {
	Current string           `json:"current"`
	Sites   map[string]*Site `json:"sites"`
}

// Dir 配置目录
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".modstart-cli"), nil
}

// Path 配置文件路径
func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// Load 读取配置
func Load() (*Config, error) {
	p, err := Path()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Sites: map[string]*Site{}}, nil
		}
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("配置文件格式错误: %w", err)
	}
	if c.Sites == nil {
		c.Sites = map[string]*Site{}
	}
	return &c, nil
}

// Save 保存配置（目录 0700、文件 0600）
func Save(c *Config) error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "config.json"), data, 0600)
}

// NormalizeURL 规范化站点地址
func NormalizeURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		raw = "https://" + raw
	}
	return strings.TrimRight(raw, "/")
}

// Find 查找站点，key 为空时使用当前站点
func (c *Config) Find(key string) (*Site, error) {
	key = NormalizeURL(key)
	if key == "" {
		key = c.Current
	}
	if key == "" {
		return nil, errors.New("未指定站点，请先执行 modstart-cli login")
	}
	s, ok := c.Sites[key]
	if !ok {
		return nil, fmt.Errorf("站点未登记: %s", key)
	}
	return s, nil
}

// Put 写入站点并设为当前站点
func (c *Config) Put(s *Site) {
	s.URL = NormalizeURL(s.URL)
	if s.App == "" {
		s.App = "admin"
	}
	if s.AdminPath == "" {
		s.AdminPath = "admin"
	}
	if s.AdminURL == "" {
		s.AdminURL = s.URL + "/" + s.AdminPath
	}
	s.AdminURL = NormalizeURL(s.AdminURL)
	if s.AddedAt == "" {
		s.AddedAt = time.Now().Format("2006-01-02 15:04:05")
	}
	c.Sites[s.URL] = s
	c.Current = s.URL
}

// Remove 移除站点
func (c *Config) Remove(key string) error {
	key = NormalizeURL(key)
	if _, ok := c.Sites[key]; !ok {
		return fmt.Errorf("站点未登记: %s", key)
	}
	delete(c.Sites, key)
	if c.Current == key {
		c.Current = ""
	}
	return nil
}
