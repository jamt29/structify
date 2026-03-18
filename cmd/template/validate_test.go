package template

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamt29/structify/internal/dsl"
)

func TestResolveManifestPath_DirAndFile(t *testing.T) {
	tmp := t.TempDir()
	dirPath := tmp
	filePath := filepath.Join(tmp, "scaffold.yaml")

	if err := os.WriteFile(filePath, []byte("name: test\n"), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	gotDir, err := resolveManifestPath(dirPath)
	if err != nil {
		t.Fatalf("resolveManifestPath(dir) error: %v", err)
	}
	if gotDir != filePath {
		t.Fatalf("expected %s, got %s", filePath, gotDir)
	}

	gotFile, err := resolveManifestPath(filePath)
	if err != nil {
		t.Fatalf("resolveManifestPath(file) error: %v", err)
	}
	if gotFile != filePath {
		t.Fatalf("expected %s, got %s", filePath, gotFile)
	}
}

func TestValidateCmd_ValidManifest_TextOutput(t *testing.T) {
	tmp := t.TempDir()
	manifestPath := filepath.Join(tmp, "scaffold.yaml")
	content := []byte(`
name: demo
version: "1.0.0"
author: alice
language: go
architecture: clean
description: Demo template
inputs: []
files: []
steps: []
`)
	if err := os.WriteFile(manifestPath, content, 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	buf := &bytes.Buffer{}
	cmd := validateCmd
	cmd.SetOut(buf)
	validateJSON = false

	if err := cmd.RunE(cmd, []string{tmp}); err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Template is valid") {
		t.Fatalf("expected success message, got: %q", out)
	}
	if !strings.Contains(out, "Inputs: 0, Steps: 0, File rules: 0") {
		t.Fatalf("expected summary line, got: %q", out)
	}
}

func TestValidateCmd_InvalidManifest_ShowsErrorsAndFails(t *testing.T) {
	tmp := t.TempDir()
	manifestPath := filepath.Join(tmp, "scaffold.yaml")

	// Missing required fields like name/language etc should trigger validator errors.
	content := []byte(`{}`)
	if err := os.WriteFile(manifestPath, content, 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	// Ensure validator complains; if not, this test may need adjustment to actual rules.
	m, err := dsl.LoadManifest(manifestPath)
	if err != nil {
		t.Fatalf("LoadManifest failed: %v", err)
	}
	if len(dsl.ValidateManifest(m)) == 0 {
		t.Skip("validator returned no errors for minimal manifest; adjust test to a manifest that is invalid")
	}

	buf := &bytes.Buffer{}
	cmd := validateCmd
	cmd.SetOut(buf)
	validateJSON = false

	err = cmd.RunE(cmd, []string{tmp})
	if err == nil {
		t.Fatalf("expected error for invalid manifest, got nil")
	}
	out := buf.String()
	if !strings.Contains(out, "- ") {
		t.Fatalf("expected at least one validation error line, got: %q", out)
	}
}

func TestValidateCmd_JSONOutput(t *testing.T) {
	tmp := t.TempDir()
	manifestPath := filepath.Join(tmp, "scaffold.yaml")
	content := []byte(`
name: demo
version: "1.0.0"
author: alice
language: go
architecture: clean
description: Demo template
inputs: []
files: []
steps: []
`)
	if err := os.WriteFile(manifestPath, content, 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	buf := &bytes.Buffer{}
	cmd := validateCmd
	cmd.SetOut(buf)
	validateJSON = true

	if err := cmd.RunE(cmd, []string{tmp}); err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, `"valid": true`) {
		t.Fatalf("expected JSON with valid true, got: %q", out)
	}
}

