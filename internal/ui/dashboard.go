package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"mogura/internal/clean"
	"mogura/internal/i18n"
)

// MenuItem 是總覽選單的一個項目;Label/Desc 存原文,顯示時才翻譯,
// 讓面板內切語言立即生效
type MenuItem struct {
	ID    string
	Label string
	Desc  string
}

type dashModel struct {
	items    []MenuItem
	cursor   int
	choice   string
	prog     *clean.Progress
	total    func() (int64, bool) // (可估算合計, 掃描是否完成)
	settings *Settings
}

type dashTickMsg struct{}

func dashTick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg { return dashTickMsg{} })
}

// RunDashboard 顯示總覽選單,背景掃描的進度即時更新
// 回傳使用者選擇的項目 ID,離開時回傳空字串
func RunDashboard(items []MenuItem, prog *clean.Progress, total func() (int64, bool)) (string, error) {
	m := dashModel{items: items, prog: prog, total: total}
	final, err := tea.NewProgram(m).Run()
	if err != nil {
		return "", fmt.Errorf(i18n.T("互動介面啟動失敗: %w"), err)
	}
	return final.(dashModel).choice, nil
}

func (m dashModel) Init() tea.Cmd { return dashTick() }

func (m dashModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case dashTickMsg:
		if _, done := m.total(); !done {
			return m, dashTick() // 掃描中每 0.1 秒重繪進度
		}
		return m, nil
	case tea.KeyMsg:
		if m.settings != nil {
			s, closed := m.settings.HandleKey(msg)
			if closed {
				m.settings = nil // 顯示走 render-time 翻譯,語言切換即時生效
			} else {
				m.settings = &s
			}
			return m, nil
		}
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "enter":
			m.choice = m.items[m.cursor].ID
			return m, tea.Quit
		case ",":
			s := NewSettings()
			m.settings = &s
		}
	}
	return m, nil
}

func (m dashModel) View() string {
	if m.settings != nil {
		return m.settings.View()
	}
	var b strings.Builder
	b.WriteString(titleStyle.Render(i18n.Brand("Mogura — 總覽")) + "\n\n")

	if sum, done := m.total(); done {
		b.WriteString("  " + i18n.T("可回收空間 ") + totalStyle.Render(clean.Humanize(sum)) +
			descStyle.Render(i18n.T("(可估算項目)")) + "\n\n")
	} else {
		b.WriteString("  " + descStyle.Render(i18n.Tf("可回收空間 掃描中... %s · %s 檔",
			clean.Humanize(m.prog.Bytes()), clean.GroupDigits(m.prog.Files()))) + "\n\n")
	}

	for i, it := range m.items {
		cursor := "  "
		if i == m.cursor {
			cursor = cursorStyle.Render("❯ ")
		}
		label := i18n.T(it.Label)
		line := cursor + label
		if it.Desc != "" {
			// CJK 全形字佔兩格,用顯示寬度對齊說明欄
			if pad := 21 - lipgloss.Width(label); pad > 0 {
				line += strings.Repeat(" ", pad)
			} else {
				line += "  "
			}
			line += descStyle.Render(i18n.T(it.Desc))
		}
		b.WriteString(line + "\n")
	}

	b.WriteString(helpStyle.Render(i18n.T("\n↑↓ 移動 · enter 進入 · , 設定 · q 離開")))
	return b.String()
}
