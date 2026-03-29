package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamt29/structify/internal/dsl"
	"github.com/jamt29/structify/internal/template"
)

func TestResolveImportSource_LocalAndInvalid(t *testing.T) {
	dir := t.TempDir()

	got, cleanup, err := resolveImportSource(dir)
	if err != nil {
		t.Fatalf("resolveImportSource(local) error: %v", err)
	}
	if cleanup != nil {
		t.Fatalf("expected nil cleanup for local directory")
	}
	if got == "" {
		t.Fatalf("expected absolute path")
	}

	_, _, err = resolveImportSource("not-a-dir-and-not-github-url")
	if err == nil {
		t.Fatalf("expected error for invalid source")
	}
}

func TestMaterializeImportedTemplate_ReplacesAndIgnores(t *testing.T) {
	src := t.TempDir()
	dest := t.TempDir()

	if err := os.WriteFile(filepath.Join(src, "README.md"), []byte("hello testproject"), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(src, "vendor"), 0o755); err != nil {
		t.Fatalf("mkdir vendor: %v", err)
	}
	if err := os.WriteFile(filepath.Join(src, "vendor", "ignored.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write ignored file: %v", err)
	}

	vars := []template.DetectedVar{
		{ID: "project_name", SuggestAs: "testproject"},
	}
	included, ignored, err := materializeImportedTemplate(src, dest, []string{"vendor"}, vars)
	if err != nil {
		t.Fatalf("materializeImportedTemplate error: %v", err)
	}
	if included == 0 {
		t.Fatalf("expected included files > 0")
	}
	if ignored == 0 {
		t.Fatalf("expected ignored files > 0")
	}

	rendered := filepath.Join(dest, "README.md.tmpl")
	b, err := os.ReadFile(rendered)
	if err != nil {
		t.Fatalf("read rendered file: %v", err)
	}
	if !strings.Contains(string(b), "{{ project_name }}") {
		t.Fatalf("expected variable replacement in %s", rendered)
	}
}

func TestImportHelperFunctions(t *testing.T) {
	ignoredSet := map[string]struct{}{"a/b": {}}
	if !isIgnoredBySet("a/b", ignoredSet) {
		t.Fatalf("expected direct ignore")
	}
	if !isIgnoredBySet("a/b/c.txt", ignoredSet) {
		t.Fatalf("expected nested ignore")
	}
	if isIgnoredBySet("a/c", ignoredSet) {
		t.Fatalf("expected non-ignored path")
	}

	manifestGo := buildImportedManifest("n", &template.AnalysisResult{
		ProjectName: "n",
		Language:    "go",
		SuggestedInputs: []template.SuggestedInput{
			{Input: dsl.Input{ID: "orm", Prompt: "ORM?", Type: "enum", Options: []string{"gorm", "sqlx", "none"}, Default: "gorm"}},
		},
	}, []template.DetectedVar{{ID: "module_path", Description: "Module", Type: "string", SuggestAs: "github.com/acme/n"}})
	if len(manifestGo.Steps) == 0 {
		t.Fatalf("expected go steps")
	}
	if manifestGo.Version != "0.1.0" {
		t.Fatalf("expected version 0.1.0, got %q", manifestGo.Version)
	}
	manifestRust := buildImportedManifest("n", &template.AnalysisResult{ProjectName: "n", Language: "rust"}, nil)
	if len(manifestRust.Steps) == 0 {
		t.Fatalf("expected rust steps")
	}

	if got := joinVarIDs(nil); got != "-" {
		t.Fatalf("expected '-', got %q", got)
	}
	if got := joinVarIDs([]template.DetectedVar{{ID: "a"}, {ID: "b"}}); got != "a, b" {
		t.Fatalf("unexpected joined vars: %q", got)
	}

	_ = hasTTYImport() // cover function; environment-dependent.
}
