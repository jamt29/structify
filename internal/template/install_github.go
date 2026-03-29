package template

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jamt29/structify/internal/dsl"
)

// GitHubCloner clones a repository reference (implemented by *GitHubClient and test fakes).
type GitHubCloner interface {
	Clone(ref *GitHubRef, destDir string) error
}

// InstallFromGitHubOptions configures installation from a cloned GitHub repo.
type InstallFromGitHubOptions struct {
	Force     bool
	LocalName string // optional; defaults to repo name from ref
}

// InstallFromGitHub clones ref into a temp dir, validates scaffold.yaml, and copies into the local template store.
func InstallFromGitHub(client GitHubCloner, ref *GitHubRef, opts InstallFromGitHubOptions) error {
	if client == nil {
		return fmt.Errorf("github client is nil")
	}
	if ref == nil {
		return fmt.Errorf("github ref is nil")
	}

	tmpDir, err := os.MkdirTemp("", "structify-template-add-*")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := client.Clone(ref, tmpDir); err != nil {
		return err
	}

	manifestPath := filepath.Join(tmpDir, manifestFileName)
	m, err := dsl.LoadManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("loading scaffold.yaml from cloned repo: %w", err)
	}
	if verrs := dsl.ValidateManifest(m); len(verrs) > 0 {
		return fmt.Errorf("cloned template is invalid")
	}

	if strings.TrimSpace(m.Name) == "" {
		return fmt.Errorf("cloned template has empty name in manifest")
	}

	localName := strings.TrimSpace(opts.LocalName)
	if localName == "" {
		localName = ref.Repo
	}

	templatesRoot := TemplatesDir()
	if err := os.MkdirAll(templatesRoot, 0o755); err != nil {
		return fmt.Errorf("creating templates dir %s: %w", templatesRoot, err)
	}

	destDir := filepath.Join(templatesRoot, localName)
	if _, err := os.Stat(destDir); err == nil {
		if !opts.Force {
			return fmt.Errorf("template %q already exists (use force to overwrite)", localName)
		}
		if err := os.RemoveAll(destDir); err != nil {
			return fmt.Errorf("removing existing template %q: %w", localName, err)
		}
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("stat destination %s: %w", destDir, err)
	}

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("creating destination dir %s: %w", destDir, err)
	}

	if err := CopyDirForTest(tmpDir, destDir); err != nil {
		return fmt.Errorf("copying cloned template to store: %w", err)
	}

	meta := &TemplateMeta{
		SourceURL:   fmt.Sprintf("github.com/%s/%s", ref.Owner, ref.Repo),
		SourceRef:   ref.Ref,
		InstalledAt: time.Now().UTC().Format(time.RFC3339),
	}
	if err := WriteTemplateMeta(destDir, meta); err != nil {
		return fmt.Errorf("writing template metadata: %w", err)
	}

	return nil
}
