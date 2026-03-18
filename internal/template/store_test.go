package template

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStore_ListEmpty(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	got, err := List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if got == nil {
		t.Fatalf("List() = nil, want empty slice")
	}
	if len(got) != 0 {
		t.Fatalf("List() len=%d, want 0", len(got))
	}
}

func TestStore_AddGetRemoveCycle(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	// Create a source template directory with a valid scaffold.yaml.
	srcRoot := filepath.Join(t.TempDir(), "my-template")
	if err := os.MkdirAll(filepath.Join(srcRoot, "template"), 0o755); err != nil {
		t.Fatalf("mkdir src template: %v", err)
	}
	// Add a regular file + symlink to cover copyDir symlink handling.
	if err := os.WriteFile(filepath.Join(srcRoot, "template", "target.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write target.txt: %v", err)
	}
	if err := os.Symlink("target.txt", filepath.Join(srcRoot, "template", "link.txt")); err != nil {
		t.Fatalf("symlink link.txt: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcRoot, "scaffold.yaml"), []byte(minValidManifestYAML("my-template")), 0o644); err != nil {
		t.Fatalf("write scaffold.yaml: %v", err)
	}

	if err := Add(srcRoot); err != nil {
		t.Fatalf("Add() error: %v", err)
	}

	if ok, err := Exists("my-template"); err != nil {
		t.Fatalf("Exists() error: %v", err)
	} else if !ok {
		t.Fatalf("Exists()=false, want true")
	}

	got, err := Get("my-template")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if got == nil || got.Manifest == nil {
		t.Fatalf("Get() returned nil template/manifest")
	}
	if got.Manifest.Name != "my-template" {
		t.Fatalf("manifest name=%q, want %q", got.Manifest.Name, "my-template")
	}

	linkPath := filepath.Join(TemplatesDir(), "my-template", "template", "link.txt")
	if fi, err := os.Lstat(linkPath); err != nil {
		t.Fatalf("lstat link.txt: %v", err)
	} else if fi.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("expected link.txt to be symlink, mode=%v", fi.Mode())
	}

	if err := Remove("my-template"); err != nil {
		t.Fatalf("Remove() error: %v", err)
	}

	if ok, err := Exists("my-template"); err != nil {
		t.Fatalf("Exists() error: %v", err)
	} else if ok {
		t.Fatalf("Exists()=true after remove, want false")
	}

	if _, err := Get("my-template"); err == nil {
		t.Fatalf("Get() after Remove() expected error, got nil")
	}
}

func TestStore_GetMissing(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if _, err := Get("does-not-exist"); err == nil {
		t.Fatalf("Get() expected error, got nil")
	}
}

func TestStore_ListIgnoresFolderWithoutManifest(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	root := TemplatesDir()
	if err := os.MkdirAll(filepath.Join(root, "no-manifest"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	got, err := List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("List() len=%d, want 0", len(got))
	}
}

func TestStore_AddDuplicateErrors(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	srcRoot := filepath.Join(t.TempDir(), "dup-template")
	if err := os.MkdirAll(filepath.Join(srcRoot, "template"), 0o755); err != nil {
		t.Fatalf("mkdir src template: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcRoot, "scaffold.yaml"), []byte(minValidManifestYAML("dup-template")), 0o644); err != nil {
		t.Fatalf("write scaffold.yaml: %v", err)
	}

	if err := Add(srcRoot); err != nil {
		t.Fatalf("Add() error: %v", err)
	}
	if err := Add(srcRoot); err == nil {
		t.Fatalf("expected duplicate Add() error")
	}
}

func TestStore_RemoveMissingErrors(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if err := Remove("nope"); err == nil {
		t.Fatalf("expected Remove() to error for missing template")
	}
}

func TestStore_BadInputs(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if _, err := Get(""); err == nil {
		t.Fatalf("expected Get(\"\") error")
	}
	if _, err := Exists(""); err == nil {
		t.Fatalf("expected Exists(\"\") error")
	}
	if err := Remove(""); err == nil {
		t.Fatalf("expected Remove(\"\") error")
	}
	if err := Add(""); err == nil {
		t.Fatalf("expected Add(\"\") error")
	}
}

func TestStore_AddInvalidManifestErrors(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	srcRoot := filepath.Join(t.TempDir(), "bad-template")
	if err := os.MkdirAll(filepath.Join(srcRoot, "template"), 0o755); err != nil {
		t.Fatalf("mkdir src template: %v", err)
	}
	// Missing required name field.
	if err := os.WriteFile(filepath.Join(srcRoot, "scaffold.yaml"), []byte("version: \"0.0.1\"\nlanguage: \"go\"\narchitecture: \"clean\"\n"), 0o644); err != nil {
		t.Fatalf("write scaffold.yaml: %v", err)
	}

	if err := Add(srcRoot); err == nil {
		t.Fatalf("expected Add() error for invalid manifest")
	}
}

func TestStore_ListInvalidTemplateErrors(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	root := TemplatesDir()
	bad := filepath.Join(root, "bad")
	if err := os.MkdirAll(filepath.Join(bad, "template"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// invalid scaffold.yaml (missing name/version etc)
	if err := os.WriteFile(filepath.Join(bad, "scaffold.yaml"), []byte("language: go\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	if _, err := List(); err == nil {
		t.Fatalf("expected List() to error on invalid template")
	}
}

func TestStore_AddBadYAMLErrors(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	srcRoot := filepath.Join(t.TempDir(), "bad-yaml")
	if err := os.MkdirAll(filepath.Join(srcRoot, "template"), 0o755); err != nil {
		t.Fatalf("mkdir src template: %v", err)
	}
	// Invalid YAML.
	if err := os.WriteFile(filepath.Join(srcRoot, "scaffold.yaml"), []byte("name: [\n"), 0o644); err != nil {
		t.Fatalf("write scaffold.yaml: %v", err)
	}
	if err := Add(srcRoot); err == nil {
		t.Fatalf("expected Add() to error on invalid YAML")
	}
}

func minValidManifestYAML(name string) string {
	return "" +
		"name: \"" + name + "\"\n" +
		"version: \"0.0.1\"\n" +
		"author: \"test\"\n" +
		"language: \"go\"\n" +
		"architecture: \"clean\"\n" +
		"description: \"test\"\n" +
		"tags: [\"test\"]\n" +
		"inputs: []\n" +
		"files: []\n" +
		"steps: []\n"
}

