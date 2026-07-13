package clean

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// Progress 讓掃描過程即時回報累計量,供 UI 顯示;nil 時所有操作皆為 no-op。
type Progress struct {
	bytes atomic.Int64
	files atomic.Int64
}

func (p *Progress) Bytes() int64 {
	if p == nil {
		return 0
	}
	return p.bytes.Load()
}

func (p *Progress) Files() int64 {
	if p == nil {
		return 0
	}
	return p.files.Load()
}

func (p *Progress) add(size int64) {
	if p != nil {
		p.bytes.Add(size)
		p.files.Add(1)
	}
}

type walker struct {
	sem      chan struct{}
	wg       sync.WaitGroup
	total    atomic.Int64
	latestNs atomic.Int64
	prog     *Progress
}

// Walk 平行走訪 path,回傳總大小與整棵樹最新的 mtime。
// 不追蹤 symlink,無權限的子項直接略過。
func Walk(path string, prog *Progress) (int64, time.Time) {
	info, err := os.Lstat(path)
	if err != nil {
		return 0, time.Time{}
	}
	w := &walker{sem: make(chan struct{}, runtime.NumCPU()), prog: prog}
	w.noteMtime(info.ModTime())
	if info.IsDir() {
		w.walkDir(path)
		w.wg.Wait()
	} else {
		w.total.Add(info.Size())
		prog.add(info.Size())
	}
	return w.total.Load(), time.Unix(0, w.latestNs.Load())
}

// SizeOf 計算檔案或目錄的總大小。
func SizeOf(path string) int64 {
	size, _ := Walk(path, nil)
	return size
}

// walkDir 對子目錄採「有空位就開 goroutine,沒有就原地遞迴」,
// 避免深樹的 goroutine 爆炸,同時吃滿磁碟平行度。
func (w *walker) walkDir(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		w.noteMtime(info.ModTime())
		if e.IsDir() {
			sub := filepath.Join(dir, e.Name())
			select {
			case w.sem <- struct{}{}:
				w.wg.Add(1)
				go func() {
					defer w.wg.Done()
					w.walkDir(sub)
					<-w.sem
				}()
			default:
				w.walkDir(sub)
			}
		} else {
			w.total.Add(info.Size())
			w.prog.add(info.Size())
		}
	}
}

func (w *walker) noteMtime(t time.Time) {
	ns := t.UnixNano()
	for {
		cur := w.latestNs.Load()
		if ns <= cur || w.latestNs.CompareAndSwap(cur, ns) {
			return
		}
	}
}
