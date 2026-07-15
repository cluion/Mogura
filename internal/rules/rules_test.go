package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEmbeddedRules(t *testing.T) {
	rs, err := Load(Options{})
	if err != nil {
		t.Fatalf("內嵌規則載入失敗: %v", err)
	}
	if len(rs) == 0 {
		t.Fatal("內嵌規則不應為空")
	}
	seen := map[string]bool{}
	for _, r := range rs {
		if seen[r.ID] {
			t.Errorf("規則 id 重複: %s", r.ID)
		}
		seen[r.ID] = true
	}
}

func TestValidate(t *testing.T) {
	cases := []struct {
		name    string
		rule    Rule
		wantErr bool
	}{
		{"合法路徑規則", Rule{ID: "a", Name: "A", Paths: []string{"~/x"}, Risk: "low"}, false},
		{"合法指令規則", Rule{ID: "b", Name: "B", Action: "true", Risk: "high"}, false},
		{"缺 id", Rule{Name: "A", Paths: []string{"~/x"}, Risk: "low"}, true},
		{"缺 name", Rule{ID: "a", Paths: []string{"~/x"}, Risk: "low"}, true},
		{"risk 不合法", Rule{ID: "a", Name: "A", Paths: []string{"~/x"}, Risk: "危"}, true},
		{"paths 與 action 皆空", Rule{ID: "a", Name: "A", Risk: "low"}, true},
		{"paths 估算大小 + action 執行", Rule{ID: "a", Name: "A", Paths: []string{"~/x"}, Action: "true", Risk: "low"}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := validate(c.rule)
			if (err != nil) != c.wantErr {
				t.Errorf("validate(%+v) 錯誤 = %v, 預期錯誤 = %v", c.rule, err, c.wantErr)
			}
		})
	}
}

func TestExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("無家目錄環境")
	}
	if got := ExpandHome("~/.cache"); got != filepath.Join(home, ".cache") {
		t.Errorf("ExpandHome(~/.cache) = %s", got)
	}
	if got := ExpandHome("/var/log"); got != "/var/log" {
		t.Errorf("絕對路徑不應被改寫: %s", got)
	}
	if got := ExpandHome("~user/x"); got != "~user/x" {
		t.Errorf("~user 形式不應被展開: %s", got)
	}
}

func TestWithOptions(t *testing.T) {
	r := Rule{
		ID: "journal-logs", Description: "清除 {days} 天以前的 journal 日誌",
		Action: "journalctl --vacuum-time={days}d",
	}
	got := r.withOptions(Options{JournalDays: 14})
	if got.Action != "journalctl --vacuum-time=14d" {
		t.Errorf("Action = %q", got.Action)
	}
	if got.Description != "清除 14 天以前的 journal 日誌" {
		t.Errorf("Description = %q", got.Description)
	}

	// 未設定時用預設 7
	if got := r.withOptions(Options{}); got.Action != "journalctl --vacuum-time=7d" {
		t.Errorf("預設 Action = %q", got.Action)
	}

	// 全域排除只併入路徑型規則
	p := Rule{ID: "p", Paths: []string{"~/x"}, Exclude: []string{"~/x/keep"}}
	got = p.withOptions(Options{Exclude: []string{"~/y"}})
	if len(got.Exclude) != 2 || got.Exclude[1] != "~/y" {
		t.Errorf("Exclude = %v", got.Exclude)
	}
	if got := r.withOptions(Options{Exclude: []string{"~/y"}}); len(got.Exclude) != 0 {
		t.Errorf("action 型規則不應併入排除: %v", got.Exclude)
	}
}
