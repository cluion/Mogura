package monitor

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"mogura/internal/clean"
	"mogura/internal/i18n"
)

const refreshInterval = 2 * time.Second

var (
	headerStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	sectionStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	faintStyle   = lipgloss.NewStyle().Faint(true)
	okStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	dangerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	trackStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
)

type snapMsg Snapshot
type tickMsg struct{}

type dashboard struct {
	snap  Snapshot
	ready bool
}

// Run 啟動即時監控儀表板,每 2 秒更新
func Run() error {
	_, err := tea.NewProgram(dashboard{}, tea.WithAltScreen()).Run()
	return err
}

func take(prev *Snapshot) tea.Cmd {
	return func() tea.Msg { return snapMsg(Take(prev)) }
}

func (d dashboard) Init() tea.Cmd { return take(nil) }

func (d dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case snapMsg:
		d.snap = Snapshot(msg)
		d.ready = true
		return d, tea.Tick(refreshInterval, func(time.Time) tea.Msg { return tickMsg{} })
	case tickMsg:
		prev := d.snap
		return d, take(&prev)
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return d, tea.Quit
		}
	}
	return d, nil
}

func bar(pct float64, width int) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	filled := int(pct/100*float64(width) + 0.5)
	style := okStyle
	switch {
	case pct >= 85:
		style = dangerStyle
	case pct >= 60:
		style = warnStyle
	}
	return style.Render(strings.Repeat("█", filled)) +
		trackStyle.Render(strings.Repeat("░", width-filled))
}

func (d dashboard) View() string {
	if !d.ready {
		return i18n.T("🦡 取樣中...\n")
	}
	s := d.snap
	var b strings.Builder

	up := s.Uptime.Round(time.Minute)
	days := int(up.Hours()) / 24
	b.WriteString(headerStyle.Render(i18n.T("🦡 Mogura 系統監控")) + "  " +
		faintStyle.Render(i18n.Tf("%s · 開機 %d 天 %s · 負載 %.2f %.2f %.2f",
			s.Hostname, days, up-time.Duration(days)*24*time.Hour, s.Load1, s.Load5, s.Load15)) + "\n\n")

	b.WriteString(sectionStyle.Render("CPU") + fmt.Sprintf("  %5.1f%%  ", s.CPUTotal) + bar(s.CPUTotal, 30) + "\n")
	for i := 0; i < len(s.CPUCores); i += 4 {
		b.WriteString("  ")
		for j := i; j < i+4 && j < len(s.CPUCores); j++ {
			b.WriteString(fmt.Sprintf("%2d %s %5.1f%%   ", j, bar(s.CPUCores[j], 10), s.CPUCores[j]))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n" + sectionStyle.Render(i18n.T("記憶體")) + fmt.Sprintf("  %5.1f%%  ", s.MemPercent) + bar(s.MemPercent, 30) +
		faintStyle.Render(i18n.Tf("  %s / %s · 可用 %s",
			clean.Humanize(int64(s.MemUsed)), clean.Humanize(int64(s.MemTotal)), clean.Humanize(int64(s.MemAvailable)))) + "\n")
	if s.SwapTotal > 0 {
		swapPct := float64(s.SwapUsed) / float64(s.SwapTotal) * 100
		b.WriteString("  swap   " + fmt.Sprintf("%5.1f%%  ", swapPct) + bar(swapPct, 30) +
			faintStyle.Render(fmt.Sprintf("  %s / %s",
				clean.Humanize(int64(s.SwapUsed)), clean.Humanize(int64(s.SwapTotal)))) + "\n")
	}

	if len(s.Disks) > 0 {
		b.WriteString("\n" + sectionStyle.Render(i18n.T("磁碟")) + "\n")
		for _, d := range s.Disks {
			b.WriteString(fmt.Sprintf("  %-16s %5.1f%%  %s  %s\n",
				d.Mount, d.Percent, bar(d.Percent, 30),
				faintStyle.Render(fmt.Sprintf("%s / %s",
					clean.Humanize(int64(d.Used)), clean.Humanize(int64(d.Total))))))
		}
	}

	b.WriteString("\n" + sectionStyle.Render(i18n.T("網路")) +
		fmt.Sprintf("  ↓ %s/s  ↑ %s/s\n",
			clean.Humanize(int64(s.RxRate)), clean.Humanize(int64(s.TxRate))))

	b.WriteString(faintStyle.Render(i18n.T("\n每 2 秒更新 · q 離開")))
	return b.String()
}
