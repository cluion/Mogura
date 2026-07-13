package analyze

import (
	"os"
	"path/filepath"
	"testing"
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

	// 快取:再次查詢同路徑不應出錯且結果一致
	again, err := s.List(root)
	if err != nil || again[0].Size != entries[0].Size {
		t.Errorf("快取後結果不一致: %+v, err=%v", again, err)
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
