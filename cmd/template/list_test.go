package template

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/jamt29/structify/internal/dsl"
	tmpl "github.com/jamt29/structify/internal/template"
)

func withListAllStub(t *testing.T, stub func() ([]*tmpl.Template, error)) func() {
	t.Helper()
	orig := engineListAll
	engineListAll = stub
	return func() { engineListAll = orig }
}

func TestList_NoTemplates_ShowsFriendlyMessage(t *testing.T) {
	restore := withListAllStub(t, func() ([]*tmpl.Template, error) {
		return []*tmpl.Template{}, nil
	})
	defer restore()

	buf := &bytes.Buffer{}
	cmd := listCmd
	cmd.SetOut(buf)
	cmd.Flags().Lookup("json").Value.Set("false")

	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "No templates found.") {
		t.Fatalf("expected friendly message, got: %q", out)
	}
	if !strings.Contains(out, "structify template add") {
		t.Fatalf("expected suggestion to add template, got: %q", out)
	}
}

func TestList_PrintsGroupsAndTable(t *testing.T) {
	restore := withListAllStub(t, func() ([]*tmpl.Template, error) {
		return []*tmpl.Template{
			{
				Manifest: &dsl.Manifest{
					Name:         "clean-go",
					Language:     "go",
					Architecture: "clean",
					Description:  "Clean Architecture in Go",
				},
				Source: "local",
			},
			{
				Manifest: &dsl.Manifest{
					Name:         "minimal-go",
					Language:     "go",
					Architecture: "",
					Description:  "Minimal Go project",
				},
				Source: "builtin",
			},
		}, nil
	})
	defer restore()

	buf := &bytes.Buffer{}
	cmd := listCmd
	cmd.SetOut(buf)
	cmd.Flags().Lookup("json").Value.Set("false")

	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Local templates:") {
		t.Fatalf("expected Local templates section, got: %q", out)
	}
	if !strings.Contains(out, "Built-in templates:") {
		t.Fatalf("expected Built-in templates section, got: %q", out)
	}
	if !strings.Contains(out, "clean-go") || !strings.Contains(out, "Clean Architecture in Go") {
		t.Fatalf("expected local template row, got: %q", out)
	}
	if !strings.Contains(out, "minimal-go") || !strings.Contains(out, "Minimal Go project") {
		t.Fatalf("expected builtin template row, got: %q", out)
	}
	if strings.Contains(out, "\t\t-\t") && !strings.Contains(out, "minimal-go") {
		t.Fatalf("expected architecture placeholder '-' only for builtin row, got: %q", out)
	}
}

func TestList_JSONOutput(t *testing.T) {
	restore := withListAllStub(t, func() ([]*tmpl.Template, error) {
		return []*tmpl.Template{
			{
				Manifest: &dsl.Manifest{
					Name:         "clean-go",
					Language:     "go",
					Architecture: "clean",
					Description:  "Clean Architecture in Go",
				},
				Source: "local",
			},
		}, nil
	})
	defer restore()

	buf := &bytes.Buffer{}
	cmd := listCmd
	cmd.SetOut(buf)
	cmd.Flags().Lookup("json").Value.Set("true")

	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, `"name": "clean-go"`) {
		t.Fatalf("expected JSON with name, got: %q", out)
	}
	if !strings.Contains(out, `"source": "local"`) {
		t.Fatalf("expected JSON with source, got: %q", out)
	}
}

func TestList_EngineError_IsWrapped(t *testing.T) {
	wantErr := errors.New("boom")
	restore := withListAllStub(t, func() ([]*tmpl.Template, error) {
		return nil, wantErr
	})
	defer restore()

	buf := &bytes.Buffer{}
	cmd := listCmd
	cmd.SetOut(buf)
	cmd.Flags().Lookup("json").Value.Set("false")

	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "listing templates") {
		t.Fatalf("expected context in error, got: %v", err)
	}
}

