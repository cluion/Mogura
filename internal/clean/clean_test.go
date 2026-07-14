package clean

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"mogura/internal/i18n"
	"mogura/internal/rules"
)

// TestMain 固定測試語言為繁中,CI 等無 zh 語系的環境才不會影響字串斷言。
func TestMain(m *testing.M) {
	i18n.SetEnglish(false)
	os.Exit(m.Run())
}

// diskSize 以 du 口徑(st_blocks×512)回傳單一路徑的磁碟佔用。
func diskSize(t *testing.T, path string) int64 {
	t.Helper()
	var st syscall.Stat_t
	if err := syscall.Lstat(path, &st); err != nil {
		t.Fatal(err)
	}
	return st.Blocks * 512
}

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
	targets, sizes := scanPaths(r, nil)
	if len(targets) != 1 || targets[0] != drop {
		t.Errorf("exclude 應排除 keep,實際 targets = %v", targets)
	}
	if want := diskSize(t, drop); len(sizes) != 1 || sizes[0] != want {
		t.Errorf("sizes = %v, 預期 [%d](僅 drop)", sizes, want)
	}
}

func TestExpandResults(t *testing.T) {
	dir := t.TempDir()
	// 10 個子目錄,大小遞增,驗證 top-8 個別列出 + 其餘 2 項合併
	for i := 1; i <= 10; i++ {
		sub := filepath.Join(dir, fmt.Sprintf("cache%02d", i))
		if err := os.Mkdir(sub, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(sub, "f"), make([]byte, i*8192), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	r := rules.Rule{
		ID: "t", Name: "快取", Risk: "low", Expand: true,
		Paths: []string{filepath.Join(dir, "*")},
	}
	results := scanRule(r, nil)
	if len(results) != 9 {
		t.Fatalf("結果數 = %d, 預期 9(8 個別 + 1 合併)", len(results))
	}
	if results[0].Rule.Name != "快取 · cache10" {
		t.Errorf("第一項應為最大的 cache10,實際 %s", results[0].Rule.Name)
	}
	last := results[8]
	if last.Rule.Name != "快取 · 其餘 2 項" || len(last.Targets) != 2 {
		t.Errorf("合併項錯誤: %s, targets=%v", last.Rule.Name, last.Targets)
	}
	for _, res := range results {
		if !res.Known || res.Size == 0 || len(res.Targets) == 0 {
			t.Errorf("每筆結果都應有大小與刪除目標: %+v", res)
		}
	}
}

func TestSizeOfDirectory(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a")
	if err := os.WriteFile(a, make([]byte, 100), 0o644); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	b := filepath.Join(sub, "b")
	if err := os.WriteFile(b, make([]byte, 50), 0o644); err != nil {
		t.Fatal(err)
	}
	want := diskSize(t, dir) + diskSize(t, a) + diskSize(t, sub) + diskSize(t, b)
	if got := SizeOf(dir); got != want {
		t.Errorf("SizeOf = %d, 預期 %d", got, want)
	}
}

func TestWalkHardlinkDedup(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a")
	if err := os.WriteFile(a, make([]byte, 8192), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Link(a, filepath.Join(dir, "b")); err != nil {
		t.Fatal(err)
	}
	// b 與 a 同 inode,只計一次
	want := diskSize(t, dir) + diskSize(t, a)
	size, _ := Walk(dir, nil)
	if size != want {
		t.Errorf("size = %d, 預期 %d(硬連結應只計一次)", size, want)
	}
}

func TestRelabel(t *testing.T) {
	i18n.SetEnglish(true)
	defer i18n.SetEnglish(false)

	results := []Result{
		{Rule: rules.Rule{ID: "c", Name: "快取 · spotify", Expand: true}, Targets: []string{"/home/u/.cache/spotify"}},
		{Rule: rules.Rule{ID: "c", Name: "快取 · 其餘 5 項", Expand: true}, Targets: []string{"/a/b/x", "/a/b/y"}},
		{Rule: rules.Rule{ID: "p", Name: "垃圾桶"}, Targets: []string{"/home/u/.local/share/Trash/files/f"}},
	}
	fresh := []rules.Rule{
		{ID: "c", Name: "Cache", Description: "caches", Expand: true},
		{ID: "p", Name: "Trash", Description: "trash"},
	}
	Relabel(results, fresh)

	if results[0].Rule.Name != "Cache · spotify" {
		t.Errorf("展開子項 = %q", results[0].Rule.Name)
	}
	if results[1].Rule.Name != "Cache · 2 others" {
		t.Errorf("合併項 = %q", results[1].Rule.Name)
	}
	if results[2].Rule.Name != "Trash" || results[2].Rule.Description != "trash" {
		t.Errorf("一般規則 = %q / %q", results[2].Rule.Name, results[2].Rule.Description)
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
	want := diskSize(t, dir) + diskSize(t, old) + diskSize(t, sub) + diskSize(t, fresh)
	if size != want {
		t.Errorf("size = %d, 預期 %d", size, want)
	}
	if time.Since(latest) > time.Hour {
		t.Errorf("latest 應來自深層新檔案,實際 %v", latest)
	}
	if prog.Bytes() != want || prog.Files() != 2 {
		t.Errorf("progress = %d bytes / %d files, 預期 %d / 2", prog.Bytes(), prog.Files(), want)
	}
}
