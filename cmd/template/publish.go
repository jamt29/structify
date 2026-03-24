package template

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamt29/structify/internal/config"
	"github.com/jamt29/structify/internal/dsl"
	"github.com/spf13/cobra"
)

var publishCmd = &cobra.Command{
	Use:   "publish [path]",
	Short: "Run the checklist to publish a template",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "."
		if len(args) == 1 && strings.TrimSpace(args[0]) != "" {
			dir = args[0]
		}

		manifestPath := filepath.Join(dir, "scaffold.yaml")
		out := cmd.OutOrStdout()
		structured := config.UseStructuredLogOut(out)
		wln := func(s string) {
			if structured {
				tmplStructuredLog(cmd).Info(s)
			} else {
				fmt.Fprintln(out, s)
			}
		}
		wf := func(format string, args ...interface{}) {
			if structured {
				tmplStructuredLog(cmd).Info(fmt.Sprintf(format, args...))
			} else {
				fmt.Fprintf(out, format, args...)
			}
		}

		// 1) scaffold.yaml exists
		exists := true
		if _, err := os.Stat(manifestPath); err != nil {
			if os.IsNotExist(err) {
				wln("[✗] scaffold.yaml exists")
				exists = false
			} else {
				return fmt.Errorf("stat scaffold.yaml: %w", err)
			}
		} else {
			wln("[✓] scaffold.yaml exists")
		}

		// 2) scaffold.yaml is valid
		var manifest *dsl.Manifest
		valid := false
		if exists {
			m, err := dsl.LoadManifest(manifestPath)
			if err != nil {
				wf("[✗] scaffold.yaml is valid (%v)\n", err)
			} else {
				if verrs := dsl.ValidateManifest(m); len(verrs) > 0 {
					wln("[✗] scaffold.yaml is valid (validation errors present)")
				} else {
					wf("[✓] scaffold.yaml is valid (%d inputs, %d steps)\n", len(m.Inputs), len(m.Steps))
					valid = true
					manifest = m
				}
			}
		} else {
			wln("[✗] scaffold.yaml is valid (missing file)")
		}

		// 3) README exists
		readmeOK := false
		readmePaths := []string{
			filepath.Join(dir, "README.md"),
			filepath.Join(dir, "template", "README.md"),
		}
		for _, p := range readmePaths {
			if _, err := os.Stat(p); err == nil {
				readmeOK = true
				break
			}
		}
		if readmeOK {
			wln("[✓] README.md exists")
		} else {
			wln("[✗] README.md is missing — add documentation for your template")
		}

		// 4) template/ has files
		templateDir := filepath.Join(dir, "template")
		hasFiles := false
		if entries, err := os.ReadDir(templateDir); err == nil {
			for _, e := range entries {
				if !e.IsDir() {
					hasFiles = true
					break
				}
				sub := filepath.Join(templateDir, e.Name())
				_ = filepath.WalkDir(sub, func(_ string, d os.DirEntry, _ error) error {
					if !d.IsDir() {
						hasFiles = true
						return fmt.Errorf("stop")
					}
					return nil
				})
				if hasFiles {
					break
				}
			}
		}
		if hasFiles {
			wln("[✓] template/ directory has files")
		} else {
			wln("[✗] template/ directory has no files")
		}

		// 5) version is reasonable
		versionOK := true
		if manifest != nil {
			v := strings.TrimSpace(manifest.Version)
			if v == "" || v == "0.0.0" {
				wln("[✗] version field looks default — consider bumping before publishing")
				versionOK = false
			} else {
				wf("[✓] version field is %q\n", v)
			}
		} else {
			wln("[✗] version field cannot be checked (invalid manifest)")
			versionOK = false
		}

		wln("")
		wln("To share your template, push it to a public GitHub repo.")
		wln("Others can then install it with:")
		wln("  structify template add github.com/<your-user>/<repo-name>")

		// Critical failures affect exit code.
		if !exists || !valid || !hasFiles {
			_ = versionOK
			return fmt.Errorf("template does not meet all critical checklist items")
		}

		return nil
	},
}

func init() {
	Cmd.AddCommand(publishCmd)
}
