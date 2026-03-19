package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"

	"github.com/jamt29/structify/internal/dsl"
)

// RunInputs runs an interactive prompt for the provided manifest inputs.
func RunInputs(inputs []dsl.Input) (dsl.Context, error) {
	return RunInputsWithInitial(inputs, nil)
}

// RunInputsWithInitial is like RunInputs but seeds the context with pre-filled values.
// This is used by the `structify new` command to support a mixed mode (flags + TUI).
func RunInputsWithInitial(inputs []dsl.Input, initial dsl.Context) (dsl.Context, error) {
	answers := map[string]string{}
	for k, v := range initial {
		answers[k] = fmt.Sprint(v)
	}

	for idx, in := range inputs {
		id := strings.TrimSpace(in.ID)
		if id == "" {
			continue
		}

		// Build context for the already processed inputs.
		ctxPrefix, err := BuildContext(inputs[:idx], answers)
		if err != nil {
			return nil, fmt.Errorf("building context for %q: %w", id, err)
		}

		ok, err := ShouldAskInput(in, ctxPrefix)
		if err != nil {
			return nil, fmt.Errorf("evaluating when for input %q: %w", id, err)
		}
		if !ok {
			// When false => no prompt; BuildContext(inputs, answers) will apply defaults.
			continue
		}

		if ans, exists := answers[id]; exists {
			if err := ValidateInputValue(in, ans); err != nil {
				return nil, fmt.Errorf("invalid value for %q: %w", id, err)
			}
			continue
		}

		label := strings.TrimSpace(in.Prompt)
		if label == "" {
			label = id
		}

		var val any
		switch strings.ToLower(strings.TrimSpace(in.Type)) {
		case "string":
			s, err := promptString(label, in)
			if err != nil {
				return nil, err
			}
			val = s
		case "enum":
			s, err := promptEnum(label, in)
			if err != nil {
				return nil, err
			}
			val = s
		case "bool":
			b, err := promptBool(label, in)
			if err != nil {
				return nil, err
			}
			val = b
		default:
			return nil, fmt.Errorf("unsupported input type %q for %q", in.Type, id)
		}

		if err := ValidateInputValue(in, fmt.Sprint(val)); err != nil {
			return nil, fmt.Errorf("invalid value for %q: %w", id, err)
		}
		answers[id] = fmt.Sprint(val)
	}

	return BuildContext(inputs, answers)
}

type stringPromptModel struct {
	label    string
	input    textinput.Model
	required bool
	errMsg   string
	done     bool
	value    string
}

func (m stringPromptModel) Init() tea.Cmd { return textinput.Blink }

func (m stringPromptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			v := strings.TrimSpace(m.input.Value())
			if v == "" {
				v = strings.TrimSpace(m.input.Placeholder)
			}
			if m.required && strings.TrimSpace(v) == "" {
				m.errMsg = "required"
				return m, nil
			}
			m.value = v
			m.done = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m stringPromptModel) View() string {
	if m.done {
		return ""
	}
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Bold(true).Render(m.label))
	b.WriteString("\n")
	b.WriteString(m.input.View())
	if strings.TrimSpace(m.errMsg) != "" {
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(m.errMsg))
	}
	b.WriteString("\n\n(enter to confirm, esc to cancel)")
	return b.String()
}

func promptString(label string, in dsl.Input) (string, error) {
	ti := textinput.New()
	ti.Prompt = "> "
	ti.Placeholder = ""
	if in.Default != nil {
		ti.Placeholder = fmt.Sprintf("%v", in.Default)
	}
	ti.Focus()

	m := stringPromptModel{
		label:    label,
		input:    ti,
		required: in.Required && strings.TrimSpace(ti.Placeholder) == "",
	}
	p := tea.NewProgram(m)
	final, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("prompting %q: %w", in.ID, err)
	}
	fm, ok := final.(stringPromptModel)
	if !ok || !fm.done {
		return "", fmt.Errorf("input cancelled")
	}
	return fm.value, nil
}

type enumItem struct{ value string }

func (i enumItem) Title() string       { return i.value }
func (i enumItem) Description() string { return "" }
func (i enumItem) FilterValue() string { return "" }

type enumPromptModel struct {
	list     list.Model
	required bool
	selected string
	done     bool
}

func (m enumPromptModel) Init() tea.Cmd { return nil }

func (m enumPromptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			if it, ok := m.list.SelectedItem().(enumItem); ok {
				m.selected = it.value
				m.done = true
				return m, tea.Quit
			}
			if m.required {
				return m, nil
			}
			m.done = true
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m enumPromptModel) View() string {
	if m.done {
		return ""
	}
	return m.list.View()
}

func promptEnum(label string, in dsl.Input) (string, error) {
	if len(in.Options) == 0 {
		return "", fmt.Errorf("enum %q has no options", in.ID)
	}
	items := make([]list.Item, 0, len(in.Options))
	for _, opt := range in.Options {
		items = append(items, enumItem{value: opt})
	}

	l := list.New(items, list.NewDefaultDelegate(), 60, min(12, len(items)+4))
	l.Title = label
	l.SetFilteringEnabled(false)
	l.SetShowHelp(true)
	l.SetShowStatusBar(false)
	l.DisableQuitKeybindings()

	// preselect default if present
	def := ""
	if in.Default != nil {
		def = fmt.Sprintf("%v", in.Default)
	}
	if def != "" {
		for idx, opt := range in.Options {
			if opt == def {
				l.Select(idx)
				break
			}
		}
	}

	p := tea.NewProgram(enumPromptModel{list: l, required: in.Required})
	final, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("prompting %q: %w", in.ID, err)
	}
	fm, ok := final.(enumPromptModel)
	if !ok || !fm.done {
		return "", fmt.Errorf("input cancelled")
	}
	if strings.TrimSpace(fm.selected) == "" && def != "" {
		return def, nil
	}
	return fm.selected, nil
}

func promptBool(label string, in dsl.Input) (bool, error) {
	def := false
	if in.Default != nil {
		if b, ok := in.Default.(bool); ok {
			def = b
		} else if s, ok := in.Default.(string); ok {
			if v, ok := parseBoolString(s); ok {
				def = v
			}
		}
	}
	hint := "y/n"
	if def {
		hint = "Y/n"
	} else {
		hint = "y/N"
	}
	ti := textinput.New()
	ti.Prompt = "> "
	ti.Placeholder = hint
	ti.Focus()

	m := stringPromptModel{
		label: label,
		input: ti,
		// bool is always answerable due to default
		required: false,
	}
	p := tea.NewProgram(m)
	final, err := p.Run()
	if err != nil {
		return false, fmt.Errorf("prompting %q: %w", in.ID, err)
	}
	fm, ok := final.(stringPromptModel)
	if !ok || !fm.done {
		return false, fmt.Errorf("input cancelled")
	}
	raw := strings.TrimSpace(fm.input.Value())
	if raw == "" {
		return def, nil
	}
	if v, ok := parseBoolString(raw); ok {
		return v, nil
	}
	return false, fmt.Errorf("expected y/n (or true/false)")
}

func parseBoolString(s string) (bool, bool) {
	v := strings.ToLower(strings.TrimSpace(s))
	switch v {
	case "y", "yes", "true", "1":
		return true, true
	case "n", "no", "false", "0":
		return false, true
	default:
		return false, false
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

