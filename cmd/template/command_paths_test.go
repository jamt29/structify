package template

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tmpl "github.com/jamt29/structify/internal/template"
)

func TestCreateCmd_RunE_NonInteractive(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfgDir := filepath.Join(home, ".structify")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte("nonInteractive: true\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	outRoot := t.TempDir()
	origOutput := createOutputPath
	createOutputPath = outRoot
	defer func() { createOutputPath = origOutput }()

	in := strings.NewReader("mytpl\nA sample\ngo\nclean\nme\n")
	out := &bytes.Buffer{}
	createCmd.SetIn(in)
	createCmd.SetOut(out)

	if err := createCmd.RunE(createCmd, nil); err != nil {
		t.Fatalf("create run error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(outRoot, "mytpl", "scaffold.yaml")); err != nil {
		t.Fatalf("expected scaffold.yaml created: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outRoot, "mytpl", "template", ".gitkeep")); err != nil {
		t.Fatalf("expected .gitkeep created: %v", err)
	}
}

func TestOpenInEditor_Branches(t *testing.T) {
	f := filepath.Join(t.TempDir(), "scaffold.yaml")
	if err := os.WriteFile(f, []byte("name: t"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	// No editor available branch.
	t.Setenv("EDITOR", "")
	t.Setenv("PATH", "")
	if err := openInEditor(f); err == nil {
		t.Fatalf("expected no-editor error")
	}

	// Editor command available but failing branch.
	t.Setenv("EDITOR", "false")
	if err := openInEditor(f); err == nil {
		t.Fatalf("expected failing editor command error")
	}
}

func TestImportCmd_RunE_LocalSourceYes(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfgDir := filepath.Join(home, ".structify")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte("nonInteractive: true\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	origName, origYes := importName, importYes
	importName, importYes = "imported-local", true
	defer func() {
		importName, importYes = origName, origYes
	}()

	out := &bytes.Buffer{}
	importCmd.SetOut(out)
	importCmd.SetIn(strings.NewReader(""))

	if err := importCmd.RunE(importCmd, []string{src}); err != nil {
		t.Fatalf("import run error: %v", err)
	}

	dest := filepath.Join(tmpl.TemplatesDir(), "imported-local")
	if _, err := os.Stat(filepath.Join(dest, "scaffold.yaml")); err != nil {
		t.Fatalf("expected imported scaffold.yaml: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dest, "template")); err != nil {
		t.Fatalf("expected imported template dir: %v", err)
	}
}
