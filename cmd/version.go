package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// These values are intended to be overridden at build time via -ldflags.
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of structify",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "structify %s (commit %s, built at %s)\n", Version, Commit, Date)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}


