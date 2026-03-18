package template

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	git "github.com/go-git/go-git/v5"
	"github.com/jamt29/structify/internal/template"
)

func TestRunAddFromGit_Success(t *testing.T) {
	// Stub git clone to just create a fake repo with scaffold.yaml.
	origClone := gitCloneFunc
	defer func() { gitCloneFunc = origClone }()

	gitCloneFunc = func(path string, bare bool, o *git.CloneOptions) (*git.Repository, error) {
		if err := os.MkdirAll(path, 0o755); err != nil {
			return nil, err
		}
		content := []byte(templateMinValidManifestYAML("from-git"))
		if err := os.WriteFile(filepath.Join(path, "scaffold.yaml"), content, 0o644); err != nil {
			return nil, err
		}
		return &git.Repository{}, nil
	}

	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	src := &template.GitSource{
		SourceURL: "github.com/user/repo",
		Ref:       "v1.0.0",
	}

	buf := &bytes.Buffer{}
	cmd := addCmd
	cmd.SetOut(buf)
	addForce = false

	if err := runAddFromGit(cmd, src); err != nil {
		t.Fatalf("runAddFromGit error: %v", err)
	}

	dest := filepath.Join(template.TemplatesDir(), "from-git")
	if _, err := os.Stat(dest); err != nil {
		t.Fatalf("expected template dir %s to exist: %v", dest, err)
	}
}

func TestRunAddFromGit_DuplicateWithoutForce_Fails(t *testing.T) {
	origClone := gitCloneFunc
	defer func() { gitCloneFunc = origClone }()

	gitCloneFunc = func(path string, bare bool, o *git.CloneOptions) (*git.Repository, error) {
		if err := os.MkdirAll(path, 0o755); err != nil {
			return nil, err
		}
		content := []byte(templateMinValidManifestYAML("dup-git"))
		if err := os.WriteFile(filepath.Join(path, "scaffold.yaml"), content, 0o644); err != nil {
			return nil, err
		}
		return &git.Repository{}, nil
	}

	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Precreate destination dir to simulate existing template.
	dest := filepath.Join(template.TemplatesDir(), "dup-git")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatalf("mkdir dest: %v", err)
	}

	src := &template.GitSource{
		SourceURL: "github.com/user/repo",
	}

	buf := &bytes.Buffer{}
	cmd := addCmd
	cmd.SetOut(buf)
	addForce = false

	err := runAddFromGit(cmd, src)
	if err == nil {
		t.Fatalf("expected error for duplicate template without --force")
	}
}

func TestRunAddFromGit_WithForce_Overwrites(t *testing.T) {
	origClone := gitCloneFunc
	defer func() { gitCloneFunc = origClone }()

	gitCloneFunc = func(path string, bare bool, o *git.CloneOptions) (*git.Repository, error) {
		if err := os.MkdirAll(path, 0o755); err != nil {
			return nil, err
		}
		content := []byte(templateMinValidManifestYAML("force-git"))
		if err := os.WriteFile(filepath.Join(path, "scaffold.yaml"), content, 0o644); err != nil {
			return nil, err
		}
		return &git.Repository{}, nil
	}

	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	dest := filepath.Join(template.TemplatesDir(), "force-git")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatalf("mkdir dest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dest, "old.txt"), []byte("old"), 0o644); err != nil {
		t.Fatalf("write old.txt: %v", err)
	}

	src := &template.GitSource{
		SourceURL: "github.com/user/repo",
		Ref:       "main",
	}

	buf := &bytes.Buffer{}
	cmd := addCmd
	cmd.SetOut(buf)
	addForce = true

	if err := runAddFromGit(cmd, src); err != nil {
		t.Fatalf("runAddFromGit error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dest, "old.txt")); !os.IsNotExist(err) {
		t.Fatalf("expected old.txt to be removed when using --force")
	}
}

func TestRunAddFromGit_CloneError_Wrapped(t *testing.T) {
	origClone := gitCloneFunc
	defer func() { gitCloneFunc = origClone }()

	gitCloneFunc = func(path string, bare bool, o *git.CloneOptions) (*git.Repository, error) {
		return nil, errors.New("network error")
	}

	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	src := &template.GitSource{
		SourceURL: "github.com/user/repo",
	}

	buf := &bytes.Buffer{}
	cmd := addCmd
	cmd.SetOut(buf)
	addForce = false

	err := runAddFromGit(cmd, src)
	if err == nil {
		t.Fatalf("expected error from clone failure")
	}
}

// helper to build a minimal manifest used in tests.
func templateMinValidManifestYAML(name string) string {
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

