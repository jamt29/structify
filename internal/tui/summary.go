package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/jamt29/structify/internal/template"
)

// ShowSummary prints a final summary after successful generation.
func ShowSummary(result *template.ScaffoldResult, req *template.ScaffoldRequest) {
	if result == nil || req == nil || req.Template == nil || req.Template.Manifest == nil {
		return
	}

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10")).Render("✓ Project created successfully")
	kvKey := lipgloss.NewStyle().Bold(true)

	outAbs, _ := filepath.Abs(req.OutputDir)
	if outAbs == "" {
		outAbs = req.OutputDir
	}

	fmt.Println(title)
	fmt.Println()
	fmt.Println(kvKey.Render("Path") + " : " + outAbs)
	fmt.Println(kvKey.Render("Files") + " : " + fmt.Sprintf("%d", len(result.FilesCreated)))
	fmt.Println()

	if len(result.StepsExecuted) > 0 {
		fmt.Println(kvKey.Render("Steps"))
		for _, s := range result.StepsExecuted {
			if s.Skipped {
				fmt.Printf("  ─ %s (skipped)\n", s.Name)
				continue
			}
			if s.Error != nil {
				fmt.Printf("  ✗ %s (%s)\n", s.Name, s.Error.Error())
				continue
			}
			fmt.Printf("  ✓ %s\n", s.Name)
		}
		fmt.Println()
	}

	fmt.Println(kvKey.Render("Next steps"))
	for _, line := range nextSteps(req.Template.Manifest.Language, ctxString(req.Variables, "project_name")) {
		fmt.Println("  " + line)
	}
}

func ctxString(ctx map[string]interface{}, key string) string {
	if ctx == nil {
		return ""
	}
	v, ok := ctx[key]
	if !ok || v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func nextSteps(language string, projectName string) []string {
	name := strings.TrimSpace(projectName)
	if name == "" {
		name = "<name>"
	}
	switch strings.ToLower(strings.TrimSpace(language)) {
	case "go":
		return []string{
			"cd " + name,
			"go run ./cmd/...",
		}
	case "typescript", "ts", "node":
		return []string{
			"cd " + name,
			"npm install",
			"npm run dev",
		}
	case "rust":
		return []string{
			"cd " + name,
			"cargo run",
		}
	default:
		return []string{
			"cd " + name,
		}
	}
}

