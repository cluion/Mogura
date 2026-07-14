package analyze

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestListSortedBySize(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "small"), make([]byte, 10), 0o644); err != nil {
		t.Fatal(err)
	}
	big := filepath.Join(root, "bigdir")
	if err := os.Mkdir(big, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(big, "data"), make([]byte, 100*1024), 0o644); err != nil {
		t.Fatal(err)
	}

	s := NewSizer()
	entries, err := s.List(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("項目數 = %d, 預期 2", len(entries))
	}
	if entries[0].Name != "bigdir" || !entries[0].IsDir {
		t.Errorf("第一項應為 bigdir,實際 %+v", entries[0])
	}
	if entries[0].Size <= entries[1].Size || entries[1].Size == 0 {
		t.Errorf("大小排序錯誤: %d vs %d", entries[0].Size, entries[1].Size)
	}
	if entries[0].Files != 1 {
		t.Errorf("bigdir 遞迴檔案數 = %d, 預期 1", entries[0].Files)
	}
	if entries[0].ModTime.IsZero() {
		t.Error("ModTime 不應為零值")
	}

	// 快取:再次查詢同路徑不應出錯且結果一致
	again, err := s.List(root)
	if err != nil || again[0].Size != entries[0].Size {
		t.Errorf("快取後結果不一致: %+v, err=%v", again, err)
	}
}

func TestListStream(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"a", "b", "c"} {
		if err := os.WriteFile(filepath.Join(root, name), make([]byte, 100), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	s := NewSizer()
	entries, ch, err := s.ListStream(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("初始項目數 = %d, 預期 3", len(entries))
	}
	for _, e := range entries {
		if e.Size != SizeUnknown {
			t.Errorf("初始大小應為 SizeUnknown,實際 %d", e.Size)
		}
	}

	var got int
	for e := range ch {
		if e.Size == SizeUnknown || e.Size == 0 {
			t.Errorf("串流結果 %s 的大小應已算出,實際 %d", e.Name, e.Size)
		}
		got++
	}
	if got != 3 {
		t.Errorf("串流筆數 = %d, 預期 3(channel 應在送完後關閉)", got)
	}
}

func TestInvalidate(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(sub, "data")
	if err := os.WriteFile(target, make([]byte, 4096), 0o644); err != nil {
		t.Fatal(err)
	}

	s := NewSizer()
	before, err := s.List(root)
	if err != nil || before[0].Size == 0 {
		t.Fatalf("前置掃描失敗: %+v, err=%v", before, err)
	}

	// 刪除檔案後失效快取,重新查詢應反映新大小
	if err := os.Remove(target); err != nil {
		t.Fatal(err)
	}
	s.Invalidate(target)
	after, err := s.List(root)
	if err != nil {
		t.Fatal(err)
	}
	if after[0].Size >= before[0].Size {
		t.Errorf("Invalidate 後大小應變小: %d → %d", before[0].Size, after[0].Size)
	}
}

func TestSortEntries(t *testing.T) {
	now := time.Now()
	entries := []Entry{
		{Name: "beta", Size: 100, ModTime: now.Add(-2 * time.Hour)},
		{Name: "Alpha", Size: 300, ModTime: now.Add(-5 * time.Hour)},
		{Name: "gamma", Size: 200, ModTime: now},
	}
	sortEntries(entries, sortSize)
	if entries[0].Name != "Alpha" || entries[2].Name != "beta" {
		t.Errorf("大小排序錯誤: %v", names(entries))
	}
	sortEntries(entries, sortName)
	if entries[0].Name != "Alpha" || entries[1].Name != "beta" || entries[2].Name != "gamma" {
		t.Errorf("名稱排序錯誤(應不分大小寫): %v", names(entries))
	}
	sortEntries(entries, sortMtime)
	if entries[0].Name != "gamma" || entries[2].Name != "Alpha" {
		t.Errorf("mtime 排序錯誤(新到舊): %v", names(entries))
	}
}

func names(entries []Entry) []string {
	out := make([]string, len(entries))
	for i, e := range entries {
		out[i] = e.Name
	}
	return out
}

func TestFillCells(t *testing.T) {
	cases := []struct {
		size, max int64
		want      int
	}{
		{0, 100, 0},
		{100, 100, barWidth},
		{50, 100, barWidth / 2},
		{300, 100, barWidth}, // size > max(非大小排序時會發生)不可超出寬度
		{100, 0, 0},          // 防除以零
		{-5, 100, 0},
	}
	for _, c := range cases {
		if got := fillCells(c.size, c.max); got != c.want {
			t.Errorf("fillCells(%d, %d) = %d, 預期 %d", c.size, c.max, got, c.want)
		}
	}
}

func TestDeleteGuard(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("無家目錄環境")
	}
	cases := []struct {
		path    string
		wantErr bool
	}{
		{"/", true},
		{"/usr", true},
		{home, true},
		{"relative/path", true},
		{"/tmp/hardlink-lab", false},
		{filepath.Join(home, ".cache", "foo"), false},
	}
	for _, c := range cases {
		if err := deleteGuard(c.path); (err != nil) != c.wantErr {
			t.Errorf("deleteGuard(%s) 錯誤 = %v, 預期錯誤 = %v", c.path, err, c.wantErr)
		}
	}
}
