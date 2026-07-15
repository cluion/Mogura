package trash

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupTrash(t *testing.T) string {
	t.Helper()
	data := t.TempDir()
	t.Setenv("XDG_DATA_HOME", data)
	return filepath.Join(data, "Trash")
}

func writeFile(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestXdgPutMovesFileAndWritesInfo(t *testing.T) {
	trashDir := setupTrash(t)
	src := filepath.Join(t.TempDir(), "報 告.txt")
	writeFile(t, src)

	if err := xdgPut(src); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Fatalf("原檔應已搬走: %v", err)
	}
	if _, err := os.Stat(filepath.Join(trashDir, "files", "報 告.txt")); err != nil {
		t.Fatalf("files/ 應有搬入的檔案: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(trashDir, "info", "報 告.txt.trashinfo"))
	if err != nil {
		t.Fatal(err)
	}
	info := string(raw)
	if !strings.HasPrefix(info, "[Trash Info]\n") {
		t.Errorf("trashinfo 缺標頭: %q", info)
	}
	if !strings.Contains(info, "Path=") || strings.Contains(info, "報 告") {
		t.Errorf("Path 應為 URL 編碼: %q", info)
	}
	if !strings.Contains(info, "DeletionDate=") {
		t.Errorf("缺 DeletionDate: %q", info)
	}
}

func TestXdgPutCollisionRenames(t *testing.T) {
	trashDir := setupTrash(t)
	dir := t.TempDir()
	for i := 0; i < 2; i++ {
		src := filepath.Join(dir, "same.txt")
		writeFile(t, src)
		if err := xdgPut(src); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"same.txt", "same.txt.2"} {
		if _, err := os.Stat(filepath.Join(trashDir, "files", name)); err != nil {
			t.Errorf("撞名應遞增改名,缺 %s: %v", name, err)
		}
	}
}

func TestXdgPutInsideTrashDeletesDirectly(t *testing.T) {
	trashDir := setupTrash(t)
	files := filepath.Join(trashDir, "files")
	if err := os.MkdirAll(files, 0o700); err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(files, "old.txt")
	writeFile(t, src)

	if err := xdgPut(src); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Fatal("垃圾桶內的路徑應直接刪除")
	}
	if _, err := os.Stat(filepath.Join(files, "old.txt.2")); err == nil {
		t.Fatal("不應在垃圾桶內再套疊一層")
	}
}
