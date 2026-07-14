package clean

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// Progress 讓掃描過程即時回報累計量,供 UI 顯示;nil 時所有操作皆為 no-op。
type Progress struct {
	bytes  atomic.Int64
	files  atomic.Int64
	parent *Progress
}

// ChildProgress 建立子進度:累計時同步轉發給 parent(可為 nil),
// 讓局部計數與全域即時顯示共用同一趟走訪。
func ChildProgress(parent *Progress) *Progress {
	return &Progress{parent: parent}
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

func (p *Progress) add(bytes int64, file bool) {
	if p == nil {
		return
	}
	p.bytes.Add(bytes)
	if file {
		p.files.Add(1)
	}
	p.parent.add(bytes, file)
}

type inodeKey struct {
	dev uint64
	ino uint64
}

type walker struct {
	sem      chan struct{}
	wg       sync.WaitGroup
	total    atomic.Int64
	latestNs atomic.Int64
	prog     *Progress

	mu   sync.Mutex
	seen map[inodeKey]struct{} // 已計數的多硬連結 inode
}

// Walk 平行走訪 path,回傳實際磁碟佔用(du 口徑:st_blocks、
// 硬連結只計一次、含目錄本身)與整棵樹最新的 mtime。
// 不追蹤 symlink,無權限的子項直接略過。
func Walk(path string, prog *Progress) (int64, time.Time) {
	info, err := os.Lstat(path)
	if err != nil {
		return 0, time.Time{}
	}
	w := &walker{
		sem:  make(chan struct{}, runtime.NumCPU()),
		prog: prog,
		seen: map[inodeKey]struct{}{},
	}
	w.noteMtime(info.ModTime())
	w.account(info)
	if info.IsDir() {
		w.walkDir(path)
		w.wg.Wait()
	}
	return w.total.Load(), time.Unix(0, w.latestNs.Load())
}

// SizeOf 計算檔案或目錄的實際磁碟佔用。
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
		w.account(info)
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
		}
	}
}

// account 累計一個項目的磁碟佔用。一般檔案若有多個硬連結,
// 同一 inode 只在首次遇到時計數(目錄的 nlink 天生 >1,不去重)。
func (w *walker) account(info os.FileInfo) {
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		if !info.IsDir() {
			w.total.Add(info.Size())
			w.prog.add(info.Size(), true)
		}
		return
	}
	if info.Mode().IsRegular() && st.Nlink > 1 {
		key := inodeKey{dev: uint64(st.Dev), ino: uint64(st.Ino)}
		w.mu.Lock()
		_, dup := w.seen[key]
		if !dup {
			w.seen[key] = struct{}{}
		}
		w.mu.Unlock()
		if dup {
			return
		}
	}
	size := st.Blocks * 512
	w.total.Add(size)
	w.prog.add(size, !info.IsDir())
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
