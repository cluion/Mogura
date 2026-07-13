// Mogura — Linux 系統清理工具。像鼴鼠一樣,把磁碟裡的垃圾挖出來。
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"

	"mogura/internal/clean"
	"mogura/internal/rules"
	"mogura/internal/ui"
)

var version = "0.1.0-dev"

func main() {
	args := os.Args[1:]
	cmd := "clean"
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		cmd, args = args[0], args[1:]
	}

	switch cmd {
	case "clean":
		if err := runClean(args); err != nil {
			fmt.Fprintln(os.Stderr, "錯誤:", err)
			os.Exit(1)
		}
	case "version":
		fmt.Println("mogura", version)
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Println(`用法: mogura [指令] [選項]

指令:
  clean      掃描並清理系統垃圾(預設)
  version    顯示版本

clean 選項:
  --list     只列出掃描結果,不進入互動清理`)
}

func runClean(args []string) error {
	listOnly := false
	for _, a := range args {
		switch a {
		case "--list":
			listOnly = true
		default:
			usage()
			return fmt.Errorf("未知選項: %s", a)
		}
	}

	rs, err := rules.Load()
	if err != nil {
		return err
	}

	fmt.Println("🦡 掃描中...")
	results := clean.ScanAll(rs)

	if listOnly || !term.IsTerminal(int(os.Stdin.Fd())) {
		printList(results)
		return nil
	}

	selected, err := ui.Select(results)
	if err != nil {
		return err
	}
	if len(selected) == 0 {
		fmt.Println("未選擇任何項目,結束。")
		return nil
	}

	if !confirm(selected) {
		fmt.Println("已取消。")
		return nil
	}

	freed, outcomes := clean.Execute(selected)
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

func printList(results []clean.Result) {
	var total int64
	for _, r := range results {
		size := "—"
		if r.Known {
			size = clean.Humanize(r.Size)
			total += r.Size
		}
		root := " "
		if r.Rule.Root {
			root = "🔒"
		}
		fmt.Printf("  %10s  %s %-24s %s\n", size, root, r.Rule.Name, r.Rule.Description)
	}
	fmt.Printf("\n合計可回收(可估算項目): %s\n", clean.Humanize(total))
}

func confirm(selected []clean.Result) bool {
	var total int64
	needRoot := false
	fmt.Println("\n將清理以下項目:")
	for _, r := range selected {
		size := "—"
		if r.Known {
			size = clean.Humanize(r.Size)
			total += r.Size
		}
		fmt.Printf("  · %s(%s)\n", r.Rule.Name, size)
		if r.Rule.Root {
			needRoot = true
		}
	}
	fmt.Printf("預估釋放: %s\n", clean.Humanize(total))
	if needRoot {
		fmt.Println("部分項目需要 sudo,執行時可能要求輸入密碼。")
	}
	fmt.Print("確定執行?[y/N] ")
	line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes"
}
