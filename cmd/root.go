package cmd

import (
	"fmt"
	"os"

	templatecmd "github.com/jamt29/structify/cmd/template"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "structify",
	Short: "Structify scaffolds projects from opinionated templates.",
	Long: "Structify is a CLI to scaffold projects based on software architectures and languages.\n" +
		"Choose an architecture and language, then generate a ready-to-extend project structure.",
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// Only panic in main; here we print and exit with non‑zero status.
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.structify/config.yaml)")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "enable verbose output")

	rootCmd.AddCommand(templatecmd.Cmd)
}



