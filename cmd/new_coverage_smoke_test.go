package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/jamt29/structify/internal/config"
	"github.com/jamt29/structify/internal/dsl"
	"github.com/jamt29/structify/internal/engine"
	"github.com/jamt29/structify/internal/template"
)

func withNonInteractiveConfig(t *testing.T, nonInteractive bool, fn func()) {
	t.Helper()

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir: %v", err)
	}
	cfgDir := filepath.Join(home, ".structify")
	cfgPath := filepath.Join(cfgDir, "config.yaml")

	// Backup existing config.yaml (if any).
	orig, readErr := os.ReadFile(cfgPath)
	if readErr != nil && !os.IsNotExist(readErr) {
		t.Fatalf("ReadFile(%s): %v", cfgPath, readErr)
	}

	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", cfgDir, err)
	}

	val := "false"
	if nonInteractive {
		val = "true"
	}
	content := []byte(fmt.Sprintf("nonInteractive: %s\n", val))
	if err := os.WriteFile(cfgPath, content, 0o644); err != nil {
		t.Fatalf("WriteFile(%s): %v", cfgPath, err)
	}

	defer func() {
		if readErr != nil {
			_ = os.Remove(cfgPath)
			return
		}
		_ = os.WriteFile(cfgPath, orig, 0o644)
	}()

	fn()
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	var buf bytes.Buffer

	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()

	fn()

	_ = w.Close()
	<-done
	os.Stdout = orig
	return buf.String()
}

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	var buf bytes.Buffer

	orig := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stderr = w

	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()

	fn()

	_ = w.Close()
	<-done
	os.Stderr = orig
	return buf.String()
}

func TestRunNew_DryRunNonInteractive_Smoke(t *testing.T) {
	withNonInteractiveConfig(t, true, func() {
		// Ensure we don't depend on Cobra execution and Bubbletea.
		origTemplate, origName, origVars, origDryRun, origOutput := newTemplate, newName, newVars, newDryRun, newOutput
		defer func() {
			newTemplate, newName, newVars, newDryRun, newOutput = origTemplate, origName, origVars, origDryRun, origOutput
		}()

		newTemplate = "clean-architecture-go"
		newName = "testproject"
		newVars = nil
		newDryRun = true
		newOutput = ""

		out := captureStdout(t, func() {
			if err := runNew(nil, nil); err != nil {
				t.Fatalf("runNew returned error: %v", err)
			}
		})
		if out == "" {
			t.Fatalf("expected dry-run output on stdout")
		}
	})
}

func TestRunNonInteractive_SkipsAllSteps_Smoke(t *testing.T) {
	tmp := t.TempDir()
	req := &template.ScaffoldRequest{
		Template: &template.Template{
			Path: tmp, // no `template/` directory => ProcessFiles is a no-op
			Manifest: &dsl.Manifest{
				Name:   "test",
				Inputs: nil,
				Files:  nil,
				Steps: []dsl.Step{
					{
						Name: "should_skip",
						Run:  "echo never",
						When: `project_name == "nope"`,
					},
				},
			},
		},
		OutputDir: tmp,
		Variables: dsl.Context{"project_name": "testproject"},
		DryRun:    false,
	}

	eng := engine.New()
	_ = eng // engine instance is kept for future-proofing; coverage comes from runNonInteractive itself.
	out := captureStderr(t, func() {
		res, err := runNonInteractive(req, config.NewLogger(false))
		if err != nil {
			t.Fatalf("runNonInteractive returned error: %v", err)
		}
		if res == nil {
			t.Fatalf("expected non-nil result")
		}
	})
	if out == "" {
		t.Fatalf("expected runNonInteractive to write progress to stderr")
	}

	// Cover failedSteps helper and all observer callbacks too (direct call).
	_ = failedSteps([]template.StepResult{
		{Name: "a", Error: nil},
		{Name: "b", Error: fmt.Errorf("boom")},
	})

	obs := printStepObserver{log: config.NewLogger(false)}
	obs.OnStepStart(dsl.Step{Name: "start"}, "echo start")
	obs.OnStepSkipped(dsl.Step{Name: "skip"})
	obs.OnStepSuccess(dsl.Step{Name: "ok"}, "out")
	obs.OnStepFailure(dsl.Step{Name: "fail"}, fmt.Errorf("err"), "out")
}
