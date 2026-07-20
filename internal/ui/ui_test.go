package ui

import (
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"mogura/internal/i18n"
)

// TestMain 把語系固定成中文原文,i18n.T 便原樣返回,
// 斷言可以直接寫程式碼裡的中文字串而不受環境語系影響
func TestMain(m *testing.M) {
	i18n.SetEnglish(false)
	os.Exit(m.Run())
}

// keyMsg 把按鍵名稱轉成 bubbletea 訊息;名稱與 KeyMsg.String 一致,
// 測試才能用 "up"、"enter" 這種可讀的字面值描述操作
func keyMsg(name string) tea.KeyMsg {
	switch name {
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "left":
		return tea.KeyMsg{Type: tea.KeyLeft}
	case "right":
		return tea.KeyMsg{Type: tea.KeyRight}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case " ":
		return tea.KeyMsg{Type: tea.KeySpace}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(name)}
	}
}

// TestKeyMsgHelper 守住輔助函式本身:名稱與 bubbletea 的 String 對不上時,
// 其他測試會用錯的按鍵通過,反而測不到東西
func TestKeyMsgHelper(t *testing.T) {
	for _, name := range []string{"up", "down", "left", "right", "enter", "esc", " ", "ctrl+c", "q", "a", "n", "x", ","} {
		if got := keyMsg(name).String(); got != name {
			t.Errorf("keyMsg(%q).String() = %q", name, got)
		}
	}
}
