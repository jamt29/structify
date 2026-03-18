package template

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/jamt29/structify/internal/dsl"
	tmpl "github.com/jamt29/structify/internal/template"
	"github.com/spf13/cobra"
)

var (
	addForce    bool
	gitCloneFunc = git.PlainClone
)

var addCmd = &cobra.Command{
	Use:   "add <source>",
	Short: "Add a template from a local path or Git repository",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		source := args[0]

		// Try to parse as GitHub source first.
		gitSource, err := tmpl.ParseGitSource(source)
		if err == nil {
			return runAddFromGit(cmd, gitSource)
		}

		// Fallback: treat as local path and delegate to store.Add.
		if err := tmpl.Add(source); err != nil {
			return fmt.Errorf("adding local template: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Template added from %s\n", source)
		return nil
	},
}

func init() {
	addCmd.Flags().BoolVar(&addForce, "force", false, "overwrite existing template with the same name")
	Cmd.AddCommand(addCmd)
}

func runAddFromGit(cmd *cobra.Command, src *tmpl.GitSource) error {
	tmpDir, err := os.MkdirTemp("", "structify-template-add-*")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	cloneOpts := &git.CloneOptions{
		URL:      "https://" + src.SourceURL,
		Progress: cmd.ErrOrStderr(),
	}
	if _, err := gitCloneFunc(tmpDir, false, cloneOpts); err != nil {
		return fmt.Errorf("cloning %s: %w", src.SourceURL, err)
	}

	manifestPath := filepath.Join(tmpDir, "scaffold.yaml")
	m, err := dsl.LoadManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("loading scaffold.yaml from cloned repo: %w", err)
	}
	if verrs := dsl.ValidateManifest(m); len(verrs) > 0 {
		return fmt.Errorf("cloned template is invalid")
	}

	name := m.Name
	if name == "" {
		return fmt.Errorf("cloned template has empty name in manifest")
	}

	templatesRoot := tmpl.TemplatesDir()
	if err := os.MkdirAll(templatesRoot, 0o755); err != nil {
		return fmt.Errorf("creating templates dir %s: %w", templatesRoot, err)
	}

	destDir := filepath.Join(templatesRoot, name)
	if _, err := os.Stat(destDir); err == nil {
		if !addForce {
			return fmt.Errorf("template %q already exists (use --force to overwrite)", name)
		}
		if err := os.RemoveAll(destDir); err != nil {
			return fmt.Errorf("removing existing template %q: %w", name, err)
		}
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("stat destination %s: %w", destDir, err)
	}

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("creating destination dir %s: %w", destDir, err)
	}

	if err := tmpl.CopyDirForTest(tmpDir, destDir); err != nil {
		return fmt.Errorf("copying cloned template to store: %w", err)
	}

	meta := &tmpl.TemplateMeta{
		SourceURL:   src.SourceURL,
		SourceRef:   src.Ref,
		InstalledAt: time.Now().UTC().Format(time.RFC3339),
	}
	if err := tmpl.WriteTemplateMeta(destDir, meta); err != nil {
		return fmt.Errorf("writing template metadata: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "✓ Template %q added successfully\n", name)
	return nil
}


