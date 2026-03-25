package template

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateMinimalLocalTemplate_DuplicateFails(t *testing.T) {
	dir := t.TempDir()
	if err := CreateMinimalLocalTemplate(dir, "dup", "", "go", "clean", "a"); err != nil {
		t.Fatal(err)
	}
	if err := CreateMinimalLocalTemplate(dir, "dup", "", "go", "clean", "a"); err == nil {
		t.Fatal("expected error for duplicate template")
	}
}

func TestCreateMinimalLocalTemplate_WritesScaffold(t *testing.T) {
	dir := t.TempDir()
	if err := CreateMinimalLocalTemplate(dir, "x", "hi", "rust", "hexagonal", "bob"); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, "x", manifestFileName)
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) < 50 {
		t.Fatalf("unexpected short scaffold: %q", string(b))
	}
}
