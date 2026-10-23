package tui

import (
	"fmt"
	"strings"

	"github.com/aethelreborn/firstboot/internal/catalog"
	"github.com/aethelreborn/firstboot/internal/install"
	tea "github.com/charmbracelet/bubbletea"
)

type installEvent struct {
	progress *install.Progress
	failures []install.Failure
	done     bool
}
type installStreamMsg struct{ events <-chan installEvent }

type installModel struct {
	platform         catalog.Platform
	packages         []catalog.Package
	password         string
	passwordRequired bool
	installing       bool
	done             bool
	current          int
	active           string
	lastError        string
	failures         []install.Failure
	failureCursor    int
	showError        bool
	quitConfirm      bool
	blink            bool
	events           <-chan installEvent
}

func RunInstall(platform catalog.Platform, packages []catalog.Package) error {
	m := installModel{platform: platform, packages: packages, passwordRequired: install.NeedsPassword(platform)}
	_, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	return err
}

func (m installModel) Init() tea.Cmd {
	if m.passwordRequired {
		return blinkCmd()
	}
	return tea.Batch(blinkCmd(), m.startInstall())
}

func (m installModel) startInstall() tea.Cmd {
	platform, packages, password := m.platform, m.packages, m.password
	return func() tea.Msg {
		events := make(chan installEvent)
		go func() {
			failures := install.RunWithProgress(platform, packages, password, func(progress install.Progress) {
				copy := progress
				events <- installEvent{progress: &copy}
			})
			events <- installEvent{failures: failures, done: true}
			close(events)
		}()
		return installStreamMsg{events: events}
	}
}

func nextInstallEvent(events <-chan installEvent) tea.Cmd {
	return func() tea.Msg {
		event, ok := <-events
		if !ok {
			return installEvent{done: true}
		}
		return event
	}
}

func (m installModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		m.blink = !m.blink
		return m, blinkCmd()
	case installStreamMsg:
		m.installing = true
		m.events = msg.events
		return m, nextInstallEvent(msg.events)
	case installEvent:
		if msg.progress != nil {
			m.current = msg.progress.Index + 1
			m.active = msg.progress.Package.Name
			if msg.progress.Error != nil {
				m.lastError = msg.progress.Error.Error()
			}
		}
		if msg.done {
			m.done = true
			m.installing = false
			m.failures = msg.failures
			return m, nil
		}
		return m, nextInstallEvent(m.events)
	case tea.KeyMsg:
		if m.passwordRequired && !m.installing && !m.done {
			switch msg.Type {
			case tea.KeyRunes:
				m.password += string(msg.Runes)
			case tea.KeyBackspace:
				if len(m.password) > 0 {
					m.password = m.password[:len(m.password)-1]
				}
			case tea.KeyEnter:
				m.passwordRequired = false
				m.installing = true
				return m, m.startInstall()
			case tea.KeyEsc:
				return m, tea.Quit
			}
			return m, nil
		}
		if m.installing {
			return m, nil
		}
		if m.quitConfirm {
			switch msg.String() {
			case "enter":
				return m, tea.Quit
			case "esc", "q", "n":
				m.quitConfirm = false
			}
			return m, nil
		}
		if m.showError {
			switch msg.String() {
			case "esc", "enter":
				m.showError = false
			}
			return m, nil
		}
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitConfirm = true
		case "up", "k":
			if m.failureCursor > 0 {
				m.failureCursor--
			}
		case "down", "j":
			if m.failureCursor < len(m.failures)-1 {
				m.failureCursor++
			}
		case "enter":
			if len(m.failures) > 0 {
				m.showError = true
			} else {
				return m, tea.Quit
			}
		case "esc":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m installModel) View() string {
	var b strings.Builder
	b.WriteString(accent.Render(banner))
	b.WriteString("\n\n")
	if m.passwordRequired && !m.installing {
		b.WriteString(accent.Render("  Administrator access required"))
		b.WriteString("\n")
		b.WriteString(muted.Render("  Enter your sudo password. It is used once and never stored."))
		b.WriteString("\n\n")
		b.WriteString(focused.Render("  sudo password: " + strings.Repeat("•", len(m.password)) + "▌"))
		b.WriteString("\n\n")
		b.WriteString(muted.Render("  enter continue · esc cancel"))
		return b.String()
	}
	b.WriteString(accent.Render("  Installing your workstation toolkit"))
	b.WriteString("\n")
	if m.installing {
		b.WriteString(muted.Render(fmt.Sprintf("  %d of %d complete · %s", m.current, len(m.packages), m.active)))
		b.WriteString("\n\n")
		b.WriteString(progressBar(m.current, len(m.packages)))
		b.WriteString("\n\n")
		b.WriteString(yellow.Render("  ⠋ Package manager is working — keep this terminal open"))
		return b.String()
	}
	if m.showError {
		return m.errorView()
	}
	if len(m.failures) == 0 {
		b.WriteString(green.Render("  ✓ Installation complete"))
		b.WriteString("\n\n")
		b.WriteString(muted.Render("  Press enter or esc to return to your shell"))
		return b.String()
	}
	b.WriteString(red.Render(fmt.Sprintf("  %d installation(s) failed", len(m.failures))))
	b.WriteString("\n\n")
	for i, failure := range m.failures {
		prefix := "  "
		if i == m.failureCursor {
			prefix = "▸ "
		}
		b.WriteString(prefix + red.Render("✗ "+failure.Package.Name) + "  " + yellow.Render("[ Show Error ]"))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(muted.Render("  ↑/↓ select failure · enter Show Error · q quit"))
	return b.String()
}

func (m installModel) errorView() string {
	failure := m.failures[m.failureCursor]
	return accent.Render("  Installation error") + "\n\n" + red.Render("  "+failure.Package.Name) + "\n\n" + muted.Render("  "+failure.Error.Error()) + "\n\n" + yellow.Render("  Press enter or esc to go back")
}

func progressBar(current, total int) string {
	width := 42
	filled := 0
	if total > 0 {
		filled = current * width / total
	}
	return cyan.Render("  [" + strings.Repeat("█", filled) + strings.Repeat("░", width-filled) + fmt.Sprintf("] %d%%", current*100/max(1, total)))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
