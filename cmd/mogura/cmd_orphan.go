package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"mogura/internal/clean"
	"mogura/internal/orphan"
	"mogura/internal/rules"
	"mogura/internal/ui"
)

func runOrphan(args []string) error {
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
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("無法取得家目錄: %w", err)
	}

	fmt.Println("🦡 蒐集已安裝軟體清單...")
	sys := orphan.Detect()
	prog := &clean.Progress{}
	var cands []orphan.Candidate
	withProgress("比對設定目錄中...", prog, func() {
		cands = orphan.ScanBases(orphan.DefaultBases(), sys.Installed, prog)
	})
	sort.SliceStable(cands, func(a, b int) bool { return cands[a].Size > cands[b].Size })

	if len(cands) == 0 && len(sys.RemovedConfigs) == 0 {
		fmt.Println("沒有找到孤兒設定,很乾淨!")
		return nil
	}

	if listOnly || !isTTY() {
		printOrphanList(cands, sys.RemovedConfigs, home)
		return nil
	}

	var opts []ui.Option
	if len(sys.RemovedConfigs) > 0 {
		opts = append(opts, ui.Option{
			Label: fmt.Sprintf("dpkg 殘留設定(%d 個已移除套件)", len(sys.RemovedConfigs)),
			Desc:  strings.Join(sys.RemovedConfigs, " "),
			Risk:  "low",
			Root:  true,
			Value: clean.Result{Rule: rules.Rule{
				ID:     "dpkg-rc",
				Name:   fmt.Sprintf("dpkg 殘留設定(%d 個套件)", len(sys.RemovedConfigs)),
				Action: "dpkg --purge -- " + strings.Join(sys.RemovedConfigs, " "),
				Risk:   "low",
				Root:   true,
			}},
		})
	}
	for _, c := range cands {
		opts = append(opts, ui.Option{
			Label: relPath(c.Path, home),
			Desc:  fmt.Sprintf("找不到對應的已安裝軟體 · %s", idleDaysLabel(c.IdleDays())),
			Size:  c.Size,
			Known: true,
			Risk:  "medium",
			Value: clean.Result{
				Rule:    rules.Rule{ID: "orphan", Name: relPath(c.Path, home), Risk: "medium"},
				Targets: []string{c.Path},
				Size:    c.Size,
				Known:   true,
			},
		})
	}

	selected, err := ui.MultiSelect("Mogura — 孤兒設定檔(啟發式判斷,刪前請確認)", opts)
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

func printOrphanList(cands []orphan.Candidate, rc []string, home string) {
	if len(rc) > 0 {
		fmt.Printf("\ndpkg 殘留設定(可用 sudo dpkg --purge 清除):\n")
		for _, p := range rc {
			fmt.Printf("  · %s\n", p)
		}
	}
	if len(cands) > 0 {
		fmt.Println("\n找不到對應軟體的設定目錄(啟發式,刪前請確認):")
		var total int64
		for _, c := range cands {
			fmt.Printf("  %10s  %-12s %s\n",
				clean.Humanize(c.Size), idleDaysLabel(c.IdleDays()), relPath(c.Path, home))
			total += c.Size
		}
		fmt.Printf("\n合計: %s\n", clean.Humanize(total))
	}
}

func idleDaysLabel(days int) string {
	if days < 0 {
		return "未知"
	}
	if days == 0 {
		return "今天有動"
	}
	return fmt.Sprintf("閒置 %d 天", days)
}
