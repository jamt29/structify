package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"

	"github.com/jamt29/structify/internal/dsl"
	"github.com/jamt29/structify/internal/engine"
	tmpl "github.com/jamt29/structify/internal/template"
)

type state int

const (
	stateSelectTemplate state = iota
	stateInputs
	stateConfirm
	stateProgress
	stateDone
	stateError
)

type progressLine struct {
	name    string
	status  string
	command string
}

type inputEntry struct {
	in      dsl.Input
	kind    string
	ti      textinput.Model
	enum    list.Model
	boolVal bool
}

type App struct {
	state state

	templates []*tmpl.Template
	selected  *tmpl.Template
	answers   dsl.Context
	result    *tmpl.ScaffoldResult
	err       error
	done      bool

	selector    list.Model
	inputs      []inputEntry
	activeInput int
	compactForm bool

	spin        spinner.Model
	progressLog []progressLine
	progressCh  <-chan tea.Msg

	engine *engine.Engine
	width  int
	height int
}

// Done retorna true cuando el usuario presionó cualquier tecla en stateDone/stateError.
// Se usa por RootModel para volver al menú sin salir del programa.
func (a *App) Done() bool { return a.done }

type msgFilesDone struct{ count int }
type msgStepStart struct {
	name    string
	command string
}
type msgStepDone struct {
	name    string
	skipped bool
}
type msgStepError struct {
	name string
	err  error
}
type msgScaffoldDone struct{ result *tmpl.ScaffoldResult }
type msgScaffoldError struct{ err error }
type msgProgressReady struct{ ch <-chan tea.Msg }
type msgProgressClosed struct{}

func RunApp(templates []*tmpl.Template, eng *engine.Engine) error {
	if eng == nil {
		return fmt.Errorf("engine is nil")
	}
	m, err := newApp(templates, eng)
	if err != nil {
		return err
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return fmt.Errorf("running tui app: %w", err)
	}
	fm, ok := final.(*App)
	if !ok {
		return fmt.Errorf("unexpected final model")
	}
	return fm.err
}

func newApp(templates []*tmpl.Template, eng *engine.Engine) (*App, error) {
	items := make([]list.Item, 0, len(templates))
	for _, t := range templates {
		if t == nil || t.Manifest == nil || strings.TrimSpace(t.Manifest.Name) == "" {
			continue
		}
		items = append(items, templateItem{t: t})
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("no templates found")
	}

	sel := list.New(items, list.NewDefaultDelegate(), 80, 20)
	sel.SetFilteringEnabled(true)
	sel.SetShowStatusBar(false)
	sel.SetShowHelp(false)
	sel.Title = "Selecciona un template"

	spin := spinner.New()
	spin.Spinner = spinner.Line

	return &App{
		state:       stateSelectTemplate,
		templates:   templates,
		answers:     dsl.Context{},
		selector:    sel,
		compactForm: true,
		spin:        spin,
		engine:      eng,
		width:       100,
		height:      30,
	}, nil
}

func (a *App) Init() tea.Cmd { return nil }

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		a.width, a.height = m.Width, m.Height
		a.resizeComponents()
		return a, nil
	case msgProgressReady:
		a.progressCh = m.ch
		return a, waitProgressMsg(a.progressCh)
	}

	switch a.state {
	case stateSelectTemplate:
		return a.updateSelect(msg)
	case stateInputs:
		return a.updateInputs(msg)
	case stateConfirm:
		return a.updateConfirm(msg)
	case stateProgress:
		return a.updateProgress(msg)
	case stateDone, stateError:
		if k, ok := msg.(tea.KeyMsg); ok {
			// Permite salir globalmente con Ctrl+C.
			if k.String() == "ctrl+c" {
				a.err = fmt.Errorf("cancelled")
				return a, tea.Quit
			}
			a.done = true
		}
		return a, nil
	default:
		return a, nil
	}
}

func (a *App) updateSelect(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "ctrl+c", "esc":
			a.err = fmt.Errorf("cancelled")
			return a, tea.Quit
		case "enter":
			it, ok := a.selector.SelectedItem().(templateItem)
			if !ok || it.t == nil {
				return a, nil
			}
			a.selected = it.t
			a.prepareInputs()
			a.state = stateInputs
			return a, nil
		}
	}
	var cmd tea.Cmd
	a.selector, cmd = a.selector.Update(msg)
	return a, cmd
}

func (a *App) updateInputs(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "ctrl+c":
			a.err = fmt.Errorf("cancelled")
			return a, tea.Quit
		case "esc":
			a.state = stateSelectTemplate
			return a, nil
		case "tab":
			if len(a.inputs) > 0 {
				a.activeInput = (a.activeInput + 1) % len(a.inputs)
				return a, nil
			}
		case "shift+tab":
			if len(a.inputs) > 0 {
				a.activeInput--
				if a.activeInput < 0 {
					a.activeInput = len(a.inputs) - 1
				}
				return a, nil
			}
		case "left":
			if e := a.currentInput(); e != nil && e.kind == "bool" {
				e.boolVal = false
				a.inputs[a.activeInput] = *e
				return a, nil
			}
		case "right":
			if e := a.currentInput(); e != nil && e.kind == "bool" {
				e.boolVal = true
				a.inputs[a.activeInput] = *e
				return a, nil
			}
		case "y":
			if e := a.currentInput(); e != nil && e.kind == "bool" {
				e.boolVal = true
				a.inputs[a.activeInput] = *e
				return a, nil
			}
		case "n":
			if e := a.currentInput(); e != nil && e.kind == "bool" {
				e.boolVal = false
				a.inputs[a.activeInput] = *e
				return a, nil
			}
		case "enter":
			if a.compactForm {
				if err := a.captureCompactAnswers(); err != nil {
					return a, nil
				}
				ctx, err := a.buildContextFromInputs()
				if err != nil {
					return a.toError(err), nil
				}
				a.answers = ctx
				a.state = stateConfirm
				return a, nil
			}
			if err := a.captureCurrentSequential(); err != nil {
				return a, nil
			}
			a.activeInput++
			if a.activeInput >= len(a.inputs) {
				ctx, err := a.buildContextFromInputs()
				if err != nil {
					return a.toError(err), nil
				}
				a.answers = ctx
				a.state = stateConfirm
			}
			return a, nil
		}
	}

	if len(a.inputs) == 0 {
		return a, nil
	}
	idx := a.activeInput
	if idx < 0 {
		idx = 0
	}
	if idx >= len(a.inputs) {
		idx = len(a.inputs) - 1
	}
	a.activeInput = idx
	entry := a.inputs[idx]
	var cmd tea.Cmd
	switch entry.kind {
	case "string":
		entry.ti, cmd = entry.ti.Update(msg)
	case "enum":
		entry.enum, cmd = entry.enum.Update(msg)
	}
	a.inputs[idx] = entry
	return a, cmd
}

func (a *App) updateConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "ctrl+c":
			a.err = fmt.Errorf("cancelled")
			return a, tea.Quit
		case "esc", "b":
			a.state = stateInputs
			return a, nil
		case "enter":
			a.state = stateProgress
			return a, tea.Batch(a.spin.Tick, startScaffoldCmd(a.buildRequest(), a.engine))
		}
	}
	return a, nil
}

func (a *App) updateProgress(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.KeyMsg:
		if m.String() == "ctrl+c" {
			a.err = fmt.Errorf("cancelled")
			return a, tea.Quit
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		a.spin, cmd = a.spin.Update(m)
		return a, cmd
	case msgFilesDone:
		a.progressLog = append(a.progressLog, progressLine{name: "files", status: "done", command: fmt.Sprintf("Archivos generados (%d archivos)", m.count)})
		return a, waitProgressMsg(a.progressCh)
	case msgStepStart:
		a.progressLog = append(a.progressLog, progressLine{name: m.name, status: "running", command: m.command})
		return a, waitProgressMsg(a.progressCh)
	case msgStepDone:
		for i := len(a.progressLog) - 1; i >= 0; i-- {
			if a.progressLog[i].name == m.name && a.progressLog[i].status == "running" {
				if m.skipped {
					a.progressLog[i].status = "skipped"
				} else {
					a.progressLog[i].status = "done"
				}
				break
			}
		}
		return a, waitProgressMsg(a.progressCh)
	case msgStepError:
		a.err = m.err
		a.state = stateError
		a.done = false
		return a, nil
	case msgScaffoldDone:
		a.result = m.result
		a.state = stateDone
		a.done = false
		return a, nil
	case msgScaffoldError:
		a.err = m.err
		a.state = stateError
		a.done = false
		return a, nil
	case msgProgressClosed:
		return a, nil
	}
	return a, nil
}

func (a *App) View() string {
	if a.width < 80 || a.height < 24 {
		return stylePending.Render("Terminal too small. Minimum 80x24.")
	}
	content := strings.Join([]string{
		a.renderHeader(),
		a.renderBody(),
		styleHelpBar.Render(" " + a.helpText()),
	}, "\n")
	return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, content)
}

func (a *App) renderHeader() string {
	parts := []string{styleHeader.Render("structify")}
	if name := strings.TrimSpace(a.templateName()); name != "" {
		parts = append(parts, styleCompletedValue.Render(name))
	}
	step := strings.TrimSpace(a.stepLabel())
	if step != "" {
		if step == "✓ listo" {
			parts = append(parts, styleCheckmark.Render(step))
		} else if step == "error" {
			parts = append(parts, lipgloss.NewStyle().Foreground(colorError).Render(step))
		} else {
			parts = append(parts, stylePending.Render(step))
		}
	}
	return strings.Join(parts, stylePending.Render("  ·  "))
}

func (a *App) renderBody() string {
	switch a.state {
	case stateSelectTemplate:
		return a.selector.View()
	case stateInputs:
		return a.renderInputs()
	case stateConfirm:
		return a.renderConfirm()
	case stateProgress:
		return a.renderProgress()
	case stateDone:
		return a.renderDone()
	case stateError:
		return a.renderError()
	default:
		return ""
	}
}

func (a *App) renderInputs() string {
	if len(a.inputs) == 0 {
		return stylePending.Render("No hay inputs activos. Presiona Enter para continuar.")
	}
	var b strings.Builder
	if !a.compactForm {
		b.WriteString(stylePending.Render(fmt.Sprintf("Input %d de %d", a.activeInput+1, len(a.inputs))))
		b.WriteString("\n\n")
	}
	for i, e := range a.inputs {
		active := i == a.activeInput
		completed := a.inputCompleted(i)
		if !a.compactForm && !active && !completed {
			b.WriteString(stylePending.Render(strings.TrimSpace(e.in.Prompt)))
		} else {
			b.WriteString(a.renderInputBlock(e, active, completed))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func (a *App) renderInputBlock(e inputEntry, active bool, completed bool) string {
	prompt := strings.TrimSpace(e.in.Prompt)
	if prompt == "" {
		prompt = strings.TrimSpace(e.in.ID)
	}
	if completed && !active {
		val, _ := entryValue(e)
		return styleCompletedLabel.Render(prompt) + "\n" + styleCompletedValue.Render(val+"  ") + styleCheckmark.Render("✓")
	}
	if !active {
		return stylePending.Render(prompt)
	}

	var body string
	switch e.kind {
	case "string":
		v := strings.TrimSpace(e.ti.Value())
		if v == "" && strings.TrimSpace(e.ti.Placeholder) != "" {
			v = e.ti.Placeholder
		}
		if v == "" {
			v = stylePending.Render(prompt)
		}
		body = fmt.Sprint(v)
	case "enum":
		lines := make([]string, 0, len(e.in.Options))
		cur, _ := entryValue(e)
		for _, opt := range e.in.Options {
			if opt == cur {
				lines = append(lines, styleCompletedValue.Render("> "+opt))
			} else {
				lines = append(lines, stylePending.Render("  "+opt))
			}
		}
		body = strings.Join(lines, "\n")
	case "bool":
		no := stylePending.Render("No")
		yes := stylePending.Render("Yes")
		if !e.boolVal {
			no = lipgloss.NewStyle().Foreground(colorText).Background(colorBorder).Padding(0, 1).Render("No")
		} else {
			yes = lipgloss.NewStyle().Foreground(colorText).Background(colorBorder).Padding(0, 1).Render("Yes")
		}
		body = "[ " + no + " ]  /  [ " + yes + " ]"
	}
	return styleActiveBox.Render(prompt + "\n" + body)
}

func (a *App) renderConfirm() string {
	var b strings.Builder
	b.WriteString("Confirmar creación\n\n")
	b.WriteString(styleCompletedLabel.Render("Template  ") + styleCompletedValue.Render(a.templateName()) + "\n")
	b.WriteString(styleCompletedLabel.Render("Output    ") + styleCompletedValue.Render(prettyPath(a.outputDir())) + "\n\n")
	b.WriteString(styleCompletedLabel.Render("Variables") + "\n")
	b.WriteString(stylePending.Render(strings.Repeat("─", 44)) + "\n")
	for _, kv := range sortedContextPairs(a.answers) {
		b.WriteString(styleCompletedLabel.Render(padRight(kv.key, 14)) + " " + styleCompletedValue.Render(kv.value) + "\n")
	}
	return styleActiveBox.Width(min(a.width-8, 72)).Render(b.String())
}

func (a *App) renderProgress() string {
	var b strings.Builder
	b.WriteString(styleCompletedValue.Render("Creando " + ctxStringMap(a.answers, "project_name") + "..."))
	b.WriteString("\n\n")
	for _, line := range a.progressLog {
		switch line.status {
		case "running":
			b.WriteString(lipgloss.NewStyle().Foreground(colorActive).Render(a.spin.View()) + " " + styleCompletedValue.Render(line.command) + "\n")
		case "done":
			b.WriteString(styleCheckmark.Render("✓ ") + styleCompletedValue.Render(line.command) + "\n")
		case "skipped":
			b.WriteString(stylePending.Render("─ " + line.command + " (skipped)") + "\n")
		case "error":
			b.WriteString(lipgloss.NewStyle().Foreground(colorError).Render("✗ " + line.command) + "\n")
		}
	}
	return b.String()
}

func (a *App) renderDone() string {
	var b strings.Builder
	b.WriteString(styleCheckmark.Render("✓  Proyecto creado exitosamente"))
	b.WriteString("\n\n")
	if a.result != nil {
		b.WriteString(styleCompletedLabel.Render("Ruta     ") + styleCompletedValue.Render(a.outputDir()) + "\n")
		b.WriteString(styleCompletedLabel.Render("Archivos ") + styleCompletedValue.Render(fmt.Sprintf("%d", len(a.result.FilesCreated))) + "\n\n")
		b.WriteString(styleCompletedValue.Render("Steps") + "\n")
		for _, s := range a.result.StepsExecuted {
			cmd := strings.TrimSpace(s.Command)
			if cmd == "" {
				cmd = strings.TrimSpace(s.Name)
			}
			if s.Skipped {
				b.WriteString(stylePending.Render("─  " + cmd + "  (skipped)") + "\n")
			} else if s.Error != nil {
				b.WriteString(lipgloss.NewStyle().Foreground(colorError).Render("✗  " + cmd) + "\n")
			} else {
				b.WriteString(styleCheckmark.Render("✓  ") + styleCompletedValue.Render(cmd) + "\n")
			}
		}
	}
	b.WriteString("\n")
	b.WriteString(styleCompletedValue.Render("Próximos pasos") + "\n")
	b.WriteString(stylePending.Render(strings.Repeat("─", 14)) + "\n")
	var block strings.Builder
	for _, line := range nextSteps(a.language(), ctxStringMap(a.answers, "project_name")) {
		block.WriteString(line + "\n")
	}
	b.WriteString(lipgloss.NewStyle().BorderLeft(true).BorderForeground(colorActive).PaddingLeft(1).Render(strings.TrimSpace(block.String())))
	b.WriteString("\n\n")
	b.WriteString(stylePending.Render("(presiona cualquier tecla para salir)"))
	return b.String()
}

func (a *App) renderError() string {
	msg := "unknown error"
	if a.err != nil {
		msg = a.err.Error()
	}
	return lipgloss.NewStyle().Foreground(colorError).Bold(true).Render("✗ " + msg)
}

func (a *App) prepareInputs() {
	if a.selected == nil || a.selected.Manifest == nil {
		return
	}
	entries := make([]inputEntry, 0, len(a.selected.Manifest.Inputs))
	for _, in := range a.selected.Manifest.Inputs {
		id := strings.TrimSpace(in.ID)
		if id == "" {
			continue
		}
		ok, err := ShouldAskInput(in, a.answers)
		if err != nil || !ok {
			continue
		}
		entry := inputEntry{in: in, kind: strings.ToLower(strings.TrimSpace(in.Type))}
		switch entry.kind {
		case "string":
			entry.ti = textinput.New()
			entry.ti.Prompt = ""
			def, _ := ApplyDefault(in, a.answers)
			if strings.TrimSpace(def) != "" {
				entry.ti.Placeholder = def
			} else {
				entry.ti.Placeholder = ""
			}
			entry.ti.SetValue(answerString(a.answers, id))
			entry.ti.Focus()
		case "bool":
			def, _ := ApplyDefault(in, a.answers)
			if ans, ok := a.answers[id]; ok && ans != nil {
				s := strings.ToLower(strings.TrimSpace(fmt.Sprint(ans)))
				entry.boolVal = s == "true" || s == "1" || s == "yes" || s == "y"
			} else {
				s := strings.ToLower(strings.TrimSpace(def))
				entry.boolVal = s == "true" || s == "1" || s == "yes" || s == "y"
			}
		case "enum":
			items := make([]list.Item, 0, len(in.Options))
			for _, opt := range in.Options {
				items = append(items, enumItem{value: opt})
			}
			l := list.New(items, list.NewDefaultDelegate(), 60, max(3, len(items)+1))
			l.SetFilteringEnabled(false)
			l.SetShowStatusBar(false)
			l.SetShowHelp(false)
			l.Title = strings.TrimSpace(in.Prompt)
			cur := strings.TrimSpace(answerString(a.answers, id))
			if cur == "" {
				def, _ := ApplyDefault(in, a.answers)
				cur = strings.TrimSpace(def)
			}
			for i, opt := range in.Options {
				if opt == cur {
					l.Select(i)
					break
				}
			}
			entry.enum = l
		default:
			continue
		}
		entries = append(entries, entry)
	}
	a.inputs = entries
	a.activeInput = 0
	a.compactForm = len(entries) <= 3
}

func (a *App) captureCompactAnswers() error {
	for _, e := range a.inputs {
		id := strings.TrimSpace(e.in.ID)
		val, err := entryValue(e)
		if err != nil {
			return err
		}
		if err := ValidateInputValue(e.in, val); err != nil {
			return err
		}
		a.answers[id] = val
	}
	return nil
}

func (a *App) captureCurrentSequential() error {
	if len(a.inputs) == 0 || a.activeInput >= len(a.inputs) {
		return nil
	}
	e := a.inputs[a.activeInput]
	id := strings.TrimSpace(e.in.ID)
	val, err := entryValue(e)
	if err != nil {
		return err
	}
	if err := ValidateInputValue(e.in, val); err != nil {
		return err
	}
	a.answers[id] = val
	return nil
}

func (a *App) buildContextFromInputs() (dsl.Context, error) {
	answers := map[string]string{}
	for k, v := range a.answers {
		answers[k] = fmt.Sprint(v)
	}
	return BuildContext(a.selected.Manifest.Inputs, answers)
}

func (a *App) buildRequest() *tmpl.ScaffoldRequest {
	return &tmpl.ScaffoldRequest{
		Template:  a.selected,
		OutputDir: a.outputDir(),
		Variables: a.answers,
		DryRun:    false,
	}
}

func (a *App) outputDir() string {
	name := strings.TrimSpace(ctxStringMap(a.answers, "project_name"))
	cwd, err := os.Getwd()
	if err != nil {
		return name
	}
	return filepath.Join(cwd, name)
}

func (a *App) templateName() string {
	if a.selected == nil || a.selected.Manifest == nil {
		return ""
	}
	return strings.TrimSpace(a.selected.Manifest.Name)
}

func (a *App) language() string {
	if a.selected == nil || a.selected.Manifest == nil {
		return ""
	}
	return strings.TrimSpace(a.selected.Manifest.Language)
}

func (a *App) stepLabel() string {
	switch a.state {
	case stateSelectTemplate:
		return ""
	case stateInputs:
		return "paso 2 de 4"
	case stateConfirm:
		return "paso 3 de 4"
	case stateProgress:
		return "paso 4 de 4"
	case stateDone:
		return "✓ listo"
	case stateError:
		return "error"
	default:
		return ""
	}
}

func (a *App) helpText() string {
	switch a.state {
	case stateSelectTemplate:
		return "↑/↓ navegar  / filtrar  enter seleccionar  ctrl+c salir"
	case stateInputs:
		return "tab/enter siguiente  esc volver  ctrl+c salir"
	case stateConfirm:
		return "enter confirmar  esc volver  ctrl+c salir"
	case stateProgress:
		return "(procesando...)"
	case stateDone, stateError:
		return "cualquier tecla para salir"
	default:
		return ""
	}
}

func (a *App) toError(err error) *App {
	a.err = err
	a.state = stateError
	return a
}

func (a *App) resizeComponents() {
	h := a.height - 4
	if h < 5 {
		h = 5
	}
	a.selector.SetSize(a.width-2, h)
	for i := range a.inputs {
		if a.inputs[i].kind == "enum" {
			a.inputs[i].enum.SetSize(a.width-8, min(max(4, h-6), 10))
		}
	}
}

func (a *App) currentInput() *inputEntry {
	if len(a.inputs) == 0 || a.activeInput < 0 || a.activeInput >= len(a.inputs) {
		return nil
	}
	return &a.inputs[a.activeInput]
}

func (a *App) inputCompleted(i int) bool {
	if i < 0 || i >= len(a.inputs) {
		return false
	}
	id := strings.TrimSpace(a.inputs[i].in.ID)
	_, ok := a.answers[id]
	return ok
}

func entryValue(e inputEntry) (string, error) {
	switch e.kind {
	case "string":
		v := strings.TrimSpace(e.ti.Value())
		if v == "" {
			v = strings.TrimSpace(e.ti.Placeholder)
		}
		return v, nil
	case "bool":
		if e.boolVal {
			return "true", nil
		}
		return "false", nil
	case "enum":
		it, ok := e.enum.SelectedItem().(enumItem)
		if !ok {
			return "", fmt.Errorf("enum value missing")
		}
		return it.value, nil
	default:
		return "", fmt.Errorf("unsupported input type")
	}
}

func startScaffoldCmd(req *tmpl.ScaffoldRequest, _ *engine.Engine) tea.Cmd {
	return func() tea.Msg {
		ch := make(chan tea.Msg, 16)
		go func() {
			defer close(ch)
			res, err := runScaffoldWithProgress(req, ch)
			if err != nil {
				ch <- msgScaffoldError{err: err}
				return
			}
			ch <- msgScaffoldDone{result: res}
		}()
		return msgProgressReady{ch: ch}
	}
}

func waitProgressMsg(ch <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return msgProgressClosed{}
		}
		return msg
	}
}

type observer struct{ ch chan<- tea.Msg }

func (o observer) OnStepStart(step dsl.Step, cmd string) {
	o.ch <- msgStepStart{name: step.Name, command: cmd}
}
func (o observer) OnStepSkipped(step dsl.Step) {
	o.ch <- msgStepDone{name: step.Name, skipped: true}
}
func (o observer) OnStepSuccess(step dsl.Step, _ string) {
	o.ch <- msgStepDone{name: step.Name}
}
func (o observer) OnStepFailure(step dsl.Step, err error, _ string) {
	o.ch <- msgStepError{name: step.Name, err: err}
}

func runScaffoldWithProgress(req *tmpl.ScaffoldRequest, ch chan<- tea.Msg) (*tmpl.ScaffoldResult, error) {
	if req == nil || req.Template == nil || req.Template.Manifest == nil {
		return nil, fmt.Errorf("invalid scaffold request")
	}
	outAbs, err := filepath.Abs(req.OutputDir)
	if err != nil {
		return nil, fmt.Errorf("abs outputDir: %w", err)
	}
	req.OutputDir = outAbs

	exists, empty, err := dirExistsAndEmpty(req.OutputDir)
	if err != nil {
		return nil, fmt.Errorf("checking outputDir: %w", err)
	}
	if exists && !empty {
		return nil, fmt.Errorf("output directory %s already exists and is not empty", req.OutputDir)
	}

	rb := engine.NewRollbackManager(req.DryRun)
	if !req.DryRun {
		if err := os.MkdirAll(req.OutputDir, 0o755); err != nil {
			return nil, fmt.Errorf("creating outputDir %s: %w", req.OutputDir, err)
		}
		rb.TrackDir(req.OutputDir)
	}

	created, skipped, err := engine.ProcessFiles(req)
	if err != nil {
		_ = rb.Rollback()
		return nil, err
	}
	ch <- msgFilesDone{count: len(created)}

	obs := observer{ch: ch}
	steps, err := engine.ExecuteStepsWithObserver(req.Template.Manifest.Steps, req.Variables, req.OutputDir, req.DryRun, obs)
	if err != nil {
		_ = rb.Rollback()
		return &tmpl.ScaffoldResult{
			FilesCreated:  created,
			FilesSkipped:  skipped,
			StepsExecuted: steps,
			StepsFailed:   failedSteps(steps),
		}, err
	}
	rb.Commit()
	return &tmpl.ScaffoldResult{
		FilesCreated:  created,
		FilesSkipped:  skipped,
		StepsExecuted: steps,
	}, nil
}

func ctxStringMap(ctx map[string]any, key string) string {
	v, ok := ctx[key]
	if !ok || v == nil {
		return ""
	}
	return fmt.Sprint(v)
}

func prettyPath(path string) string {
	cwd, err := os.Getwd()
	if err != nil {
		return path
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	rel, err := filepath.Rel(cwd, abs)
	if err != nil {
		return path
	}
	if rel == "." {
		return "."
	}
	if !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) && rel != ".." {
		return "." + string(os.PathSeparator) + rel
	}
	return path
}

type kvPair struct {
	key   string
	value string
}

func sortedContextPairs(ctx dsl.Context) []kvPair {
	keys := make([]string, 0, len(ctx))
	for k := range ctx {
		keys = append(keys, k)
	}
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[j] < keys[i] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	out := make([]kvPair, 0, len(keys))
	for _, k := range keys {
		out = append(out, kvPair{key: k, value: fmt.Sprint(ctx[k])})
	}
	return out
}

func padRight(s string, w int) string {
	if len(s) >= w {
		return s
	}
	return s + strings.Repeat(" ", w-len(s))
}

func answerString(ctx dsl.Context, key string) string {
	v, ok := ctx[key]
	if !ok || v == nil {
		return ""
	}
	return fmt.Sprint(v)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
