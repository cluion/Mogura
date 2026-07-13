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
	riskLabels = map[string]string{"low": "風險低", "medium": "風險中", "high": "風險高"}
)

// Option 是多選清單中的一個項目,Value 由呼叫端夾帶原始資料。
type Option struct {
	Label string
	Desc  string
	Size  int64
	Known bool
	Risk  string // low/medium/high,空字串不顯示
	Root  bool
	Value any
}

type item struct {
	opt      Option
	selected bool
}

type model struct {
	title   string
	items   []item
	cursor  int
	aborted bool
}

// MultiSelect 顯示互動多選清單,回傳使用者勾選的項目;取消時回傳 nil。
func MultiSelect(title string, opts []Option) ([]Option, error) {
	items := make([]item, len(opts))
	for i, o := range opts {
		items[i] = item{opt: o}
	}
	final, err := tea.NewProgram(model{title: title, items: items}).Run()
	if err != nil {
		return nil, fmt.Errorf("互動介面啟動失敗: %w", err)
	}
	m := final.(model)
	if m.aborted {
		return nil, nil
	}
	var selected []Option
	for _, it := range m.items {
		if it.selected {
			selected = append(selected, it.opt)
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
	b.WriteString(titleStyle.Render("🦡 "+m.title) + "\n\n")

	var total int64
	for _, it := range m.items {
		if it.selected && it.opt.Known {
			total += it.opt.Size
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
		if it.opt.Known {
			size = clean.Humanize(it.opt.Size)
		}
		risk := ""
		if it.opt.Risk != "" {
			risk = riskStyles[it.opt.Risk].Render(riskLabels[it.opt.Risk]) + " "
		}
		lock := "  "
		if it.opt.Root {
			lock = "🔒"
		}
		b.WriteString(fmt.Sprintf("%s%s %s %s%s %s\n",
			cursor, check, sizeStyle.Render(size), risk, lock, it.opt.Label))
		if i == m.cursor && it.opt.Desc != "" {
			b.WriteString("        " + descStyle.Render(it.opt.Desc) + "\n")
		}
	}

	b.WriteString("\n" + totalStyle.Render("已選擇可回收: "+clean.Humanize(total)))
	b.WriteString(helpStyle.Render("\n空白鍵 勾選 · a 全選 · n 全不選 · enter 執行 · q 離開 · 🔒 需要 sudo"))
	return b.String()
}
