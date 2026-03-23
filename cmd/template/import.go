package template

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/term"
	"gopkg.in/yaml.v3"

	"github.com/jamt29/structify/internal/dsl"
	tmpl "github.com/jamt29/structify/internal/template"
	"github.com/spf13/cobra"
)

var (
	importName string
	importYes  bool
)

type importReview struct {
	vars     []tmpl.DetectedVar
	ignored  []string
	included []string
}

var importCmd = &cobra.Command{
	Use:   "import <source>",
	Short: "Import a local project or GitHub repo as template",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		source := strings.TrimSpace(args[0])
		if source == "" {
			return fmt.Errorf("source is required")
		}

		projectPath, cleanup, err := resolveImportSource(source)
		if err != nil {
			return err
		}
		if cleanup != nil {
			defer cleanup()
		}

		analysis, err := tmpl.AnalyzeProject(projectPath)
		if err != nil {
			return fmt.Errorf("analyzing source project: %w", err)
		}

		name := strings.TrimSpace(importName)
		if name == "" {
			name = analysis.ProjectName
		}
		if name == "" {
			return fmt.Errorf("could not determine template name")
		}

		review := importReview{
			vars:     append([]tmpl.DetectedVar{}, analysis.DetectedVars...),
			ignored:  append([]string{}, analysis.FilesToIgnore...),
			included: append([]string{}, analysis.FilesToInclude...),
		}
		if !importYes && hasTTYImport() {
			// v0.2.0 keeps review simple: confirmation over detected defaults.
			fmt.Fprintf(cmd.OutOrStdout(), "structify · Importar template\n")
			fmt.Fprintf(cmd.OutOrStdout(), "Proyecto detectado: %s\n", analysis.ProjectName)
			fmt.Fprintf(cmd.OutOrStdout(), "Lenguaje: %s · Archivos: %d · Ignorados: %d\n", analysis.Language, len(review.included), len(review.ignored))
			fmt.Fprintln(cmd.OutOrStdout(), "Confirmar importación con valores detectados? [Y/n]")
			var answer string
			fmt.Fscanln(cmd.InOrStdin(), &answer)
			answer = strings.ToLower(strings.TrimSpace(answer))
			if answer == "n" || answer == "no" {
				return fmt.Errorf("import cancelled")
			}
		}

		destRoot := filepath.Join(tmpl.TemplatesDir(), name)
		templateDir := filepath.Join(destRoot, "template")
		if err := os.MkdirAll(templateDir, 0o755); err != nil {
			return fmt.Errorf("creating template destination: %w", err)
		}

		selectedVars := make([]tmpl.DetectedVar, 0, len(review.vars))
		for _, v := range review.vars {
			if strings.TrimSpace(v.ID) != "" && strings.TrimSpace(v.SuggestAs) != "" {
				selectedVars = append(selectedVars, v)
			}
		}

		includedCount, ignoredCount, err := materializeImportedTemplate(projectPath, templateDir, review.ignored, selectedVars)
		if err != nil {
			return err
		}

		manifest := buildImportedManifest(name, analysis.Language, selectedVars)
		manifestBytes, err := yaml.Marshal(manifest)
		if err != nil {
			return fmt.Errorf("marshal scaffold.yaml: %w", err)
		}
		if err := os.WriteFile(filepath.Join(destRoot, "scaffold.yaml"), manifestBytes, 0o644); err != nil {
			return fmt.Errorf("write scaffold.yaml: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "✓ Template '%s' creado en %s/\n", name, destRoot)
		fmt.Fprintf(cmd.OutOrStdout(), "Archivos: %d incluidos, %d ignorados\n", includedCount, ignoredCount)
		fmt.Fprintf(cmd.OutOrStdout(), "Variables: %s\n", joinVarIDs(selectedVars))
		fmt.Fprintf(cmd.OutOrStdout(), "Inputs: %d detectados\n", len(manifest.Inputs))
		fmt.Fprintf(cmd.OutOrStdout(), "Para usarlo:\nstructify new --template %s\n", name)
		fmt.Fprintf(cmd.OutOrStdout(), "Para editarlo:\nstructify template edit %s\n", name)
		return nil
	},
}

func init() {
	importCmd.Flags().StringVar(&importName, "name", "", "Nombre del template (default: nombre de carpeta/repo)")
	importCmd.Flags().BoolVar(&importYes, "yes", false, "Saltar confirmación interactiva")
	Cmd.AddCommand(importCmd)
}

func resolveImportSource(source string) (string, func(), error) {
	if st, err := os.Stat(source); err == nil && st.IsDir() {
		abs, err := filepath.Abs(source)
		if err != nil {
			return "", nil, err
		}
		return abs, nil, nil
	}

	ref, err := tmpl.ParseGitHubURL(source)
	if err == nil {
		tmpDir, mkErr := os.MkdirTemp("", "structify-template-import-*")
		if mkErr != nil {
			return "", nil, fmt.Errorf("creating temp dir: %w", mkErr)
		}
		client := tmpl.NewGitHubClient()
		if clErr := client.Clone(ref, tmpDir); clErr != nil {
			_ = os.RemoveAll(tmpDir)
			return "", nil, clErr
		}
		return tmpDir, func() { _ = os.RemoveAll(tmpDir) }, nil
	}

	return "", nil, fmt.Errorf("source %q is neither a local directory nor a valid github.com URL", source)
}

func materializeImportedTemplate(srcRoot, destTemplateDir string, ignored []string, vars []tmpl.DetectedVar) (int, int, error) {
	ignoredSet := map[string]struct{}{}
	for _, p := range ignored {
		ignoredSet[strings.TrimSuffix(filepath.ToSlash(strings.TrimSpace(p)), "/")] = struct{}{}
	}

	includeCount := 0
	ignoredCount := 0
	err := filepath.WalkDir(srcRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == srcRoot {
			return nil
		}
		rel, err := filepath.Rel(srcRoot, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		if isIgnoredBySet(rel, ignoredSet) {
			if d.IsDir() {
				ignoredCount++
				return filepath.SkipDir
			}
			ignoredCount++
			return nil
		}

		target := filepath.Join(destTemplateDir, filepath.FromSlash(rel))
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		content := string(b)
		replaced := false
		for _, v := range vars {
			if strings.TrimSpace(v.SuggestAs) == "" {
				continue
			}
			tag := "{{ " + v.ID + " }}"
			if strings.Contains(content, v.SuggestAs) {
				content = strings.ReplaceAll(content, v.SuggestAs, tag)
				replaced = true
			}
		}

		if replaced && !strings.HasSuffix(target, ".tmpl") {
			target += ".tmpl"
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
			return err
		}
		includeCount++
		return nil
	})
	if err != nil {
		return 0, 0, fmt.Errorf("copying imported template files: %w", err)
	}
	return includeCount, ignoredCount, nil
}

func isIgnoredBySet(rel string, ignoredSet map[string]struct{}) bool {
	trim := strings.TrimSuffix(rel, "/")
	if _, ok := ignoredSet[trim]; ok {
		return true
	}
	for k := range ignoredSet {
		if strings.HasPrefix(trim, k+"/") {
			return true
		}
	}
	return false
}

func buildImportedManifest(name, lang string, vars []tmpl.DetectedVar) *dsl.Manifest {
	inputs := make([]dsl.Input, 0, len(vars))
	for _, v := range vars {
		inputs = append(inputs, dsl.Input{
			ID:       v.ID,
			Prompt:   v.Description,
			Type:     v.Type,
			Required: true,
			Default:  v.SuggestAs,
		})
	}
	steps := []dsl.Step{}
	switch lang {
	case "go":
		steps = append(steps,
			dsl.Step{Name: "Init module", Run: "go mod init {{ module_path }}"},
			dsl.Step{Name: "Tidy", Run: "go mod tidy"},
		)
	case "typescript", "javascript":
		steps = append(steps, dsl.Step{Name: "Install dependencies", Run: "npm install"})
	case "rust":
		steps = append(steps, dsl.Step{Name: "Build", Run: "cargo build"})
	}

	return &dsl.Manifest{
		Name:         name,
		Version:      "1.0.0",
		Language:     lang,
		Architecture: "unknown",
		Description:  "Imported template",
		Inputs:       inputs,
		Steps:        steps,
	}
}

func joinVarIDs(vars []tmpl.DetectedVar) string {
	if len(vars) == 0 {
		return "-"
	}
	out := make([]string, 0, len(vars))
	for _, v := range vars {
		out = append(out, v.ID)
	}
	return strings.Join(out, ", ")
}

func hasTTYImport() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}
