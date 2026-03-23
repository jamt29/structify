package tui

import (
	"testing"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jamt29/structify/internal/dsl"
	"github.com/jamt29/structify/internal/engine"
	"github.com/jamt29/structify/internal/template"
)

func TestStateTransition_SelectToInputs(t *testing.T) {
	tpl := &template.Template{
		Manifest: &dsl.Manifest{
			Name: "clean-architecture-go",
			Inputs: []dsl.Input{
				{ID: "project_name", Type: "string", Prompt: "Project name?", Required: true, Validate: `^[a-zA-Z][a-zA-Z0-9_-]*$`},
			},
		},
	}
	app, err := newApp([]*template.Template{tpl}, engine.New())
	if err != nil {
		t.Fatalf("newApp error: %v", err)
	}
	if app.state != stateSelectTemplate {
		t.Fatalf("expected select state, got %v", app.state)
	}

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if app.state != stateInputs {
		t.Fatalf("expected inputs state, got %v", app.state)
	}
	if app.selected == nil || app.selected.Manifest == nil || app.selected.Manifest.Name != "clean-architecture-go" {
		t.Fatalf("expected selected template to be set")
	}
}

func TestStateTransition_InputsToDone(t *testing.T) {
	tpl := &template.Template{
		Manifest: &dsl.Manifest{
			Name: "clean-architecture-go",
			Inputs: []dsl.Input{
				{ID: "project_name", Type: "string", Prompt: "Project name?", Required: true, Validate: `^[a-zA-Z][a-zA-Z0-9_-]*$`},
			},
		},
	}
	app, err := newApp([]*template.Template{tpl}, engine.New())
	if err != nil {
		t.Fatalf("newApp error: %v", err)
	}
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnter}) // select -> inputs
	if app.state != stateInputs {
		t.Fatalf("expected inputs state, got %v", app.state)
	}

	if len(app.inputs) == 0 {
		t.Fatalf("expected at least one input")
	}
	app.inputs[0].ti.SetValue("my-api")

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnter}) // inputs -> confirm
	if app.state != stateConfirm {
		t.Fatalf("expected confirm state, got %v", app.state)
	}

	_, _ = app.Update(msgScaffoldDone{result: &template.ScaffoldResult{}})
	if app.state != stateConfirm {
		t.Fatalf("scaffold done should not change confirm state")
	}

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnter}) // confirm -> progress
	if app.state != stateProgress {
		t.Fatalf("expected progress state, got %v", app.state)
	}

	_, _ = app.Update(msgScaffoldDone{result: &template.ScaffoldResult{}})
	if app.state != stateDone {
		t.Fatalf("expected done state, got %v", app.state)
	}
}

func TestApp_RenderAndHelpers(t *testing.T) {
	tpl := &template.Template{
		Manifest: &dsl.Manifest{
			Name:     "clean-architecture-go",
			Language: "go",
			Inputs: []dsl.Input{
				{ID: "project_name", Type: "string", Prompt: "Project name?", Required: true, Default: "my-api"},
				{ID: "use_prisma", Type: "bool", Prompt: "Include Prisma?", Default: false},
				{ID: "runtime", Type: "enum", Prompt: "Runtime?", Options: []string{"express", "fastify"}, Default: "express"},
			},
			Steps: []dsl.Step{
				{Name: "Init go module", Run: "go mod init github.com/user/my-api"},
			},
		},
	}
	app, err := newApp([]*template.Template{tpl}, engine.New())
	if err != nil {
		t.Fatalf("newApp error: %v", err)
	}

	// Select template and prepare inputs.
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if len(app.inputs) == 0 {
		t.Fatalf("expected inputs")
	}
	app.width = 120
	app.height = 40
	app.inputs[0].ti.SetValue("my-api")

	// Render inputs and confirm.
	_ = app.renderInputs()
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if app.state != stateConfirm {
		t.Fatalf("expected confirm state")
	}
	_ = app.renderConfirm()

	// Header / step labels / help text.
	_ = app.renderHeader()
	if app.stepLabel() == "" {
		t.Fatalf("expected step label in confirm")
	}
	if app.helpText() == "" {
		t.Fatalf("expected help text")
	}

	// Progress rendering with command lines.
	app.state = stateProgress
	app.progressLog = []progressLine{
		{name: "files", status: "done", command: "Archivos generados (11 archivos)"},
		{name: "Init go module", status: "running", command: "go mod init github.com/user/my-api"},
	}
	_ = app.renderProgress()
	_, _ = app.Update(msgStepDone{name: "Init go module"})
	_ = app.renderProgress()

	// Done rendering and utility helpers.
	app.state = stateDone
	app.result = &template.ScaffoldResult{
		FilesCreated: []string{"a", "b"},
		StepsExecuted: []template.StepResult{
			{Name: "Init go module", Command: "go mod init github.com/user/my-api"},
		},
	}

	doneBody := app.renderDone()
	if strings.Contains(doneBody, "(presiona cualquier tecla para salir)") {
		t.Fatalf("renderDone() must not include inline exit text")
	}

	view := app.ViewContent()
	if got := strings.Count(view, "cualquier tecla para salir"); got != 1 {
		t.Fatalf("expected exactly one helpText occurrence, got %d", got)
	}

	_ = app.View()
	_ = prettyPath(app.outputDir())
	_ = sortedContextPairs(app.answers)
	_ = padRight("x", 4)
	_ = max(1, 2)
}
