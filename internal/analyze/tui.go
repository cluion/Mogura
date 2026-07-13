package analyze

import (
	"errors"
	"fmt"
	"os"
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
	dangerLine  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
	okLine      = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
)

type loadedMsg struct {
	dir     string
	entries []Entry
	err     error
}

type deletedMsg struct {
	entry Entry
	err   error
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

	confirm  *Entry // 非 nil 時處於刪除確認狀態
	deleting bool
	status   string
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

func deleteEntry(e Entry) tea.Cmd {
	return func() tea.Msg {
		if err := deleteGuard(e.Path); err != nil {
			return deletedMsg{entry: e, err: err}
		}
		return deletedMsg{entry: e, err: os.RemoveAll(e.Path)}
	}
}

// deleteGuard 是 TUI 刪除的防呆:擋根目錄、第一層系統目錄與家目錄本身。
func deleteGuard(path string) error {
	if !filepath.IsAbs(path) || path == "/" {
		return errors.New("拒絕刪除")
	}
	if strings.Count(path, "/") < 2 {
		return errors.New("拒絕刪除第一層系統目錄")
	}
	if home, err := os.UserHomeDir(); err == nil && filepath.Clean(path) == filepath.Clean(home) {
		return errors.New("拒絕刪除家目錄")
	}
	return nil
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
		if msg.dir != b.cwd {
			b.cursor = 0 // 換目錄回到頂端;原地刷新(如刪除後)保留游標
		}
		b.cwd = msg.dir
		b.entries = msg.entries
		if b.cursor >= len(b.entries) {
			b.cursor = 0
		}
		b.loading = false
		b.errMsg = ""
		return b, nil
	case deletedMsg:
		b.deleting = false
		if msg.err != nil {
			b.status = dangerLine.Render("刪除失敗:" + msg.err.Error())
			return b, nil
		}
		b.status = okLine.Render(fmt.Sprintf("已刪除 %s,釋放 %s", msg.entry.Name, clean.Humanize(msg.entry.Size)))
		b.sizer.Invalidate(msg.entry.Path)
		b.loading = true
		return b, b.load(b.cwd)
	case tea.KeyMsg:
		return b.handleKey(msg)
	}
	return b, nil
}

func (b browser) handleKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	// 刪除確認狀態:只接受 y / 其他鍵取消
	if b.confirm != nil {
		target := *b.confirm
		b.confirm = nil
		if key.String() == "y" {
			b.deleting = true
			b.status = ""
			return b, deleteEntry(target)
		}
		b.status = "已取消刪除。"
		return b, nil
	}

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
			target := b.entries[b.cursor].Path
			b.loading = true
			return b, b.load(target)
		}
	case "backspace", "left", "h":
		if !b.loading && b.cwd != b.root {
			b.loading = true
			return b, b.load(filepath.Dir(b.cwd))
		}
	case "d":
		if !b.loading && !b.deleting && b.cursor < len(b.entries) {
			e := b.entries[b.cursor]
			b.confirm = &e
			b.status = ""
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

	visible := b.height - 7
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

	switch {
	case b.confirm != nil:
		sb.WriteString("\n" + dangerLine.Render(fmt.Sprintf("刪除 %s(%s)?此操作無法復原  y 確認 · 其他鍵取消",
			b.confirm.Name, clean.Humanize(b.confirm.Size))))
	case b.deleting:
		sb.WriteString("\n刪除中...")
	case b.status != "":
		sb.WriteString("\n" + b.status)
	}

	sb.WriteString(helpStyle.Render("\nenter 進入 · backspace 上層 · d 刪除 · q 離開"))
	return sb.String()
}
