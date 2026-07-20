package main

import (
	"fmt"

	"mogura/internal/clean"
	"mogura/internal/i18n"
	"mogura/internal/rules"
	"mogura/internal/ui"
)

func runClean(args []string) error {
	listOnly, jsonOut := false, false
	for _, a := range args {
		switch a {
		case "--list":
			listOnly = true
		case "--json":
			jsonOut = true
		default:
			usage()
			return fmt.Errorf(i18n.T("未知選項: %s"), a)
		}
	}

	rs, err := rules.Load(ruleOptions())
	if err != nil {
		return err
	}

	prog := &clean.Progress{}
	var results []clean.Result
	withProgress(i18n.T("掃描系統垃圾中..."), prog, func() {
		results = clean.ScanAll(rs, prog)
	})

	if jsonOut {
		return printCleanJSON(results)
	}
	if listOnly || !isTTY() {
		printCleanList(results)
		return nil
	}
	return cleanInteract(results)
}

// cleanInteract 對已掃描的結果跑互動選擇與清理(dashboard 沿用同一份掃描)
func cleanInteract(results []clean.Result) error {
	for {
		results = dropExcluded(results)
		opts := make([]ui.Option, len(results))
		for i, r := range results {
			opts[i] = ui.Option{
				Label: r.Rule.Name,
				Desc:  r.Rule.Description,
				Size:  r.Size,
				Known: r.Known,
				Risk:  r.Rule.Risk,
				Root:  r.Rule.Root,
				Path:  singleTarget(r),
				Value: r,
			}
		}
		selected, restart, err := ui.MultiSelect(i18n.T("Mogura — 選擇要清理的項目"), opts)
		if err != nil {
			return err
		}
		if restart {
			// 語言或 journal 天數已變:重載規則取得新文字,原地重貼標籤,不重掃
			if fresh, err := rules.Load(ruleOptions()); err == nil {
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

// singleTarget 回傳規則唯一的目標路徑;多目標或 action 型規則不可用 x 排除
func singleTarget(r clean.Result) string {
	if len(r.Targets) == 1 && r.Rule.Action == "" {
		return r.Targets[0]
	}
	return ""
}

// dropExcluded 濾掉已被全域排除的單一路徑項目(x 排除後語言切換等重建時生效)
func dropExcluded(results []clean.Result) []clean.Result {
	ex := excludePaths()
	if len(ex) == 0 {
		return results
	}
	out := make([]clean.Result, 0, len(results))
	for _, r := range results {
		if t := singleTarget(r); t != "" && clean.Excluded(t, ex) {
			continue
		}
		out = append(out, r)
	}
	return out
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
