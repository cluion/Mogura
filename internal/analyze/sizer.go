// Package analyze 提供磁碟空間分析:計算目錄大小並以 TUI 瀏覽。
package analyze

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"mogura/internal/clean"
)

// Entry 是目錄下的一個項目及其統計。
type Entry struct {
	Name    string
	Path    string
	Size    int64
	Files   int64 // 遞迴檔案數
	IsDir   bool
	ModTime time.Time // 整棵樹最新的 mtime
}

type stat struct {
	size  int64
	files int64
	mtime time.Time
}

// Sizer 提供帶快取的目錄統計,同一路徑只完整走訪一次。
type Sizer struct {
	mu    sync.Mutex
	cache map[string]stat
	live  *clean.Progress // 所有走訪的即時匯流,供 UI 顯示
}

func NewSizer() *Sizer {
	return &Sizer{cache: map[string]stat{}}
}

// SetProgress 設定即時進度匯流點。
func (s *Sizer) SetProgress(p *clean.Progress) { s.live = p }

// SizeUnknown 表示該項目的統計尚未算出(串流中)。
const SizeUnknown int64 = -1

// ListStream 立即回傳項目清單(統計未知),統計算完一項就往 channel 送一項,
// 全部完成後關閉 channel。
func (s *Sizer) ListStream(dir string) ([]Entry, <-chan Entry, error) {
	dirents, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, err
	}

	entries := make([]Entry, len(dirents))
	for i, de := range dirents {
		entries[i] = Entry{
			Name: de.Name(), Path: filepath.Join(dir, de.Name()),
			IsDir: de.IsDir(), Size: SizeUnknown,
		}
	}

	ch := make(chan Entry, len(entries))
	go func() {
		defer close(ch)
		sem := make(chan struct{}, runtime.NumCPU())
		var wg sync.WaitGroup
		for _, e := range entries {
			wg.Add(1)
			go func(e Entry) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()
				st := s.stat(e.Path)
				e.Size, e.Files, e.ModTime = st.size, st.files, st.mtime
				ch <- e
			}(e)
		}
		wg.Wait()
	}()
	return entries, ch, nil
}

// List 列出目錄下所有項目並平行計算統計,依大小遞減排序(阻塞至全部完成)。
func (s *Sizer) List(dir string) ([]Entry, error) {
	entries, ch, err := s.ListStream(dir)
	if err != nil {
		return nil, err
	}
	byPath := make(map[string]int, len(entries))
	for i, e := range entries {
		byPath[e.Path] = i
	}
	for e := range ch {
		entries[byPath[e.Path]] = e
	}
	sort.SliceStable(entries, func(a, b int) bool {
		return entries[a].Size > entries[b].Size
	})
	return entries, nil
}

// Invalidate 移除 path 本身、其子孫與所有祖先的快取(刪除後統計全變了)。
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

func (s *Sizer) stat(path string) stat {
	s.mu.Lock()
	if st, ok := s.cache[path]; ok {
		s.mu.Unlock()
		return st
	}
	s.mu.Unlock()

	prog := clean.ChildProgress(s.live)
	size, mtime := clean.Walk(path, prog)
	st := stat{size: size, files: prog.Files(), mtime: mtime}

	s.mu.Lock()
	s.cache[path] = st
	s.mu.Unlock()
	return st
}
