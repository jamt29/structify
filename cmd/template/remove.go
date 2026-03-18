package template

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	tmpl "github.com/jamt29/structify/internal/template"
	"github.com/spf13/cobra"
)

var removeYes bool

var removeCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a local template",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := strings.TrimSpace(args[0])
		if name == "" {
			return fmt.Errorf("template name is required")
		}

		// Check if there is a local template.
		exists, err := tmpl.Exists(name)
		if err != nil {
			return err
		}
		if !exists {
			// If it's only a built-in, report accordingly.
			allBuiltins, err := tmpl.LoadBuiltins()
			if err == nil {
				for _, t := range allBuiltins {
					if t.Manifest != nil && t.Manifest.Name == name {
						return fmt.Errorf("built-in templates cannot be removed")
					}
				}
			}
			return fmt.Errorf("template %q not found", name)
		}

		if !removeYes {
			if ok, err := confirmRemoval(cmd.InOrStdin(), cmd.OutOrStdout(), name); err != nil {
				return err
			} else if !ok {
				fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
				return nil
			}
		}

		if err := tmpl.Remove(name); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Template %q removed.\n", name)
		return nil
	},
}

func init() {
	removeCmd.Flags().BoolVarP(&removeYes, "yes", "y", false, "remove without confirmation")
	Cmd.AddCommand(removeCmd)
}

func confirmRemoval(r io.Reader, w io.Writer, name string) (bool, error) {
	fmt.Fprintf(w, "Are you sure you want to remove template %q? [y/N]: ", name)
	reader := bufio.NewReader(r)
	line, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("reading confirmation: %w", err)
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes", nil
}

