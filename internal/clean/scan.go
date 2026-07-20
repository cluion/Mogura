// Package clean 實作清理引擎:掃描規則對應的空間佔用,並執行清理
package clean

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"

	"mogura/internal/i18n"
	"mogura/internal/rules"
)

// Result 是一條規則的掃描結果
type Result struct {
	Rule    rules.Rule
	Targets []string // 路徑規則展開後的實際刪除目標
	Size    int64    // 可回收位元組數
	Known   bool     // 大小是否可估算(指令規則無 probe 時為 false)
	Err     error
}

// ScanAll 平行掃描所有規則,結果依大小遞減排序(大小未知者排最後)
// expand 規則會攤開成多筆結果;prog 可為 nil,非 nil 時即時累計掃描進度
func ScanAll(rs []rules.Rule, prog *Progress) []Result {
	perRule := make([][]Result, len(rs))
	sem := make(chan struct{}, runtime.NumCPU())
	var wg sync.WaitGroup
	for i, r := range rs {
		wg.Add(1)
		go func(i int, r rules.Rule) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			perRule[i] = scanRule(r, prog)
		}(i, r)
	}
	wg.Wait()

	var results []Result
	for _, rs := range perRule {
		results = append(results, rs...)
	}
	sort.SliceStable(results, func(a, b int) bool {
		if results[a].Known != results[b].Known {
			return results[a].Known
		}
		return results[a].Size > results[b].Size
	})
	return results
}

// expandTopN 是 expand 規則個別列出的子項數量,其餘打包成一項
const expandTopN = 8

func scanRule(r rules.Rule, prog *Progress) []Result {
	res := Result{Rule: r}
	if len(r.Paths) > 0 {
		targets, sizes := scanPaths(r, prog)
		if r.Expand && len(targets) > 0 {
			return expandResults(r, targets, sizes)
		}
		for _, s := range sizes {
			res.Size += s
		}
		res.Known = true
		if r.Action == "" {
			res.Targets = targets // 有 action 時 paths 僅用於估算大小
		}
		return []Result{res}
	}
	if r.Probe != "" {
		// Probe 來自 go:embed 的內建規則,非使用者輸入;若未來支援
		// 使用者自訂規則檔,須重新審視 sh -c 的注入風險
		out, err := exec.Command("sh", "-c", r.Probe).Output()
		if err == nil {
			if n, perr := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64); perr == nil {
				res.Size = n
				res.Known = true
			}
		}
	}
	return []Result{res}
}

// expandResults 把子項依大小排序,前 expandTopN 名個別成列,其餘合併為一項
func expandResults(r rules.Rule, targets []string, sizes []int64) []Result {
	idx := make([]int, len(targets))
	for i := range idx {
		idx[i] = i
	}
	sort.SliceStable(idx, func(a, b int) bool { return sizes[idx[a]] > sizes[idx[b]] })

	var results []Result
	for n, i := range idx {
		if n >= expandTopN {
			break
		}
		child := r
		child.Name = r.Name + " · " + filepath.Base(targets[i]) // r.Name 已於載入點翻譯
		child.Description = displayPath(targets[i])
		results = append(results, Result{
			Rule: child, Targets: []string{targets[i]}, Size: sizes[i], Known: true,
		})
	}
	if rest := idx[min(expandTopN, len(idx)):]; len(rest) > 0 {
		agg := r
		agg.Name = r.Name + " · " + i18n.Tf("其餘 %d 項", len(rest))
		var restTargets []string
		var restSize int64
		for _, i := range rest {
			restTargets = append(restTargets, targets[i])
			restSize += sizes[i]
		}
		results = append(results, Result{Rule: agg, Targets: restTargets, Size: restSize, Known: true})
	}
	return results
}

// Relabel 以新載入的規則(語言可能已切換)重建結果的顯示文字,不重新掃描
func Relabel(results []Result, rs []rules.Rule) {
	byID := map[string]rules.Rule{}
	for _, r := range rs {
		byID[r.ID] = r
	}
	for i := range results {
		fresh, ok := byID[results[i].Rule.ID]
		if !ok {
			continue
		}
		if results[i].Rule.Expand && len(results[i].Targets) > 0 {
			if len(results[i].Targets) == 1 {
				fresh.Description = displayPath(results[i].Targets[0])
				fresh.Name += " · " + filepath.Base(results[i].Targets[0])
			} else {
				fresh.Name += " · " + i18n.Tf("其餘 %d 項", len(results[i].Targets))
			}
		}
		results[i].Rule = fresh
	}
}

// displayPath 把家目錄縮寫成 ~,讓子項描述短一點
func displayPath(p string) string {
	if home, err := os.UserHomeDir(); err == nil && strings.HasPrefix(p, home+"/") {
		return "~" + strings.TrimPrefix(p, home)
	}
	return p
}

func scanPaths(r rules.Rule, prog *Progress) (targets []string, sizes []int64) {
	var excluded []string
	for _, ex := range r.Exclude {
		p := rules.ExpandHome(ex)
		if matches, _ := filepath.Glob(p); len(matches) > 0 {
			excluded = append(excluded, matches...)
		} else {
			excluded = append(excluded, p) // 路徑尚不存在也保留,前綴比對仍有效
		}
	}
	for _, p := range r.Paths {
		matches, _ := filepath.Glob(rules.ExpandHome(p))
		for _, m := range matches {
			if Excluded(m, excluded) {
				continue
			}
			targets = append(targets, m)
			size, _ := Walk(m, prog)
			sizes = append(sizes, size)
		}
	}
	return targets, sizes
}

// Excluded 回報 path 是否等於任一排除路徑、或位於其之下
// excludes 須為已展開的絕對路徑
func Excluded(path string, excludes []string) bool {
	for _, ex := range excludes {
		if path == ex || strings.HasPrefix(path, ex+string(os.PathSeparator)) {
			return true
		}
	}
	return false
}

// GroupDigits 加上千分位,大數字才讀得出量級
func GroupDigits(n int64) string {
	s := strconv.FormatInt(n, 10)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}

// Humanize 將位元組數轉為人類可讀格式
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
