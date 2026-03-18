package template

import (
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a local template",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Fase 1: solo estructura, sin lógica real.
		return nil
	},
}

func init() {
	Cmd.AddCommand(removeCmd)
}


