package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/term"

	"github.com/jamt29/structify/internal/config"
	"github.com/jamt29/structify/internal/dsl"
	"github.com/jamt29/structify/internal/engine"
	"github.com/jamt29/structify/internal/template"
	"github.com/jamt29/structify/internal/tui"
	"github.com/spf13/cobra"
)

var (
	newTemplate string
	newName     string
	newVars     []string
	newDryRun   bool
	newOutput   string
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new project from a template",
	Long: "Create a new project from a Structify template.\n" +
		"Use flags to select the template, project name, additional variables, and output directory.",
	RunE: runNew,
}

func init() {
	rootCmd.AddCommand(newCmd)

	newCmd.Flags().StringVar(&newTemplate, "template", "", "template name or path to use")
	newCmd.Flags().StringVar(&newName, "name", "", "name of the project to create")
	newCmd.Flags().StringArrayVar(&newVars, "var", nil, "additional variables in key=value form (repeatable)")
	newCmd.Flags().BoolVar(&newDryRun, "dry-run", false, "show what would be generated without writing files")
	newCmd.Flags().StringVar(&newOutput, "output", "", "output directory for the generated project")
}

func runNew(cmd *cobra.Command, args []string) error {
	// 1. Load config (ensures ~/.structify dirs exist).
	if _, err := config.Load(); err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// 2. Resolve templates.
	all, err := engine.ListAll()
	if err != nil {
		return fmt.Errorf("listing templates: %w", err)
	}

	interactive := hasTTY() && !configIsNonInteractive()

	if interactive {
		eng := engine.New()
		return tui.RunApp(all, eng)
	}

	// 3. Select template (flags-only / no TTY).
	tpl, err := resolveTemplate(newTemplate, all, false)
	if err != nil {
		return err
	}

	// 4. Gather inputs (flags and/or TUI).
	ctx, err := buildInitialContext(newName, newVars)
	if err != nil {
		return err
	}

	manifestInputs := []dsl.Input{}
	if tpl != nil && tpl.Manifest != nil {
		manifestInputs = tpl.Manifest.Inputs
	}

	// No TTY => flags-only mode.
	if strings.TrimSpace(newTemplate) == "" {
		return fmt.Errorf("no TTY detected: --template is required")
	}
	ctx, err = coerceProvidedVarsToTypes(manifestInputs, ctx)
	if err != nil {
		return err
	}
	ctx, err = finalizeContextNonInteractive(manifestInputs, ctx)
	if err != nil {
		return err
	}
	if tpl != nil && tpl.Manifest != nil {
		if err := applyComputedValues(tpl.Manifest.Computed, ctx); err != nil {
			return err
		}
	}

	// Resolve nested interpolations inside string input values (e.g. module_path defaults).
	if err := resolveContextInterpolations(ctx); err != nil {
		return fmt.Errorf("resolving nested input interpolations: %w", err)
	}

	projectName := ctxString(ctx, "project_name")
	if strings.TrimSpace(projectName) == "" {
		return fmt.Errorf("project_name is required")
	}

	// Ensure `validate` regexes from scaffold.yaml are enforced in both
	// interactive and flag-only modes.
	if err := validateManifestInputs(manifestInputs, ctx); err != nil {
		return err
	}

	// 5. Build request.
	outputDir, err := resolveOutputDir(newOutput, projectName)
	if err != nil {
		return err
	}
	req := &template.ScaffoldRequest{
		Template:  tpl,
		OutputDir: outputDir,
		Variables: ctx,
		DryRun:    newDryRun,
	}

	eng := engine.New()

	// 6. Execute engine.
	if req.DryRun {
		return runDryRun(req, eng)
	}

	res, err := runNonInteractive(req)
	if err != nil {
		return err
	}

	// 7. Summary (non-interactive: keep it simple).
	fmt.Printf("  ✓ Created %d files\n", len(res.FilesCreated))
	return nil
}

func configIsNonInteractive() bool {
	cfg, err := config.Load()
	if err != nil {
		return false
	}
	return cfg.NonInteractive
}

func hasTTY() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}

func resolveTemplate(flag string, templates []*template.Template, interactive bool) (*template.Template, error) {
	if strings.TrimSpace(flag) != "" {
		// Allow path to template directory.
		if st, err := os.Stat(flag); err == nil && st.IsDir() {
			t, err := template.LoadFromPath(flag)
			if err != nil {
				return nil, fmt.Errorf("loading template from path %q: %w", flag, err)
			}
			return t, nil
		}
		t, err := engine.Resolve(flag)
		if err != nil {
			return nil, err
		}
		return t, nil
	}
	if !interactive {
		return nil, fmt.Errorf("no TTY detected: --template is required")
	}
	return tui.RunSelector(templates)
}

func buildInitialContext(name string, vars []string) (dsl.Context, error) {
	ctx := dsl.Context{}
	if strings.TrimSpace(name) != "" {
		ctx["project_name"] = strings.TrimSpace(name)
	}
	for _, raw := range vars {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		k, v, ok := strings.Cut(raw, "=")
		if !ok || strings.TrimSpace(k) == "" {
			return nil, fmt.Errorf("invalid --var %q (expected key=value)", raw)
		}
		ctx[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	return ctx, nil
}

func resolveOutputDir(flag string, projectName string) (string, error) {
	if strings.TrimSpace(flag) != "" {
		return flag, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getwd: %w", err)
	}
	return filepath.Join(cwd, projectName), nil
}

func ctxString(ctx dsl.Context, key string) string {
	if ctx == nil {
		return ""
	}
	v, ok := ctx[key]
	if !ok || v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func needsTUI(inputs []dsl.Input, provided dsl.Context) bool {
	// Needs TUI if any active input is missing and has no default and is required.
	_, err := finalizeContextNonInteractive(inputs, provided)
	return err != nil
}

func coerceProvidedVarsToTypes(inputs []dsl.Input, ctx dsl.Context) (dsl.Context, error) {
	out := dsl.Context{}
	for k, v := range ctx {
		out[k] = v
	}

	byID := map[string]dsl.Input{}
	for _, in := range inputs {
		if strings.TrimSpace(in.ID) != "" {
			byID[in.ID] = in
		}
	}

	for k, v := range out {
		in, ok := byID[k]
		if !ok {
			continue
		}
		// Flags provide strings; convert to expected type when needed.
		if s, ok := v.(string); ok {
			switch strings.ToLower(strings.TrimSpace(in.Type)) {
			case "bool":
				b, ok := parseBool(s)
				if !ok {
					return nil, fmt.Errorf("invalid bool for %q: %q", k, s)
				}
				out[k] = b
			case "string", "enum", "path":
				out[k] = s
			case "multiselect":
				parts := strings.Split(s, ",")
				values := make([]string, 0, len(parts))
				for _, p := range parts {
					p = strings.TrimSpace(p)
					if p != "" {
						values = append(values, p)
					}
				}
				out[k] = values
			}
		}
	}

	return out, nil
}

func finalizeContextNonInteractive(inputs []dsl.Input, provided dsl.Context) (dsl.Context, error) {
	ctx := dsl.Context{}
	for k, v := range provided {
		ctx[k] = v
	}

	for _, in := range inputs {
		id := strings.TrimSpace(in.ID)
		if id == "" {
			continue
		}

		ok, err := evalWhen(in.When, ctx)
		if err != nil {
			return nil, fmt.Errorf("evaluating when for input %q: %w", id, err)
		}
		if !ok {
			if _, exists := ctx[id]; !exists {
				ctx[id] = defaultOrZero(in)
			}
			continue
		}

		if v, exists := ctx[id]; exists {
			// If empty string and default exists, apply default (matches TUI behavior).
			if s, ok := v.(string); ok && strings.TrimSpace(s) == "" && in.Default != nil {
				ctx[id] = fmt.Sprintf("%v", in.Default)
			}
			continue
		}

		if in.Default != nil {
			ctx[id] = defaultOrZero(in)
			continue
		}
		if in.Required {
			return nil, fmt.Errorf("missing required input %q (use --var %s=...)", id, id)
		}
		ctx[id] = defaultOrZero(in)
	}
	return ctx, nil
}

func evalWhen(expr string, ctx dsl.Context) (bool, error) {
	when := strings.TrimSpace(expr)
	if when == "" {
		return true, nil
	}
	ast, err := dsl.NewParser(when).Parse()
	if err != nil {
		return false, err
	}
	return dsl.Evaluate(ast, ctx)
}

func defaultOrZero(in dsl.Input) any {
	switch strings.ToLower(strings.TrimSpace(in.Type)) {
	case "string", "enum", "path":
		if in.Default == nil {
			return ""
		}
		if s, ok := in.Default.(string); ok {
			return s
		}
		return fmt.Sprintf("%v", in.Default)
	case "bool":
		if in.Default == nil {
			return false
		}
		if b, ok := in.Default.(bool); ok {
			return b
		}
		if s, ok := in.Default.(string); ok {
			v, _ := parseBool(s)
			return v
		}
		return false
	case "multiselect":
		if in.Default == nil {
			return []string{}
		}
		if arr, ok := in.Default.([]any); ok {
			out := make([]string, 0, len(arr))
			for _, v := range arr {
				out = append(out, fmt.Sprint(v))
			}
			return out
		}
		if s, ok := in.Default.(string); ok {
			if strings.TrimSpace(s) == "" {
				return []string{}
			}
			parts := strings.Split(s, ",")
			out := make([]string, 0, len(parts))
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" {
					out = append(out, p)
				}
			}
			return out
		}
		return []string{}
	default:
		return nil
	}
}

func parseBool(s string) (bool, bool) {
	v := strings.ToLower(strings.TrimSpace(s))
	switch v {
	case "y", "yes", "true", "1":
		return true, true
	case "n", "no", "false", "0":
		return false, true
	default:
		return false, false
	}
}

func resolveContextInterpolations(ctx dsl.Context) error {
	if ctx == nil {
		return nil
	}
	for k, v := range ctx {
		s, ok := v.(string)
		if !ok {
			continue
		}
		if !strings.Contains(s, "{{") {
			continue
		}
		resolved, err := dsl.Interpolate(s, ctx)
		if err != nil {
			return fmt.Errorf("input %q: %w", k, err)
		}
		ctx[k] = resolved
	}
	return nil
}

func applyComputedValues(computed []dsl.Computed, ctx dsl.Context) error {
	for _, c := range computed {
		id := strings.TrimSpace(c.ID)
		if id == "" {
			continue
		}
		value, err := dsl.Interpolate(c.Value, ctx)
		if err != nil {
			return fmt.Errorf("computing %q: %w", id, err)
		}
		ctx[id] = value
	}
	return nil
}

func runNonInteractive(req *template.ScaffoldRequest) (*template.ScaffoldResult, error) {
	fmt.Println("  → Creating project...")

	outAbs, err := filepath.Abs(req.OutputDir)
	if err != nil {
		return nil, fmt.Errorf("abs outputDir: %w", err)
	}
	req.OutputDir = outAbs

	rb := engine.NewRollbackManager(req.DryRun)

	exists, empty, err := dirExistsAndEmpty(outAbs)
	if err != nil {
		return nil, fmt.Errorf("checking outputDir: %w", err)
	}
	if exists && !empty {
		return nil, fmt.Errorf("output directory %s already exists and is not empty", outAbs)
	}

	if !req.DryRun {
		if err := os.MkdirAll(outAbs, 0o755); err != nil {
			return nil, fmt.Errorf("creating outputDir %s: %w", outAbs, err)
		}
		rb.TrackDir(outAbs)
	}

	created, skipped, err := engine.ProcessFiles(req)
	if err != nil {
		_ = rb.Rollback()
		return nil, err
	}
	_ = skipped

	obs := printStepObserver{}
	stepResults, err := engine.ExecuteStepsWithObserver(req.Template.Manifest.Steps, req.Variables, outAbs, req.DryRun, obs)
	if err != nil {
		_ = rb.Rollback()
		return &template.ScaffoldResult{
			FilesCreated:  created,
			FilesSkipped:  skipped,
			StepsExecuted: stepResults,
			StepsFailed:   failedSteps(stepResults),
		}, err
	}

	rb.Commit()

	return &template.ScaffoldResult{
		FilesCreated:  created,
		FilesSkipped:  skipped,
		StepsExecuted: stepResults,
	}, nil
}

type printStepObserver struct{}

func (printStepObserver) OnStepStart(step dsl.Step, _ string) {
	// no-op; keep logs clean
	_ = step
}
func (printStepObserver) OnStepSkipped(step dsl.Step) {
	fmt.Printf("  ─ %s (skipped)\n", step.Name)
}
func (printStepObserver) OnStepSuccess(step dsl.Step, _ string) {
	fmt.Printf("  ✓ %s\n", step.Name)
}
func (printStepObserver) OnStepFailure(step dsl.Step, err error, _ string) {
	fmt.Printf("  ✗ %s (%s)\n", step.Name, err.Error())
}

func dirExistsAndEmpty(path string) (exists bool, empty bool, err error) {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, true, nil
		}
		return false, false, err
	}
	if !fi.IsDir() {
		return true, false, fmt.Errorf("%s exists and is not a directory", path)
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return true, false, err
	}
	return true, len(entries) == 0, nil
}

func failedSteps(all []template.StepResult) []template.StepResult {
	var failed []template.StepResult
	for _, r := range all {
		if r.Error != nil {
			failed = append(failed, r)
		}
	}
	return failed
}

func runDryRun(req *template.ScaffoldRequest, eng *engine.Engine) error {
	outAbs, err := filepath.Abs(req.OutputDir)
	if err == nil {
		req.OutputDir = outAbs
	}

	res, err := eng.Scaffold(req)
	if err != nil {
		return err
	}

	fmt.Println("Dry run — no files will be written.")
	fmt.Printf("Template : %s\n", req.Template.Manifest.Name)
	fmt.Printf("Output   : %s\n", prettyPath(req.OutputDir))
	fmt.Printf("Variables: %s\n", formatVars(req.Variables))
	fmt.Println("Files that would be created:")
	for _, f := range res.FilesCreated {
		fmt.Println(f)
	}
	fmt.Println("Steps that would run:")
	for _, line := range dryRunSteps(req.Template.Manifest.Steps, req.Variables) {
		fmt.Println(line)
	}
	fmt.Println("No files were written.")
	return nil
}

func formatVars(ctx dsl.Context) string {
	if ctx == nil {
		return ""
	}
	keys := make([]string, 0, len(ctx))
	for k := range ctx {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	// show project_name first if present
	for _, k := range keys {
		if k == "project_name" {
			parts = append(parts, fmt.Sprintf("%s=%v", k, ctx[k]))
		}
	}
	for _, k := range keys {
		if k == "project_name" {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s=%v", k, ctx[k]))
	}
	return strings.Join(parts, ", ")
}

func prettyPath(absOrRel string) string {
	cwd, err := os.Getwd()
	if err != nil {
		return absOrRel
	}
	abs, err := filepath.Abs(absOrRel)
	if err != nil {
		return absOrRel
	}
	rel, err := filepath.Rel(cwd, abs)
	if err != nil {
		return absOrRel
	}
	if rel == "." {
		return "."
	}
	if !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) && rel != ".." {
		return "." + string(os.PathSeparator) + rel
	}
	return absOrRel
}

func dryRunSteps(steps []dsl.Step, ctx dsl.Context) []string {
	lines := make([]string, 0, len(steps))
	for _, s := range steps {
		when := strings.TrimSpace(s.When)
		ok := true
		if when != "" {
			ast, err := dsl.NewParser(when).Parse()
			if err != nil {
				lines = append(lines, fmt.Sprintf("✗ %s (invalid when: %s)", s.Name, err.Error()))
				continue
			}
			v, err := dsl.Evaluate(ast, ctx)
			if err != nil {
				lines = append(lines, fmt.Sprintf("✗ %s (when eval error: %s)", s.Name, err.Error()))
				continue
			}
			ok = v
		}

		cmdStr, err := dsl.Interpolate(s.Run, ctx)
		if err != nil {
			lines = append(lines, fmt.Sprintf("✗ %s (interpolation error: %s)", s.Name, err.Error()))
			continue
		}

		if ok {
			lines = append(lines, "✓ "+cmdStr)
		} else if when != "" {
			lines = append(lines, "─ "+cmdStr+"  (skipped: "+when+")")
		} else {
			lines = append(lines, "─ "+cmdStr+"  (skipped)")
		}
	}
	return lines
}

func validateManifestInputs(inputs []dsl.Input, ctx dsl.Context) error {
	if ctx == nil {
		return nil
	}

	for _, in := range inputs {
		id := strings.TrimSpace(in.ID)
		if id == "" {
			continue
		}
		if strings.TrimSpace(in.Validate) == "" {
			continue
		}

		v, ok := ctx[id]
		if !ok {
			if in.Required {
				return fmt.Errorf("invalid value for %q: value is required", id)
			}
			continue
		}

		if err := tui.ValidateInputValue(in, fmt.Sprint(v)); err != nil {
			return fmt.Errorf("invalid value for %q: %w", id, err)
		}
	}

	return nil
}


