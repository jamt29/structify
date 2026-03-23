package cmd

import (
	"fmt"
	"os"

	templatecmd "github.com/jamt29/structify/cmd/template"
	"github.com/jamt29/structify/internal/config"
	"github.com/jamt29/structify/internal/engine"
	"github.com/jamt29/structify/internal/template"
	"github.com/jamt29/structify/internal/tui"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	verbose bool

	runRootFn = tui.Run
)

var rootCmd = &cobra.Command{
	Use:   "structify",
	Short: "Structify scaffolds projects from opinionated templates.",
	Long: "Structify is a CLI to scaffold projects based on software architectures and languages.\n" +
		"Choose an architecture and language, then generate a ready-to-extend project structure.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInteractive()
	},
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

func runInteractive() error {
	// Asegura que el setup de config/dirs exista antes de renderizar.
	_, _ = config.Load()

	templates, err := resolveAllTemplates()
	if err != nil {
		return err
	}
	return runRootFn(templates, engine.New())
}

func resolveAllTemplates() ([]*template.Template, error) {
	all, err := engine.ListAll()
	if err != nil {
		return nil, fmt.Errorf("listing templates: %w", err)
	}
	return all, nil
}


