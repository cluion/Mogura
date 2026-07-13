package orphan

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsActive(t *testing.T) {
	installed := map[string]bool{
		"firefox":              true,
		"google-chrome-stable": true,
		"python3-pip":          true,
		"code":                 true,
	}
	cases := []struct {
		name   string
		active bool
	}{
		{"firefox", true},              // 完全相符
		{"google-chrome", true},        // 子字串相符(套件名較長)
		{"pip", true},                  // 發行版前綴 python3-pip
		{"code", true},                 // 完全相符
		{"gh", true},                   // 名稱太短,保守視為使用中
		{"ghost-editor-legacy", false}, // 無任何相符 → 孤兒
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isActive(c.name, installed); got != c.active {
				t.Errorf("isActive(%s) = %v, 預期 %v", c.name, got, c.active)
			}
		})
	}
}

func TestScanBases(t *testing.T) {
	base := t.TempDir()
	installed := map[string]bool{"firefox": true}

	mk := func(name string) {
		dir := filepath.Join(base, name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "conf"), []byte("data"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mk("firefox")     // 使用中 → 跳過
	mk("autostart")   // 白名單 → 跳過
	mk("deadapp2020") // 孤兒
	if err := os.WriteFile(filepath.Join(base, "plainfile"), []byte("x"), 0o644); err != nil {
		t.Fatal(err) // 檔案不列入掃描
	}

	cands := ScanBases([]string{base}, installed)
	if len(cands) != 1 {
		t.Fatalf("孤兒數 = %d, 預期 1(%+v)", len(cands), cands)
	}
	if cands[0].Name != "deadapp2020" {
		t.Errorf("孤兒 = %s, 預期 deadapp2020", cands[0].Name)
	}
	if cands[0].Size == 0 {
		t.Error("孤兒大小不應為 0")
	}
}
