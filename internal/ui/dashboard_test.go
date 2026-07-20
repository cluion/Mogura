package ui

import (
	"strings"
	"testing"

	"mogura/internal/clean"
)

func newDash(total func() (int64, bool)) dashModel {
	return dashModel{
		items: []MenuItem{
			{ID: "clean", Label: "清理系統垃圾", Desc: "快取、垃圾桶"},
			{ID: "analyze", Label: "磁碟空間分析", Desc: "互動瀏覽"},
			{ID: "quit", Label: "離開"},
		},
		prog:  &clean.Progress{},
		total: total,
	}
}

func scanned(sum int64) func() (int64, bool) { return func() (int64, bool) { return sum, true } }
func scanning() func() (int64, bool)         { return func() (int64, bool) { return 0, false } }

func pressDash(m dashModel, keys ...string) dashModel {
	for _, k := range keys {
		next, _ := m.Update(keyMsg(k))
		m = next.(dashModel)
	}
	return m
}

func TestDashCursorStaysInBounds(t *testing.T) {
	for _, tc := range []struct {
		name string
		keys []string
		want int
	}{
		{"頂端再上移不動", []string{"up"}, 0},
		{"底端再下移不動", []string{"down", "down", "down", "down"}, 2},
		{"vim 鍵同樣有效", []string{"j", "j", "k"}, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := pressDash(newDash(scanned(0)), tc.keys...).cursor; got != tc.want {
				t.Errorf("游標 = %d, 預期 %d", got, tc.want)
			}
		})
	}
}

func TestDashEnterRecordsChoice(t *testing.T) {
	next, cmd := pressDash(newDash(scanned(0)), "down").Update(keyMsg("enter"))
	m := next.(dashModel)
	if m.choice != "analyze" {
		t.Errorf("選擇 = %q, 預期游標所在的 analyze", m.choice)
	}
	if !isQuit(cmd) {
		t.Error("enter 應結束總覽,交還給呼叫端執行子功能")
	}
}

func TestDashQuitLeavesChoiceEmpty(t *testing.T) {
	for _, key := range []string{"q", "esc", "ctrl+c"} {
		t.Run(key, func(t *testing.T) {
			next, cmd := newDash(scanned(0)).Update(keyMsg(key))
			if choice := next.(dashModel).choice; choice != "" {
				t.Errorf("離開不該產生選擇,實際 %q", choice)
			}
			if !isQuit(cmd) {
				t.Errorf("%s 應結束總覽", key)
			}
		})
	}
}

func TestDashTickStopsOnceScanDone(t *testing.T) {
	if _, cmd := newDash(scanning()).Update(dashTickMsg{}); cmd == nil {
		t.Error("掃描中應持續排下一次 tick 以更新進度")
	}
	if _, cmd := newDash(scanned(1024)).Update(dashTickMsg{}); cmd != nil {
		t.Error("掃描完成後應停止 tick,不必再空轉重繪")
	}
}

func TestDashSettingsPanelOpensAndCloses(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	m := pressDash(newDash(scanned(0)), ",")
	if m.settings == nil {
		t.Fatal("逗號應開啟設定面板")
	}
	if m = pressDash(m, "down"); m.cursor != 0 {
		t.Errorf("面板開啟時方向鍵屬於面板,選單游標不該動,實際 %d", m.cursor)
	}
	if m = pressDash(m, "enter"); m.settings != nil {
		t.Error("enter 應關閉設定面板")
	}
	if m.choice != "" {
		t.Error("關閉面板的 enter 不該被當成選擇項目")
	}
}

func TestDashViewReflectsScanState(t *testing.T) {
	scanningOut := newDash(scanning()).View()
	if !strings.Contains(scanningOut, "掃描中") {
		t.Error("掃描未完成時應顯示進度而非合計")
	}

	doneOut := newDash(scanned(2 * 1024 * 1024 * 1024)).View()
	if strings.Contains(doneOut, "掃描中") {
		t.Error("掃描完成後不該還說掃描中")
	}
	for _, want := range []string{"可回收空間", "2.0 GiB", "清理系統垃圾", "磁碟空間分析"} {
		if !strings.Contains(doneOut, want) {
			t.Errorf("畫面應包含 %q", want)
		}
	}
}
