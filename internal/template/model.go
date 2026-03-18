package template

import "github.com/jamt29/structify/internal/dsl"

// Template represents a scaffolding template loaded from some source.
type Template struct {
	Manifest *dsl.Manifest
	Path     string
	Source   string // "builtin" | "local" | "github"
}

// ScaffoldRequest contains everything needed to scaffold a project.
type ScaffoldRequest struct {
	Template   *Template
	OutputDir  string
	Variables  dsl.Context
	DryRun     bool
}

// ScaffoldResult summarizes what the engine did.
type ScaffoldResult struct {
	FilesCreated  []string
	FilesSkipped  []string
	StepsExecuted []StepResult
	StepsFailed   []StepResult
}

// StepResult is the outcome of executing a single step.
type StepResult struct {
	Name    string
	Command string
	Output  string
	Error   error
	Skipped bool
}

