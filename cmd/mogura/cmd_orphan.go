package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"mogura/internal/clean"
	"mogura/internal/i18n"
	"mogura/internal/orphan"
	"mogura/internal/rules"
	"mogura/internal/ui"
)

func runOrphan(args []string) error {
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
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf(i18n.T("無法取得家目錄: %w"), err)
	}

	fmt.Println(i18n.T("🦡 蒐集已安裝軟體清單..."))
	sys := orphan.Detect()
	prog := &clean.Progress{}
	var cands []orphan.Candidate
	withProgress(i18n.T("比對設定目錄中..."), prog, func() {
		cands = orphan.ScanBases(orphan.DefaultBases(), sys.Installed, prog)
	})
	cands = dropExcludedCands(cands)
	sort.SliceStable(cands, func(a, b int) bool { return cands[a].Size > cands[b].Size })

	if jsonOut {
		return printOrphanJSON(cands, sys.RemovedConfigs)
	}
	if len(cands) == 0 && len(sys.RemovedConfigs) == 0 {
		fmt.Println(i18n.T("沒有找到孤兒設定,很乾淨!"))
		return nil
	}

	if listOnly || !isTTY() {
		printOrphanList(cands, sys.RemovedConfigs, home)
		return nil
	}

	for {
		cands = dropExcludedCands(cands)
		var opts []ui.Option
		if len(sys.RemovedConfigs) > 0 {
			opts = append(opts, ui.Option{
				Label: i18n.Tf("dpkg 殘留設定(%d 個已移除套件)", len(sys.RemovedConfigs)),
				Desc:  strings.Join(sys.RemovedConfigs, " "),
				Risk:  "low",
				Root:  true,
				Value: clean.Result{Rule: rules.Rule{
					ID:     "dpkg-rc",
					Name:   i18n.Tf("dpkg 殘留設定(%d 個套件)", len(sys.RemovedConfigs)),
					Action: "dpkg --purge -- " + strings.Join(sys.RemovedConfigs, " "),
					Risk:   "low",
					Root:   true,
				}},
			})
		}
		for _, c := range cands {
			opts = append(opts, ui.Option{
				Label: relPath(c.Path, home),
				Desc:  i18n.Tf("找不到對應的已安裝軟體 · %s", idleDaysLabel(c.IdleDays())),
				Size:  c.Size,
				Known: true,
				Risk:  "medium",
				Path:  c.Path,
				Value: clean.Result{
					Rule:    rules.Rule{ID: "orphan", Name: relPath(c.Path, home), Risk: "medium"},
					Targets: []string{c.Path},
					Size:    c.Size,
					Known:   true,
				},
			})
		}

		selected, restart, err := ui.MultiSelect(i18n.T("Mogura — 孤兒設定檔(啟發式判斷,刪前請確認)"), opts)
		if err != nil {
			return err
		}
		if restart {
			continue // 語言已切換,選單用新語言重建(比對結果沿用)
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

// dropExcludedCands 濾掉已被全域排除的候選目錄
func dropExcludedCands(cands []orphan.Candidate) []orphan.Candidate {
	ex := excludePaths()
	if len(ex) == 0 {
		return cands
	}
	out := make([]orphan.Candidate, 0, len(cands))
	for _, c := range cands {
		if clean.Excluded(c.Path, ex) {
			continue
		}
		out = append(out, c)
	}
	return out
}

func printOrphanList(cands []orphan.Candidate, rc []string, home string) {
	if len(rc) > 0 {
		fmt.Print(i18n.T("\ndpkg 殘留設定(可用 sudo dpkg --purge 清除):\n"))
		for _, p := range rc {
			fmt.Printf("  · %s\n", p)
		}
	}
	if len(cands) > 0 {
		fmt.Println(i18n.T("\n找不到對應軟體的設定目錄(啟發式,刪前請確認):"))
		var total int64
		for _, c := range cands {
			fmt.Printf("  %10s  %-12s %s\n",
				clean.Humanize(c.Size), idleDaysLabel(c.IdleDays()), relPath(c.Path, home))
			total += c.Size
		}
		fmt.Print(i18n.Tf("\n合計: %s\n", clean.Humanize(total)))
	}
}

func idleDaysLabel(days int) string {
	if days < 0 {
		return i18n.T("未知")
	}
	if days == 0 {
		return i18n.T("今天有動")
	}
	return i18n.Tf("閒置 %d 天", days)
}
