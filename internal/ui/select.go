// Package ui 提供終端機互動元件。
package ui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"mogura/internal/clean"
	"mogura/internal/config"
	"mogura/internal/i18n"
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
// Path 非空時,該項目可用 x 加入全域排除清單。
type Option struct {
	Label string
	Desc  string
	Size  int64
	Known bool
	Risk  string // low/medium/high,空字串不顯示
	Root  bool
	Path  string
	Value any
}

type item struct {
	opt      Option
	selected bool
}

type model struct {
	title    string
	items    []item
	cursor   int
	height   int
	aborted  bool
	restart  bool
	status   string
	settings *Settings
}

// MultiSelect 顯示互動多選清單,回傳使用者勾選的項目;取消時回傳 nil。
// restart 為 true 表示使用者在面板中切換了語言,呼叫端應重建流程再進來。
func MultiSelect(title string, opts []Option) (selected []Option, restart bool, err error) {
	items := make([]item, len(opts))
	for i, o := range opts {
		items[i] = item{opt: o}
	}
	final, err := tea.NewProgram(model{title: title, items: items, height: 24}).Run()
	if err != nil {
		return nil, false, fmt.Errorf(i18n.T("互動介面啟動失敗: %w"), err)
	}
	m := final.(model)
	if m.restart {
		return nil, true, nil
	}
	if m.aborted {
		return nil, false, nil
	}
	for _, it := range m.items {
		if it.selected {
			selected = append(selected, it.opt)
		}
	}
	return selected, false, nil
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		m.height = size.Height
		return m, nil
	}
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	if m.settings != nil {
		s, closed := m.settings.HandleKey(key)
		if closed {
			if s.Changed() {
				m.restart = true // 語言變了,請宿主重建整個選單
				return m, tea.Quit
			}
			m.settings = nil
		} else {
			m.settings = &s
		}
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
		if len(m.items) > 0 {
			m.items[m.cursor].selected = !m.items[m.cursor].selected
		}
	case "x":
		m.excludeCurrent()
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
	case ",":
		s := NewSettings()
		m.settings = &s
	}
	return m, nil
}

// excludeCurrent 把游標項目的路徑寫進全域排除清單,並從清單移除。
func (m *model) excludeCurrent() {
	if len(m.items) == 0 {
		return
	}
	opt := m.items[m.cursor].opt
	if opt.Path == "" {
		m.status = i18n.T("此項目無法排除(非單一路徑)")
		return
	}
	if err := config.AddExclude(shortenHome(opt.Path)); err != nil {
		m.status = i18n.T("設定儲存失敗:") + err.Error()
		return
	}
	m.items = append(m.items[:m.cursor], m.items[m.cursor+1:]...)
	if m.cursor >= len(m.items) && m.cursor > 0 {
		m.cursor--
	}
	m.status = i18n.Tf("已排除 %s,之後掃描不再顯示(設定檔可移除)", shortenHome(opt.Path))
}

// shortenHome 把家目錄前綴縮寫成 ~,排除清單存起來跨機器可攜。
func shortenHome(p string) string {
	if home, err := os.UserHomeDir(); err == nil && strings.HasPrefix(p, home+"/") {
		return "~" + strings.TrimPrefix(p, home)
	}
	return p
}

func (m model) View() string {
	if m.settings != nil {
		return m.settings.View()
	}
	var b strings.Builder
	b.WriteString(titleStyle.Render("🦡 "+m.title) + "\n\n")

	var total int64
	for _, it := range m.items {
		if it.selected && it.opt.Known {
			total += it.opt.Size
		}
	}

	// 視窗捲動:游標行加上描述行,保證永遠在可視範圍內
	visible := m.height - 7
	if visible < 5 {
		visible = 5
	}
	start := 0
	if m.cursor >= visible {
		start = m.cursor - visible + 1
	}
	end := start + visible
	if end > len(m.items) {
		end = len(m.items)
	}

	for i := start; i < end; i++ {
		it := m.items[i]
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
			risk = riskStyles[it.opt.Risk].Render(i18n.T(riskLabels[it.opt.Risk])) + " "
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

	b.WriteString("\n" + totalStyle.Render(i18n.T("已選擇可回收: ")+clean.Humanize(total)))
	if m.status != "" {
		b.WriteString("\n" + descStyle.Render(m.status))
	}
	b.WriteString(helpStyle.Render(i18n.T("\n空白鍵 勾選 · a 全選 · n 全不選 · enter 執行 · x 排除 · , 設定 · q 離開 · 🔒 需要 sudo")))
	return b.String()
}
