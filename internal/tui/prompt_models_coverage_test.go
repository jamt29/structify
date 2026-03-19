package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
)

type dummyEnumItem struct{ value string }

func (i dummyEnumItem) Title() string       { return i.value }
func (i dummyEnumItem) Description() string { return "" }
func (i dummyEnumItem) FilterValue() string { return "" }

func TestStringPromptModel_UpdateAndView(t *testing.T) {
	ti := textinput.New()
	ti.Placeholder = ""

	m := stringPromptModel{
		label:    "Label",
		input:    ti,
		required: true,
	}

	// Press enter with empty value and required=true => show required error, no done.
	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		// stringPromptModel returns nil cmd in the "required" error path
		t.Fatalf("expected nil cmd, got %v", cmd)
	}
	pm := m2.(stringPromptModel)
	if pm.errMsg != "required" {
		t.Fatalf("expected errMsg=required, got %q", pm.errMsg)
	}
	if pm.done {
		t.Fatalf("expected done=false")
	}

	view := pm.View()
	if !strings.Contains(view, "required") {
		t.Fatalf("expected view to contain error msg, got %q", view)
	}

	// Press enter again with placeholder set => confirm and done.
	pm.input.Placeholder = "fallback"
	m3, cmd := pm.Update(tea.KeyMsg{Type: tea.KeyEnter})
	pm = m3.(stringPromptModel)
	if cmd == nil {
		t.Fatalf("expected non-nil cmd on confirm")
	}
	if !pm.done {
		t.Fatalf("expected done=true")
	}
	if pm.View() != "" {
		t.Fatalf("expected empty view when done")
	}
}

func TestEnumItemAndEnumPromptModel(t *testing.T) {
	ei := enumItem{value: "gorm"}
	if ei.Title() != "gorm" {
		t.Fatalf("unexpected enumItem Title: %q", ei.Title())
	}
	if ei.Description() != "" || ei.FilterValue() != "" {
		t.Fatalf("unexpected enumItem Description/FilterValue")
	}

	// Case 1: SelectedItem is enumItem and required=true => should set selected + done.
	items := []list.Item{enumItem{value: "grpc"}}
	l := list.New(items, list.NewDefaultDelegate(), 80, 10)
	m := enumPromptModel{list: l, required: true}
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	pm := m2.(enumPromptModel)
	if pm.selected != "grpc" || !pm.done {
		t.Fatalf("expected selected=grpc done=true, got selected=%q done=%v", pm.selected, pm.done)
	}
	if pm.View() != "" {
		t.Fatalf("expected empty View when done")
	}

	// Case 2: SelectedItem is not enumItem => required=true should not complete.
	items2 := []list.Item{dummyEnumItem{value: "x"}}
	l2 := list.New(items2, list.NewDefaultDelegate(), 80, 10)
	m = enumPromptModel{list: l2, required: true}
	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	pm = m2.(enumPromptModel)
	if cmd != nil {
		// required=true path returns m,nil (no Quit)
		t.Fatalf("expected nil cmd, got %v", cmd)
	}
	if pm.done {
		t.Fatalf("expected done=false when selection type mismatch and required=true")
	}
}

