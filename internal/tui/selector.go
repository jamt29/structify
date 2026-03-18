package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"

	"github.com/jamt29/structify/internal/template"
)

type templateItem struct {
	t *template.Template
}

func (i templateItem) Title() string {
	if i.t == nil || i.t.Manifest == nil {
		return "(invalid template)"
	}
	return strings.TrimSpace(i.t.Manifest.Name)
}

func (i templateItem) Description() string {
	if i.t == nil || i.t.Manifest == nil {
		return ""
	}
	m := i.t.Manifest
	arch := strings.TrimSpace(m.Architecture)
	lang := strings.TrimSpace(m.Language)
	desc := strings.TrimSpace(m.Description)

	meta := strings.TrimSpace(strings.Join([]string{arch, lang}, " · "))
	if meta == "" {
		return desc
	}
	if desc == "" {
		return meta
	}
	return meta + "\n" + desc
}

func (i templateItem) FilterValue() string {
	if i.t == nil || i.t.Manifest == nil {
		return ""
	}
	m := i.t.Manifest
	parts := []string{
		m.Name,
		m.Architecture,
		m.Language,
		m.Description,
		strings.Join(m.Tags, " "),
	}
	return strings.ToLower(strings.TrimSpace(strings.Join(parts, " ")))
}

type selectorModel struct {
	list     list.Model
	selected *template.Template
}

func (m selectorModel) Init() tea.Cmd { return nil }

func (m selectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if it, ok := m.list.SelectedItem().(templateItem); ok {
				m.selected = it.t
				return m, tea.Quit
			}
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m selectorModel) View() string {
	if m.selected != nil {
		return ""
	}
	return m.list.View()
}

// RunSelector shows a searchable list of templates and returns the selected one.
//
// Edge cases:
// - if templates is empty => returns an actionable error
// - if only one template => returns it without launching TUI
func RunSelector(templates []*template.Template) (*template.Template, error) {
	if len(templates) == 0 {
		return nil, fmt.Errorf("No templates found. Run 'structify template add <github-url>' to add one.")
	}
	if len(templates) == 1 {
		return templates[0], nil
	}

	items := make([]list.Item, 0, len(templates))
	for _, t := range templates {
		if t == nil || t.Manifest == nil || strings.TrimSpace(t.Manifest.Name) == "" {
			continue
		}
		items = append(items, templateItem{t: t})
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("No templates found. Run 'structify template add <github-url>' to add one.")
	}
	if len(items) == 1 {
		return items[0].(templateItem).t, nil
	}

	const defaultWidth = 80
	l := list.New(items, list.NewDefaultDelegate(), defaultWidth, 18)
	l.Title = "Select a template"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(true)
	l.Styles.Title = lipgloss.NewStyle().Bold(true)

	p := tea.NewProgram(selectorModel{list: l}, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("running template selector: %w", err)
	}
	sm, ok := finalModel.(selectorModel)
	if !ok {
		return nil, fmt.Errorf("internal error: unexpected selector model")
	}
	if sm.selected == nil {
		return nil, fmt.Errorf("template selection cancelled")
	}
	return sm.selected, nil
}

