package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jamt29/structify/internal/dsl"
)

func TestExecuteSteps_DryRunDoesNotExecute(t *testing.T) {
	out := t.TempDir()

	steps := []dsl.Step{
		{Name: "write", Run: "echo hi > ran.txt"},
	}
	res, err := ExecuteSteps(steps, dsl.Context{}, out, true)
	if err != nil {
		t.Fatalf("ExecuteSteps() error: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("results len=%d, want 1", len(res))
	}
	if res[0].Command == "" {
		t.Fatalf("expected interpolated command to be recorded")
	}
	if _, err := os.Stat(filepath.Join(out, "ran.txt")); err == nil {
		t.Fatalf("expected ran.txt not to exist in dry run")
	}
}

func TestExecuteSteps_WhenFalseSkipped(t *testing.T) {
	out := t.TempDir()

	steps := []dsl.Step{
		{Name: "skip", Run: "echo hi > ran.txt", When: "flag == true"},
	}
	res, err := ExecuteSteps(steps, dsl.Context{"flag": false}, out, false)
	if err != nil {
		t.Fatalf("ExecuteSteps() error: %v", err)
	}
	if len(res) != 1 || !res[0].Skipped {
		t.Fatalf("expected step skipped, got %+v", res)
	}
	if _, err := os.Stat(filepath.Join(out, "ran.txt")); err == nil {
		t.Fatalf("expected ran.txt not to exist for skipped step")
	}
}

func TestExecuteSteps_FailureStopsAndReturnsError(t *testing.T) {
	out := t.TempDir()

	steps := []dsl.Step{
		{Name: "ok", Run: "echo hi > ok.txt"},
		{Name: "fail", Run: "false"},
		{Name: "after", Run: "echo nope > after.txt"},
	}
	res, err := ExecuteSteps(steps, dsl.Context{}, out, false)
	if err == nil {
		t.Fatalf("expected error")
	}
	if len(res) < 2 {
		t.Fatalf("expected partial results, got %d", len(res))
	}
	if _, err := os.Stat(filepath.Join(out, "ok.txt")); err != nil {
		t.Fatalf("expected ok.txt to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "after.txt")); err == nil {
		t.Fatalf("expected after.txt not to exist after failure")
	}
}

