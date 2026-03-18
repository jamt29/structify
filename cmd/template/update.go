package template

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/jamt29/structify/internal/dsl"
	tmpl "github.com/jamt29/structify/internal/template"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update <name>",
	Short: "Update a template from its original source",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := strings.TrimSpace(args[0])
		if name == "" {
			return fmt.Errorf("template name is required")
		}

		t, err := tmpl.Get(name)
		if err != nil {
			return err
		}
		if t.Meta == nil || strings.TrimSpace(t.Meta.SourceURL) == "" {
			return fmt.Errorf("Template %q was not installed from GitHub", name)
		}

		oldVersion := ""
		if t.Manifest != nil {
			oldVersion = t.Manifest.Version
		}

		tmpDir, err := os.MkdirTemp("", "structify-template-update-*")
		if err != nil {
			return fmt.Errorf("creating temp dir: %w", err)
		}
		defer os.RemoveAll(tmpDir)

		cloneOpts := &git.CloneOptions{
			URL: "https://" + t.Meta.SourceURL,
		}
		if t.Meta.SourceRef != "" {
			cloneOpts.ReferenceName = ""
		}
		if _, err := gitCloneFunc(tmpDir, false, cloneOpts); err != nil {
			return fmt.Errorf("cloning %s: %w", t.Meta.SourceURL, err)
		}

		manifestPath := filepath.Join(tmpDir, "scaffold.yaml")
		m, err := dsl.LoadManifest(manifestPath)
		if err != nil {
			return fmt.Errorf("loading scaffold.yaml from cloned repo: %w", err)
		}
		if verrs := dsl.ValidateManifest(m); len(verrs) > 0 {
			return fmt.Errorf("cloned template is invalid")
		}

		destDir := t.Path
		entries, err := os.ReadDir(destDir)
		if err != nil {
			return fmt.Errorf("reading destination dir %s: %w", destDir, err)
		}
		for _, e := range entries {
			name := e.Name()
			if name == "." || name == ".." {
				continue
			}
			if err := os.RemoveAll(filepath.Join(destDir, name)); err != nil {
				return fmt.Errorf("clearing destination %s: %w", destDir, err)
			}
		}

		if err := tmpl.CopyDirForTest(tmpDir, destDir); err != nil {
			return fmt.Errorf("copying updated template to store: %w", err)
		}

		meta := &tmpl.TemplateMeta{
			SourceURL:   t.Meta.SourceURL,
			SourceRef:   t.Meta.SourceRef,
			InstalledAt: time.Now().UTC().Format(time.RFC3339),
		}
		if err := tmpl.WriteTemplateMeta(destDir, meta); err != nil {
			return fmt.Errorf("writing template metadata: %w", err)
		}

		out := cmd.OutOrStdout()
		newVersion := m.Version
		if oldVersion != "" && newVersion != "" && oldVersion != newVersion {
			fmt.Fprintf(out, "Updated %q from %s to %s\n", name, oldVersion, newVersion)
		} else {
			fmt.Fprintf(out, "Template %q is already up to date\n", name)
		}

		return nil
	},
}

func init() {
	Cmd.AddCommand(updateCmd)
}

