package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteScaffoldYAML_CreatesMinimalManifest(t *testing.T) {
	dir := t.TempDir()

	if err := writeScaffoldYAML(dir, "my-template", "desc", "go", "clean", "alice"); err != nil {
		t.Fatalf("writeScaffoldYAML error: %v", err)
	}

	path := filepath.Join(dir, "scaffold.yaml")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read scaffold.yaml: %v", err)
	}
	s := string(b)
	if !containsAll(s, []string{"name: \"my-template\"", "language: \"go\"", "architecture: \"clean\""}) {
		t.Fatalf("unexpected scaffold.yaml content: %s", s)
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

