package dsl

import (
	"strings"
	"testing"
)

func TestInterpolate_SimpleAndMultiple(t *testing.T) {
	ctx := Context{
		"project_name": "MyProject",
		"transport":    "http",
	}

	got, err := Interpolate("Hello {{ project_name }}", ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "Hello MyProject" {
		t.Fatalf("got %q want %q", got, "Hello MyProject")
	}

	got2, err := Interpolate("{{project_name}} uses {{ transport }}", ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got2 != "MyProject uses http" {
		t.Fatalf("got %q want %q", got2, "MyProject uses http")
	}
}

func TestInterpolate_WithFilter(t *testing.T) {
	ctx := Context{"project_name": "MyProject"}

	got, err := Interpolate("{{ project_name | snake_case }}", ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "my_project" {
		t.Fatalf("got %q want %q", got, "my_project")
	}
}

func TestInterpolate_Errors(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		ctx         Context
		wantSubstrs []string
	}{
		{
			name:        "missing_variable",
			input:       "{{ missing }}",
			ctx:         Context{"x": "y"},
			wantSubstrs: []string{"variable 'missing' not defined"},
		},
		{
			name:        "unknown_filter",
			input:       "{{ project_name | nope }}",
			ctx:         Context{"project_name": "MyProject"},
			wantSubstrs: []string{"unknown filter 'nope'", "available:"},
		},
		{
			name:        "filter_chaining_not_supported",
			input:       "{{ project_name | snake_case | upper }}",
			ctx:         Context{"project_name": "MyProject"},
			wantSubstrs: []string{"filter chaining is not supported"},
		},
		{
			name:        "unterminated",
			input:       "hello {{ project_name ",
			ctx:         Context{"project_name": "MyProject"},
			wantSubstrs: []string{"unterminated interpolation"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Interpolate(tt.input, tt.ctx)
			if err == nil {
				t.Fatalf("expected error")
			}
			for _, sub := range tt.wantSubstrs {
				if !strings.Contains(err.Error(), sub) {
					t.Fatalf("error mismatch:\n got: %q\nwant contain: %q", err.Error(), sub)
				}
			}
		})
	}
}
