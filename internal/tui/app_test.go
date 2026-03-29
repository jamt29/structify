package tui

import (
	"strings"
	"testing"

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

	app.huhString["project_name"] = "my-api"
	ctx, err := app.buildContextFromHuh()
	if err != nil {
		t.Fatalf("buildContextFromHuh error: %v", err)
	}
	app.answers = ctx
	app.state = stateConfirm

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

func TestStateDone_KeyQuitsOnlyWhenTopLevel(t *testing.T) {
	tpl := &template.Template{
		Manifest: &dsl.Manifest{Name: "t"},
	}
	app, err := newApp([]*template.Template{tpl}, engine.New())
	if err != nil {
		t.Fatalf("newApp error: %v", err)
	}
	app.state = stateDone

	// Default path (embedded under RootModel): mark done, don't quit.
	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !app.done {
		t.Fatalf("expected app.done=true")
	}
	if cmd != nil {
		t.Fatalf("expected nil cmd when quitOnDoneKey=false")
	}

	// Top-level RunApp path: mark done and quit.
	app2, err := newApp([]*template.Template{tpl}, engine.New())
	if err != nil {
		t.Fatalf("newApp error: %v", err)
	}
	app2.state = stateDone
	app2.quitOnDoneKey = true
	_, cmd2 := app2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !app2.done {
		t.Fatalf("expected app2.done=true")
	}
	if cmd2 == nil {
		t.Fatalf("expected tea.Quit cmd when quitOnDoneKey=true")
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
	app.width = 120
	app.height = 40
	app.huhString["project_name"] = "my-api"

	// Render inputs and confirm.
	_ = app.renderInputs()
	ctx, err := app.buildContextFromHuh()
	if err != nil {
		t.Fatalf("buildContextFromHuh error: %v", err)
	}
	app.answers = ctx
	app.state = stateConfirm
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

func TestBuildContextFromHuh_UsesDefaultsWhenEmptyStringAnswers(t *testing.T) {
	tpl := &template.Template{
		Manifest: &dsl.Manifest{
			Name: "clean-structure-go",
			Inputs: []dsl.Input{
				{ID: "project_name", Type: "string", Required: true},
				{ID: "module_path", Type: "string", Default: "github.com/user/{{ project_name | kebab_case }}"},
				{ID: "http_framework", Type: "enum", Options: []string{"gin", "fiber"}, Default: "gin"},
				{ID: "sql_database", Type: "enum", Options: []string{"none", "postgres"}, Default: "postgres"},
			},
		},
	}
	app, err := newApp([]*template.Template{tpl}, engine.New())
	if err != nil {
		t.Fatalf("newApp error: %v", err)
	}
	app.selected = tpl
	app.huhString = map[string]string{
		"project_name":   "prueba",
		"module_path":    "",
		"http_framework": "",
		"sql_database":   "",
	}

	ctx, err := app.buildContextFromHuh()
	if err != nil {
		t.Fatalf("buildContextFromHuh error: %v", err)
	}
	if got := strings.TrimSpace(ctxStringMap(ctx, "module_path")); got != "github.com/user/prueba" {
		t.Fatalf("expected module_path default interpolation, got %q", got)
	}
	if got := strings.TrimSpace(ctxStringMap(ctx, "http_framework")); got != "gin" {
		t.Fatalf("expected enum default for http_framework, got %q", got)
	}
	if got := strings.TrimSpace(ctxStringMap(ctx, "sql_database")); got != "postgres" {
		t.Fatalf("expected enum default for sql_database, got %q", got)
	}
}

func TestUpdateInputs_EnterDoesNotSkipHuhForm(t *testing.T) {
	tpl := &template.Template{
		Manifest: &dsl.Manifest{
			Name: "clean-structure-go",
			Inputs: []dsl.Input{
				{ID: "project_name", Type: "string", Required: true},
				{ID: "module_path", Type: "string", Default: "github.com/user/{{ project_name | kebab_case }}"},
				{ID: "http_framework", Type: "enum", Options: []string{"gin", "fiber"}, Default: "gin"},
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
	if app.huhForm == nil {
		t.Fatalf("expected huh form to be initialized")
	}

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if app.state != stateInputs {
		t.Fatalf("enter should not skip to confirm while huh form is not completed")
	}
}
