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
