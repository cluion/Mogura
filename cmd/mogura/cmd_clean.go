package main

import (
	"fmt"

	"mogura/internal/clean"
	"mogura/internal/i18n"
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
			return fmt.Errorf(i18n.T("未知選項: %s"), a)
		}
	}

	rs, err := rules.Load()
	if err != nil {
		return err
	}

	prog := &clean.Progress{}
	var results []clean.Result
	withProgress(i18n.T("掃描系統垃圾中..."), prog, func() {
		results = clean.ScanAll(rs, prog)
	})

	if listOnly || !isTTY() {
		printCleanList(results)
		return nil
	}

	for {

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
		selected, restart, err := ui.MultiSelect(i18n.T("Mogura — 選擇要清理的項目"), opts)
		if err != nil {
			return err
		}
		if restart {
			// 語言已切換:重載規則取得新語言文字,原地重貼標籤,不重掃
			if fresh, err := rules.Load(); err == nil {
				clean.Relabel(results, fresh)
			}
			continue
		}
		if len(selected) == 0 {
			fmt.Println(i18n.T("未選擇任何項目,結束。"))
			return nil
		}

		var picked []clean.Result
		for _, o := range selected {
			picked = append(picked, o.Value.(clean.Result))
		}
		return confirmAndRun(picked)
	}
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
	fmt.Print(i18n.Tf("\n合計可回收(可估算項目): %s\n", clean.Humanize(total)))
}
