package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamt29/structify/internal/template"
)

type Engine struct{}

func New() *Engine {
	return &Engine{}
}

func (e *Engine) Scaffold(req *template.ScaffoldRequest) (*template.ScaffoldResult, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}
	if req.Template == nil {
		return nil, fmt.Errorf("template is nil")
	}
	if req.Variables == nil {
		return nil, fmt.Errorf("variables is nil")
	}
	if strings.TrimSpace(req.OutputDir) == "" {
		return nil, fmt.Errorf("outputDir is empty")
	}

	outAbs, err := filepath.Abs(req.OutputDir)
	if err != nil {
		return nil, fmt.Errorf("abs outputDir: %w", err)
	}

	if exists, empty, err := dirExistsAndEmpty(outAbs); err != nil {
		return nil, fmt.Errorf("checking outputDir: %w", err)
	} else if exists && !empty {
		return nil, fmt.Errorf("output directory %s already exists and is not empty", outAbs)
	}

	rb := NewRollbackManager(req.DryRun)

	if !req.DryRun {
		if err := os.MkdirAll(outAbs, 0o755); err != nil {
			return nil, fmt.Errorf("creating outputDir %s: %w", outAbs, err)
		}
		rb.TrackDir(outAbs)
	}

	created, skipped, err := ProcessFiles(req)
	if err != nil {
		_ = rb.Rollback()
		return nil, err
	}

	stepResults, err := ExecuteSteps(req.Template.Manifest.Steps, req.Variables, outAbs, req.DryRun)
	if err != nil {
		_ = rb.Rollback()
		// best-effort: include partial results from ExecuteSteps
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

