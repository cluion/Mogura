// Package devjunk 掃描開發專案的建置產物:node_modules、target、vendor 等
// 可以重新產生的目錄。
package devjunk

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mogura/internal/clean"
)

// Kind 定義一種建置產物:目錄名 + 佐證檔(存在於同層才視為專案產物,避免誤刪同名目錄)。
type Kind struct {
	Dir     string
	Sibling string
	Label   string
	Risk    string
}

var kinds = []Kind{
	{Dir: "node_modules", Sibling: "package.json", Label: "Node", Risk: "medium"},
	{Dir: "target", Sibling: "Cargo.toml", Label: "Rust", Risk: "medium"},
	{Dir: "vendor", Sibling: "composer.json", Label: "PHP", Risk: "medium"},
	{Dir: "__pycache__", Label: "Python", Risk: "low"},
	{Dir: ".pytest_cache", Label: "Python", Risk: "low"},
	{Dir: ".mypy_cache", Label: "Python", Risk: "low"},
	{Dir: ".ruff_cache", Label: "Python", Risk: "low"},
}

// Junk 是一個可清除的建置產物目錄。
type Junk struct {
	Path    string
	Kind    Kind
	Size    int64
	ModTime time.Time
}

// Scan 從 root 向下找建置產物。隱藏目錄(除了產物本身)一律跳過,
// 找到的產物目錄不再深入。prog 可為 nil。
func Scan(root string, prog *clean.Progress) ([]Junk, error) {
	var junks []Junk
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // 無權限等錯誤直接略過
		}
		if !d.IsDir() || path == root {
			return nil
		}
		if k, ok := match(path, d.Name()); ok {
			junks = append(junks, Junk{Path: path, Kind: k})
			return filepath.SkipDir
		}
		if strings.HasPrefix(d.Name(), ".") {
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	for i := range junks {
		junks[i].Size, junks[i].ModTime = clean.Walk(junks[i].Path, prog)
	}
	return junks, nil
}

func match(path, name string) (Kind, bool) {
	for _, k := range kinds {
		if name != k.Dir {
			continue
		}
		if k.Sibling == "" {
			return k, true
		}
		if _, err := os.Stat(filepath.Join(filepath.Dir(path), k.Sibling)); err == nil {
			return k, true
		}
	}
	return Kind{}, false
}

// IdleDays 回傳距離最後修改的天數,無法取得時回傳 -1。
func (j Junk) IdleDays() int {
	if j.ModTime.IsZero() {
		return -1
	}
	return int(time.Since(j.ModTime).Hours() / 24)
}
