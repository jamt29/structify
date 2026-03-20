package tui

import (
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
