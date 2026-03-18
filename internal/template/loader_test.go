package template

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromPath_MissingManifest(t *testing.T) {
	dir := t.TempDir()
	if _, err := LoadFromPath(dir); err == nil {
		t.Fatalf("expected error for missing scaffold.yaml")
	}
}

func TestLoadBuiltins_ReturnsTemplates(t *testing.T) {
	got, err := LoadBuiltins()
	if err != nil {
		t.Fatalf("LoadBuiltins() error: %v", err)
	}
	if len(got) == 0 {
		t.Fatalf("LoadBuiltins() returned 0 templates")
	}
	// Ensure at least one known builtin exists.
	var found bool
	for _, tpls := range got {
		if tpls != nil && tpls.Manifest != nil && tpls.Manifest.Name == "minimal-go" {
			found = true
			if tpls.Source != "builtin" {
				t.Fatalf("expected Source=builtin, got %q", tpls.Source)
			}
			if filepath.Base(tpls.Path) != "minimal-go" {
				t.Fatalf("expected path basename minimal-go, got %q", filepath.Base(tpls.Path))
			}
		}
	}
	if !found {
		t.Fatalf("expected to find builtin minimal-go")
	}
}

func TestLoadFromPath_SourceLocalWhenUnderStore(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	// Install a template into the store using Add, then Get, then verify source.
	srcRoot := filepath.Join(t.TempDir(), "local-template")
	if err := os.MkdirAll(filepath.Join(srcRoot, "template"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcRoot, "scaffold.yaml"), []byte(minValidManifestYAML("local-template")), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := Add(srcRoot); err != nil {
		t.Fatalf("Add() error: %v", err)
	}
	got, err := Get("local-template")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if got.Source != "local" {
		t.Fatalf("Source=%q, want local", got.Source)
	}
}

