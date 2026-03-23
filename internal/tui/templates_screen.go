package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jamt29/structify/internal/dsl"
	"github.com/jamt29/structify/internal/template"
)

type templatesItem struct {
	t     *template.Template
	group string // "Local" | "Built-in"
}

type TemplatesModel struct {
	local   []*template.Template
	builtins []*template.Template

	items []templatesItem

	cursor int
	width  int
	height int

	detail *dsl.Manifest

	transitionToNew *template.Template
	done             bool
}

func NewTemplatesModel(all []*template.Template) *TemplatesModel {
	m := &TemplatesModel{
		width:  100,
		height: 30,
		cursor: 0,
	}
	for _, t := range all {
		if t == nil || t.Manifest == nil {
			continue
		}
		if t.Source == "builtin" {
			m.builtins = append(m.builtins, t)
		} else {
			m.local = append(m.local, t)
		}
	}

	for _, t := range m.local {
		m.items = append(m.items, templatesItem{t: t, group: "Local"})
	}
	for _, t := range m.builtins {
		m.items = append(m.items, templatesItem{t: t, group: "Built-in"})
	}
	if len(m.items) == 0 {
		m.detail = nil
	}
	return m
}

func (m *TemplatesModel) Init() tea.Cmd { return nil }

func (m *TemplatesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.width, m.height = ws.Width, ws.Height
		return m, nil
	}

	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.done = true
			return m, nil
		case "up", "k":
			if len(m.items) == 0 {
				return m, nil
			}
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.items) - 1
			}
			return m, nil
		case "down", "j":
			if len(m.items) == 0 {
				return m, nil
			}
			m.cursor++
			if m.cursor >= len(m.items) {
				m.cursor = 0
			}
			return m, nil
		case "enter":
			if len(m.items) == 0 {
				return m, nil
			}
			if m.items[m.cursor].t != nil {
				m.detail = m.items[m.cursor].t.Manifest
			}
			return m, nil
		case "n":
			if len(m.items) == 0 {
				return m, nil
			}
			m.transitionToNew = m.items[m.cursor].t
			return m, nil
		}
	}

	return m, nil
}

func (m *TemplatesModel) View() string {
	return m.ViewContent()
}

// ViewContent returns the raw (non-centered) content of this screen.
// RootModel is responsible for applying any alignment/centering.
func (m *TemplatesModel) ViewContent() string {
	if m.width < 80 || m.height < 24 {
		return stylePending.Render("Terminal too small. Minimum 80x24.")
	}

	var b strings.Builder
	header := styleHeader.Render("structify") + "  " + styleCompletedLabel.Render("·  Mis templates")
	b.WriteString(header)
	b.WriteString("\n\n")

	if len(m.items) == 0 {
		b.WriteString(stylePending.Render("No templates found."))
		b.WriteString("\n")
		return b.String()
	}

	// Render groups with headings.
	localCount := len(m.local)

	renderRow := func(active bool, it templatesItem) string {
		name := strings.TrimSpace(it.t.Manifest.Name)
		lang := strings.TrimSpace(it.t.Manifest.Language)
		arch := strings.TrimSpace(it.t.Manifest.Architecture)
		meta := strings.TrimSpace(strings.TrimSpace(lang) + " · " + strings.TrimSpace(arch))
		row := fmt.Sprintf("%-34s  %s", name, meta)
		if active {
			return styleMenuItemActive.Render(row)
		}
		return styleMenuItem.Render(row)
	}

	for i, it := range m.items {
		if it.group == "Local" && i == 0 {
			b.WriteString(styleCompletedLabel.Render("Local") + "\n")
			b.WriteString(stylePending.Render(strings.Repeat("─", 55)) + "\n")
		}
		if it.group == "Built-in" && i == localCount {
			b.WriteString("\n" + styleCompletedLabel.Render("Built-in") + "\n")
			b.WriteString(stylePending.Render(strings.Repeat("─", 55)) + "\n")
		}
		active := i == m.cursor
		b.WriteString(renderRow(active, it))
		b.WriteString("\n")
	}

	if m.detail != nil {
		b.WriteString("\n")
		d := m.detail
		desc := strings.TrimSpace(d.Description)
		if desc == "" {
			desc = "-"
		}
		b.WriteString(styleActiveBox.Render(
			strings.Join([]string{
				styleCompletedLabel.Render("Detalle") + ": " + styleCompletedValue.Render(strings.TrimSpace(d.Name)),
				"Language/Arch: " + styleCompletedValue.Render(strings.TrimSpace(d.Language) + " · " + strings.TrimSpace(d.Architecture)),
				"Description: " + styleCompletedValue.Render(desc),
				fmt.Sprintf("Inputs: %d  Steps: %d", len(d.Inputs), len(d.Steps)),
			}, "\n"),
		))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(styleHelpBar.Render("enter ver detalle  n nuevo template  esc volver"))
	b.WriteString("\n")
	return b.String()
}

