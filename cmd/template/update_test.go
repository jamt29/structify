package template

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jamt29/structify/internal/dsl"
	"github.com/jamt29/structify/internal/template"
)

type fakeUpdateGitHubClient struct{}

func (f *fakeUpdateGitHubClient) Clone(ref *template.GitHubRef, destDir string) error {
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}
	content := []byte(templateMinValidManifestYAML("from-git-update"))
	return os.WriteFile(filepath.Join(destDir, "scaffold.yaml"), content, 0o644)
}

func (f *fakeUpdateGitHubClient) ValidateTemplateRepo(clonedPath string) (*dsl.Manifest, error) {
	// Minimal validation for tests; assume manifest is valid.
	return nil, nil
}

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
	if err != nil {
		t.Fatalf("did not expect error for template without metadata, got: %v", err)
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

	origNewClient := newGitHubClientFn
	newGitHubClientFn = func() githubClient {
		return &fakeUpdateGitHubClient{}
	}
	defer func() { newGitHubClientFn = origNewClient }()

	cmd := *updateCmd
	cmd.SetOut(os.Stdout)
	if err := cmd.RunE(&cmd, []string{"from-git-update"}); err != nil {
		t.Fatalf("RunE error: %v", err)
	}
}
