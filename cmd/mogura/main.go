// Mogura — Linux 系統清理工具。像鼴鼠一樣,把磁碟裡的垃圾挖出來。
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"mogura/internal/clean"
)

// 正式版本由 GoReleaser 以 ldflags 注入
var version = "dev"

func main() {
	args := os.Args[1:]
	cmd := "clean"
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		cmd, args = args[0], args[1:]
	}

	var err error
	switch cmd {
	case "clean":
		err = runClean(args)
	case "analyze":
		err = runAnalyze(args)
	case "dev":
		err = runDev(args)
	case "orphan":
		err = runOrphan(args)
	case "monitor":
		err = runMonitor(args)
	case "mem":
		err = runMem(args)
	case "version":
		fmt.Println("mogura", version)
	default:
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "錯誤:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Println(`用法: mogura [指令] [選項]

指令:
  clean      掃描並清理系統垃圾(預設)
  analyze    磁碟空間分析,互動瀏覽各目錄佔用
  dev        掃描開發專案的建置產物(node_modules、target、vendor...)
  orphan     找出已解除安裝軟體留下的孤兒設定檔
  monitor    即時系統監控(CPU、記憶體、磁碟、網路)
  mem        記憶體大戶排行;--drop-caches / --swap-reset 釋放
  version    顯示版本

選項:
  --list         只列出結果,不進入互動清理(clean、dev)
  [路徑]         analyze 與 dev 的起始目錄,預設為家目錄`)
}

// confirm 顯示選定項目摘要並要求使用者確認。
func confirm(labels []string, sizes []int64, needRoot bool) bool {
	var total int64
	known := false
	fmt.Println("\n將清理以下項目:")
	for i, label := range labels {
		size := "—"
		if sizes[i] >= 0 {
			size = clean.Humanize(sizes[i])
			total += sizes[i]
			known = true
		}
		fmt.Printf("  · %s(%s)\n", label, size)
	}
	if known {
		fmt.Printf("預估釋放: %s\n", clean.Humanize(total))
	}
	if needRoot {
		fmt.Println("部分項目需要 sudo,執行時可能要求輸入密碼。")
	}
	return promptYes()
}

// promptYes 讀取使用者的 y/N 確認。
func promptYes() bool {
	fmt.Print("確定執行?[y/N] ")
	line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes"
}

func isTTY() bool {
	fi, err := os.Stdin.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

// withProgress 在 fn 執行期間即時顯示掃描進度;
// stdout 不是終端機(如管線輸出)時安靜執行,避免 \r 汙染管線。
func withProgress(label string, prog *clean.Progress, fn func()) {
	if fi, err := os.Stdout.Stat(); err != nil || fi.Mode()&os.ModeCharDevice == 0 {
		fn()
		return
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		fn()
	}()
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-done:
			fmt.Print("\r\033[K") // 清掉進度列
			return
		case <-tick.C:
			fmt.Printf("\r\033[K🦡 %s  已掃描 %s · %s 個檔案",
				label, clean.Humanize(prog.Bytes()), groupDigits(prog.Files()))
		}
	}
}

// groupDigits 加上千分位,大數字才讀得出量級。
func groupDigits(n int64) string {
	s := fmt.Sprintf("%d", n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}

// confirmAndRun 對選定項目做最終確認、執行並逐項回報。
func confirmAndRun(picked []clean.Result) error {
	var (
		labels   []string
		sizes    []int64
		needRoot bool
	)
	for _, r := range picked {
		labels = append(labels, r.Rule.Name)
		if r.Known {
			sizes = append(sizes, r.Size)
		} else {
			sizes = append(sizes, -1)
		}
		needRoot = needRoot || r.Rule.Root
	}
	if !confirm(labels, sizes, needRoot) {
		fmt.Println("已取消。")
		return nil
	}

	freed, outcomes := clean.Execute(picked)
	fmt.Println()
	for _, o := range outcomes {
		if o.Err != nil {
			fmt.Printf("  ✗ %s — %s\n", o.Result.Rule.Name, o.Err)
		} else {
			fmt.Printf("  ✓ %s\n", o.Result.Rule.Name)
		}
	}
	fmt.Printf("\n✨ 完成,共釋放約 %s\n", clean.Humanize(freed))
	return nil
}
