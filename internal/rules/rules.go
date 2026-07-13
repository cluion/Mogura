// Package rules 定義宣告式清理規則:規則是資料不是程式碼,
// 全部以 YAML 描述並用 go:embed 編進執行檔。
package rules

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed data/*.yaml
var dataFS embed.FS

// Rule 是一條清理規則。Paths 與 Action 二選一:
// Paths 走內建的掃描+刪除;Action 交給外部指令執行,搭配 Probe 估算大小。
type Rule struct {
	ID          string   `yaml:"id"`
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Paths       []string `yaml:"paths"`
	Exclude     []string `yaml:"exclude"`
	Probe       string   `yaml:"probe"`
	Action      string   `yaml:"action"`
	Requires    string   `yaml:"requires"`
	Risk        string   `yaml:"risk"`
	Root        bool     `yaml:"root"`
}

var validRisks = map[string]bool{"low": true, "medium": true, "high": true}

// Load 讀取所有內嵌規則檔,驗證後回傳。
// Requires 指定的指令不存在時,該規則會被略過。
func Load() ([]Rule, error) {
	entries, err := dataFS.ReadDir("data")
	if err != nil {
		return nil, fmt.Errorf("讀取內嵌規則目錄失敗: %w", err)
	}

	var all []Rule
	seen := map[string]bool{}
	for _, e := range entries {
		raw, err := dataFS.ReadFile("data/" + e.Name())
		if err != nil {
			return nil, fmt.Errorf("讀取規則檔 %s 失敗: %w", e.Name(), err)
		}
		var rs []Rule
		if err := yaml.Unmarshal(raw, &rs); err != nil {
			return nil, fmt.Errorf("解析規則檔 %s 失敗: %w", e.Name(), err)
		}
		for _, r := range rs {
			if err := validate(r); err != nil {
				return nil, fmt.Errorf("規則檔 %s: %w", e.Name(), err)
			}
			if seen[r.ID] {
				return nil, fmt.Errorf("規則 id 重複: %s", r.ID)
			}
			seen[r.ID] = true
			if r.Requires != "" {
				if _, err := exec.LookPath(r.Requires); err != nil {
					continue
				}
			}
			all = append(all, r)
		}
	}
	return all, nil
}

func validate(r Rule) error {
	switch {
	case r.ID == "":
		return fmt.Errorf("規則缺少 id")
	case r.Name == "":
		return fmt.Errorf("規則 %s 缺少 name", r.ID)
	case !validRisks[r.Risk]:
		return fmt.Errorf("規則 %s 的 risk 必須是 low/medium/high", r.ID)
	case len(r.Paths) == 0 && r.Action == "":
		return fmt.Errorf("規則 %s 必須提供 paths 或 action", r.ID)
	case len(r.Paths) > 0 && r.Action != "":
		return fmt.Errorf("規則 %s 的 paths 與 action 不可同時設定", r.ID)
	}
	return nil
}

// ExpandHome 將開頭的 ~ 展開成使用者家目錄。
func ExpandHome(p string) string {
	if p == "~" || strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(p, "~"))
		}
	}
	return p
}
