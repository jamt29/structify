package template

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jamt29/structify/internal/dsl"
	tmpl "github.com/jamt29/structify/internal/template"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit <name>",
	Short: "Edit scaffold.yaml of a local template",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := strings.TrimSpace(args[0])
		if name == "" {
			return fmt.Errorf("template name is required")
		}

		exists, err := tmpl.Exists(name)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("template %q not found", name)
		}

		manifestPath := filepath.Join(tmpl.TemplatesDir(), name, "scaffold.yaml")
		original, err := os.ReadFile(manifestPath)
		if err != nil {
			return fmt.Errorf("reading scaffold.yaml: %w", err)
		}

		for {
			if err := openInEditor(manifestPath); err != nil {
				return err
			}

			m, err := dsl.LoadManifest(manifestPath)
			if err == nil {
				verrs := dsl.ValidateManifest(m)
				if len(verrs) == 0 {
					fmt.Fprintln(cmd.OutOrStdout(), "✓ scaffold.yaml actualizado y válido")
					return nil
				}
				if !handleEditValidationErrors(cmd, manifestPath, original, verrs) {
					return nil
				}
				continue
			}

			parseErr := []dsl.ValidationError{{Field: "manifest", Message: err.Error()}}
			if !handleEditValidationErrors(cmd, manifestPath, original, parseErr) {
				return nil
			}
		}
	},
}

func init() {
	Cmd.AddCommand(editCmd)
}

func openInEditor(path string) error {
	editor := os.Getenv("EDITOR")
	if strings.TrimSpace(editor) == "" {
		for _, e := range []string{"vim", "nano", "vi"} {
			if _, err := exec.LookPath(e); err == nil {
				editor = e
				break
			}
		}
	}
	if strings.TrimSpace(editor) == "" {
		return fmt.Errorf("no editor found: set $EDITOR environment variable")
	}

	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func handleEditValidationErrors(cmd *cobra.Command, manifestPath string, original []byte, verrs []dsl.ValidationError) bool {
	errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	fmt.Fprintln(cmd.OutOrStdout(), errStyle.Render("El scaffold.yaml tiene errores:"))
	for _, ve := range verrs {
		fmt.Fprintf(cmd.OutOrStdout(), "· %s: %s\n", ve.Field, ve.Message)
	}
	fmt.Fprintln(cmd.OutOrStdout(), "\n¿Qué deseas hacer?")
	fmt.Fprintln(cmd.OutOrStdout(), "1) Volver a editar")
	fmt.Fprintln(cmd.OutOrStdout(), "2) Guardar de todas formas (no recomendado)")
	fmt.Fprintln(cmd.OutOrStdout(), "3) Descartar cambios")
	fmt.Fprint(cmd.OutOrStdout(), "> ")

	var choice string
	_, _ = fmt.Fscanln(cmd.InOrStdin(), &choice)
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		return true
	case "2":
		fmt.Fprintln(cmd.OutOrStdout(), "⚠ scaffold.yaml guardado con errores")
		return false
	case "3":
		if err := os.WriteFile(manifestPath, original, 0o644); err != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "No se pudo restaurar el archivo original: %v\n", err)
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "Cambios descartados")
		}
		return false
	default:
		fmt.Fprintln(cmd.OutOrStdout(), "Opción inválida. Se vuelve a abrir el editor.")
		return true
	}
}
