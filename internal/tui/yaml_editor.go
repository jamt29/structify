package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/jamt29/structify/internal/dsl"
	"github.com/jamt29/structify/internal/template"
)

// YAMLEditorModel edita scaffold.yaml de un template local en el TUI.
type YAMLEditorModel struct {
	tpl *template.Template

	ta textarea.Model

	path     string
	original string

	validationErrs []string
	saveErr        string

	width  int
	height int

	confirmForm *huh.Form
	confirmYes  *bool
	confirmKind string // "quit" | ""

	done          bool
	reloadName    string
	pendingReload bool
}

// NewYAMLEditor carga el YAML del template local.
func NewYAMLEditor(tpl *template.Template) *YAMLEditorModel {
	m := &YAMLEditorModel{
		tpl:    tpl,
		width:  100,
		height: 24,
	}
	if tpl == nil || tpl.Manifest == nil {
		m.saveErr = "template inválido"
		return m
	}
	m.path = filepath.Join(tpl.Path, "scaffold.yaml")
	b, err := os.ReadFile(m.path)
	if err != nil {
		m.saveErr = err.Error()
		return m
	}
	m.original = string(b)

	ta := textarea.New()
	ta.SetWidth(72)
	ta.SetHeight(18)
	ta.CharLimit = 512 * 1024
	ta.SetValue(m.original)
	ta.Focus()
	m.ta = ta

	return m
}

// Done indica que el usuario salió del editor (esc o guardado final).
func (m *YAMLEditorModel) Done() bool {
	return m.done
}

// TemplateName nombre a recargar si PendingReload es true.
func (m *YAMLEditorModel) TemplateName() string {
	return m.reloadName
}

// PendingReload indica si hubo guardado exitoso y hay que refrescar el listado.
func (m *YAMLEditorModel) PendingReload() bool {
	return m.pendingReload
}

func (m *YAMLEditorModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m *YAMLEditorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.saveErr != "" && m.original == "" {
		if k, ok := msg.(tea.KeyMsg); ok && k.String() == "esc" {
			m.done = true
		}
		return m, nil
	}

	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.width, m.height = ws.Width, ws.Height
		w := m.width - 6
		if w < 40 {
			w = 40
		}
		h := m.height - 14
		if h < 8 {
			h = 8
		}
		m.ta.SetWidth(w)
		m.ta.SetHeight(h)
	}

	if m.confirmForm != nil {
		form, cmd := m.confirmForm.Update(msg)
		m.confirmForm = form.(*huh.Form)
		switch m.confirmForm.State {
		case huh.StateCompleted:
			if m.confirmYes != nil && *m.confirmYes && m.confirmKind == "quit" {
				m.done = true
			}
			m.confirmForm = nil
			m.confirmKind = ""
			m.confirmYes = nil
			return m, nil
		case huh.StateAborted:
			m.confirmForm = nil
			m.confirmKind = ""
			m.confirmYes = nil
			return m, nil
		}
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+s":
			m.trySave()
			return m, nil
		case "esc":
			if m.ta.Value() != m.original {
				return m, m.startQuitConfirm()
			}
			m.done = true
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.ta, cmd = m.ta.Update(msg)
	return m, cmd
}

func (m *YAMLEditorModel) startQuitConfirm() tea.Cmd {
	m.confirmYes = new(bool)
	*m.confirmYes = false
	m.confirmKind = "quit"
	m.confirmForm = huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Key("quit").
				Title("¿Descartar cambios sin guardar?").
				Value(m.confirmYes),
		),
	).
		WithTheme(structifyHuhTheme()).
		WithShowHelp(false).
		WithWidth(minWidth(m.width-8, 50))
	return m.confirmForm.Init()
}

func minWidth(w, floor int) int {
	if w < floor {
		return floor
	}
	return w
}

func (m *YAMLEditorModel) trySave() {
	m.validationErrs = nil
	m.saveErr = ""
	text := m.ta.Value()
	man, err := dsl.ParseManifest([]byte(text))
	if err != nil {
		m.validationErrs = []string{err.Error()}
		return
	}
	verrs := dsl.ValidateManifest(man)
	if len(verrs) > 0 {
		for _, e := range verrs {
			m.validationErrs = append(m.validationErrs, fmt.Sprintf("%s: %s", e.Field, e.Message))
		}
		return
	}
	if err := os.WriteFile(m.path, []byte(text), 0o644); err != nil {
		m.saveErr = err.Error()
		return
	}
	m.original = text
	m.reloadName = strings.TrimSpace(man.Name)
	m.pendingReload = true
	m.done = true
}

func (m *YAMLEditorModel) View() string {
	return m.ViewContent()
}

// ViewContent dibuja el panel del editor (para RootModel / centrado).
func (m *YAMLEditorModel) ViewContent() string {
	if m.saveErr != "" && m.original == "" {
		return styleHeader.Render("structify") + "\n\n" + stylePending.Render(m.saveErr) + "\n\n" + styleHelpBar.Render("esc volver")
	}

	title := "scaffold.yaml"
	if m.tpl != nil && m.tpl.Manifest != nil {
		title = strings.TrimSpace(m.tpl.Manifest.Name)
	}
	header := styleHeader.Render("structify") + "  " + styleCompletedLabel.Render("·  Mis templates  ·  Editando: "+title)

	var b strings.Builder
	b.WriteString(header)
	b.WriteString("\n\n")

	if m.confirmForm != nil {
		b.WriteString(m.confirmForm.View())
		return b.String()
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Padding(0, 1).
		Width(m.width - 4)
	b.WriteString(styleCompletedLabel.Render("┌─ scaffold.yaml " + strings.Repeat("─", maxInt(0, minInt(m.width-24, 40))) + "┐"))
	b.WriteString("\n")
	b.WriteString(box.Render(m.ta.View()))
	b.WriteString("\n")
	b.WriteString(styleCompletedLabel.Render("└" + strings.Repeat("─", maxInt(0, m.width-8)) + "┘"))

	if len(m.validationErrs) > 0 {
		b.WriteString("\n")
		for _, line := range m.validationErrs {
			b.WriteString(stylePending.Render(line))
			b.WriteString("\n")
		}
	}
	if m.saveErr != "" {
		b.WriteString("\n")
		b.WriteString(stylePending.Render(m.saveErr))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(styleHelpBar.Render("ctrl+s guardar  esc salir  (validación al guardar)"))
	b.WriteString("\n")
	return b.String()
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
