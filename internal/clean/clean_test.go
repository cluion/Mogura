package clean

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"mogura/internal/rules"
)

func TestGuardPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("無家目錄環境")
	}
	cases := []struct {
		name    string
		path    string
		root    bool
		wantErr bool
	}{
		{"家目錄內合法路徑", filepath.Join(home, ".cache", "foo"), false, false},
		{"相對路徑", ".cache/foo", false, true},
		{"根目錄", "/", false, true},
		{"層級過淺", "/var/crash", true, true},
		{"非 root 規則刪家目錄外", "/var/cache/apt/archives/x.deb", false, true},
		{"root 規則刪系統路徑", "/var/crash/dump/core.1", true, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := guardPath(c.path, c.root)
			if (err != nil) != c.wantErr {
				t.Errorf("guardPath(%s, root=%v) 錯誤 = %v, 預期錯誤 = %v", c.path, c.root, err, c.wantErr)
			}
		})
	}
}

func TestHumanize(t *testing.T) {
	cases := []struct {
		in   int64
		want string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KiB"},
		{1536, "1.5 KiB"},
		{1048576, "1.0 MiB"},
		{5368709120, "5.0 GiB"},
	}
	for _, c := range cases {
		if got := Humanize(c.in); got != c.want {
			t.Errorf("Humanize(%d) = %s, 預期 %s", c.in, got, c.want)
		}
	}
}

func TestScanPathsWithExclude(t *testing.T) {
	dir := t.TempDir()
	keep := filepath.Join(dir, "keep")
	drop := filepath.Join(dir, "drop")
	for _, p := range []string{keep, drop} {
		if err := os.WriteFile(p, []byte("12345"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	r := rules.Rule{
		ID: "t", Name: "T", Risk: "low",
		Paths:   []string{filepath.Join(dir, "*")},
		Exclude: []string{keep},
	}
	targets, total := scanPaths(r, nil)
	if len(targets) != 1 || targets[0] != drop {
		t.Errorf("exclude 應排除 keep,實際 targets = %v", targets)
	}
	if total != 5 {
		t.Errorf("total = %d, 預期 5", total)
	}
}

func TestSizeOfDirectory(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a"), make([]byte, 100), 0o644); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "b"), make([]byte, 50), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := SizeOf(dir); got != 150 {
		t.Errorf("SizeOf = %d, 預期 150", got)
	}
}

func TestWalkLatestMtimeAndProgress(t *testing.T) {
	dir := t.TempDir()
	old := filepath.Join(dir, "old")
	if err := os.WriteFile(old, make([]byte, 10), 0o644); err != nil {
		t.Fatal(err)
	}
	past := time.Now().Add(-100 * 24 * time.Hour)
	if err := os.Chtimes(old, past, past); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	fresh := filepath.Join(sub, "fresh")
	if err := os.WriteFile(fresh, make([]byte, 20), 0o644); err != nil {
		t.Fatal(err)
	}
	// 頂層目錄設成舊時間,驗證 latest 來自深層檔案而非頂層
	if err := os.Chtimes(dir, past, past); err != nil {
		t.Fatal(err)
	}

	prog := &Progress{}
	size, latest := Walk(dir, prog)
	if size != 30 {
		t.Errorf("size = %d, 預期 30", size)
	}
	if time.Since(latest) > time.Hour {
		t.Errorf("latest 應來自深層新檔案,實際 %v", latest)
	}
	if prog.Bytes() != 30 || prog.Files() != 2 {
		t.Errorf("progress = %d bytes / %d files, 預期 30 / 2", prog.Bytes(), prog.Files())
	}
}
