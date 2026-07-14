package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"mogura/internal/config"
	"mogura/internal/i18n"
)

var langValues = []string{"auto", "zh", "en"}

// Settings 是可嵌入各 TUI 的設定面板;變更即時生效並立刻存檔。
type Settings struct {
	cfg     config.Config
	initial string
	saveErr error
}

func NewSettings() Settings {
	cfg := config.Load()
	return Settings{cfg: cfg, initial: cfg.Language}
}

// Changed 回報開啟面板以來語言是否有變,供宿主決定要不要重建畫面。
func (s Settings) Changed() bool { return s.cfg.Language != s.initial }

// HandleKey 處理按鍵,回傳更新後的面板與「是否關閉」。
func (s Settings) HandleKey(key tea.KeyMsg) (Settings, bool) {
	switch key.String() {
	case "enter", "q", "esc", ",", "ctrl+c":
		return s, true // enter 確定返回;變更早已即時存檔
	case "right", "l", " ":
		s.cycleLang(1)
	case "left", "h":
		s.cycleLang(-1)
	}
	return s, false
}

func (s *Settings) cycleLang(delta int) {
	idx := 0
	for i, v := range langValues {
		if v == s.cfg.Language {
			idx = i
			break
		}
	}
	idx = (idx + delta + len(langValues)) % len(langValues)
	s.cfg.Language = langValues[idx]
	i18n.Apply(s.cfg.Language) // 即時生效,整個 TUI 下一幀就換語言
	s.saveErr = config.Save(s.cfg)
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

func (s Settings) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(i18n.T("🦡 Mogura — 設定")) + "\n\n")
	b.WriteString(cursorStyle.Render("❯ ") + i18n.T("語言") + "  " + totalStyle.Render(s.langLabel()) + "\n")
	if s.saveErr != nil {
		b.WriteString("\n" + i18n.T("設定儲存失敗:") + s.saveErr.Error() + "\n")
	}
	if p, err := config.Path(); err == nil {
		b.WriteString("\n" + descStyle.Render(i18n.Tf("設定檔:%s", p)))
	}
	b.WriteString(helpStyle.Render(i18n.T("\n←→ 切換 · enter 確定")))
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
