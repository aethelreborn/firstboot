package tui

import (
	"testing"

	"github.com/aethelreborn/firstboot/internal/catalog"
	tea "github.com/charmbracelet/bubbletea"
)

func testModel() model {
	return model{
		platform: catalog.Linux,
		packages: []catalog.Package{
			{Name: "Git", Supported: true},
			{Name: "Python", Supported: true},
			{Name: "Docker", Supported: true, Installed: true},
		},
	}
}

func sendKey(m model, key tea.KeyType) model {
	updated, _ := m.Update(tea.KeyMsg{Type: key})
	return updated.(model)
}

func TestSpaceTogglesFocusedPackage(t *testing.T) {
	m := sendKey(testModel(), tea.KeySpace)
	if !m.packages[0].Selected {
		t.Fatal("space should select the focused package")
	}
	m = sendKey(m, tea.KeySpace)
	if m.packages[0].Selected {
		t.Fatal("space should deselect the focused package")
	}
}

func TestArrowsMoveFocus(t *testing.T) {
	m := sendKey(testModel(), tea.KeyDown)
	if m.cursor != 1 {
		t.Fatalf("down moved cursor to %d, want 1", m.cursor)
	}
	m = sendKey(m, tea.KeyUp)
	if m.cursor != 0 {
		t.Fatalf("up moved cursor to %d, want 0", m.cursor)
	}
}

func TestSelectAllAndClear(t *testing.T) {
	m := sendKey(testModel(), tea.KeyRunes)
	// KeyRunes is not used for command shortcuts; exercise the public update path below.
	_ = m
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = updated.(model)
	if !m.packages[0].Selected || !m.packages[1].Selected || m.packages[2].Selected {
		t.Fatal("a should select supported missing packages only")
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = updated.(model)
	if m.packages[0].Selected || m.packages[1].Selected {
		t.Fatal("n should clear the selection")
	}
}

func TestEnterOpensAndConfirmsReview(t *testing.T) {
	m := testModel()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	if !m.confirm {
		t.Fatal("enter should open the review state")
	}
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter in review should return a quit command")
	}
	_ = updated
}
