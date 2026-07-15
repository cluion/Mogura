// Package config 管理 ~/.config/mogura/config.yaml 的使用者設定。
package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Language string `yaml:"language"` // auto | zh | en
	Delete   string `yaml:"delete"`   // direct | trash
}

// UseTrash 回報刪除是否走垃圾桶。
func (c Config) UseTrash() bool { return c.Delete == "trash" }

// Path 回傳設定檔路徑(依 XDG 慣例)。
func Path() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "mogura", "config.yaml"), nil
}

// Load 讀取設定;檔案不存在或損壞時回傳預設值,不阻擋主流程。
func Load() Config {
	cfg := Config{Language: "auto", Delete: "direct"}
	p, err := Path()
	if err != nil {
		return cfg
	}
	raw, err := os.ReadFile(p)
	if err != nil {
		return cfg
	}
	_ = yaml.Unmarshal(raw, &cfg)
	if cfg.Language == "" {
		cfg.Language = "auto"
	}
	if cfg.Delete != "trash" {
		cfg.Delete = "direct"
	}
	return cfg
}

// Save 寫入設定檔,目錄不存在時自動建立。
func Save(cfg Config) error {
	p, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	raw, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(p, raw, 0o644)
}
