package cmd

import (
	"fmt"

	"github.com/jamt29/structify/internal/buildinfo"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Mostrar version del binario actual",
	Long: "Muestra la version de Structify junto con commit y fecha de build.\n\n" +
		"Ejemplo:\n" +
		"  structify version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "structify %s (commit %s, built at %s)\n", buildinfo.Version, buildinfo.Commit, buildinfo.Date)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}


