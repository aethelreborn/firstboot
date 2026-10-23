package tui

import (
	"fmt"
	"io"

	"github.com/aethelreborn/firstboot/internal/catalog"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type item catalog.Package

func (i item) FilterValue() string { return i.Name }

type model struct {
	platform  catalog.Platform
	list      list.Model
	items     []catalog.Package
	confirm   bool
	cancelled bool
}

type delegate struct{}

func (delegate) Height() int                         { return 1 }
func (delegate) Spacing() int                        { return 0 }
func (delegate) Update(tea.Msg, *list.Model) tea.Cmd { return nil }
func (delegate) Render(w io.Writer, _ list.Model, _ int, it list.Item) {
	i, ok := it.(item)
	if !ok {
		return
	}
	mark := "[ ]"
	if i.Installed {
		mark = "[x]"
	} else if i.Selected {
		mark = "[✓]"
	}
	if !i.Supported {
		mark = "[-]"
	}
	_, _ = fmt.Fprintf(w, "%s %s — %s", mark, i.Name, i.Description)
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

func Run(platform catalog.Platform, packages []catalog.Package) ([]catalog.Package, error) {
	items := make([]list.Item, len(packages))
	for i, p := range packages {
		items[i] = item(p)
	}
	l := list.New(items, delegate{}, 76, 18)
	l.Title = "Firstboot · choose missing tools"
	m := model{platform: platform, list: l, items: packages}
	final, err := tea.NewProgram(m).Run()
	if err != nil {
		return nil, err
	}
	out := final.(model)
	if out.cancelled || !out.confirm {
		return nil, nil
	}
	selected := make([]catalog.Package, 0)
	for i, raw := range out.list.Items() {
		v := raw.(item)
		if v.Selected {
			selected = append(selected, out.items[i])
		}
	}
	return selected, nil
}

func (m model) Init() tea.Cmd { return nil }

func (m model) View() string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	hint := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	status := "space toggle · a all · n none · enter review · q quit"
	if m.confirm {
		status = "enter confirm install · q cancel"
	}
	return style.Render(banner) + "\n" + hint.Render(fmt.Sprintf("Detected platform: %s · %s", m.platform, status)) + "\n\n" + m.list.View()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "q", "ctrl+c":
			m.cancelled = true
			return m, tea.Quit
		case "space":
			if !m.confirm {
				m.toggle(m.list.Index())
			}
		case "a":
			if !m.confirm {
				for i := range m.list.Items() {
					m.toggle(i)
				}
			}
		case "n":
			if !m.confirm {
				for i, raw := range m.list.Items() {
					v := raw.(item)
					v.Selected = false
					m.list.SetItem(i, v)
				}
			}
		case "enter":
			if m.confirm {
				return m, tea.Quit
			}
			m.confirm = true
			m.list.Title = "Review selection — press enter to install"
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *model) toggle(index int) {
	if index < 0 || index >= len(m.list.Items()) {
		return
	}
	v := m.list.Items()[index].(item)
	if v.Supported && !v.Installed {
		v.Selected = !v.Selected
		m.list.SetItem(index, v)
	}
}
