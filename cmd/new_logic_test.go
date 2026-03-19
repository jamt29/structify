package cmd

import (
	"strings"
	"testing"

	"github.com/jamt29/structify/internal/dsl"
)

func TestBuildInitialContext_NameAndVars(t *testing.T) {
	ctx, err := buildInitialContext("myapp", []string{"a=x", "b=y"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if ctxString(ctx, "project_name") != "myapp" {
		t.Fatalf("expected project_name=myapp, got %v", ctx["project_name"])
	}
	if ctxString(ctx, "a") != "x" || ctxString(ctx, "b") != "y" {
		t.Fatalf("expected vars in context, got %#v", ctx)
	}
}

func TestBuildInitialContext_InvalidVar(t *testing.T) {
	_, err := buildInitialContext("", []string{"novalue"})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestCoerceProvidedVarsToTypes_Bool(t *testing.T) {
	inputs := []dsl.Input{
		{ID: "use_docker", Type: "bool"},
	}
	ctx := dsl.Context{"use_docker": "true"}
	got, err := coerceProvidedVarsToTypes(inputs, ctx)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if v, ok := got["use_docker"].(bool); !ok || v != true {
		t.Fatalf("expected bool true, got %#v", got["use_docker"])
	}
}

func TestFinalizeContextNonInteractive_RequiredMissing(t *testing.T) {
	inputs := []dsl.Input{
		{ID: "project_name", Type: "string", Required: true},
	}
	_, err := finalizeContextNonInteractive(inputs, dsl.Context{})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestFinalizeContextNonInteractive_DefaultApplied(t *testing.T) {
	inputs := []dsl.Input{
		{ID: "transport", Type: "enum", Required: true, Options: []string{"http", "grpc"}, Default: "http"},
	}
	ctx, err := finalizeContextNonInteractive(inputs, dsl.Context{})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if ctxString(ctx, "transport") != "http" {
		t.Fatalf("expected http, got %#v", ctx["transport"])
	}
}

func TestDryRunSteps_ShowsSkipReason(t *testing.T) {
	steps := []dsl.Step{
		{Name: "install", Run: "echo install", When: `orm == "gorm"`},
	}
	ctx := dsl.Context{"orm": "none"}
	lines := dryRunSteps(steps, ctx)
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	if !strings.HasPrefix(lines[0], "─ ") {
		t.Fatalf("expected skipped line, got %q", lines[0])
	}
}

func TestResolveContextInterpolations_ModulePathDefault(t *testing.T) {
	ctx := dsl.Context{
		"project_name": "MyProject",
		"module_path":  "github.com/user/{{ project_name | kebab_case }}",
	}

	if err := resolveContextInterpolations(ctx); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	got := ctxString(ctx, "module_path")
	if got != "github.com/user/my-project" {
		t.Fatalf("expected resolved module_path, got %q", got)
	}
}

