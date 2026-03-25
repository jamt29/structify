package template

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ProjectNameValidateRegex is the default regex for template and project names in minimal scaffolds.
const ProjectNameValidateRegex = `^[a-zA-Z][a-zA-Z0-9_-]*$`

// CreateMinimalLocalTemplate creates destRoot/<name>/ with scaffold.yaml and template/.gitkeep.
// If destRoot is empty, TemplatesDir() is used.
func CreateMinimalLocalTemplate(destRoot, name, description, language, architecture, author string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("template name cannot be empty")
	}
	if strings.TrimSpace(destRoot) == "" {
		destRoot = TemplatesDir()
	}
	destDir := filepath.Join(destRoot, name)
	if _, err := os.Stat(destDir); err == nil {
		return fmt.Errorf("template %q already exists at %s", name, destDir)
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("stat template dir %s: %w", destDir, err)
	}
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("creating template dir %s: %w", destDir, err)
	}
	if err := writeMinimalScaffoldYAML(destDir, name, description, language, architecture, author); err != nil {
		return err
	}
	templateDir := filepath.Join(destDir, "template")
	if err := os.MkdirAll(templateDir, 0o755); err != nil {
		return fmt.Errorf("creating template/ dir: %w", err)
	}
	if err := os.WriteFile(filepath.Join(templateDir, ".gitkeep"), []byte{}, 0o644); err != nil {
		return fmt.Errorf("writing template/.gitkeep: %w", err)
	}
	return nil
}

func writeMinimalScaffoldYAML(dir, name, description, language, architecture, author string) error {
	content := fmt.Sprintf(`name: %q
version: "0.1.0"
author: %q
language: %q
architecture: %q
description: %q
inputs:
  - id: "project_name"
    prompt: "Project name?"
    type: string
    required: true
    validate: %q
steps: []
`, name, author, language, architecture, description, ProjectNameValidateRegex)

	path := filepath.Join(dir, manifestFileName)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing scaffold.yaml: %w", err)
	}
	return nil
}
