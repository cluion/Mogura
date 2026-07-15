package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"mogura/internal/clean"
	"mogura/internal/devjunk"
	"mogura/internal/i18n"
	"mogura/internal/rules"
	"mogura/internal/ui"
)

func runDev(args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf(i18n.T("無法取得家目錄: %w"), err)
	}
	root := home
	listOnly, jsonOut := false, false
	for _, a := range args {
		switch {
		case a == "--list":
			listOnly = true
		case a == "--json":
			jsonOut = true
		case strings.HasPrefix(a, "-"):
			usage()
			return fmt.Errorf(i18n.T("未知選項: %s"), a)
		default:
			root = rules.ExpandHome(a)
		}
	}
	abs, err := filepath.Abs(root)
	if err != nil || (!strings.HasPrefix(abs+"/", home+"/") && abs != home) {
		return fmt.Errorf(i18n.T("dev 掃描僅支援家目錄內的路徑: %s"), root)
	}

	prog := &clean.Progress{}
	var junks []devjunk.Junk
	var scanErr error
	withProgress(i18n.Tf("掃描 %s 的建置產物中...", abs), prog, func() {
		junks, scanErr = devjunk.Scan(abs, excludePaths(), prog)
	})
	if scanErr != nil {
		return scanErr
	}
	if len(junks) == 0 {
		fmt.Println(i18n.T("沒有找到建置產物,很乾淨!"))
		return nil
	}
	sort.SliceStable(junks, func(a, b int) bool { return junks[a].Size > junks[b].Size })

	if jsonOut {
		return printDevJSON(junks)
	}
	if listOnly || !isTTY() {
		printDevList(junks, abs)
		return nil
	}

	for {
		junks = dropExcludedJunks(junks)
		opts := make([]ui.Option, len(junks))
		for i, j := range junks {
			opts[i] = ui.Option{
				Label: relPath(j.Path, home),
				Desc:  fmt.Sprintf("%s · %s", j.Kind.Label, idleDaysLabel(j.IdleDays())),
				Size:  j.Size,
				Known: true,
				Risk:  j.Kind.Risk,
				Path:  j.Path,
				Value: j,
			}
		}
		selected, restart, err := ui.MultiSelect(i18n.T("Mogura — 選擇要刪除的建置產物"), opts)
		if err != nil {
			return err
		}
		if restart {
			continue // 語言已切換,選單文字每輪重建,直接重開即可
		}
		if len(selected) == 0 {
			fmt.Println(i18n.T("未選擇任何項目,結束。"))
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
}

// dropExcludedJunks 濾掉已被全域排除的產物(x 排除後重建選單時生效)。
func dropExcludedJunks(junks []devjunk.Junk) []devjunk.Junk {
	ex := excludePaths()
	if len(ex) == 0 {
		return junks
	}
	out := make([]devjunk.Junk, 0, len(junks))
	for _, j := range junks {
		if clean.Excluded(j.Path, ex) {
			continue
		}
		out = append(out, j)
	}
	return out
}

func printDevList(junks []devjunk.Junk, home string) {
	var total int64
	for _, j := range junks {
		fmt.Printf("  %10s  %-8s %-12s %s\n",
			clean.Humanize(j.Size), j.Kind.Label, idleDaysLabel(j.IdleDays()), relPath(j.Path, home))
		total += j.Size
	}
	fmt.Print(i18n.Tf("\n合計可回收: %s\n", clean.Humanize(total)))
}

func relPath(path, home string) string {
	if rel, err := filepath.Rel(home, path); err == nil && !strings.HasPrefix(rel, "..") {
		return "~/" + rel
	}
	return path
}
