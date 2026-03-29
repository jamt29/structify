package tui

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/jamt29/structify/internal/dsl"
	"github.com/jamt29/structify/internal/template"
)

type templatesMode int

const (
	modeList templatesMode = iota
	modeCreate
	modeDetail
	modeDelete
	modeEdit
)

type listRowKind int

const (
	rowLocalTemplate listRowKind = iota
	rowNewTemplate
	rowBuiltin
)

type listRow struct {
	kind listRowKind
	t    *template.Template
}

type TemplatesModel struct {
	local    []*template.Template
	builtins []*template.Template

	rows   []listRow
	cursor int

	width  int
	height int

	detail *dsl.Manifest
	mode   templatesMode

	createForm *huh.Form
	createName *string
	createDesc *string
	createLang *string
	createArch *string
	deleteForm   *huh.Form
	deleteOK     *bool
	deleteTarget *template.Template

	needsReload       bool
	selectAfterReload string
	toast             string

	// PARTE C: YAML editor (nil until modeEdit)
	editor tea.Model

	done bool
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
	m.rebuildRows()
	if len(m.rows) == 0 {
		m.detail = nil
	}
	return m
}

func (m *TemplatesModel) rebuildRows() {
	m.rows = nil
	for _, t := range m.local {
		m.rows = append(m.rows, listRow{kind: rowLocalTemplate, t: t})
	}
	m.rows = append(m.rows, listRow{kind: rowNewTemplate, t: nil})
	for _, t := range m.builtins {
		m.rows = append(m.rows, listRow{kind: rowBuiltin, t: t})
	}
	if m.cursor >= len(m.rows) && len(m.rows) > 0 {
		m.cursor = len(m.rows) - 1
	}
}

func (m *TemplatesModel) Init() tea.Cmd { return nil }

// Mode returns the active templates screen mode (for RootModel centering).
func (m *TemplatesModel) Mode() templatesMode { return m.mode }

func (m *TemplatesModel) NeedsReload() bool { return m.needsReload }

func (m *TemplatesModel) ClearReload() {
	m.needsReload = false
	m.selectAfterReload = ""
}

func (m *TemplatesModel) SelectAfterReloadName() string { return m.selectAfterReload }

func (m *TemplatesModel) SelectByName(name string) {
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}
	for i, row := range m.rows {
		if row.t == nil || row.t.Manifest == nil {
			continue
		}
		if strings.TrimSpace(row.t.Manifest.Name) == name {
			m.cursor = i
			return
		}
	}
	// Match install folder name (store key), not only manifest title.
	for i, row := range m.rows {
		if row.t == nil {
			continue
		}
		if templateStoreDirName(row.t) == name {
			m.cursor = i
			return
		}
	}
}

// templateStoreDirName returns the on-disk template folder name (store key for local templates).
func templateStoreDirName(t *template.Template) string {
	if t == nil || strings.TrimSpace(t.Path) == "" {
		return ""
	}
	return filepath.Base(t.Path)
}

func (m *TemplatesModel) startCreate() tea.Cmd {
	m.createName = new(string)
	m.createDesc = new(string)
	m.createLang = new(string)
	m.createArch = new(string)
	*m.createLang = "go"
	*m.createArch = "clean"

	m.createForm = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key("name").
				Title("Nombre del template?").
				Value(m.createName).
				Validate(func(s string) error {
					return ValidateInputValue(dsl.Input{
						ID: "name", Type: "string", Required: true,
						Validate: template.ProjectNameValidateRegex,
					}, s)
				}),
			huh.NewInput().
				Key("description").
				Title("Descripción?").
				Value(m.createDesc),
			huh.NewSelect[string]().
				Key("language").
				Title("Lenguaje").
				Options(
					huh.NewOption("go", "go"),
					huh.NewOption("typescript", "typescript"),
					huh.NewOption("rust", "rust"),
					huh.NewOption("csharp", "csharp"),
					huh.NewOption("python", "python"),
				).
				Value(m.createLang),
			huh.NewSelect[string]().
				Key("architecture").
				Title("Arquitectura").
				Options(
					huh.NewOption("clean", "clean"),
					huh.NewOption("vertical-slice", "vertical-slice"),
					huh.NewOption("hexagonal", "hexagonal"),
					huh.NewOption("mvc", "mvc"),
					huh.NewOption("monorepo", "monorepo"),
				).
				Value(m.createArch),
		),
	).
		WithTheme(structifyHuhTheme()).
		WithShowHelp(false).
		WithWidth(m.formWidth())
	return m.createForm.Init()
}

func (m *TemplatesModel) formWidth() int {
	w := m.width - 8
	if w < 40 {
		w = 40
	}
	return w
}

func (m *TemplatesModel) startDelete() tea.Cmd {
	if m.deleteTarget == nil {
		return nil
	}
	m.deleteOK = new(bool)
	*m.deleteOK = false
	name := ""
	if m.deleteTarget.Manifest != nil {
		name = strings.TrimSpace(m.deleteTarget.Manifest.Name)
	}
	m.deleteForm = huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Key("confirm").
				Title(fmt.Sprintf("¿Eliminar template local %q?", name)).
				Value(m.deleteOK),
		),
	).
		WithTheme(structifyHuhTheme()).
		WithShowHelp(false).
		WithWidth(m.formWidth())
	return m.deleteForm.Init()
}

func (m *TemplatesModel) selectedRow() *listRow {
	if m.cursor < 0 || m.cursor >= len(m.rows) {
		return nil
	}
	return &m.rows[m.cursor]
}

func (m *TemplatesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.width, m.height = ws.Width, ws.Height
		if m.createForm != nil {
			m.createForm = m.createForm.WithWidth(m.formWidth())
		}
		if m.deleteForm != nil {
			m.deleteForm = m.deleteForm.WithWidth(m.formWidth())
		}
		if m.editor != nil {
			newEd, cmd := m.editor.Update(msg)
			m.editor = newEd
			return m, cmd
		}
		return m, nil
	}

	switch m.mode {
	case modeEdit:
		if m.editor != nil {
			newEd, cmd := m.editor.Update(msg)
			m.editor = newEd
			if yed, ok := newEd.(*YAMLEditorModel); ok && yed.Done() {
				if yed.PendingReload() {
					m.needsReload = true
					m.selectAfterReload = yed.TemplateName()
				}
				m.editor = nil
				m.mode = modeList
				return m, nil
			}
			return m, cmd
		}
		return m, nil

	case modeCreate:
		if m.createForm != nil {
			form, cmd := m.createForm.Update(msg)
			m.createForm = form.(*huh.Form)
			switch m.createForm.State {
			case huh.StateCompleted:
				m.finishCreate()
				m.mode = modeList
				m.createForm = nil
				return m, nil
			case huh.StateAborted:
				m.mode = modeList
				m.createForm = nil
				return m, nil
			}
			return m, cmd
		}
		return m, nil

	case modeDelete:
		if m.deleteForm != nil {
			form, cmd := m.deleteForm.Update(msg)
			m.deleteForm = form.(*huh.Form)
			switch m.deleteForm.State {
			case huh.StateCompleted:
				if m.deleteOK != nil && *m.deleteOK && m.deleteTarget != nil && m.deleteTarget.Manifest != nil {
					n := templateStoreDirName(m.deleteTarget)
					if n == "" {
						n = strings.TrimSpace(m.deleteTarget.Manifest.Name)
					}
					if err := template.Remove(n); err != nil {
						m.toast = err.Error()
					} else {
						m.toast = fmt.Sprintf("Eliminado: %s", n)
						m.needsReload = true
						m.selectAfterReload = ""
					}
				}
				m.mode = modeList
				m.deleteForm = nil
				m.deleteTarget = nil
				return m, nil
			case huh.StateAborted:
				m.mode = modeList
				m.deleteForm = nil
				m.deleteTarget = nil
				return m, nil
			}
			return m, cmd
		}
		return m, nil
	}

	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.mode == modeDetail {
				m.mode = modeList
				m.detail = nil
				return m, nil
			}
			m.done = true
			return m, nil
		case "up", "k":
			if len(m.rows) == 0 {
				return m, nil
			}
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.rows) - 1
			}
			m.detail = nil
			m.mode = modeList
			return m, nil
		case "down", "j":
			if len(m.rows) == 0 {
				return m, nil
			}
			m.cursor++
			if m.cursor >= len(m.rows) {
				m.cursor = 0
			}
			m.detail = nil
			m.mode = modeList
			return m, nil
		case "enter":
			if len(m.rows) == 0 {
				return m, nil
			}
			row := m.rows[m.cursor]
			switch row.kind {
			case rowNewTemplate:
				return m, nil
			case rowLocalTemplate, rowBuiltin:
				if row.t != nil {
					m.detail = row.t.Manifest
					m.mode = modeDetail
				}
			}
			return m, nil
		case "n":
			m.mode = modeCreate
			return m, m.startCreate()
		case "e":
			row := m.selectedRow()
			if row != nil && row.kind == rowLocalTemplate && row.t != nil {
				m.mode = modeEdit
				m.editor = NewYAMLEditor(row.t)
				return m, m.editor.Init()
			}
			return m, nil
		case "d":
			row := m.selectedRow()
			if row != nil && row.kind == rowLocalTemplate && row.t != nil {
				m.mode = modeDelete
				m.deleteTarget = row.t
				return m, m.startDelete()
			}
			return m, nil
		}
	}

	return m, nil
}

func (m *TemplatesModel) finishCreate() {
	author := detectGitUserForTemplates()
	if strings.TrimSpace(author) == "" {
		author = "structify"
	}
	name := strings.TrimSpace(*m.createName)
	desc := strings.TrimSpace(*m.createDesc)
	lang := strings.TrimSpace(*m.createLang)
	arch := strings.TrimSpace(*m.createArch)
	if err := template.CreateMinimalLocalTemplate("", name, desc, lang, arch, author); err != nil {
		m.toast = err.Error()
		return
	}
	m.toast = fmt.Sprintf("Template creado: %s", name)
	m.needsReload = true
	m.selectAfterReload = name
}

func detectGitUserForTemplates() string {
	out, err := exec.Command("git", "config", "user.name").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func (m *TemplatesModel) View() string {
	return m.ViewContent()
}

// ViewContent returns the raw (non-centered) content of this screen.
func (m *TemplatesModel) ViewContent() string {
	if m.width < 80 || m.height < 24 {
		return stylePending.Render("Terminal too small. Minimum 80x24.")
	}
	return lipgloss.NewStyle().MaxWidth(MaxWidthTemplates).Render(m.viewContentInner())
}

func (m *TemplatesModel) viewContentInner() string {
	switch m.mode {
	case modeCreate:
		if m.createForm != nil {
			header := styleHeader.Render("structify") + "  " + styleCompletedLabel.Render("·  Mis templates  ·  Nuevo template")
			return header + "\n\n" + m.createForm.View()
		}
	case modeDelete:
		if m.deleteForm != nil {
			header := styleHeader.Render("structify") + "  " + styleCompletedLabel.Render("·  Mis templates  ·  Eliminar")
			return header + "\n\n" + m.deleteForm.View()
		}
	case modeEdit:
		if m.editor != nil {
			if v, ok := m.editor.(interface{ ViewContent() string }); ok {
				return v.ViewContent()
			}
			return m.editor.View()
		}
	}

	if len(m.rows) == 0 {
		var b strings.Builder
		b.WriteString(styleHeader.Render("structify") + "  " + styleCompletedLabel.Render("·  Mis templates"))
		b.WriteString("\n\n")
		b.WriteString(stylePending.Render("No templates found."))
		b.WriteString("\n")
		return b.String()
	}

	layoutW := minInt(m.width, MaxWidthTemplates)
	leftW := m.localColWidth()
	rightW := layoutW - leftW - 4
	if rightW < 28 {
		rightW = 28
	}

	nameW := maxInt(8, leftW-m.metaColWidth()-2)
	metaW := m.metaColWidth()

	localLines := make([]string, 0, len(m.local)+1)
	for _, t := range m.local {
		localLines = append(localLines, rowLineTwoCol(t, nameW, metaW))
	}
	localLines = append(localLines, "[+] Nuevo template")

	rightLines := make([]string, 0, len(m.builtins))
	for _, t := range m.builtins {
		rightLines = append(rightLines, rowLineTwoCol(t, nameW, metaW))
	}

	var b strings.Builder
	header := styleHeader.Render("structify") + "  " + styleCompletedLabel.Render("·  Mis templates")
	b.WriteString(header)
	b.WriteString("\n\n")

	leftHead := fmt.Sprintf("Local (%d)", len(m.local))
	rightHead := fmt.Sprintf("Built-in (%d)", len(m.builtins))
	sepLeft := minInt(leftW-2, 32)
	if sepLeft < 8 {
		sepLeft = 8
	}
	sepRight := minInt(rightW-2, 40)
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(leftW).Render(
			styleCompletedLabel.Render(leftHead)+"\n"+stylePending.Render(strings.Repeat("─", sepLeft))+"\n"+m.joinRowsLocal(localLines, leftW),
		),
		lipgloss.NewStyle().Width(rightW).Render(
			styleCompletedLabel.Render(rightHead)+"\n"+stylePending.Render(strings.Repeat("─", sepRight))+"\n"+m.joinRowsBuiltin(rightLines, rightW),
		),
	))

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

	if m.toast != "" {
		b.WriteString("\n")
		b.WriteString(styleCheckmark.Render(m.toast))
		b.WriteString("\n")
	}

	main := b.String()
	help := "enter detalle  n nuevo  e editar  d eliminar  esc volver"
	return m.padHelpBarToBottom(main, help)
}

func (m *TemplatesModel) localColWidth() int {
	maxLen := 20
	for _, t := range m.local {
		if t == nil || t.Manifest == nil {
			continue
		}
		n := len(strings.TrimSpace(t.Manifest.Name))
		if n > maxLen {
			maxLen = n
		}
	}
	if len("[+] Nuevo template") > maxLen {
		maxLen = len("[+] Nuevo template")
	}
	return minInt(maxLen+4, 35)
}

func (m *TemplatesModel) metaColWidth() int {
	return 24
}

func (m *TemplatesModel) padHelpBarToBottom(main string, help string) string {
	helpBar := styleHelpBar.Render(help) + "\n"
	ch := lipgloss.Height(main)
	hh := lipgloss.Height(helpBar)
	pad := m.height - ch - hh - 1
	if pad < 0 {
		pad = 0
	}
	return main + strings.Repeat("\n", pad) + helpBar
}

func rowLineTwoCol(t *template.Template, nameW, metaW int) string {
	if t == nil || t.Manifest == nil {
		return ""
	}
	name := strings.TrimSpace(t.Manifest.Name)
	lang := strings.TrimSpace(t.Manifest.Language)
	arch := strings.TrimSpace(t.Manifest.Architecture)
	meta := strings.TrimSpace(lang + " · " + arch)
	nameSt := lipgloss.NewStyle().MaxWidth(nameW).Inline(true).Render(name)
	metaSt := lipgloss.NewStyle().MaxWidth(metaW).Inline(true).Foreground(colorMuted).Render(meta)
	return lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(nameW).Render(nameSt),
		lipgloss.NewStyle().Width(metaW).Render(metaSt),
	)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (m *TemplatesModel) cursorLeftLine() int {
	if m.cursor < 0 || m.cursor >= len(m.rows) {
		return -1
	}
	r := m.rows[m.cursor]
	switch r.kind {
	case rowLocalTemplate:
		for i, t := range m.local {
			if t == r.t {
				return i
			}
		}
	case rowNewTemplate:
		return len(m.local)
	default:
		return -1
	}
	return -1
}

func (m *TemplatesModel) cursorRightLine() int {
	if m.cursor < 0 || m.cursor >= len(m.rows) {
		return -1
	}
	r := m.rows[m.cursor]
	if r.kind != rowBuiltin {
		return -1
	}
	for i, t := range m.builtins {
		if t == r.t {
			return i
		}
	}
	return -1
}

func (m *TemplatesModel) joinRowsLocal(lines []string, width int) string {
	var b strings.Builder
	activeLine := m.cursorLeftLine()
	for i, line := range lines {
		active := i == activeLine
		row := line
		if active {
			row = styleMenuItemActive.Render("› " + line)
		} else {
			row = styleMenuItem.Render("  " + line)
		}
		b.WriteString(lipgloss.NewStyle().Width(width).Render(row))
		b.WriteString("\n")
	}
	return b.String()
}

func (m *TemplatesModel) joinRowsBuiltin(lines []string, width int) string {
	var b strings.Builder
	activeLine := m.cursorRightLine()
	for i, line := range lines {
		active := i == activeLine
		row := line
		if active {
			row = styleMenuItemActive.Render("› " + line)
		} else {
			row = styleMenuItem.Render("  " + line)
		}
		b.WriteString(lipgloss.NewStyle().Width(width).Render(row))
		b.WriteString("\n")
	}
	return b.String()
}
