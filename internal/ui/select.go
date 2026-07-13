// Package ui 提供終端機互動元件。
package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"mogura/internal/clean"
)

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	sizeStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Width(10).Align(lipgloss.Right)
	descStyle   = lipgloss.NewStyle().Faint(true)
	helpStyle   = lipgloss.NewStyle().Faint(true).MarginTop(1)
	totalStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42"))

	riskStyles = map[string]lipgloss.Style{
		"low":    lipgloss.NewStyle().Foreground(lipgloss.Color("42")),
		"medium": lipgloss.NewStyle().Foreground(lipgloss.Color("214")),
		"high":   lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
	}
	riskLabels = map[string]string{"low": "低", "medium": "中", "high": "高"}
)

type item struct {
	res      clean.Result
	selected bool
}

type model struct {
	items   []item
	cursor  int
	aborted bool
}

// Select 顯示互動多選清單,回傳使用者勾選的項目;取消時回傳 nil。
func Select(results []clean.Result) ([]clean.Result, error) {
	items := make([]item, len(results))
	for i, r := range results {
		items[i] = item{res: r}
	}
	final, err := tea.NewProgram(model{items: items}).Run()
	if err != nil {
		return nil, fmt.Errorf("互動介面啟動失敗: %w", err)
	}
	m := final.(model)
	if m.aborted {
		return nil, nil
	}
	var selected []clean.Result
	for _, it := range m.items {
		if it.selected {
			selected = append(selected, it.res)
		}
	}
	return selected, nil
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "q", "esc", "ctrl+c":
		m.aborted = true
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.items)-1 {
			m.cursor++
		}
	case " ":
		m.items[m.cursor].selected = !m.items[m.cursor].selected
	case "a":
		for i := range m.items {
			m.items[i].selected = true
		}
	case "n":
		for i := range m.items {
			m.items[i].selected = false
		}
	case "enter":
		return m, tea.Quit
	}
	return m, nil
}

func (m model) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("🦡 Mogura — 選擇要清理的項目") + "\n\n")

	var total int64
	for _, it := range m.items {
		if it.selected && it.res.Known {
			total += it.res.Size
		}
	}

	for i, it := range m.items {
		cursor := "  "
		if i == m.cursor {
			cursor = cursorStyle.Render("❯ ")
		}
		check := "[ ]"
		if it.selected {
			check = cursorStyle.Render("[✓]")
		}
		size := "—"
		if it.res.Known {
			size = clean.Humanize(it.res.Size)
		}
		risk := riskStyles[it.res.Rule.Risk].Render("風險" + riskLabels[it.res.Rule.Risk])
		lock := "  "
		if it.res.Rule.Root {
			lock = "🔒"
		}
		b.WriteString(fmt.Sprintf("%s%s %s %s %s %s\n",
			cursor, check, sizeStyle.Render(size), risk, lock, it.res.Rule.Name))
		if i == m.cursor {
			b.WriteString("        " + descStyle.Render(it.res.Rule.Description) + "\n")
		}
	}

	b.WriteString("\n" + totalStyle.Render("已選擇可回收: "+clean.Humanize(total)))
	b.WriteString(helpStyle.Render("\n空白鍵 勾選 · a 全選 · n 全不選 · enter 執行 · q 離開 · 🔒 需要 sudo"))
	return b.String()
}
