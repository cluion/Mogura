package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"mogura/internal/clean"
	"mogura/internal/devjunk"
	"mogura/internal/rules"
	"mogura/internal/ui"
)

func runDev(args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("無法取得家目錄: %w", err)
	}
	root := home
	listOnly := false
	for _, a := range args {
		switch {
		case a == "--list":
			listOnly = true
		case strings.HasPrefix(a, "-"):
			usage()
			return fmt.Errorf("未知選項: %s", a)
		default:
			root = rules.ExpandHome(a)
		}
	}
	abs, err := filepath.Abs(root)
	if err != nil || (!strings.HasPrefix(abs+"/", home+"/") && abs != home) {
		return fmt.Errorf("dev 掃描僅支援家目錄內的路徑: %s", root)
	}

	fmt.Printf("🦡 掃描 %s 的建置產物中...\n", abs)
	junks, err := devjunk.Scan(abs)
	if err != nil {
		return err
	}
	if len(junks) == 0 {
		fmt.Println("沒有找到建置產物,很乾淨!")
		return nil
	}
	sort.SliceStable(junks, func(a, b int) bool { return junks[a].Size > junks[b].Size })

	if listOnly || !isTTY() {
		printDevList(junks, abs)
		return nil
	}

	opts := make([]ui.Option, len(junks))
	for i, j := range junks {
		opts[i] = ui.Option{
			Label: relPath(j.Path, home),
			Desc:  fmt.Sprintf("%s · %s", j.Kind.Label, idleLabel(j)),
			Size:  j.Size,
			Known: true,
			Risk:  j.Kind.Risk,
			Value: j,
		}
	}
	selected, err := ui.MultiSelect("Mogura — 選擇要刪除的建置產物", opts)
	if err != nil {
		return err
	}
	if len(selected) == 0 {
		fmt.Println("未選擇任何項目,結束。")
		return nil
	}

	var (
		picked []clean.Result
		labels []string
		sizes  []int64
	)
	for _, o := range selected {
		j := o.Value.(devjunk.Junk)
		picked = append(picked, clean.Result{
			Rule:    rules.Rule{ID: "dev-junk", Name: relPath(j.Path, home), Risk: j.Kind.Risk},
			Targets: []string{j.Path},
			Size:    j.Size,
			Known:   true,
		})
		labels = append(labels, relPath(j.Path, home))
		sizes = append(sizes, j.Size)
	}
	if !confirm(labels, sizes, false) {
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

func printDevList(junks []devjunk.Junk, home string) {
	var total int64
	for _, j := range junks {
		fmt.Printf("  %10s  %-8s %-12s %s\n",
			clean.Humanize(j.Size), j.Kind.Label, idleLabel(j), relPath(j.Path, home))
		total += j.Size
	}
	fmt.Printf("\n合計可回收: %s\n", clean.Humanize(total))
}

func relPath(path, home string) string {
	if rel, err := filepath.Rel(home, path); err == nil && !strings.HasPrefix(rel, "..") {
		return "~/" + rel
	}
	return path
}

func idleLabel(j devjunk.Junk) string {
	days := j.IdleDays()
	if days < 0 {
		return "未知"
	}
	if days == 0 {
		return "今天有動"
	}
	return fmt.Sprintf("閒置 %d 天", days)
}
