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
			Desc:  fmt.Sprintf("%s · %s", j.Kind.Label, idleDaysLabel(j.IdleDays())),
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

	var picked []clean.Result
	for _, o := range selected {
		j := o.Value.(devjunk.Junk)
		picked = append(picked, clean.Result{
			Rule:    rules.Rule{ID: "dev-junk", Name: relPath(j.Path, home), Risk: j.Kind.Risk},
			Targets: []string{j.Path},
			Size:    j.Size,
			Known:   true,
		})
	}
	return confirmAndRun(picked)
}

func printDevList(junks []devjunk.Junk, home string) {
	var total int64
	for _, j := range junks {
		fmt.Printf("  %10s  %-8s %-12s %s\n",
			clean.Humanize(j.Size), j.Kind.Label, idleDaysLabel(j.IdleDays()), relPath(j.Path, home))
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
