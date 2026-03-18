package cmd

import (
	"github.com/spf13/cobra"
)

var (
	newTemplate string
	newName     string
	newVars     []string
	newDryRun   bool
	newOutput   string
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new project from a template",
	Long: "Create a new project from a Structify template.\n" +
		"Use flags to select the template, project name, additional variables, and output directory.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Fase 1: solo estructura, sin lógica de scaffolding aún.
		return nil
	},
}

func init() {
	rootCmd.AddCommand(newCmd)

	newCmd.Flags().StringVar(&newTemplate, "template", "", "template name or path to use")
	newCmd.Flags().StringVar(&newName, "name", "", "name of the project to create")
	newCmd.Flags().StringArrayVar(&newVars, "var", nil, "additional variables in key=value form (repeatable)")
	newCmd.Flags().BoolVar(&newDryRun, "dry-run", false, "show what would be generated without writing files")
	newCmd.Flags().StringVar(&newOutput, "output", "", "output directory for the generated project")
}


