package template

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tmpl "github.com/jamt29/structify/internal/template"
)

func TestRemove_LocalTemplate_WithConfirmation(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	// Create a local template.
	dir := filepath.Join(tmpl.TemplatesDir(), "to-remove")
	if err := os.MkdirAll(filepath.Join(dir, "template"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "scaffold.yaml"), []byte(templateMinValidManifestYAML("to-remove")), 0o644); err != nil {
		t.Fatalf("write scaffold.yaml: %v", err)
	}

	// Prepare fake stdin with "y\n".
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	if _, err := w.WriteString("y\n"); err != nil {
		t.Fatalf("write to pipe: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	buf := &bytes.Buffer{}
	cmd := *removeCmd
	cmd.SetIn(r)
	cmd.SetOut(buf)
	removeYes = false

	if err := cmd.RunE(&cmd, []string{"to-remove"}); err != nil {
		t.Fatalf("RunE error: %v", err)
	}

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("expected template dir to be removed")
	}
	out := buf.String()
	if !strings.Contains(out, "removed") {
		t.Fatalf("expected removal message, got: %q", out)
	}
}

func TestRemove_LocalTemplate_YesFlagSkipsPrompt(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := filepath.Join(tmpl.TemplatesDir(), "to-remove-yes")
	if err := os.MkdirAll(filepath.Join(dir, "template"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "scaffold.yaml"), []byte(templateMinValidManifestYAML("to-remove-yes")), 0o644); err != nil {
		t.Fatalf("write scaffold.yaml: %v", err)
	}

	buf := &bytes.Buffer{}
	cmd := *removeCmd
	cmd.SetOut(buf)
	removeYes = true

	if err := cmd.RunE(&cmd, []string{"to-remove-yes"}); err != nil {
		t.Fatalf("RunE error: %v", err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("expected template dir to be removed")
	}
}

func TestRemove_BuiltinTemplate_Errors(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	// There is no local template named "minimal-go"; attempt to remove should
	// detect builtin (if present) or just error as not found. For this test,
	// we only assert that we don't accidentally delete anything local.
	buf := &bytes.Buffer{}
	cmd := *removeCmd
	cmd.SetOut(buf)
	removeYes = true

	err := cmd.RunE(&cmd, []string{"minimal-go"})
	if err == nil {
		t.Fatalf("expected error when removing non-local or builtin template")
	}
}
