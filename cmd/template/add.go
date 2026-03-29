package template

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jamt29/structify/internal/config"
	"github.com/jamt29/structify/internal/dsl"
	tmpl "github.com/jamt29/structify/internal/template"
	"github.com/spf13/cobra"
)

var (
	addForce bool
	addName  string
	newGitHubClient = tmpl.NewGitHubClient
)

type githubClient interface {
	Clone(ref *tmpl.GitHubRef, destDir string) error
	ValidateTemplateRepo(clonedPath string) (*dsl.Manifest, error)
}

var addCmd = &cobra.Command{
	Use:   "add <source>",
	Short: "Instalar un template desde ruta local o GitHub",
	Long: "Instala un template en el store local de Structify.\n\n" +
		"<source> puede ser una ruta local o una URL de GitHub\n" +
		"(por ejemplo github.com/user/repo o github.com/user/repo@v1.2.0).\n\n" +
		"Ejemplos:\n" +
		"  structify template add ./mi-template\n" +
		"  structify template add github.com/user/repo\n" +
		"  structify template add github.com/user/repo@v1.2.0 --name mi-template",
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
	addCmd.Flags().BoolVar(&addForce, "force", false, "sobrescribir template local existente con el mismo nombre")
	addCmd.Flags().StringVar(&addName, "name", "", "nombre local para guardar el template instalado (default: nombre del repo)")
	Cmd.AddCommand(addCmd)
}

func runAddFromGit(cmd *cobra.Command, client githubClient, ref *tmpl.GitHubRef) error {
	if config.UseStructuredLogOut(cmd.OutOrStdout()) {
		tmplStructuredLog(cmd).Info("Fetching template from GitHub", "owner", ref.Owner, "repo", ref.Repo)
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "  → Fetching template from github.com/%s/%s...\n", ref.Owner, ref.Repo)
	}

	if err := tmpl.InstallFromGitHub(client, ref, tmpl.InstallFromGitHubOptions{
		Force:     addForce,
		LocalName: addName,
	}); err != nil {
		return err
	}

	localName := strings.TrimSpace(addName)
	if localName == "" {
		localName = ref.Repo
	}

	manifestPath := filepath.Join(tmpl.TemplatesDir(), localName, "scaffold.yaml")
	m, err := dsl.LoadManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("loading installed manifest: %w", err)
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
