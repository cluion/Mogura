// Package rules 定義宣告式清理規則:規則是資料不是程式碼,
// 全部以 YAML 描述並用 go:embed 編進執行檔。
package rules

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"mogura/internal/i18n"
)

//go:embed data/*.yaml
var dataFS embed.FS

// Rule 是一條清理規則。只有 Paths 時走內建的掃描+刪除;
// 有 Action 時交給外部指令執行,大小用 Paths(內建平行掃描)或 Probe 估算。
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
	Expand      bool     `yaml:"expand"` // 掃描時把 paths 的子項攤開成獨立選項
}

var validRisks = map[string]bool{"low": true, "medium": true, "high": true}

// Options 帶入使用者設定調整規則:全域排除清單與 {days} 佔位符的值。
type Options struct {
	Exclude     []string // 併入每條路徑型規則的 exclude
	JournalDays int      // 代入 {days},<1 時用預設 7
}

const defaultJournalDays = 7

// Load 讀取所有內嵌規則檔,驗證後依 opt 調整並回傳。
// Requires 指定的指令不存在時,該規則會被略過。
func Load(opt Options) ([]Rule, error) {
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
			// 在載入點翻譯一次,下游顯示、確認、回報全部自動生效
			r.Name = i18n.T(r.Name)
			r.Description = i18n.T(r.Description)
			all = append(all, r.withOptions(opt))
		}
	}
	return all, nil
}

// withOptions 套用使用者設定。{days} 在翻譯後才代入,
// 讓 i18n 對照表的鍵維持含佔位符的原文。
func (r Rule) withOptions(opt Options) Rule {
	days := opt.JournalDays
	if days < 1 {
		days = defaultJournalDays
	}
	d := strconv.Itoa(days)
	r.Description = strings.ReplaceAll(r.Description, "{days}", d)
	r.Action = strings.ReplaceAll(r.Action, "{days}", d)
	if len(r.Paths) > 0 {
		r.Exclude = append(r.Exclude, opt.Exclude...)
	}
	return r
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
	case r.Expand && (len(r.Paths) == 0 || r.Action != ""):
		return fmt.Errorf("規則 %s 的 expand 只能用於純 paths 規則", r.ID)
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
