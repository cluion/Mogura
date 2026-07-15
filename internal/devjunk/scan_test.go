package devjunk

import (
	"os"
	"path/filepath"
	"testing"
)

func mkdirWithFile(t *testing.T, dir, file string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if file != "" {
		if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestScan(t *testing.T) {
	root := t.TempDir()

	// 合法 Node 專案:node_modules + 同層 package.json
	node := filepath.Join(root, "app", "node_modules")
	mkdirWithFile(t, node, filepath.Join(root, "app", "package.json"))
	mkdirWithFile(t, filepath.Join(node, "lodash"), filepath.Join(node, "lodash", "index.js"))

	// target 但沒有 Cargo.toml → 不應匹配
	mkdirWithFile(t, filepath.Join(root, "notrust", "target"), "")

	// __pycache__ 不需要佐證檔
	pycache := filepath.Join(root, "py", "__pycache__")
	mkdirWithFile(t, pycache, filepath.Join(pycache, "m.pyc"))

	// 隱藏目錄下的產物 → 應被跳過
	hidden := filepath.Join(root, ".hidden", "node_modules")
	mkdirWithFile(t, hidden, filepath.Join(root, ".hidden", "package.json"))

	junks, err := Scan(root, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	found := map[string]bool{}
	for _, j := range junks {
		found[j.Path] = true
	}
	if !found[node] {
		t.Error("應找到 node_modules(有 package.json 佐證)")
	}
	if found[filepath.Join(root, "notrust", "target")] {
		t.Error("無 Cargo.toml 的 target 不應匹配")
	}
	if !found[pycache] {
		t.Error("應找到 __pycache__")
	}
	if found[hidden] {
		t.Error("隱藏目錄下的產物應被跳過")
	}
	for _, j := range junks {
		if j.Path == node && j.Size == 0 {
			t.Error("node_modules 大小不應為 0")
		}
	}
}

func TestScanSkipsInsideJunk(t *testing.T) {
	root := t.TempDir()
	outer := filepath.Join(root, "app", "node_modules")
	mkdirWithFile(t, outer, filepath.Join(root, "app", "package.json"))
	// node_modules 內部的巢狀 node_modules 不應被單獨列出
	inner := filepath.Join(outer, "pkg", "node_modules")
	mkdirWithFile(t, inner, filepath.Join(outer, "pkg", "package.json"))

	junks, err := Scan(root, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(junks) != 1 || junks[0].Path != outer {
		t.Errorf("只應列出最外層 node_modules,實際 %+v", junks)
	}
}

func TestScanExclude(t *testing.T) {
	root := t.TempDir()
	for _, p := range []string{"a/node_modules", "skip/b/node_modules"} {
		if err := os.MkdirAll(filepath.Join(root, p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, filepath.Dir(p), "package.json"), []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	junks, err := Scan(root, []string{filepath.Join(root, "skip")}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(junks) != 1 || filepath.Base(filepath.Dir(junks[0].Path)) != "a" {
		t.Errorf("排除清單下的產物不應列出: %+v", junks)
	}
}
