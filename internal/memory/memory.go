// Package memory 提供記憶體狀態、程序排行與釋放操作
package memory

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/process"
)

// Stats 是目前的記憶體與 swap 狀態
type Stats struct {
	Total     uint64
	Used      uint64
	Available uint64
	Cached    uint64
	SwapTotal uint64
	SwapUsed  uint64
}

// Proc 是一個程序的記憶體佔用
type Proc struct {
	PID  int32
	Name string
	RSS  uint64
}

// Read 讀取目前記憶體狀態
func Read() (Stats, error) {
	vm, err := mem.VirtualMemory()
	if err != nil {
		return Stats{}, fmt.Errorf("讀取記憶體狀態失敗: %w", err)
	}
	s := Stats{
		Total: vm.Total, Used: vm.Used, Available: vm.Available, Cached: vm.Cached,
	}
	if sw, err := mem.SwapMemory(); err == nil {
		s.SwapTotal, s.SwapUsed = sw.Total, sw.Used
	}
	return s, nil
}

// Top 回傳 RSS 前 n 名的程序
func Top(n int) ([]Proc, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, fmt.Errorf("讀取程序清單失敗: %w", err)
	}
	var out []Proc
	for _, p := range procs {
		info, err := p.MemoryInfo()
		if err != nil || info == nil || info.RSS == 0 {
			continue
		}
		name, err := p.Name()
		if err != nil || name == "" {
			continue
		}
		out = append(out, Proc{PID: p.Pid, Name: name, RSS: info.RSS})
	}
	return Rank(out, n), nil
}

// Rank 依 RSS 遞減排序並取前 n 名
func Rank(procs []Proc, n int) []Proc {
	sort.SliceStable(procs, func(a, b int) bool { return procs[a].RSS > procs[b].RSS })
	if len(procs) > n {
		procs = procs[:n]
	}
	return procs
}

// DropCaches 清除 page cache(需要 root);多數情況下不必要,cache 會自動回收
func DropCaches() error {
	return runPrivileged("sync && echo 3 > /proc/sys/vm/drop_caches")
}

// SwapReset 把 swap 內容搬回 RAM(需要 root 且要有足夠可用記憶體)
func SwapReset() error {
	return runPrivileged("swapoff -a && swapon -a")
}

func runPrivileged(script string) error {
	// script 為套件內固定字串,非使用者輸入
	cmd := exec.Command("sh", "-c", script)
	if os.Geteuid() != 0 {
		cmd = exec.Command("sudo", "sh", "-c", script)
	}
	cmd.Stdin = os.Stdin
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
