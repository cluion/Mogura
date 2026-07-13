package main

import (
	"fmt"

	"mogura/internal/clean"
	"mogura/internal/rules"
	"mogura/internal/ui"
)

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

	if listOnly || !isTTY() {
		printCleanList(results)
		return nil
	}

	opts := make([]ui.Option, len(results))
	for i, r := range results {
		opts[i] = ui.Option{
			Label: r.Rule.Name,
			Desc:  r.Rule.Description,
			Size:  r.Size,
			Known: r.Known,
			Risk:  r.Rule.Risk,
			Root:  r.Rule.Root,
			Value: r,
		}
	}
	selected, err := ui.MultiSelect("Mogura — 選擇要清理的項目", opts)
	if err != nil {
		return err
	}
	if len(selected) == 0 {
		fmt.Println("未選擇任何項目,結束。")
		return nil
	}

	var picked []clean.Result
	for _, o := range selected {
		picked = append(picked, o.Value.(clean.Result))
	}
	return confirmAndRun(picked)
}

func printCleanList(results []clean.Result) {
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
