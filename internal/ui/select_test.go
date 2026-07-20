package ui

import (
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"mogura/internal/config"
)

func newModel(opts ...Option) model {
	items := make([]item, len(opts))
	for i, o := range opts {
		items[i] = item{opt: o}
	}
	return model{title: "測試清單", items: items, height: 24}
}

// press 依序送入按鍵並回傳最後的 model,讓測試以操作序列描述情境
func press(m model, keys ...string) model {
	for _, k := range keys {
		next, _ := m.Update(keyMsg(k))
		m = next.(model)
	}
	return m
}

// isQuit 回報 Cmd 是否為結束指令;Cmd 是函式無法直接比較,執行後看訊息型別
func isQuit(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	_, ok := cmd().(tea.QuitMsg)
	return ok
}

func threeItems() model {
	return newModel(
		Option{Label: "甲", Size: 300, Known: true, Path: "/tmp/a"},
		Option{Label: "乙", Size: 200, Known: true, Path: "/tmp/b"},
		Option{Label: "丙", Size: 100, Known: true, Path: "/tmp/c"},
	)
}

func TestCursorStaysInBounds(t *testing.T) {
	for _, tc := range []struct {
		name string
		keys []string
		want int
	}{
		{"頂端再上移不動", []string{"up", "up"}, 0},
		{"下移一格", []string{"down"}, 1},
		{"底端再下移不動", []string{"down", "down", "down", "down"}, 2},
		{"下到底再回頭", []string{"down", "down", "up"}, 1},
		{"vim 鍵同樣有效", []string{"j", "j", "k"}, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := press(threeItems(), tc.keys...).cursor; got != tc.want {
				t.Errorf("游標 = %d, 預期 %d", got, tc.want)
			}
		})
	}
}

func TestSpaceTogglesSelection(t *testing.T) {
	m := press(threeItems(), " ")
	if !m.items[0].selected {
		t.Error("空白鍵應勾選游標項目")
	}
	if m = press(m, " "); m.items[0].selected {
		t.Error("再按一次空白鍵應取消勾選")
	}
}

func TestSelectAllAndNone(t *testing.T) {
	m := press(threeItems(), "a")
	for i, it := range m.items {
		if !it.selected {
			t.Errorf("a 應全選,第 %d 項未選", i)
		}
	}
	m = press(m, "n")
	for i, it := range m.items {
		if it.selected {
			t.Errorf("n 應全不選,第 %d 項仍選著", i)
		}
	}
}

func TestAbortKeys(t *testing.T) {
	for _, key := range []string{"q", "esc", "ctrl+c"} {
		t.Run(key, func(t *testing.T) {
			next, cmd := threeItems().Update(keyMsg(key))
			m := next.(model)
			if !m.aborted {
				t.Errorf("%s 應標記為中止", key)
			}
			if !isQuit(cmd) {
				t.Errorf("%s 應回傳結束指令", key)
			}
		})
	}
}

func TestEnterConfirmsWithoutAborting(t *testing.T) {
	next, cmd := press(threeItems(), " ").Update(keyMsg("enter"))
	m := next.(model)
	if m.aborted {
		t.Error("enter 是確認,不該標記中止")
	}
	if !isQuit(cmd) {
		t.Error("enter 應回傳結束指令")
	}
	if !m.items[0].selected {
		t.Error("enter 不該清掉既有勾選")
	}
}

func TestExcludeRemovesItemAndPersists(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	m := press(threeItems(), "down", "x") // 排除第二項

	if len(m.items) != 2 {
		t.Fatalf("排除後應剩 2 項,實際 %d 項", len(m.items))
	}
	if m.items[0].opt.Label != "甲" || m.items[1].opt.Label != "丙" {
		t.Errorf("移除的應是游標所在的乙,實際剩 %q、%q", m.items[0].opt.Label, m.items[1].opt.Label)
	}
	if ex := config.Load().Exclude; len(ex) != 1 || ex[0] != "/tmp/b" {
		t.Errorf("排除清單 = %v, 預期寫入 /tmp/b", ex)
	}
	if !strings.Contains(m.status, "已排除") {
		t.Errorf("狀態列應提示已排除,實際 %q", m.status)
	}
}

func TestExcludeClampsCursorAtTail(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	m := press(threeItems(), "down", "down", "x") // 排除最後一項

	if m.cursor != 1 {
		t.Errorf("排除末項後游標 = %d, 預期退回 1 以免指向不存在的項目", m.cursor)
	}
	if m.cursor >= len(m.items) {
		t.Errorf("游標 %d 超出項目數 %d", m.cursor, len(m.items))
	}
}

func TestExcludeRejectsItemWithoutPath(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	m := press(newModel(Option{Label: "多路徑合併項"}), "x")

	if len(m.items) != 1 {
		t.Error("無單一路徑的項目不該被移除")
	}
	if len(config.Load().Exclude) != 0 {
		t.Error("無單一路徑的項目不該寫進排除清單")
	}
	if !strings.Contains(m.status, "無法排除") {
		t.Errorf("狀態列應說明無法排除,實際 %q", m.status)
	}
}

func TestExcludeOnEmptyListIsNoop(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	m := press(newModel(), "x", " ", "a", "n") // 空清單按任何鍵都不該 panic

	if len(m.items) != 0 {
		t.Errorf("空清單不該生出項目,實際 %d 項", len(m.items))
	}
}

func TestShortenHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	for _, tc := range []struct{ in, want string }{
		{filepath.Join(home, ".cache/pip"), "~/.cache/pip"},
		{"/opt/data", "/opt/data"},
		{home, home}, // 家目錄本身不縮寫,避免排除清單寫成裸 ~
	} {
		if got := shortenHome(tc.in); got != tc.want {
			t.Errorf("shortenHome(%q) = %q, 預期 %q", tc.in, got, tc.want)
		}
	}
}

func TestSettingsPanelTakesOverKeys(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	m := press(threeItems(), ",")
	if m.settings == nil {
		t.Fatal("逗號應開啟設定面板")
	}
	if m = press(m, "down"); m.cursor != 0 {
		t.Errorf("面板開啟時方向鍵屬於面板,清單游標不該動,實際 %d", m.cursor)
	}
}

func TestViewRendersItemsAndTotal(t *testing.T) {
	out := press(threeItems(), " ").View()
	for _, want := range []string{"測試清單", "甲", "乙", "丙", "已選擇可回收"} {
		if !strings.Contains(out, want) {
			t.Errorf("畫面應包含 %q", want)
		}
	}
}

func TestViewScrollsToKeepCursorVisible(t *testing.T) {
	opts := make([]Option, 40)
	for i := range opts {
		opts[i] = Option{Label: string(rune('A' + i%26))}
	}
	m := newModel(opts...)
	m.cursor = 39
	if out := m.View(); !strings.Contains(out, "❯") {
		t.Error("游標捲到末端時,畫面仍應顯示游標所在列")
	}
}
