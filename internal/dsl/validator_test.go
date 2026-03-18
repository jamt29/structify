package dsl

import (
	"testing"
)

func TestValidateManifest_Valid(t *testing.T) {
	m := &Manifest{
		Name:     "t",
		Version:  "1.0.0",
		Language: "go",
		Inputs: []Input{
			{ID: "project_name", Type: "string", Prompt: "p", Required: true},
			{ID: "use_docker", Type: "bool", Prompt: "p"},
		},
		Files: []FileRule{
			{Include: "docker/**", When: `use_docker == true`},
		},
		Steps: []Step{
			{Name: "Tidy", Run: "go mod tidy"},
		},
	}

	errs := ValidateManifest(m)
	if len(errs) != 0 {
		t.Fatalf("expected zero errors, got: %#v", errs)
	}
}

func TestValidateManifest_MultipleErrors(t *testing.T) {
	m := &Manifest{
		Name:     "",
		Version:  "1",
		Language: "unknown",
		Inputs: []Input{
			{ID: "project-name", Type: "enum"}, // bad id + enum without options
			{ID: "project-name", Type: "wat"},  // duplicate + bad type
			{ID: "a", Type: "string", When: `b == "x"`},
			{ID: "b", Type: "string", When: `a == "y"`}, // cycle a <-> b
		},
		Files: []FileRule{
			{Include: "a", Exclude: "b"}, // both set
			{When: `transport = "http"`}, // neither include nor exclude + bad when
		},
		Steps: []Step{
			{Name: "", Run: ""},                   // missing name/run
			{Name: "X", Run: "echo", When: `a =`}, // bad when
		},
	}

	errs := ValidateManifest(m)
	if len(errs) < 8 {
		t.Fatalf("expected multiple errors, got %d: %#v", len(errs), errs)
	}
}

func TestValidateManifest_InvalidWhenFieldPath(t *testing.T) {
	m := &Manifest{
		Name:     "t",
		Version:  "1.0.0",
		Language: "go",
		Inputs: []Input{
			{ID: "transport", Type: "string", When: `transport = "http"`},
		},
	}
	errs := ValidateManifest(m)
	found := false
	for _, e := range errs {
		if e.Field == "inputs[0].when" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected error on inputs[0].when, got: %#v", errs)
	}
}
