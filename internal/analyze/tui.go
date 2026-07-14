package analyze

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

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

type prefetchDoneMsg struct{ path string }

type scanTickMsg struct{}

func scanTick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg { return scanTickMsg{} })
}

type sortMode int

const (
	sortSize sortMode = iota
	sortName
	sortMtime
	sortModeCount
)

func (m sortMode) label() string {
	switch m {
	case sortName:
		return "名稱"
	case sortMtime:
		return "修改時間"
	default:
		return "大小"
	}
}

func sortEntries(entries []Entry, mode sortMode) {
	switch mode {
	case sortName:
		sort.SliceStable(entries, func(a, b int) bool {
			return strings.ToLower(entries[a].Name) < strings.ToLower(entries[b].Name)
		})
	case sortMtime:
		sort.SliceStable(entries, func(a, b int) bool {
			return entries[a].ModTime.After(entries[b].ModTime)
		})
	default:
		sort.SliceStable(entries, func(a, b int) bool {
			return entries[a].Size > entries[b].Size
		})
	}
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
	sort    sortMode

	confirm     *Entry // 非 nil 時處於刪除確認狀態
	deleting    bool
	status      string
	prefetching map[string]bool // 背景預取中的目錄(map 為參考型別,跨複本共用)
	scanProg    *clean.Progress
}

// Browse 啟動磁碟分析瀏覽器,從 root 開始向下鑽。
func Browse(root string) error {
	live := &clean.Progress{}
	sizer := NewSizer()
	sizer.SetProgress(live)
	b := browser{
		sizer: sizer, root: root, cwd: root,
		loading: true, height: 24, prefetching: map[string]bool{},
		scanProg: live,
	}
	_, err := tea.NewProgram(b, tea.WithAltScreen()).Run()
	return err
}

// prefetch 在游標停到目錄上時背景先算它的下一層,enter 時就有快取可用。
func (b browser) prefetch() tea.Cmd {
	if b.loading || b.cursor >= len(b.entries) {
		return nil
	}
	e := b.entries[b.cursor]
	if !e.IsDir || b.prefetching[e.Path] {
		return nil
	}
	b.prefetching[e.Path] = true
	sizer := b.sizer
	return func() tea.Msg {
		sizer.List(e.Path)
		return prefetchDoneMsg{path: e.Path}
	}
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

func (b browser) Init() tea.Cmd { return tea.Batch(b.load(b.cwd), scanTick()) }

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
		sortEntries(b.entries, b.sort)
		if b.cursor >= len(b.entries) {
			b.cursor = 0
		}
		b.loading = false
		b.errMsg = ""
		return b, b.prefetch()
	case prefetchDoneMsg:
		delete(b.prefetching, msg.path)
		return b, nil
	case scanTickMsg:
		if b.loading {
			return b, scanTick() // 載入中每 0.1 秒重繪一次進度
		}
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
		return b, tea.Batch(b.load(b.cwd), scanTick())
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
		return b, b.prefetch()
	case "down", "j":
		if b.cursor < len(b.entries)-1 {
			b.cursor++
		}
		return b, b.prefetch()
	case "s":
		if len(b.entries) > 0 {
			current := b.entries[b.cursor].Path
			b.sort = (b.sort + 1) % sortModeCount
			sortEntries(b.entries, b.sort)
			for i, e := range b.entries {
				if e.Path == current { // 排序後游標跟著原項目走
					b.cursor = i
					break
				}
			}
		}
	case "enter", "right", "l":
		if !b.loading && b.cursor < len(b.entries) && b.entries[b.cursor].IsDir {
			target := b.entries[b.cursor].Path
			b.loading = true
			return b, tea.Batch(b.load(target), scanTick())
		}
	case "backspace", "left", "h":
		if !b.loading && b.cwd != b.root {
			b.loading = true
			return b, tea.Batch(b.load(filepath.Dir(b.cwd)), scanTick())
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
	sb.WriteString(headerStyle.Render("🦡 Mogura 磁碟分析") + "  " + pathStyle.Render(b.cwd) +
		pathStyle.Render("  排序:"+b.sort.label()) + "\n\n")

	if b.loading {
		sb.WriteString(fmt.Sprintf("掃描中...  已掃描 %s · %s 檔\n",
			clean.Humanize(b.scanProg.Bytes()), clean.GroupDigits(b.scanProg.Files())))
		return sb.String()
	}
	if b.errMsg != "" {
		sb.WriteString("讀取失敗: " + b.errMsg + "\n")
		sb.WriteString(helpStyle.Render("backspace 返回上層 · q 離開"))
		return sb.String()
	}

	// 分母取全清單最大值:排序模式下第一名不一定是最大的
	var max int64 = 1
	for _, e := range b.entries {
		if e.Size > max {
			max = e.Size
		}
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
		filled := fillCells(e.Size, max)
		bar := barStyle.Render(strings.Repeat("█", filled)) +
			barBgStyle.Render(strings.Repeat("░", barWidth-filled))
		name := e.Name
		if e.IsDir {
			name = dirStyle.Render(name + "/")
		}
		extra := ""
		if b.sort == sortMtime {
			extra = pathStyle.Render(" · " + ageLabel(e.ModTime))
		} else if e.IsDir {
			extra = pathStyle.Render(" · " + clean.GroupDigits(e.Files) + " 檔")
		}
		sb.WriteString(fmt.Sprintf("%s%s %s %s%s\n", cursor, sizeStyle.Render(clean.Humanize(e.Size)), bar, name, extra))
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

	sb.WriteString(helpStyle.Render("\nenter 進入 · backspace 上層 · s 排序 · d 刪除 · q 離開"))
	return sb.String()
}

// fillCells 計算長條圖填滿格數,夾限在 [0, barWidth] 防止負數 Repeat。
func fillCells(size, max int64) int {
	if max <= 0 || size <= 0 {
		return 0
	}
	filled := int(int64(barWidth) * size / max)
	if filled > barWidth {
		filled = barWidth
	}
	return filled
}

// ageLabel 把 mtime 轉成相對時間描述。
func ageLabel(t time.Time) string {
	if t.IsZero() {
		return "未知"
	}
	d := time.Since(t)
	switch {
	case d < time.Hour:
		return "剛剛"
	case d < 24*time.Hour:
		return fmt.Sprintf("%d 小時前", int(d.Hours()))
	default:
		return fmt.Sprintf("%d 天前", int(d.Hours()/24))
	}
}
