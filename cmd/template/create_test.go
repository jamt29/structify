package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tmpl "github.com/jamt29/structify/internal/template"
)

func TestCreateMinimalLocalTemplate_CreatesMinimalManifest(t *testing.T) {
	dir := t.TempDir()

	if err := tmpl.CreateMinimalLocalTemplate(dir, "my-template", "desc", "go", "clean", "alice"); err != nil {
		t.Fatalf("CreateMinimalLocalTemplate error: %v", err)
	}

	path := filepath.Join(dir, "my-template", "scaffold.yaml")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read scaffold.yaml: %v", err)
	}
	s := string(b)
	if !containsAll(s, []string{
		`name: "my-template"`,
		`language: "go"`,
		`architecture: "clean"`,
		`description: "desc"`,
		`author: "alice"`,
		`id: "project_name"`,
		`required: true`,
		`validate: "^[a-zA-Z][a-zA-Z0-9_-]*$"`,
		`steps: []`,
	}) {
		t.Fatalf("unexpected scaffold.yaml content: %s", s)
	}
	gitkeep := filepath.Join(dir, "my-template", "template", ".gitkeep")
	if _, err := os.Stat(gitkeep); err != nil {
		t.Fatalf("expected template/.gitkeep: %v", err)
	}
}

func containsAll(s string, subs []string) bool {
	for _, sub := range subs {
		if !strings.Contains(s, sub) {
			return false
		}
	}
	return true
}
