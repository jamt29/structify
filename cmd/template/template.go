package template

import (
	"github.com/spf13/cobra"
)

// Cmd is the base command for template management: `structify template`.
var Cmd = &cobra.Command{
	Use:   "template",
	Short: "Manage Structify templates",
	Long:  "Manage Structify templates: list, add, import, create, edit, validate, remove, inspect, update, and publish templates.",
}
