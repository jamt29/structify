package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"

	"github.com/jamt29/structify/internal/dsl"
	"github.com/jamt29/structify/internal/engine"
	"github.com/jamt29/structify/internal/template"
)

type progressPhase string

const (
	phaseFiles progressPhase = "files"
	phaseSteps progressPhase = "steps"
	phaseDone  progressPhase = "done"
)

type progressMsgFilesDone struct {
	created []string
	skipped []string
}

type progressMsgStepStart struct {
	name string
}

type progressMsgStepSkipped struct {
	name string
}

type progressMsgStepSuccess struct {
	name string
}

type progressMsgStepFailure struct {
	name string
	err  error
}

type progressMsgDone struct {
	result *template.ScaffoldResult
	err    error
}

type progressModel struct {
	spin   spinner.Model
	phase  progressPhase
	status string

	steps []string
	done  map[string]string // name -> "ok"|"skipped"|"fail"
	fail  error

	req    *template.ScaffoldRequest
	result *template.ScaffoldResult
}

func (m progressModel) Init() tea.Cmd {
	return m.spin.Tick
}

func (m progressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		return m, cmd

	case progressMsgFilesDone:
		m.phase = phaseSteps
		m.status = fmt.Sprintf("Files generated (%d created, %d skipped). Running steps...", len(msg.created), len(msg.skipped))
		return m, nil

	case progressMsgStepStart:
		m.status = msg.name
		return m, nil

	case progressMsgStepSkipped:
		m.done[msg.name] = "skipped"
		return m, nil

	case progressMsgStepSuccess:
		m.done[msg.name] = "ok"
		return m, nil

	case progressMsgStepFailure:
		m.done[msg.name] = "fail"
		m.fail = msg.err
		return m, tea.Quit

	case progressMsgDone:
		m.phase = phaseDone
		m.result = msg.result
		m.fail = msg.err
		return m, tea.Quit

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.fail = fmt.Errorf("cancelled")
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m progressModel) View() string {
	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Render("Creating project")
	b.WriteString(title)
	b.WriteString("\n\n")

	if m.phase == phaseFiles {
		b.WriteString(m.spin.View())
		b.WriteString(" ")
		b.WriteString("Generating files...")
		b.WriteString("\n")
		return b.String()
	}

	if m.phase == phaseSteps {
		b.WriteString(m.spin.View())
		b.WriteString(" ")
		b.WriteString(strings.TrimSpace(m.status))
		b.WriteString("\n\n")

		for _, s := range m.steps {
			state, ok := m.done[s]
			if !ok {
				continue
			}
			switch state {
			case "ok":
				b.WriteString("✓ ")
			case "skipped":
				b.WriteString("─ ")
			case "fail":
				b.WriteString("✗ ")
			}
			b.WriteString(s)
			if state == "skipped" {
				b.WriteString(" (skipped)")
			}
			b.WriteString("\n")
		}
		return b.String()
	}

	if m.fail != nil {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("✗ "))
		b.WriteString(m.fail.Error())
		b.WriteString("\n")
		return b.String()
	}

	b.WriteString("✓ Done\n")
	return b.String()
}

type teaStepObserver struct {
	send func(tea.Msg)
}

func (o teaStepObserver) safeSend(msg tea.Msg) {
	if o.send == nil {
		return
	}
	defer func() { _ = recover() }()
	o.send(msg)
}

func (o teaStepObserver) OnStepStart(step dsl.Step, _ string) {
	o.safeSend(progressMsgStepStart{name: step.Name})
}
func (o teaStepObserver) OnStepSkipped(step dsl.Step) {
	o.safeSend(progressMsgStepSkipped{name: step.Name})
}
func (o teaStepObserver) OnStepSuccess(step dsl.Step, _ string) {
	o.safeSend(progressMsgStepSuccess{name: step.Name})
}
func (o teaStepObserver) OnStepFailure(step dsl.Step, err error, _ string) {
	o.safeSend(progressMsgStepFailure{name: step.Name, err: err})
}

// RunProgress executes scaffolding while rendering a progress TUI.
func RunProgress(req *template.ScaffoldRequest, eng *engine.Engine) (*template.ScaffoldResult, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}
	if eng == nil {
		return nil, fmt.Errorf("engine is nil")
	}
	if req.Template == nil || req.Template.Manifest == nil {
		return nil, fmt.Errorf("template is nil")
	}
	if req.Variables == nil {
		return nil, fmt.Errorf("variables is nil")
	}

	outAbs, err := filepath.Abs(req.OutputDir)
	if err != nil {
		return nil, fmt.Errorf("abs outputDir: %w", err)
	}
	req.OutputDir = outAbs

	spin := spinner.New()
	spin.Spinner = spinner.Line
	spin.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))

	stepNames := make([]string, 0, len(req.Template.Manifest.Steps))
	for _, s := range req.Template.Manifest.Steps {
		if strings.TrimSpace(s.Name) != "" {
			stepNames = append(stepNames, s.Name)
		}
	}

	m := progressModel{
		spin:   spin,
		phase: phaseFiles,
		done:  map[string]string{},
		steps: stepNames,
		req:   req,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())

	go func() {
		res, runErr := runScaffoldWithObserver(req, eng, func(msg tea.Msg) {
			p.Send(msg)
		})
		p.Send(progressMsgDone{result: res, err: runErr})
	}()

	final, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("running progress ui: %w", err)
	}
	fm, ok := final.(progressModel)
	if !ok {
		return nil, fmt.Errorf("internal error: unexpected progress model")
	}
	if fm.fail != nil {
		return fm.result, fm.fail
	}
	return fm.result, nil
}

func runScaffoldWithObserver(req *template.ScaffoldRequest, _ *engine.Engine, send func(tea.Msg)) (*template.ScaffoldResult, error) {
	// Mirror engine.Engine.Scaffold behavior, but with per-step progress.
	if strings.TrimSpace(req.OutputDir) == "" {
		return nil, fmt.Errorf("outputDir is empty")
	}
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
	if send != nil {
		send(progressMsgFilesDone{created: created, skipped: skipped})
	}

	obs := teaStepObserver{send: send}
	stepResults, err := engine.ExecuteStepsWithObserver(req.Template.Manifest.Steps, req.Variables, req.OutputDir, req.DryRun, obs)
	if err != nil {
		_ = rb.Rollback()
		return &template.ScaffoldResult{
			FilesCreated:  created,
			FilesSkipped:  skipped,
			StepsExecuted: stepResults,
			StepsFailed:   failedSteps(stepResults),
		}, err
	}

	rb.Commit()
	return &template.ScaffoldResult{
		FilesCreated:  created,
		FilesSkipped:  skipped,
		StepsExecuted: stepResults,
	}, nil
}

func dirExistsAndEmpty(path string) (exists bool, empty bool, err error) {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, true, nil
		}
		return false, false, err
	}
	if !fi.IsDir() {
		return true, false, fmt.Errorf("%s exists and is not a directory", path)
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return true, false, err
	}
	return true, len(entries) == 0, nil
}

func failedSteps(all []template.StepResult) []template.StepResult {
	var failed []template.StepResult
	for _, r := range all {
		if r.Error != nil {
			failed = append(failed, r)
		}
	}
	return failed
}

