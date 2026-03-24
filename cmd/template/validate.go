package template

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamt29/structify/internal/config"
	"github.com/jamt29/structify/internal/dsl"
	"github.com/spf13/cobra"
)

var validateJSON bool

type validateError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type validateResult struct {
	Valid  bool            `json:"valid"`
	Errors []validateError `json:"errors,omitempty"`
}

var validateCmd = &cobra.Command{
	Use:   "validate <path>",
	Short: "Validate a template directory or scaffold.yaml file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := strings.TrimSpace(args[0])
		if path == "" {
			return fmt.Errorf("path is required")
		}

		manifestPath, err := resolveManifestPath(path)
		if err != nil {
			return err
		}

		m, err := dsl.LoadManifest(manifestPath)
		if err != nil {
			if validateJSON {
				res := validateResult{
					Valid: false,
					Errors: []validateError{
						{Field: "manifest", Message: err.Error()},
					},
				}
				return printValidateJSON(cmd.OutOrStdout(), res)
			}
			return err
		}

		verrs := dsl.ValidateManifest(m)
		if validateJSON {
			res := validateResult{Valid: len(verrs) == 0}
			for _, ve := range verrs {
				res.Errors = append(res.Errors, validateError{
					Field:   ve.Field,
					Message: ve.Message,
				})
			}
			return printValidateJSON(cmd.OutOrStdout(), res)
		}

		out := cmd.OutOrStdout()
		structured := config.UseStructuredLogOut(out)
		if len(verrs) == 0 {
			if structured {
				log := tmplStructuredLog(cmd)
				log.Info("Template is valid")
				log.Info("manifest summary", "inputs", len(m.Inputs), "steps", len(m.Steps), "file_rules", len(m.Files))
			} else {
				fmt.Fprintln(out, "✓ Template is valid")
				fmt.Fprintf(out, "Inputs: %d, Steps: %d, File rules: %d\n", len(m.Inputs), len(m.Steps), len(m.Files))
			}
			return nil
		}

		for _, ve := range verrs {
			if structured {
				tmplStructuredLog(cmd).Error("validation error", "field", ve.Field, "message", ve.Message)
			} else {
				fmt.Fprintf(out, "- %s: %s\n", ve.Field, ve.Message)
			}
		}
		return fmt.Errorf("template has %d validation error(s)", len(verrs))
	},
}

func init() {
	validateCmd.Flags().BoolVar(&validateJSON, "json", false, "print validation result as JSON")
	Cmd.AddCommand(validateCmd)
}

func resolveManifestPath(p string) (string, error) {
	info, err := os.Stat(p)
	if err != nil {
		return "", fmt.Errorf("stat path %s: %w", p, err)
	}
	if info.IsDir() {
		return filepath.Join(p, "scaffold.yaml"), nil
	}
	return p, nil
}

func printValidateJSON(w io.Writer, res validateResult) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(res)
}
