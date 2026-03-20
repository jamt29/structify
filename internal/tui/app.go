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
	in   dsl.Input
	kind string
	ti   textinput.Model
	enum list.Model
}

type App struct {
	state state

	templates []*tmpl.Template
	selected  *tmpl.Template
	answers   dsl.Context
	result    *tmpl.ScaffoldResult
	err       error

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
		if _, ok := msg.(tea.KeyMsg); ok {
			return a, tea.Quit
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
			if a.compactForm && len(a.inputs) > 0 {
				a.activeInput = (a.activeInput + 1) % len(a.inputs)
				return a, nil
			}
		case "shift+tab":
			if a.compactForm && len(a.inputs) > 0 {
				a.activeInput--
				if a.activeInput < 0 {
					a.activeInput = len(a.inputs) - 1
				}
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
	case "string", "bool":
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
		return a, nil
	case msgScaffoldDone:
		a.result = m.result
		a.state = stateDone
		return a, nil
	case msgScaffoldError:
		a.err = m.err
		a.state = stateError
		return a, nil
	case msgProgressClosed:
		return a, nil
	case msgFilesDone:
		return a, waitProgressMsg(a.progressCh)
	}
	return a, nil
}

func (a *App) View() string {
	if a.width < 80 || a.height < 24 {
		return "Terminal too small. Minimum 80x24."
	}
	parts := []string{
		lipgloss.NewStyle().Bold(true).Render(twoCols(a.width, " structify · "+a.templateName(), " "+a.stepLabel()+" ")),
		a.renderBody(),
		lipgloss.NewStyle().Faint(true).Render(" " + a.helpText()),
	}
	return strings.Join(parts, "\n")
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
		return "No hay inputs activos. Presiona Enter para continuar."
	}
	var b strings.Builder
	b.WriteString("\n")
	if a.compactForm {
		for i, e := range a.inputs {
			mark := "  "
			if i == a.activeInput {
				mark = "> "
			}
			b.WriteString(mark + strings.TrimSpace(e.in.Prompt) + "\n")
			switch e.kind {
			case "string", "bool":
				b.WriteString("  " + e.ti.View() + "\n\n")
			case "enum":
				b.WriteString(e.enum.View() + "\n\n")
			}
		}
		return b.String()
	}
	idx := a.activeInput
	if idx >= len(a.inputs) {
		idx = len(a.inputs) - 1
	}
	e := a.inputs[idx]
	b.WriteString(fmt.Sprintf("Input %d de %d\n\n", idx+1, len(a.inputs)))
	b.WriteString(strings.TrimSpace(e.in.Prompt) + "\n")
	if e.kind == "enum" {
		b.WriteString(e.enum.View())
	} else {
		b.WriteString(e.ti.View())
	}
	return b.String()
}

func (a *App) renderConfirm() string {
	var b strings.Builder
	b.WriteString("\nSe va a crear:\n\n")
	b.WriteString("Template: " + a.templateName() + "\n")
	b.WriteString("Output  : " + prettyPath(a.outputDir()) + "\n")
	b.WriteString("Variables:\n")
	for k, v := range a.answers {
		b.WriteString(fmt.Sprintf("  - %s=%v\n", k, v))
	}
	return b.String()
}

func (a *App) renderProgress() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(a.spin.View() + " Creando " + ctxStringMap(a.answers, "project_name") + "...\n\n")
	for _, line := range a.progressLog {
		switch line.status {
		case "running":
			b.WriteString(a.spin.View() + " " + line.command + "\n")
		case "done":
			b.WriteString("✓ " + line.command + "\n")
		case "skipped":
			b.WriteString("─ " + line.command + " (skipped)\n")
		case "error":
			b.WriteString("✗ " + line.command + "\n")
		}
	}
	return b.String()
}

func (a *App) renderDone() string {
	var b strings.Builder
	b.WriteString("\n✓ Proyecto creado exitosamente\n\n")
	if a.result != nil {
		b.WriteString("Ruta   : " + a.outputDir() + "\n")
		b.WriteString(fmt.Sprintf("Archivos: %d\n\n", len(a.result.FilesCreated)))
		b.WriteString("Steps ejecutados:\n")
		for _, s := range a.result.StepsExecuted {
			if s.Skipped {
				b.WriteString("  ─ " + s.Command + " (skipped)\n")
			} else if s.Error != nil {
				b.WriteString("  ✗ " + s.Command + "\n")
			} else {
				b.WriteString("  ✓ " + s.Command + "\n")
			}
		}
	}
	b.WriteString("\nPróximos pasos:\n")
	for _, line := range nextSteps(a.language(), ctxStringMap(a.answers, "project_name")) {
		b.WriteString("  " + line + "\n")
	}
	return b.String()
}

func (a *App) renderError() string {
	msg := "unknown error"
	if a.err != nil {
		msg = a.err.Error()
	}
	return "\n✗ " + msg + "\n"
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
			entry.ti.Prompt = "> "
			def, _ := ApplyDefault(in, a.answers)
			entry.ti.Placeholder = def
			entry.ti.SetValue(fmt.Sprint(a.answers[id]))
			entry.ti.Focus()
		case "bool":
			entry.ti = textinput.New()
			entry.ti.Prompt = "> "
			def, _ := ApplyDefault(in, a.answers)
			if strings.TrimSpace(def) == "" {
				def = "n"
			}
			entry.ti.Placeholder = def
			entry.ti.SetValue(fmt.Sprint(a.answers[id]))
			entry.ti.Focus()
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
			cur := strings.TrimSpace(fmt.Sprint(a.answers[id]))
			if cur == "" {
				def, _ := ApplyDefault(in, a.answers)
				cur = def
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
		return "paso 1 de 4"
	case stateInputs:
		return "paso 2 de 4"
	case stateConfirm:
		return "paso 3 de 4"
	case stateProgress:
		return "paso 3 de 4"
	case stateDone, stateError:
		return "paso 4 de 4"
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
			a.inputs[i].enum.SetSize(a.width-4, min(max(4, h-4), 12))
		}
	}
}

func entryValue(e inputEntry) (string, error) {
	switch e.kind {
	case "string", "bool":
		v := strings.TrimSpace(e.ti.Value())
		if v == "" {
			v = strings.TrimSpace(e.ti.Placeholder)
		}
		return v, nil
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

func twoCols(width int, left, right string) string {
	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)
	total := len(left) + len(right)
	if total+1 >= width {
		return left + " " + right
	}
	return left + strings.Repeat(" ", width-total) + right
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
