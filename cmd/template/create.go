package template

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jamt29/structify/internal/template"
	"github.com/spf13/cobra"
)

var createOutputPath string

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Start a wizard to create a new template",
	RunE: func(cmd *cobra.Command, args []string) error {
		reader := bufio.NewReader(cmd.InOrStdin())
		out := cmd.OutOrStdout()

		ask := func(prompt string) (string, error) {
			fmt.Fprint(out, prompt)
			line, err := reader.ReadString('\n')
			if err != nil {
				return "", err
			}
			return strings.TrimSpace(line), nil
		}

		name, err := ask("Template name: ")
		if err != nil {
			return fmt.Errorf("reading name: %w", err)
		}
		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("template name cannot be empty")
		}

		description, err := ask("Description: ")
		if err != nil {
			return fmt.Errorf("reading description: %w", err)
		}

		language, err := ask("Language (go/typescript/rust/...): ")
		if err != nil {
			return fmt.Errorf("reading language: %w", err)
		}

		architecture, err := ask("Architecture (clean/vertical-slice/...): ")
		if err != nil {
			return fmt.Errorf("reading architecture: %w", err)
		}

		authorDefault := detectGitUserName()
		authorPrompt := "Author"
		if authorDefault != "" {
			authorPrompt += fmt.Sprintf(" [%s]", authorDefault)
		}
		authorPrompt += ": "
		author, err := ask(authorPrompt)
		if err != nil {
			return fmt.Errorf("reading author: %w", err)
		}
		if strings.TrimSpace(author) == "" {
			author = authorDefault
		}

		destRoot := createOutputPath
		if strings.TrimSpace(destRoot) == "" {
			destRoot = template.TemplatesDir()
		}
		destDir := filepath.Join(destRoot, name)

		if err := os.MkdirAll(destDir, 0o755); err != nil {
			return fmt.Errorf("creating template dir %s: %w", destDir, err)
		}

		if err := writeScaffoldYAML(destDir, name, description, language, architecture, author); err != nil {
			return err
		}

		if err := os.MkdirAll(filepath.Join(destDir, "template"), 0o755); err != nil {
			return fmt.Errorf("creating template/ dir: %w", err)
		}

		readmePath := filepath.Join(destDir, "template", "README.md.tmpl")
		readmeContent := "# {{ project_name }}\n\nGenerated from Structify template \"" + name + "\".\n"
		if err := os.WriteFile(readmePath, []byte(readmeContent), 0o644); err != nil {
			return fmt.Errorf("writing README.md.tmpl: %w", err)
		}

		fmt.Fprintf(out, "Template created at %s\n", destDir)
		fmt.Fprintln(out, "You can now add files under the template/ directory.")
		fmt.Fprintf(out, "Then run: structify new --template %s --dry-run\n", name)

		return nil
	},
}

func init() {
	createCmd.Flags().StringVar(&createOutputPath, "output", "", "output directory for the new template")
	Cmd.AddCommand(createCmd)
}

func detectGitUserName() string {
	out, err := exec.Command("git", "config", "user.name").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func writeScaffoldYAML(dir, name, description, language, architecture, author string) error {
	content := "" +
		"name: \"" + name + "\"\n" +
		"version: \"0.1.0\"\n" +
		"author: \"" + author + "\"\n" +
		"language: \"" + language + "\"\n" +
		"architecture: \"" + architecture + "\"\n" +
		"description: \"" + description + "\"\n" +
		"tags: []\n" +
		"inputs: []\n" +
		"files: []\n" +
		"steps: []\n"
	path := filepath.Join(dir, "scaffold.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing scaffold.yaml: %w", err)
	}
	return nil
}

