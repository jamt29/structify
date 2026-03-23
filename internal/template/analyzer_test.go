package template

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyzeProject_GoModuleAndProjectName(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module github.com/acme/my-api\n\ngo 1.21\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "internal"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "internal", "x.go"), []byte("package internal\nimport \"github.com/acme/my-api/pkg\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := AnalyzeProject(root)
	if err != nil {
		t.Fatalf("AnalyzeProject error: %v", err)
	}
	if res.Language != "go" {
		t.Fatalf("expected go language, got %q", res.Language)
	}
	if len(res.DetectedVars) == 0 {
		t.Fatalf("expected detected vars")
	}
}
