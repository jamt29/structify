package template

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/jamt29/structify/internal/dsl"
	tmpl "github.com/jamt29/structify/internal/template"
)

func withResolveStub(t *testing.T, stub func(string) (*tmpl.Template, error)) func() {
	t.Helper()
	orig := engineResolve
	engineResolve = stub
	return func() { engineResolve = orig }
}

func TestInfo_PrintsDetails(t *testing.T) {
	restore := withResolveStub(t, func(name string) (*tmpl.Template, error) {
		return &tmpl.Template{
			Manifest: &dsl.Manifest{
				Name:         "minimal-go",
				Version:      "1.0.0",
				Author:       "Alice",
				Language:     "go",
				Architecture: "",
				Description:  "Minimal Go project",
				Tags:         []string{"minimal", "example"},
				Inputs: []dsl.Input{
					{
						ID:       "project_name",
						Type:     "string",
						Prompt:   "Project name",
						Required: true,
						Default:  "myapp",
						When:     "",
					},
				},
				Steps: []dsl.Step{
					{
						Name: "Init go module",
						Run:  "go mod tidy",
						When: "",
					},
				},
			},
			Source: "builtin",
		}, nil
	})
	defer restore()

	buf := &bytes.Buffer{}
	cmd := infoCmd
	cmd.SetOut(buf)

	if err := cmd.RunE(cmd, []string{"minimal-go"}); err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "minimal-go") {
		t.Fatalf("expected template name in output, got: %q", out)
	}
	if !strings.Contains(out, "Minimal Go project") {
		t.Fatalf("expected description in output, got: %q", out)
	}
	if !strings.Contains(out, "Version: 1.0.0") {
		t.Fatalf("expected version in output, got: %q", out)
	}
	if !strings.Contains(out, "Author: Alice") {
		t.Fatalf("expected author in output, got: %q", out)
	}
	if !strings.Contains(out, "Tags: minimal, example") {
		t.Fatalf("expected tags in output, got: %q", out)
	}
	if !strings.Contains(out, "Inputs") || !strings.Contains(out, "project_name (string)") {
		t.Fatalf("expected Inputs section with project_name, got: %q", out)
	}
	if !strings.Contains(out, "Steps") || !strings.Contains(out, "Init go module") {
		t.Fatalf("expected Steps section, got: %q", out)
	}
}

func TestInfo_ResolveError_Propagated(t *testing.T) {
	restore := withResolveStub(t, func(name string) (*tmpl.Template, error) {
		return nil, errors.New("not found")
	})
	defer restore()

	buf := &bytes.Buffer{}
	cmd := infoCmd
	cmd.SetOut(buf)

	err := cmd.RunE(cmd, []string{"does-not-exist"})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected underlying error message, got: %v", err)
	}
}
