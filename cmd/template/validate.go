package template

import (
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate <path>",
	Short: "Validate a template directory or scaffold.yaml file",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Fase 1: solo estructura, sin lógica real.
		return nil
	},
}

func init() {
	Cmd.AddCommand(validateCmd)
}


