package template

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"

	"github.com/jamt29/structify/internal/config"
	"github.com/jamt29/structify/internal/dsl"
	"github.com/jamt29/structify/internal/template"
	"github.com/jamt29/structify/internal/tui"
	"github.com/spf13/cobra"
)

var createOutputPath string

const projectNameValidateRegex = "^[a-zA-Z][a-zA-Z0-9_-]*$"

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Start a wizard to create a new template",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		interactive := hasTTY() && !cfg.NonInteractive
		authorDefault := detectGitUserName()

		var name, description, language, architecture, author string

		if interactive {
			wizardInputs := []dsl.Input{
				{
					ID:       "name",
					Prompt:   "Template name?",
					Type:     "string",
					Required: true,
					Validate: projectNameValidateRegex,
				},
				{
					ID:       "description",
					Prompt:   "Description?",
					Type:     "string",
					Required: false,
				},
				{
					ID:       "language",
					Prompt:   "Language?",
					Type:     "enum",
					Options:  []string{"go", "typescript", "rust", "csharp", "python"},
					Default:  "go",
					Required: true,
				},
				{
					ID:       "architecture",
					Prompt:   "Architecture?",
					Type:     "enum",
					Options:  []string{"clean", "vertical-slice", "hexagonal", "mvc", "monorepo"},
					Default:  "clean",
					Required: true,
				},
				{
					ID:       "author",
					Prompt:   "Author?",
					Type:     "string",
					Required: true,
					Default:  authorDefault,
				},
			}

			ctx, err := tui.RunInputs(wizardInputs)
			if err != nil {
				return err
			}

			name = strings.TrimSpace(fmt.Sprint(ctx["name"]))
			description = strings.TrimSpace(fmt.Sprint(ctx["description"]))
			language = strings.TrimSpace(fmt.Sprint(ctx["language"]))
			architecture = strings.TrimSpace(fmt.Sprint(ctx["architecture"]))
			author = strings.TrimSpace(fmt.Sprint(ctx["author"]))
		} else {
			// Fallback non-TUI: allows piping stdin in CI/smoke tests.
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

			name, err = ask("Template name? ")
			if err != nil {
				return fmt.Errorf("reading name: %w", err)
			}
			if err := tui.ValidateInputValue(
				dsl.Input{ID: "name", Type: "string", Required: true, Validate: projectNameValidateRegex},
				name,
			); err != nil {
				return fmt.Errorf("invalid value for %q: %w", "name", err)
			}

			description, err = ask("Description? ")
			if err != nil {
				return fmt.Errorf("reading description: %w", err)
			}

			language, err = ask("Language? (go/typescript/rust/csharp/python) ")
			if err != nil {
				return fmt.Errorf("reading language: %w", err)
			}
			if err := tui.ValidateInputValue(
				dsl.Input{ID: "language", Type: "enum", Options: []string{"go", "typescript", "rust", "csharp", "python"}, Required: true},
				language,
			); err != nil {
				return fmt.Errorf("invalid value for %q: %w", "language", err)
			}

			architecture, err = ask("Architecture? (clean/vertical-slice/hexagonal/mvc/monorepo) ")
			if err != nil {
				return fmt.Errorf("reading architecture: %w", err)
			}
			if err := tui.ValidateInputValue(
				dsl.Input{ID: "architecture", Type: "enum", Options: []string{"clean", "vertical-slice", "hexagonal", "mvc", "monorepo"}, Required: true},
				architecture,
			); err != nil {
				return fmt.Errorf("invalid value for %q: %w", "architecture", err)
			}

			authorPrompt := "Author? "
			if strings.TrimSpace(authorDefault) != "" {
				authorPrompt = fmt.Sprintf("Author? [%s] ", authorDefault)
			}
			author, err = ask(authorPrompt)
			if err != nil {
				return fmt.Errorf("reading author: %w", err)
			}
			author = strings.TrimSpace(author)
			if author == "" {
				author = authorDefault
			}
			if err := tui.ValidateInputValue(
				dsl.Input{ID: "author", Type: "string", Required: true},
				author,
			); err != nil {
				return fmt.Errorf("invalid value for %q: %w", "author", err)
			}
		}

		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("template name cannot be empty")
		}
		if strings.TrimSpace(author) == "" {
			author = authorDefault
		}

		destRoot := createOutputPath
		if strings.TrimSpace(destRoot) == "" {
			destRoot = template.TemplatesDir()
		}
		destDir := filepath.Join(destRoot, name)

		if _, err := os.Stat(destDir); err == nil {
			return fmt.Errorf("template %q already exists at %s", name, destDir)
		} else if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("stat template dir %s: %w", destDir, err)
		}

		if err := os.MkdirAll(destDir, 0o755); err != nil {
			return fmt.Errorf("creating template dir %s: %w", destDir, err)
		}

		if err := writeScaffoldYAML(destDir, name, description, language, architecture, author); err != nil {
			return err
		}

		templateDir := filepath.Join(destDir, "template")
		if err := os.MkdirAll(templateDir, 0o755); err != nil {
			return fmt.Errorf("creating template/ dir: %w", err)
		}
		if err := os.WriteFile(filepath.Join(templateDir, ".gitkeep"), []byte{}, 0o644); err != nil {
			return fmt.Errorf("writing template/.gitkeep: %w", err)
		}

		if config.UseStructuredLogOut(cmd.OutOrStdout()) {
			log := tmplStructuredLog(cmd)
			log.Info("Template created", "name", name, "path", destDir)
			log.Info("Next: add files under template/", "dir", filepath.ToSlash(filepath.Join(destDir, "template"))+"/")
			log.Info("Next: edit scaffold.yaml for inputs and steps")
			log.Info("Next: validate", "cmd", "structify template validate "+filepath.ToSlash(destDir)+"/")
			log.Info("Next: use template", "cmd", "structify new --template "+name)
		} else {
			title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10")).Render(
				fmt.Sprintf("✓ Template '%s' created at %s", name, destDir),
			)
			fmt.Println(title)
			fmt.Println()
			fmt.Println(lipgloss.NewStyle().Bold(true).Render("Next steps"))
			fmt.Println("  1. Add your files to: " + filepath.ToSlash(filepath.Join(destDir, "template")) + "/")
			fmt.Println("  2. Edit scaffold.yaml to add inputs and steps")
			fmt.Println("  3. Test it: structify template validate " + filepath.ToSlash(destDir) + "/")
			fmt.Println("  4. Use it: structify new --template " + name)
		}

		return nil
	},
}

func init() {
	createCmd.Flags().StringVar(&createOutputPath, "output", "", "output directory for the new template")
	Cmd.AddCommand(createCmd)
}

func hasTTY() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}

func detectGitUserName() string {
	out, err := exec.Command("git", "config", "user.name").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func writeScaffoldYAML(dir, name, description, language, architecture, author string) error {
	content := fmt.Sprintf(`name: %q
version: "0.1.0"
author: %q
language: %q
architecture: %q
description: %q
inputs:
  - id: "project_name"
    prompt: "Project name?"
    type: string
    required: true
    validate: %q
steps: []
`, name, author, language, architecture, description, projectNameValidateRegex)

	path := filepath.Join(dir, "scaffold.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing scaffold.yaml: %w", err)
	}
	return nil
}
