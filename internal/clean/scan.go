// Package clean 實作清理引擎:掃描規則對應的空間佔用,並執行清理。
package clean

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"

	"mogura/internal/rules"
)

// Result 是一條規則的掃描結果。
type Result struct {
	Rule    rules.Rule
	Targets []string // 路徑規則展開後的實際刪除目標
	Size    int64    // 可回收位元組數
	Known   bool     // 大小是否可估算(指令規則無 probe 時為 false)
	Err     error
}

// ScanAll 平行掃描所有規則,結果依大小遞減排序(大小未知者排最後)。
func ScanAll(rs []rules.Rule) []Result {
	results := make([]Result, len(rs))
	sem := make(chan struct{}, runtime.NumCPU())
	var wg sync.WaitGroup
	for i, r := range rs {
		wg.Add(1)
		go func(i int, r rules.Rule) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			results[i] = scanRule(r)
		}(i, r)
	}
	wg.Wait()

	sort.SliceStable(results, func(a, b int) bool {
		if results[a].Known != results[b].Known {
			return results[a].Known
		}
		return results[a].Size > results[b].Size
	})
	return results
}

func scanRule(r rules.Rule) Result {
	res := Result{Rule: r}
	if len(r.Paths) > 0 {
		res.Targets, res.Size = scanPaths(r)
		res.Known = true
		return res
	}
	if r.Probe != "" {
		// Probe 來自 go:embed 的內建規則,非使用者輸入;若未來支援
		// 使用者自訂規則檔,須重新審視 sh -c 的注入風險。
		out, err := exec.Command("sh", "-c", r.Probe).Output()
		if err == nil {
			if n, perr := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64); perr == nil {
				res.Size = n
				res.Known = true
			}
		}
	}
	return res
}

func scanPaths(r rules.Rule) (targets []string, total int64) {
	excluded := map[string]bool{}
	for _, ex := range r.Exclude {
		matches, _ := filepath.Glob(rules.ExpandHome(ex))
		for _, m := range matches {
			excluded[m] = true
		}
	}
	for _, p := range r.Paths {
		matches, _ := filepath.Glob(rules.ExpandHome(p))
		for _, m := range matches {
			if excluded[m] {
				continue
			}
			targets = append(targets, m)
			total += sizeOf(m)
		}
	}
	return targets, total
}

// sizeOf 計算檔案或目錄的總大小,不追蹤 symlink,忽略無權限的子項。
func sizeOf(path string) int64 {
	info, err := os.Lstat(path)
	if err != nil {
		return 0
	}
	if !info.IsDir() {
		return info.Size()
	}
	var total int64
	filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // 無權限等錯誤直接略過該項
		}
		if fi, err := d.Info(); err == nil && !d.IsDir() {
			total += fi.Size()
		}
		return nil
	})
	return total
}

// Humanize 將位元組數轉為人類可讀格式。
func Humanize(n int64) string {
	const unit = 1024
	if n < unit {
		return strconv.FormatInt(n, 10) + " B"
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return strconv.FormatFloat(float64(n)/float64(div), 'f', 1, 64) + " " + string("KMGTPE"[exp]) + "iB"
}
