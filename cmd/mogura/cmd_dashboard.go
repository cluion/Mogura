package main

import (
	"fmt"

	"mogura/internal/clean"
	"mogura/internal/i18n"
	"mogura/internal/rules"
	"mogura/internal/ui"
)

// 選單項目存原文,ui 端顯示時才翻譯,面板內切語言立即生效
var dashMenu = []ui.MenuItem{
	{ID: "clean", Label: "清理系統垃圾", Desc: "快取、垃圾桶、套件快取與日誌"},
	{ID: "analyze", Label: "磁碟空間分析", Desc: "互動瀏覽各目錄佔用"},
	{ID: "dev", Label: "開發垃圾", Desc: "node_modules、target 等建置產物"},
	{ID: "orphan", Label: "孤兒設定檔", Desc: "已解除安裝軟體留下的殘留設定"},
	{ID: "monitor", Label: "系統監控", Desc: "CPU、記憶體、磁碟、網路"},
	{ID: "mem", Label: "記憶體", Desc: "大戶排行與釋放操作"},
	{ID: "config", Label: "設定", Desc: "語言、刪除方式、journal 保留"},
	{ID: "quit", Label: "離開", Desc: ""},
}

// runDashboard 是無參數時的總覽:選單秒開,可回收總量背景掃描即時更新
// 子功能結束後回到總覽並重新掃描
func runDashboard() error {
	for {
		prog := &clean.Progress{}
		done := make(chan struct{})
		var results []clean.Result
		go func() {
			defer close(done)
			if rs, err := rules.Load(ruleOptions()); err == nil {
				results = clean.ScanAll(rs, prog)
			}
		}()
		total := func() (int64, bool) {
			select {
			case <-done:
			default:
				return 0, false
			}
			var sum int64
			for _, r := range results {
				if r.Known {
					sum += r.Size
				}
			}
			return sum, true
		}

		choice, err := ui.RunDashboard(dashMenu, prog, total)
		if err != nil {
			return err
		}

		var ferr error
		switch choice {
		case "", "quit":
			return nil
		case "clean":
			progressLoop(i18n.T("掃描系統垃圾中..."), prog, done) // 通常已完成,秒過
			ferr = cleanInteract(results)
			pause()
		case "analyze":
			ferr = runAnalyze(nil)
		case "dev":
			ferr = runDev(nil)
			pause()
		case "orphan":
			ferr = runOrphan(nil)
			pause()
		case "monitor":
			ferr = runMonitor(nil)
		case "mem":
			ferr = runMem(nil)
			pause()
		case "config":
			ferr = runConfig(nil)
		}
		if ferr != nil {
			return ferr
		}
	}
}

// pause 讓有列印輸出的子功能結果停留在畫面上,再返回總覽
func pause() {
	fmt.Print(i18n.T("\n按 Enter 返回總覽..."))
	_, _ = stdin.ReadString('\n')
}
