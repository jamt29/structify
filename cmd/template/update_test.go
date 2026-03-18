package template

import (
	"os"
	"path/filepath"
	"testing"

	git "github.com/go-git/go-git/v5"
	"github.com/jamt29/structify/internal/template"
)

func TestUpdate_TemplateWithoutMetadataErrors(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	// Create a local template without metadata.
	dir := filepath.Join(template.TemplatesDir(), "no-meta")
	if err := os.MkdirAll(filepath.Join(dir, "template"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "scaffold.yaml"), []byte(templateMinValidManifestYAML("no-meta")), 0o644); err != nil {
		t.Fatalf("write scaffold.yaml: %v", err)
	}

	cmd := *updateCmd
	cmd.SetOut(os.Stdout)
	err := cmd.RunE(&cmd, []string{"no-meta"})
	if err == nil {
		t.Fatalf("expected error for template without metadata")
	}
}

func TestUpdate_ReclonesFromGitSource(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	// Prepare local template with metadata.
	dir := filepath.Join(template.TemplatesDir(), "from-git-update")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "scaffold.yaml"), []byte(templateMinValidManifestYAML("from-git-update")), 0o644); err != nil {
		t.Fatalf("write scaffold.yaml: %v", err)
	}
	meta := &template.TemplateMeta{
		SourceURL: "github.com/user/repo",
		SourceRef: "v1.0.0",
	}
	if err := template.WriteTemplateMeta(dir, meta); err != nil {
		t.Fatalf("WriteTemplateMeta: %v", err)
	}

	origClone := gitCloneFunc
	defer func() { gitCloneFunc = origClone }()

	gitCloneFunc = func(path string, bare bool, o *git.CloneOptions) (*git.Repository, error) {
		if err := os.MkdirAll(path, 0o755); err != nil {
			return nil, err
		}
		content := []byte(templateMinValidManifestYAML("from-git-update"))
		if err := os.WriteFile(filepath.Join(path, "scaffold.yaml"), content, 0o644); err != nil {
			return nil, err
		}
		return &git.Repository{}, nil
	}

	cmd := *updateCmd
	cmd.SetOut(os.Stdout)
	if err := cmd.RunE(&cmd, []string{"from-git-update"}); err != nil {
		t.Fatalf("RunE error: %v", err)
	}
}

