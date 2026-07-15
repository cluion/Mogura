package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"mogura/internal/config"
	"mogura/internal/i18n"
)

var (
	langValues       = []string{"auto", "zh", "en"}
	deleteValues     = []string{"direct", "trash"}
	journalDayValues = []int{3, 7, 14, 30}
)

const settingsRows = 3

// Settings 是可嵌入各 TUI 的設定面板;變更即時生效並立刻存檔。
type Settings struct {
	cfg     config.Config
	initial config.Config
	row     int
	saveErr error
}

func NewSettings() Settings {
	cfg := config.Load()
	return Settings{cfg: cfg, initial: cfg}
}

// Changed 回報開啟面板以來,會影響已顯示規則文字的設定是否有變,
// 供宿主決定要不要重建畫面(語言、journal 天數都會反映在規則名稱與說明)。
func (s Settings) Changed() bool {
	return s.cfg.Language != s.initial.Language || s.cfg.JournalDays != s.initial.JournalDays
}

// HandleKey 處理按鍵,回傳更新後的面板與「是否關閉」。
func (s Settings) HandleKey(key tea.KeyMsg) (Settings, bool) {
	switch key.String() {
	case "enter", "q", "esc", ",", "ctrl+c":
		return s, true // enter 確定返回;變更早已即時存檔
	case "up", "k":
		s.row = (s.row + settingsRows - 1) % settingsRows
	case "down", "j":
		s.row = (s.row + 1) % settingsRows
	case "right", "l", " ":
		s.cycle(1)
	case "left", "h":
		s.cycle(-1)
	}
	return s, false
}

func (s *Settings) cycle(delta int) {
	switch s.row {
	case 0:
		s.cfg.Language = cycleValue(langValues, s.cfg.Language, delta)
		i18n.Apply(s.cfg.Language) // 即時生效,整個 TUI 下一幀就換語言
	case 1:
		s.cfg.Delete = cycleValue(deleteValues, s.cfg.Delete, delta)
	case 2:
		idx := 1 // 手動改過的自訂天數不在清單裡,從預設 7 起跳
		for i, v := range journalDayValues {
			if v == s.cfg.JournalDays {
				idx = i
				break
			}
		}
		s.cfg.JournalDays = journalDayValues[(idx+delta+len(journalDayValues))%len(journalDayValues)]
	}
	s.saveErr = config.Save(s.cfg)
}

func cycleValue(values []string, current string, delta int) string {
	idx := 0
	for i, v := range values {
		if v == current {
			idx = i
			break
		}
	}
	return values[(idx+delta+len(values))%len(values)]
}

func (s Settings) langLabel() string {
	switch s.cfg.Language {
	case "zh":
		return "繁體中文"
	case "en":
		return "English"
	default:
		return i18n.T("自動(跟隨系統)")
	}
}

func (s Settings) deleteLabel() string {
	if s.cfg.UseTrash() {
		return i18n.T("移至垃圾桶")
	}
	return i18n.T("直接刪除")
}

func (s Settings) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(i18n.T("🦡 Mogura — 設定")) + "\n\n")
	rows := []struct{ label, value string }{
		{i18n.T("語言"), s.langLabel()},
		{i18n.T("刪除方式"), s.deleteLabel()},
		{i18n.T("journal 保留"), i18n.Tf("%d 天", s.cfg.JournalDays)},
	}
	for i, r := range rows {
		prefix := "  "
		if i == s.row {
			prefix = cursorStyle.Render("❯ ")
		}
		b.WriteString(prefix + r.label + "  " + totalStyle.Render(r.value) + "\n")
	}
	if s.saveErr != nil {
		b.WriteString("\n" + i18n.T("設定儲存失敗:") + s.saveErr.Error() + "\n")
	}
	if p, err := config.Path(); err == nil {
		b.WriteString("\n" + descStyle.Render(i18n.Tf("設定檔:%s", p)))
		b.WriteString("\n" + descStyle.Render(i18n.T("排除清單(exclude)等進階設定請直接編輯設定檔")))
	}
	b.WriteString(helpStyle.Render(i18n.T("\n↑↓ 選擇 · ←→ 切換 · enter 確定")))
	return b.String()
}

type settingsProgram struct{ s Settings }

func (p settingsProgram) Init() tea.Cmd { return nil }

func (p settingsProgram) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		var closed bool
		p.s, closed = p.s.HandleKey(key)
		if closed {
			return p, tea.Quit
		}
	}
	return p, nil
}

func (p settingsProgram) View() string { return p.s.View() }

// RunSettings 以獨立程式執行設定面板(mogura config)。
func RunSettings() error {
	_, err := tea.NewProgram(settingsProgram{s: NewSettings()}).Run()
	return err
}
