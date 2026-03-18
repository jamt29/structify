package template

import (
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <name>",
	Short: "Show detailed information about a template",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Fase 1: solo estructura, sin lógica real.
		return nil
	},
}

func init() {
	Cmd.AddCommand(infoCmd)
}


