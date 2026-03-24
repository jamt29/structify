package template

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jamt29/structify/internal/config"
	"github.com/jamt29/structify/internal/dsl"
	tmpl "github.com/jamt29/structify/internal/template"
	"github.com/spf13/cobra"
)

var (
	addForce        bool
	addName         string
	newGitHubClient = tmpl.NewGitHubClient
)

type githubClient interface {
	Clone(ref *tmpl.GitHubRef, destDir string) error
	ValidateTemplateRepo(clonedPath string) (*dsl.Manifest, error)
}

var addCmd = &cobra.Command{
	Use:   "add <source>",
	Short: "Add a template from a local path or GitHub repository",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		source := args[0]

		// Try to parse as GitHub URL first.
		ref, err := tmpl.ParseGitHubURL(source)
		if err == nil {
			client := newGitHubClient()
			return runAddFromGit(cmd, client, ref)
		}

		// Fallback: treat as local path and delegate to store.Add.
		if err := tmpl.Add(source); err != nil {
			return fmt.Errorf("adding local template: %w", err)
		}
		if config.UseStructuredLogOut(cmd.OutOrStdout()) {
			tmplStructuredLog(cmd).Info("Template added", "source", source)
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "✓ Template added from %s\n", source)
		}
		return nil
	},
}

func init() {
	addCmd.Flags().BoolVar(&addForce, "force", false, "overwrite existing template with the same name")
	addCmd.Flags().StringVar(&addName, "name", "", "local name to use for the installed template (defaults to repository name)")
	Cmd.AddCommand(addCmd)
}

func runAddFromGit(cmd *cobra.Command, client githubClient, ref *tmpl.GitHubRef) error {
	tmpDir, err := os.MkdirTemp("", "structify-template-add-*")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	if config.UseStructuredLogOut(cmd.OutOrStdout()) {
		tmplStructuredLog(cmd).Info("Fetching template from GitHub", "owner", ref.Owner, "repo", ref.Repo)
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "  → Fetching template from github.com/%s/%s...\n", ref.Owner, ref.Repo)
	}

	if err := client.Clone(ref, tmpDir); err != nil {
		return err
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

	localName := addName
	if strings.TrimSpace(localName) == "" {
		localName = ref.Repo
	}

	templatesRoot := tmpl.TemplatesDir()
	if err := os.MkdirAll(templatesRoot, 0o755); err != nil {
		return fmt.Errorf("creating templates dir %s: %w", templatesRoot, err)
	}

	destDir := filepath.Join(templatesRoot, localName)
	if _, err := os.Stat(destDir); err == nil {
		if !addForce {
			return fmt.Errorf("template %q already exists (use --force to overwrite)", localName)
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

	if err := tmpl.CopyDirForTest(tmpDir, destDir); err != nil {
		return fmt.Errorf("copying cloned template to store: %w", err)
	}

	meta := &tmpl.TemplateMeta{
		SourceURL:   fmt.Sprintf("github.com/%s/%s", ref.Owner, ref.Repo),
		SourceRef:   ref.Ref,
		InstalledAt: time.Now().UTC().Format(time.RFC3339),
	}
	if err := tmpl.WriteTemplateMeta(destDir, meta); err != nil {
		return fmt.Errorf("writing template metadata: %w", err)
	}

	if config.UseStructuredLogOut(cmd.OutOrStdout()) {
		log := tmplStructuredLog(cmd)
		log.Info("template metadata", "name", m.Name, "language", m.Language, "architecture", m.Architecture)
		log.Info("Template installed", "localName", localName)
		log.Info("Run structify new", "template", localName)
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "  ✓ Found: %s (%s / %s)\n", m.Name, m.Language, m.Architecture)
		fmt.Fprintf(cmd.OutOrStdout(), "  ✓ Template '%s' installed successfully\n", localName)
		fmt.Fprintf(cmd.OutOrStdout(), "  Run: structify new --template %s\n", localName)
	}
	return nil
}
