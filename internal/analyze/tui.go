package analyze

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"mogura/internal/clean"
)

const barWidth = 20

var (
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	pathStyle   = lipgloss.NewStyle().Faint(true)
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	sizeStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Width(10).Align(lipgloss.Right)
	barStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	barBgStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	dirStyle    = lipgloss.NewStyle().Bold(true)
	helpStyle   = lipgloss.NewStyle().Faint(true).MarginTop(1)
)

type loadedMsg struct {
	dir     string
	entries []Entry
	err     error
}

type browser struct {
	sizer   *Sizer
	root    string
	cwd     string
	entries []Entry
	cursor  int
	loading bool
	errMsg  string
	height  int
}

// Browse 啟動磁碟分析瀏覽器,從 root 開始向下鑽。
func Browse(root string) error {
	b := browser{sizer: NewSizer(), root: root, cwd: root, loading: true, height: 24}
	_, err := tea.NewProgram(b, tea.WithAltScreen()).Run()
	return err
}

func (b browser) load(dir string) tea.Cmd {
	return func() tea.Msg {
		entries, err := b.sizer.List(dir)
		return loadedMsg{dir: dir, entries: entries, err: err}
	}
}

func (b browser) Init() tea.Cmd { return b.load(b.cwd) }

func (b browser) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		b.height = msg.Height
		return b, nil
	case loadedMsg:
		if msg.err != nil {
			b.errMsg = msg.err.Error()
			b.loading = false
			return b, nil
		}
		b.cwd = msg.dir
		b.entries = msg.entries
		b.cursor = 0
		b.loading = false
		b.errMsg = ""
		return b, nil
	case tea.KeyMsg:
		return b.handleKey(msg)
	}
	return b, nil
}

func (b browser) handleKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "q", "esc", "ctrl+c":
		return b, tea.Quit
	case "up", "k":
		if b.cursor > 0 {
			b.cursor--
		}
	case "down", "j":
		if b.cursor < len(b.entries)-1 {
			b.cursor++
		}
	case "enter", "right", "l":
		if !b.loading && b.cursor < len(b.entries) && b.entries[b.cursor].IsDir {
			b.loading = true
			return b, b.load(b.entries[b.cursor].Path)
		}
	case "backspace", "left", "h":
		if !b.loading && b.cwd != b.root {
			b.loading = true
			return b, b.load(filepath.Dir(b.cwd))
		}
	}
	return b, nil
}

func (b browser) View() string {
	var sb strings.Builder
	sb.WriteString(headerStyle.Render("🦡 Mogura 磁碟分析") + "  " + pathStyle.Render(b.cwd) + "\n\n")

	if b.loading {
		sb.WriteString("掃描中,大目錄需要一點時間...\n")
		return sb.String()
	}
	if b.errMsg != "" {
		sb.WriteString("讀取失敗: " + b.errMsg + "\n")
		sb.WriteString(helpStyle.Render("backspace 返回上層 · q 離開"))
		return sb.String()
	}

	var max int64 = 1
	if len(b.entries) > 0 && b.entries[0].Size > 0 {
		max = b.entries[0].Size
	}

	visible := b.height - 6
	if visible < 5 {
		visible = 5
	}
	start := 0
	if b.cursor >= visible {
		start = b.cursor - visible + 1
	}
	end := start + visible
	if end > len(b.entries) {
		end = len(b.entries)
	}

	for i := start; i < end; i++ {
		e := b.entries[i]
		cursor := "  "
		if i == b.cursor {
			cursor = cursorStyle.Render("❯ ")
		}
		filled := int(int64(barWidth) * e.Size / max)
		bar := barStyle.Render(strings.Repeat("█", filled)) +
			barBgStyle.Render(strings.Repeat("░", barWidth-filled))
		name := e.Name
		if e.IsDir {
			name = dirStyle.Render(name + "/")
		}
		sb.WriteString(fmt.Sprintf("%s%s %s %s\n", cursor, sizeStyle.Render(clean.Humanize(e.Size)), bar, name))
	}
	if len(b.entries) == 0 {
		sb.WriteString(pathStyle.Render("(空目錄)") + "\n")
	}

	sb.WriteString(helpStyle.Render("enter 進入 · backspace 上層 · q 離開"))
	return sb.String()
}
