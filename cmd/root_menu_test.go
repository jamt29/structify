package cmd

import (
	"errors"
	"testing"

	"github.com/jamt29/structify/internal/engine"
	"github.com/jamt29/structify/internal/tui"

	templ "github.com/jamt29/structify/internal/template"
)

func TestRunInteractive_CallsTUI(t *testing.T) {
	orig := runRootFn
	defer func() { runRootFn = orig }()

	called := false
	runRootFn = func(templates []*templ.Template, eng *engine.Engine) error {
		called = true
		if eng == nil {
			t.Fatalf("engine should not be nil")
		}
		if templates == nil {
			t.Fatalf("templates should not be nil")
		}
		return nil
	}

	if err := runInteractive(); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !called {
		t.Fatalf("expected runRootFn to be called")
	}
}

func TestRunInteractive_ErrorPath(t *testing.T) {
	orig := runRootFn
	defer func() { runRootFn = orig }()

	runRootFn = func(_ []*templ.Template, _ *engine.Engine) error {
		return errors.New("boom")
	}

	if err := runInteractive(); err == nil {
		t.Fatalf("expected error")
	}
}

// Ensure these exported symbols keep compiling after menu/root refactors.
func TestMenuExportsSmoke(t *testing.T) {
	_ = tui.ActionNew
	_ = tui.ActionTemplates
	_ = tui.ActionGitHub
	_ = tui.ActionConfig
}
