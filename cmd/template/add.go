package template

import (
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <source>",
	Short: "Add a template from a local path or Git repository",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Fase 1: solo estructura, sin lógica real.
		return nil
	},
}

func init() {
	Cmd.AddCommand(addCmd)
}


