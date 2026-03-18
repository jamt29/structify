package engine

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/jamt29/structify/internal/dsl"
	"github.com/jamt29/structify/internal/template"
)

// StepObserver is an optional hook used to report step execution progress.
// All callbacks are best-effort; observers must be fast and must not panic.
type StepObserver interface {
	OnStepStart(step dsl.Step, interpolatedCommand string)
	OnStepSkipped(step dsl.Step)
	OnStepSuccess(step dsl.Step, output string)
	OnStepFailure(step dsl.Step, err error, output string)
}

// ExecuteSteps runs post-generation steps in order.
func ExecuteSteps(steps []dsl.Step, ctx dsl.Context, outputDir string, dryRun bool) ([]template.StepResult, error) {
	return ExecuteStepsWithObserver(steps, ctx, outputDir, dryRun, nil)
}

// ExecuteStepsWithObserver runs steps and notifies an observer of progress.
func ExecuteStepsWithObserver(steps []dsl.Step, ctx dsl.Context, outputDir string, dryRun bool, obs StepObserver) ([]template.StepResult, error) {
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
				if obs != nil {
					obs.OnStepFailure(s, err, "")
				}
				return results, fmt.Errorf("parsing when for step %q: %w", s.Name, err)
			}
			ok, err := dsl.Evaluate(ast, ctx)
			if err != nil {
				res.Error = err
				results = append(results, res)
				if obs != nil {
					obs.OnStepFailure(s, err, "")
				}
				return results, fmt.Errorf("evaluating when for step %q: %w", s.Name, err)
			}
			if !ok {
				res.Skipped = true
				results = append(results, res)
				if obs != nil {
					obs.OnStepSkipped(s)
				}
				continue
			}
		}

		cmdStr, err := dsl.Interpolate(s.Run, ctx)
		if err != nil {
			res.Error = err
			results = append(results, res)
			if obs != nil {
				obs.OnStepFailure(s, err, "")
			}
			return results, fmt.Errorf("interpolating run for step %q: %w", s.Name, err)
		}
		res.Command = cmdStr

		if obs != nil {
			obs.OnStepStart(s, cmdStr)
		}

		if dryRun {
			results = append(results, res)
			if obs != nil {
				obs.OnStepSuccess(s, "")
			}
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
			if obs != nil {
				obs.OnStepFailure(s, err, res.Output)
			}
			return results, fmt.Errorf("step %q failed: %w\n%s", s.Name, err, res.Output)
		}

		res.Output = buf.String()
		results = append(results, res)
		if obs != nil {
			obs.OnStepSuccess(s, res.Output)
		}
	}

	return results, nil
}

