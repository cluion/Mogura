package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEmbeddedRules(t *testing.T) {
	rs, err := Load()
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
