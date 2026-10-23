package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/aethelreborn/firstboot/internal/catalog"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tickMsg time.Time

type model struct {
	platform        catalog.Platform
	packages        []catalog.Package
	defaultPackages []catalog.Package
	cursor          int
	confirm         bool
	searching       bool
	searchInput     string
	notice          string
	blink           bool
	cancelled       bool
}

const banner = `
······················································
:     _                 _       _                 __ :
:    /.\       ____    FJ_     FJ___      ____    LJ :
:   //_\\     F __ J  J  _|   J  __ '.   F __ J   FJ :
:  / ___ \   | _____J | |-'   | |--| |  | _____J J  L:
: / L___J \  F L___--.F |__-. F L  J J  F L___--.J  L:
:J__L   J__LJ\_*____/F\_____/J__L  J__LJ\_____/FJ__L:
:|__L   J__| J______F J_____F|__L  J__| J______F |__|:
:   ___      _        __                             :
: ,"___".   FJ        FJ                             :
: FJ---L]  J |       J  L                            :
:J |   LJ  | |       |  |                            :
:| \___--. F L_____  F  J                            :
:J\_____/FJ________LJ____L                           :
: J_____F |________||____|                           :
······················································
`

var (
	accent   = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	muted    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	cyan     = lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	green    = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	yellow   = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	red      = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	focused  = lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Background(lipgloss.Color("57")).Bold(true).PaddingLeft(1).PaddingRight(1)
	selected = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
)

func Run(platform catalog.Platform, packages []catalog.Package) ([]catalog.Package, error) {
	m := model{platform: platform, packages: packages, defaultPackages: packages}
	final, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		return nil, err
	}
	out := final.(model)
	if out.cancelled || !out.confirm {
		return nil, nil
	}
	result := make([]catalog.Package, 0)
	for _, p := range out.packages {
		if p.Selected {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m model) Init() tea.Cmd { return blinkCmd() }

func blinkCmd() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		m.blink = !m.blink
		return m, blinkCmd()
	case tea.KeyMsg:
		if m.searching {
			m.handleSearchKey(msg)
			return m, nil
		}
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.cancelled = true
			return m, tea.Quit
		case "up", "k":
			if !m.confirm && m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if !m.confirm && m.cursor < len(m.packages)-1 {
				m.cursor++
			}
		case "home", "g":
			if !m.confirm {
				m.cursor = 0
			}
		case "end", "G":
			if !m.confirm && len(m.packages) > 0 {
				m.cursor = len(m.packages) - 1
			}
		case " ", "space":
			if !m.confirm {
				m.toggle()
			}
		case "a":
			if !m.confirm {
				m.selectAll()
			}
		case "n":
			if !m.confirm {
				m.clearSelection()
			}
		case "/":
			if !m.confirm {
				m.searching = true
				m.searchInput = ""
				m.notice = ""
			}
		case "b":
			if !m.confirm {
				m.packages = append([]catalog.Package(nil), m.defaultPackages...)
				m.cursor = 0
				m.notice = "Back to recommended tools"
			}
		case "enter":
			if m.confirm {
				return m, tea.Quit
			}
			m.confirm = true
		}
	}
	return m, nil
}

func (m *model) handleSearchKey(msg tea.KeyMsg) {
	switch msg.Type {
	case tea.KeyEsc:
		m.searching = false
	case tea.KeyBackspace:
		if len(m.searchInput) > 0 {
			m.searchInput = m.searchInput[:len(m.searchInput)-1]
		}
	case tea.KeyEnter:
		query := strings.TrimSpace(m.searchInput)
		if query == "" {
			m.notice = "Type an app or package name first"
			return
		}
		results, err := catalog.Search(m.platform, query)
		m.searching = false
		m.cursor = 0
		if err != nil {
			m.packages = nil
			m.notice = err.Error()
			return
		}
		m.packages = results
		m.notice = fmt.Sprintf("%d results for %q · space queues · b returns to recommendations", len(results), query)
	case tea.KeyRunes:
		m.searchInput += string(msg.Runes)
	}
}

func (m *model) toggle() {
	if m.cursor < 0 || m.cursor >= len(m.packages) {
		return
	}
	p := &m.packages[m.cursor]
	if p.Supported && !p.Installed {
		p.Selected = !p.Selected
	}
}

func (m *model) selectAll() {
	for i := range m.packages {
		if m.packages[i].Supported && !m.packages[i].Installed {
			m.packages[i].Selected = true
		}
	}
}

func (m *model) clearSelection() {
	for i := range m.packages {
		m.packages[i].Selected = false
	}
}

func (m model) View() string {
	var b strings.Builder
	b.WriteString(accent.Render(banner))
	b.WriteString("\n")
	b.WriteString(muted.Render(fmt.Sprintf("  %s  %s", badge(string(m.platform)), m.subtitle())))
	b.WriteString("\n\n")
	if m.confirm {
		b.WriteString(accent.Render("  Review installation"))
		b.WriteString("\n")
		b.WriteString(muted.Render("  Press enter to install selected tools · esc to cancel"))
	} else {
		b.WriteString(accent.Render("  Select your workstation toolkit"))
		b.WriteString("\n")
		b.WriteString(muted.Render("  ↑/↓ or j/k move · space toggle · / search apps · enter review · q quit"))
	}
	b.WriteString("\n\n")
	if m.searching {
		b.WriteString(accent.Render("  Search native package sources"))
		b.WriteString("\n")
		b.WriteString(focused.Render("  / " + m.searchInput + "▌"))
		b.WriteString("\n")
		b.WriteString(muted.Render("  Try discord, vscode, android studio, vlc, or any package name"))
		b.WriteString("\n\n")
	}
	if m.notice != "" {
		b.WriteString(yellow.Render("  " + m.notice))
		b.WriteString("\n\n")
	}
	if len(m.packages) == 0 && !m.searching {
		b.WriteString(muted.Render("  No results. Press / to search again or b to return to recommendations."))
		b.WriteString("\n")
	}
	for i, p := range m.packages {
		b.WriteString(m.renderRow(i, p))
		b.WriteByte('\n')
	}
	b.WriteString("\n")
	b.WriteString(m.footer())
	return b.String()
}

func (m model) subtitle() string {
	missing, chosen := 0, 0
	for _, p := range m.packages {
		if !p.Installed && p.Supported {
			missing++
			if p.Selected {
				chosen++
			}
		}
	}
	return fmt.Sprintf("%d available · %d selected", missing, chosen)
}

func badge(platform string) string { return cyan.Render("● " + strings.ToUpper(platform)) }

func (m model) renderRow(i int, p catalog.Package) string {
	indicator := "  "
	if i == m.cursor && !m.confirm {
		if m.blink {
			indicator = "▸ "
		} else {
			indicator = "  "
		}
	}
	mark := muted.Render("○")
	status := muted.Render("missing")
	if p.Installed {
		mark, status = green.Render("✓"), green.Render("installed")
	} else if !p.Supported {
		mark, status = red.Render("×"), red.Render("unsupported")
	} else if p.Selected {
		mark, status = selected.Render("●"), yellow.Render("queued")
	}
	name := p.Name
	if i == m.cursor && !m.confirm {
		name = focused.Render(name)
	}
	return fmt.Sprintf("%s%s  %-17s %s  %s", indicator, mark, name, muted.Render(p.Description), status)
}

func (m model) footer() string {
	if m.confirm {
		return yellow.Render("  ⚡ Ready to install selected tools")
	}
	return muted.Render("  firstboot · / searches native package sources · no changes happen until confirmation")
}
