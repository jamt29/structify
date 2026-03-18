package engine

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/jamt29/structify/internal/dsl"
	"github.com/jamt29/structify/internal/template"
)

// ExecuteSteps runs post-generation steps in order.
func ExecuteSteps(steps []dsl.Step, ctx dsl.Context, outputDir string, dryRun bool) ([]template.StepResult, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}
	if strings.TrimSpace(outputDir) == "" {
		return nil, fmt.Errorf("outputDir is empty")
	}

	results := make([]template.StepResult, 0, len(steps))

	for _, s := range steps {
		res := template.StepResult{Name: s.Name}

		when := strings.TrimSpace(s.When)
		if when != "" {
			ast, err := dsl.NewParser(when).Parse()
			if err != nil {
				res.Error = err
				results = append(results, res)
				return results, fmt.Errorf("parsing when for step %q: %w", s.Name, err)
			}
			ok, err := dsl.Evaluate(ast, ctx)
			if err != nil {
				res.Error = err
				results = append(results, res)
				return results, fmt.Errorf("evaluating when for step %q: %w", s.Name, err)
			}
			if !ok {
				res.Skipped = true
				results = append(results, res)
				continue
			}
		}

		cmdStr, err := dsl.Interpolate(s.Run, ctx)
		if err != nil {
			res.Error = err
			results = append(results, res)
			return results, fmt.Errorf("interpolating run for step %q: %w", s.Name, err)
		}
		res.Command = cmdStr

		if dryRun {
			results = append(results, res)
			continue
		}

		cmd := exec.Command("bash", "-lc", cmdStr)
		cmd.Dir = outputDir

		var buf bytes.Buffer
		cmd.Stdout = &buf
		cmd.Stderr = &buf

		if err := cmd.Run(); err != nil {
			res.Output = buf.String()
			res.Error = err
			results = append(results, res)
			return results, fmt.Errorf("step %q failed: %w\n%s", s.Name, err, res.Output)
		}

		res.Output = buf.String()
		results = append(results, res)
	}

	return results, nil
}

