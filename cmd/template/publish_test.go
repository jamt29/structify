package template

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPublish_FullChecklist_Passes(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "scaffold.yaml"), []byte(templateMinValidManifestYAML("pub-template")), 0o644); err != nil {
		t.Fatalf("write scaffold.yaml: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "template"), 0o755); err != nil {
		t.Fatalf("mkdir template/: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "template", "file.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write file.txt: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# readme"), 0o644); err != nil {
		t.Fatalf("write README.md: %v", err)
	}

	buf := &bytes.Buffer{}
	cmd := *publishCmd
	cmd.SetOut(buf)

	if err := cmd.RunE(&cmd, []string{dir}); err != nil {
		t.Fatalf("RunE error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "[✓] scaffold.yaml exists") {
		t.Fatalf("expected scaffold.yaml exists check, got: %q", out)
	}
	if !strings.Contains(out, "[✓] template/ directory has files") {
		t.Fatalf("expected template/ files check, got: %q", out)
	}
}

func TestPublish_MissingScaffold_Fails(t *testing.T) {
	dir := t.TempDir()

	buf := &bytes.Buffer{}
	cmd := *publishCmd
	cmd.SetOut(buf)

	err := cmd.RunE(&cmd, []string{dir})
	if err == nil {
		t.Fatalf("expected error when scaffold.yaml is missing")
	}
	out := buf.String()
	if !strings.Contains(out, "[✗] scaffold.yaml exists") {
		t.Fatalf("expected missing scaffold message, got: %q", out)
	}
}

func TestPublish_NoTemplateFiles_Fails(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "scaffold.yaml"), []byte(templateMinValidManifestYAML("no-files")), 0o644); err != nil {
		t.Fatalf("write scaffold.yaml: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "template"), 0o755); err != nil {
		t.Fatalf("mkdir template/: %v", err)
	}

	buf := &bytes.Buffer{}
	cmd := *publishCmd
	cmd.SetOut(buf)

	err := cmd.RunE(&cmd, []string{dir})
	if err == nil {
		t.Fatalf("expected error when template/ has no files")
	}
	out := buf.String()
	if !strings.Contains(out, "template/ directory has no files") {
		t.Fatalf("expected no files message, got: %q", out)
	}
}

