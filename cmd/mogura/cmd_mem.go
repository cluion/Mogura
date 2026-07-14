package main

import (
	"fmt"
	"time"

	"mogura/internal/clean"
	"mogura/internal/i18n"
	"mogura/internal/memory"
)

const topN = 15

func runMem(args []string) error {
	action := ""
	for _, a := range args {
		switch a {
		case "--drop-caches", "--swap-reset":
			action = a
		default:
			usage()
			return fmt.Errorf(i18n.T("未知選項: %s"), a)
		}
	}

	before, err := memory.Read()
	if err != nil {
		return err
	}
	printMemStats(before)

	switch action {
	case "":
		return printTop()
	case "--drop-caches":
		fmt.Println(i18n.T("\n提醒:page cache 平常會自動回收,清除通常不必要,且會讓系統短暫變慢。"))
		return runMemAction(i18n.T("清除 page cache"), memory.DropCaches, before)
	case "--swap-reset":
		if before.SwapUsed == 0 {
			fmt.Println(i18n.T("\nswap 未被使用,不需要重置。"))
			return nil
		}
		if before.Available < before.SwapUsed*2 {
			return fmt.Errorf(i18n.T("可用記憶體不足以安全收回 swap(需要約 %s)"), clean.Humanize(int64(before.SwapUsed*2)))
		}
		fmt.Println(i18n.T("\n將把 swap 內容搬回 RAM,期間系統可能短暫變慢。"))
		return runMemAction(i18n.T("重置 swap"), memory.SwapReset, before)
	}
	return nil
}

func printMemStats(s memory.Stats) {
	fmt.Print(i18n.Tf("記憶體  %s / %s(可用 %s · cache %s)\n",
		clean.Humanize(int64(s.Used)), clean.Humanize(int64(s.Total)),
		clean.Humanize(int64(s.Available)), clean.Humanize(int64(s.Cached))))
	if s.SwapTotal > 0 {
		fmt.Print(i18n.Tf("swap    %s / %s\n",
			clean.Humanize(int64(s.SwapUsed)), clean.Humanize(int64(s.SwapTotal))))
	}
	fmt.Println(i18n.T("「可用」才是真實可用量,cache 由 kernel 自動回收。"))
}

func printTop() error {
	procs, err := memory.Top(topN)
	if err != nil {
		return err
	}
	fmt.Print(i18n.Tf("\n記憶體佔用前 %d 名:\n", len(procs)))
	for i, p := range procs {
		fmt.Printf("  %2d. %10s  %-30s pid %d\n",
			i+1, clean.Humanize(int64(p.RSS)), p.Name, p.PID)
	}
	fmt.Println(i18n.T("\n釋放操作(需要 sudo):mogura mem --drop-caches · mogura mem --swap-reset"))
	return nil
}

func runMemAction(name string, fn func() error, before memory.Stats) error {
	fmt.Print(i18n.Tf("%s 需要 sudo,執行時可能要求輸入密碼。\n", name))
	if !promptYes() {
		fmt.Println(i18n.T("已取消。"))
		return nil
	}
	fmt.Println(i18n.T("執行中,資料量大時可能需要數十秒..."))
	start := time.Now()
	if err := fn(); err != nil {
		return err
	}
	after, err := memory.Read()
	if err != nil {
		return err
	}

	fmt.Print(i18n.Tf("\n✨ 完成(耗時 %s)\n", time.Since(start).Round(time.Second)))
	fmt.Print(i18n.Tf("  可用記憶體 %s → %s(%s)\n",
		clean.Humanize(int64(before.Available)), clean.Humanize(int64(after.Available)),
		signedDiff(int64(after.Available)-int64(before.Available))))
	if before.SwapTotal > 0 && before.SwapUsed != after.SwapUsed {
		fmt.Print(i18n.Tf("  swap 使用   %s → %s(內容已搬回 RAM)\n",
			clean.Humanize(int64(before.SwapUsed)), clean.Humanize(int64(after.SwapUsed))))
	}
	return nil
}

func signedDiff(diff int64) string {
	sign := "+"
	if diff < 0 {
		sign = "-"
		diff = -diff
	}
	return sign + clean.Humanize(diff)
}
