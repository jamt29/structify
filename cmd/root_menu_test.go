package cmd

import (
	"errors"
	"testing"

	"github.com/jamt29/structify/internal/config"
	"github.com/jamt29/structify/internal/dsl"
	"github.com/jamt29/structify/internal/engine"
	"github.com/jamt29/structify/internal/template"
	"github.com/jamt29/structify/internal/tui"
)

func TestRunInteractive_MenuExitIsClean(t *testing.T) {
	origRunMenu := runMenuFn
	defer func() { runMenuFn = origRunMenu }()
	runMenuFn = func() (tui.MenuAction, error) { return tui.ActionNew, tui.ErrMenuExit }
	if err := runInteractive(); err != nil {
		t.Fatalf("expected nil err on menu exit, got %v", err)
	}
}

func TestRunInteractive_ActionNew(t *testing.T) {
	origRunMenu, origResolve, origRunApp := runMenuFn, resolveAllFn, runAppFn
	defer func() {
		runMenuFn, resolveAllFn, runAppFn = origRunMenu, origResolve, origRunApp
	}()

	runMenuFn = func() (tui.MenuAction, error) { return tui.ActionNew, nil }
	resolveAllFn = func() ([]*template.Template, error) {
		return []*template.Template{}, nil
	}
	called := false
	runAppFn = func(_ []*template.Template, _ *engine.Engine) error {
		called = true
		return nil
	}

	if err := runInteractive(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatalf("expected runAppFn to be called")
	}
}

func TestRunInteractive_ActionTemplatesAndConfig(t *testing.T) {
	origRunMenu, origList, origLoad := runMenuFn, runTemplateListFn, loadConfigFn
	defer func() {
		runMenuFn, runTemplateListFn, loadConfigFn = origRunMenu, origList, origLoad
	}()

	runMenuFn = func() (tui.MenuAction, error) { return tui.ActionTemplates, nil }
	calledList := false
	runTemplateListFn = func() error {
		calledList = true
		return nil
	}
	if err := runInteractive(); err != nil {
		t.Fatalf("templates action err: %v", err)
	}
	if !calledList {
		t.Fatalf("expected runTemplateListFn called")
	}

	runMenuFn = func() (tui.MenuAction, error) { return tui.ActionConfig, nil }
	loadConfigFn = func() (config.Config, error) {
		return config.Config{ConfigDir: "/tmp/.structify", ConfigFile: ""}, nil
	}
	if err := runInteractive(); err != nil {
		t.Fatalf("config action err: %v", err)
	}

	runMenuFn = func() (tui.MenuAction, error) { return tui.ActionGitHub, nil }
	if err := runInteractive(); err != nil {
		t.Fatalf("github action err: %v", err)
	}
}

func TestRunInteractive_ConfigLoadError(t *testing.T) {
	origRunMenu, origLoad := runMenuFn, loadConfigFn
	defer func() { runMenuFn, loadConfigFn = origRunMenu, origLoad }()
	runMenuFn = func() (tui.MenuAction, error) { return tui.ActionConfig, nil }
	loadConfigFn = func() (config.Config, error) { return config.Config{}, errors.New("load failed") }
	if err := runInteractive(); err == nil {
		t.Fatalf("expected error")
	}
}

func TestRunInteractive_ErrorPath(t *testing.T) {
	origRunMenu := runMenuFn
	defer func() { runMenuFn = origRunMenu }()
	runMenuFn = func() (tui.MenuAction, error) { return tui.ActionNew, errors.New("boom") }
	if err := runInteractive(); err == nil {
		t.Fatalf("expected error")
	}
}

func TestRunTemplateList_Smoke(t *testing.T) {
	origResolve := resolveAllFn
	defer func() { resolveAllFn = origResolve }()

	resolveAllFn = func() ([]*template.Template, error) {
		return []*template.Template{
			{Source: "builtin", Manifest: &dsl.Manifest{Name: "clean-go"}},
			{Source: "local", Manifest: &dsl.Manifest{Name: "my-template"}},
		}, nil
	}
	if err := runTemplateList(); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestRunTemplateList_EmptyAndResolve(t *testing.T) {
	origResolve := resolveAllFn
	defer func() { resolveAllFn = origResolve }()

	resolveAllFn = func() ([]*template.Template, error) { return []*template.Template{}, nil }
	if err := runTemplateList(); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if _, err := resolveAllTemplates(); err != nil {
		t.Fatalf("resolveAllTemplates should work with current fixtures: %v", err)
	}
}
