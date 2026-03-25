package dsl

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadManifest_Valid(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "scaffold.yaml")

	y := `
name: "clean-architecture-go"
version: "1.0.0"
author: "me"
language: "go"
architecture: "clean"
description: "desc"
tags: ["go"]
inputs:
  - id: project_name
    prompt: "Project name?"
    type: string
    required: true
    default: "my-project"
files:
  - include: "docker/**"
    when: use_docker == true
steps:
  - name: "Tidy"
    run: "go mod tidy"
`
	if err := os.WriteFile(p, []byte(y), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	m, err := LoadManifest(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Name != "clean-architecture-go" {
		t.Fatalf("got name %q", m.Name)
	}
	if m.Language != "go" {
		t.Fatalf("got language %q", m.Language)
	}
}

func TestParseManifest_Valid(t *testing.T) {
	y := `
name: "t"
version: "1.0.0"
author: "me"
language: "go"
architecture: "clean"
description: "d"
`
	m, err := ParseManifest([]byte(y))
	if err != nil {
		t.Fatalf("ParseManifest: %v", err)
	}
	if m.Name != "t" {
		t.Fatalf("name: %q", m.Name)
	}
}

func TestLoadManifest_FileNotFound(t *testing.T) {
	_, err := LoadManifest("/does/not/exist/scaffold.yaml")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestLoadManifest_BadYAML(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "scaffold.yaml")
	if err := os.WriteFile(p, []byte("name: ["), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	_, err := LoadManifest(p)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "parsing manifest") {
		t.Fatalf("unexpected error: %v", err)
	}
}
