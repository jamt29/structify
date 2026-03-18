package template

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"text/tabwriter"

	"github.com/jamt29/structify/internal/engine"
	"github.com/spf13/cobra"
)

var (
	listJSON     = false
	engineListAll = engine.ListAll
)

type listTemplate struct {
	Name         string `json:"name"`
	Language     string `json:"language"`
	Architecture string `json:"architecture"`
	Description  string `json:"description"`
	Source       string `json:"source"`
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available templates",
	RunE: func(cmd *cobra.Command, args []string) error {
		templates, err := engineListAll()
		if err != nil {
			return fmt.Errorf("listing templates: %w", err)
		}

		if len(templates) == 0 {
			if listJSON {
				return printListJSON(cmd.OutOrStdout(), nil)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "No templates found.")
			fmt.Fprintln(cmd.OutOrStdout(), "You can add one with: structify template add <github-url>")
			return nil
		}

		locals := make([]listTemplate, 0)
		builtins := make([]listTemplate, 0)

		for _, t := range templates {
			if t == nil || t.Manifest == nil {
				continue
			}
			item := listTemplate{
				Name:         t.Manifest.Name,
				Language:     t.Manifest.Language,
				Architecture: t.Manifest.Architecture,
				Description:  t.Manifest.Description,
				Source:       t.Source,
			}
			switch t.Source {
			case "local":
				locals = append(locals, item)
			case "builtin":
				builtins = append(builtins, item)
			default:
				// Treat unknown sources as local for display purposes.
				locals = append(locals, item)
			}
		}

		sort.Slice(locals, func(i, j int) bool { return locals[i].Name < locals[j].Name })
		sort.Slice(builtins, func(i, j int) bool { return builtins[i].Name < builtins[j].Name })

		if listJSON {
			all := make([]listTemplate, 0, len(locals)+len(builtins))
			all = append(all, locals...)
			all = append(all, builtins...)
			return printListJSON(cmd.OutOrStdout(), all)
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)

		printGroup := func(title string, items []listTemplate) {
			if len(items) == 0 {
				return
			}
			fmt.Fprintln(w, title+":")
			fmt.Fprintln(w, "Name\tLanguage\tArchitecture\tDescription")
			for _, it := range items {
				arch := it.Architecture
				if arch == "" {
					arch = "-"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", it.Name, it.Language, arch, it.Description)
			}
			fmt.Fprintln(w)
		}

		printGroup("Local templates", locals)
		printGroup("Built-in templates", builtins)

		return w.Flush()
	},
}

func printListJSON(w io.Writer, items []listTemplate) error {
	if items == nil {
		items = []listTemplate{}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(items)
}

func init() {
	listCmd.Flags().BoolVar(&listJSON, "json", false, "print templates as JSON")
	Cmd.AddCommand(listCmd)
}

