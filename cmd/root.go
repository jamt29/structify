package cmd

import (
	"fmt"
	"os"
	"strings"

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

	runMenuFn         = tui.RunMenu
	runAppFn          = tui.RunApp
	resolveAllFn      = resolveAllTemplates
	runTemplateListFn = runTemplateList
	loadConfigFn      = config.Load
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
	action, err := runMenuFn()
	if err != nil {
		if err == tui.ErrMenuExit {
			return nil
		}
		return err
	}

	switch action {
	case tui.ActionNew:
		templates, err := resolveAllFn()
		if err != nil {
			return err
		}
		return runAppFn(templates, engine.New())
	case tui.ActionTemplates:
		return runTemplateListFn()
	case tui.ActionGitHub:
		fmt.Println("Usa: structify template add github.com/<user>/<repo>")
		return nil
	case tui.ActionConfig:
		cfg, err := loadConfigFn()
		if err != nil {
			return err
		}
		if strings.TrimSpace(cfg.ConfigFile) == "" {
			fmt.Printf("Config: %s\n", cfg.ConfigDir+"/config.yaml")
			return nil
		}
		fmt.Printf("Config: %s\n", cfg.ConfigFile)
		return nil
	default:
		return nil
	}
}

func resolveAllTemplates() ([]*template.Template, error) {
	all, err := engine.ListAll()
	if err != nil {
		return nil, fmt.Errorf("listing templates: %w", err)
	}
	return all, nil
}

func runTemplateList() error {
	all, err := resolveAllFn()
	if err != nil {
		return err
	}
	if len(all) == 0 {
		fmt.Println("No templates found.")
		return nil
	}
	fmt.Println("Templates disponibles:")
	for _, t := range all {
		if t == nil || t.Manifest == nil {
			continue
		}
		fmt.Printf("- %s (%s)\n", t.Manifest.Name, t.Source)
	}
	return nil
}


