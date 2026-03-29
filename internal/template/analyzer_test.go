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
	if res.Confidence <= 0 {
		t.Fatalf("expected confidence > 0")
	}
	if len(res.SuggestedInputs) == 0 {
		t.Fatalf("expected suggested inputs")
	}
	foundORM := false
	foundTransport := false
	foundGoVersion := false
	for _, s := range res.SuggestedInputs {
		if s.Input.ID == "orm" && s.Default == "gorm" {
			foundORM = true
		}
		if s.Input.ID == "transport" && s.Default == "gin" {
			foundTransport = true
		}
		if s.Input.ID == "go_version" {
			foundGoVersion = true
		}
	}
	if !foundGoVersion {
		t.Fatalf("expected go_version suggested input")
	}
	if foundORM || foundTransport {
		t.Fatalf("did not expect orm/transport without deps")
	}
}

func TestAnalyzeProject_GoDependenciesSuggestInputs(t *testing.T) {
	root := t.TempDir()
	goMod := `module github.com/acme/my-api

go 1.21

require (
	gorm.io/gorm v1.25.0
	github.com/gin-gonic/gin v1.9.0
)`
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "cmd", "main"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "cmd", "main", "main.go"), []byte("package main\nfunc main(){}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := AnalyzeProject(root)
	if err != nil {
		t.Fatalf("AnalyzeProject error: %v", err)
	}
	if len(res.DetectedDeps) == 0 {
		t.Fatalf("expected detected dependencies")
	}
	seenORM := false
	seenTransport := false
	for _, s := range res.SuggestedInputs {
		if s.Input.ID == "orm" && s.Default == "gorm" {
			seenORM = true
		}
		if s.Input.ID == "transport" && s.Default == "gin" {
			seenTransport = true
		}
	}
	if !seenORM {
		t.Fatalf("expected orm suggestion for gorm")
	}
	if !seenTransport {
		t.Fatalf("expected transport suggestion for gin")
	}
}
