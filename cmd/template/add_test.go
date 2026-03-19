package template

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jamt29/structify/internal/dsl"
	"github.com/jamt29/structify/internal/template"
)

type fakeGitHubClient struct {
	cloneErr error
}

func (f *fakeGitHubClient) Clone(ref *template.GitHubRef, destDir string) error {
	if f.cloneErr != nil {
		return f.cloneErr
	}
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}
	content := []byte(templateMinValidManifestYAML("from-git"))
	return os.WriteFile(filepath.Join(destDir, "scaffold.yaml"), content, 0o644)
}

func (f *fakeGitHubClient) ValidateTemplateRepo(clonedPath string) (*dsl.Manifest, error) {
	// not used in add tests; validation is done directly via dsl in runAddFromGit
	return nil, nil
}

func TestRunAddFromGit_Success(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	buf := &bytes.Buffer{}
	cmd := addCmd
	cmd.SetOut(buf)
	addForce = false

	client := &fakeGitHubClient{}
	ref := &template.GitHubRef{
		Owner: "user",
		Repo:  "from-git",
		Ref:   "v1.0.0",
	}
	if err := runAddFromGit(cmd, client, ref); err != nil {
		t.Fatalf("runAddFromGit error: %v", err)
	}

	dest := filepath.Join(template.TemplatesDir(), "from-git")
	if _, err := os.Stat(dest); err != nil {
		t.Fatalf("expected template dir %s to exist: %v", dest, err)
	}
}

func TestRunAddFromGit_DuplicateWithoutForce_Fails(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Precreate destination dir to simulate existing template.
	dest := filepath.Join(template.TemplatesDir(), "dup-git")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatalf("mkdir dest: %v", err)
	}

	buf := &bytes.Buffer{}
	cmd := addCmd
	cmd.SetOut(buf)
	addForce = false

	client := &fakeGitHubClient{}
	ref := &template.GitHubRef{
		Owner: "user",
		Repo:  "dup-git",
	}

	err := runAddFromGit(cmd, client, ref)
	if err == nil {
		t.Fatalf("expected error for duplicate template without --force")
	}
}

func TestRunAddFromGit_WithForce_Overwrites(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	dest := filepath.Join(template.TemplatesDir(), "force-git")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatalf("mkdir dest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dest, "old.txt"), []byte("old"), 0o644); err != nil {
		t.Fatalf("write old.txt: %v", err)
	}

	buf := &bytes.Buffer{}
	cmd := addCmd
	cmd.SetOut(buf)
	addForce = true

	client := &fakeGitHubClient{}
	ref := &template.GitHubRef{
		Owner: "user",
		Repo:  "force-git",
		Ref:   "main",
	}

	if err := runAddFromGit(cmd, client, ref); err != nil {
		t.Fatalf("runAddFromGit error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dest, "old.txt")); !os.IsNotExist(err) {
		t.Fatalf("expected old.txt to be removed when using --force")
	}
}

func TestRunAddFromGit_CloneError_Wrapped(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	buf := &bytes.Buffer{}
	cmd := addCmd
	cmd.SetOut(buf)
	addForce = false

	client := &fakeGitHubClient{cloneErr: fmt.Errorf("network error")}
	ref := &template.GitHubRef{
		Owner: "user",
		Repo:  "repo",
	}

	err := runAddFromGit(cmd, client, ref)
	if err == nil {
		t.Fatalf("expected error from clone failure")
	}
}

