package cmd

import (
	"fmt"

	"github.com/jamt29/structify/internal/buildinfo"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of structify",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "structify %s (commit %s, built at %s)\n", buildinfo.Version, buildinfo.Commit, buildinfo.Date)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}


