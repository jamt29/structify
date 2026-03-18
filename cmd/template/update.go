package template

import (
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update <name>",
	Short: "Update a template from its original source",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Fase 1: solo estructura, sin lógica real.
		return nil
	},
}

func init() {
	Cmd.AddCommand(updateCmd)
}


