package template

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jamt29/structify/internal/config"
	tmpl "github.com/jamt29/structify/internal/template"
	"github.com/spf13/cobra"
)

var (
	updateDryRun      bool
	newGitHubClientFn = func() githubClient { return tmpl.NewGitHubClient() }
)

var updateCmd = &cobra.Command{
	Use:   "update [name]",
	Short: "Update one or all templates installed from GitHub",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newGitHubClientFn()

		var templates []*tmpl.Template
		if len(args) == 0 {
			all, err := tmpl.List()
			if err != nil {
				return fmt.Errorf("listing templates: %w", err)
			}
			templates = all
		} else {
			name := strings.TrimSpace(args[0])
			if name == "" {
				return fmt.Errorf("template name is required")
			}
			t, err := tmpl.Get(name)
			if err != nil {
				return err
			}
			templates = []*tmpl.Template{t}
		}

		var updated, skipped, failed int
		structured := config.UseStructuredLogOut(cmd.OutOrStdout())

		for _, t := range templates {
			name := filepath.Base(t.Path)
			if t.Meta == nil || strings.TrimSpace(t.Meta.SourceURL) == "" {
				if structured {
					tmplStructuredLog(cmd).Info("template skipped — not installed from GitHub", "name", name)
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "  → '%s' skipped — not installed from GitHub\n", name)
				}
				skipped++
				continue
			}

			ref, err := tmpl.ParseGitHubURL(t.Meta.SourceURL)
			if err != nil {
				if structured {
					tmplStructuredLog(cmd).Warn("template skipped — invalid source URL", "name", name, "url", t.Meta.SourceURL, "err", err)
				} else {
					fmt.Fprintf(cmd.ErrOrStderr(), "  → '%s' skipped — invalid source URL %q: %v\n", name, t.Meta.SourceURL, err)
				}
				skipped++
				continue
			}
			ref.Ref = t.Meta.SourceRef

			displaySource := t.Meta.SourceURL
			if strings.TrimSpace(t.Meta.SourceRef) != "" {
				displaySource = fmt.Sprintf("%s@%s", t.Meta.SourceURL, t.Meta.SourceRef)
			}

			if structured {
				tmplStructuredLog(cmd).Info("Updating template", "name", name, "source", displaySource)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "  → Updating '%s' from %s...\n", name, displaySource)
			}

			if updateDryRun {
				updated++
				continue
			}

			if err := updateSingleTemplate(cmd, client, t, ref); err != nil {
				if structured {
					tmplStructuredLog(cmd).Error("template update failed", "name", name, "err", err)
				} else {
					fmt.Fprintf(cmd.ErrOrStderr(), "  ✗ '%s' update failed: %v\n", name, err)
				}
				failed++
				continue
			}

			if structured {
				tmplStructuredLog(cmd).Info("template updated", "name", name)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "  ✓ '%s' updated successfully\n", name)
			}
			updated++
		}

		if structured {
			tmplStructuredLog(cmd).Info("update summary", "updated", updated, "skipped", skipped)
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "%d updated, %d skipped.\n", updated, skipped)
		}

		if failed > 0 {
			return fmt.Errorf("%d template updates failed", failed)
		}

		return nil
	},
}

func init() {
	updateCmd.Flags().BoolVar(&updateDryRun, "dry-run", false, "show what would be updated without changing templates")
	Cmd.AddCommand(updateCmd)
}

func updateSingleTemplate(cmd *cobra.Command, client githubClient, t *tmpl.Template, ref *tmpl.GitHubRef) error {
	tmpDir, err := os.MkdirTemp("", "structify-template-update-*")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := client.Clone(ref, tmpDir); err != nil {
		return err
	}

	if _, err := client.ValidateTemplateRepo(tmpDir); err != nil {
		return err
	}

	destDir := t.Path
	backupDir := destDir + ".bak"

	if err := os.RemoveAll(backupDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing old backup dir %s: %w", backupDir, err)
	}

	if err := os.Rename(destDir, backupDir); err != nil {
		return fmt.Errorf("creating backup dir %s: %w", backupDir, err)
	}

	if err := os.Rename(tmpDir, destDir); err != nil {
		_ = os.Rename(backupDir, destDir)
		return fmt.Errorf("replacing template dir %s: %w", destDir, err)
	}

	meta := &tmpl.TemplateMeta{
		SourceURL:   t.Meta.SourceURL,
		SourceRef:   t.Meta.SourceRef,
		InstalledAt: time.Now().UTC().Format(time.RFC3339),
	}
	if err := tmpl.WriteTemplateMeta(destDir, meta); err != nil {
		return fmt.Errorf("writing template metadata: %w", err)
	}

	_ = os.RemoveAll(backupDir)

	return nil
}
