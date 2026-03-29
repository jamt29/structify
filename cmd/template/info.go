package template

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jamt29/structify/internal/engine"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <name>",
	Short: "Mostrar detalle de un template",
	Long: "Muestra metadata, inputs y steps de un template por nombre.\n\n" +
		"Sirve para inspeccionar rapidamente como se comporta un template\n" +
		"antes de usarlo en `structify new`.\n\n" +
		"Ejemplos:\n" +
		"  structify template info clean-architecture-go\n" +
		"  structify template info mi-template",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := strings.TrimSpace(args[0])
		if name == "" {
			return fmt.Errorf("template name is required")
		}

		t, err := engineResolve(name)
		if err != nil {
			return err
		}
		if t == nil || t.Manifest == nil {
			return fmt.Errorf("template %q has no manifest loaded", name)
		}

		out := cmd.OutOrStdout()

		titleStyle := lipgloss.NewStyle().Bold(true)
		labelStyle := lipgloss.NewStyle().Bold(true)

		fmt.Fprintln(out, titleStyle.Render(t.Manifest.Name))
		if t.Manifest.Description != "" {
			fmt.Fprintln(out, t.Manifest.Description)
		}
		fmt.Fprintln(out)

		fmt.Fprintf(out, "%s %s\n", labelStyle.Render("Version:"), orDash(t.Manifest.Version))
		fmt.Fprintf(out, "%s %s\n", labelStyle.Render("Author:"), orDash(t.Manifest.Author))
		fmt.Fprintf(out, "%s %s\n", labelStyle.Render("Language:"), orDash(t.Manifest.Language))
		fmt.Fprintf(out, "%s %s\n", labelStyle.Render("Architecture:"), orDash(t.Manifest.Architecture))
		fmt.Fprintf(out, "%s %s\n", labelStyle.Render("Source:"), orDash(t.Source))

		if len(t.Manifest.Tags) > 0 {
			fmt.Fprintf(out, "%s %s\n", labelStyle.Render("Tags:"), strings.Join(t.Manifest.Tags, ", "))
		}

		fmt.Fprintln(out)

		if len(t.Manifest.Inputs) > 0 {
			fmt.Fprintln(out, titleStyle.Render("Inputs"))
			for _, in := range t.Manifest.Inputs {
				fmt.Fprintf(out, "- %s (%s)\n", in.ID, in.Type)
				if in.Prompt != "" {
					fmt.Fprintf(out, "  %s %s\n", labelStyle.Render("Prompt:"), in.Prompt)
				}
				fmt.Fprintf(out, "  %s %v\n", labelStyle.Render("Required:"), in.Required)
				if in.Default != nil {
					fmt.Fprintf(out, "  %s %v\n", labelStyle.Render("Default:"), in.Default)
				}
				if strings.TrimSpace(in.When) != "" {
					fmt.Fprintf(out, "  %s %s\n", labelStyle.Render("When:"), in.When)
				}
			}
			fmt.Fprintln(out)
		}

		if len(t.Manifest.Steps) > 0 {
			fmt.Fprintln(out, titleStyle.Render("Steps"))
			for _, st := range t.Manifest.Steps {
				fmt.Fprintf(out, "- %s\n", st.Name)
				if st.Run != "" {
					fmt.Fprintf(out, "  %s %s\n", labelStyle.Render("Run:"), st.Run)
				}
				if strings.TrimSpace(st.When) != "" {
					fmt.Fprintf(out, "  %s %s\n", labelStyle.Render("When:"), st.When)
				}
			}
		}

		return nil
	},
}

func init() {
	Cmd.AddCommand(infoCmd)
}

var engineResolve = engine.Resolve

func orDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}
