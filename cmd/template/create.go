package template

import (
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Start a wizard to create a new template",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Fase 1: solo estructura, sin lógica real.
		return nil
	},
}

func init() {
	Cmd.AddCommand(createCmd)
}


