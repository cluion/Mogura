// Package analyze 提供磁碟空間分析:計算目錄大小並以 TUI 瀏覽。
package analyze

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"mogura/internal/clean"
)

// Entry 是目錄下的一個項目及其總大小。
type Entry struct {
	Name  string
	Path  string
	Size  int64
	IsDir bool
}

// Sizer 提供帶快取的目錄大小計算,同一路徑只完整走訪一次。
type Sizer struct {
	mu    sync.Mutex
	cache map[string]int64
}

func NewSizer() *Sizer {
	return &Sizer{cache: map[string]int64{}}
}

// List 列出目錄下所有項目並平行計算大小,依大小遞減排序。
func (s *Sizer) List(dir string) ([]Entry, error) {
	dirents, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	entries := make([]Entry, len(dirents))
	sem := make(chan struct{}, runtime.NumCPU())
	var wg sync.WaitGroup
	for i, de := range dirents {
		wg.Add(1)
		go func(i int, name string, isDir bool) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			path := filepath.Join(dir, name)
			entries[i] = Entry{Name: name, Path: path, Size: s.size(path), IsDir: isDir}
		}(i, de.Name(), de.IsDir())
	}
	wg.Wait()

	sort.SliceStable(entries, func(a, b int) bool {
		return entries[a].Size > entries[b].Size
	})
	return entries, nil
}

// Invalidate 移除 path 本身、其子孫與所有祖先的快取(刪除後大小全變了)。
func (s *Sizer) Invalidate(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	prefix := path + string(filepath.Separator)
	for k := range s.cache {
		if k == path || strings.HasPrefix(k, prefix) {
			delete(s.cache, k)
		}
	}
	for p := filepath.Dir(path); ; p = filepath.Dir(p) {
		delete(s.cache, p)
		if p == filepath.Dir(p) {
			break
		}
	}
}

func (s *Sizer) size(path string) int64 {
	s.mu.Lock()
	if n, ok := s.cache[path]; ok {
		s.mu.Unlock()
		return n
	}
	s.mu.Unlock()

	n := clean.SizeOf(path)

	s.mu.Lock()
	s.cache[path] = n
	s.mu.Unlock()
	return n
}
