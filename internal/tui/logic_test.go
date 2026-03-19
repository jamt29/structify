package tui

import (
	"testing"

	"github.com/jamt29/structify/internal/dsl"
)

func TestShouldAskInput(t *testing.T) {
	tests := []struct {
		name    string
		input   dsl.Input
		ctx     dsl.Context
		wantAsk bool
		wantErr bool
	}{
		{
			name: "when true",
			input: dsl.Input{
				ID:   "orm",
				Type: "string",
				When: `transport == "grpc"`,
			},
			ctx:     dsl.Context{"transport": "grpc"},
			wantAsk: true,
		},
		{
			name: "when false",
			input: dsl.Input{
				ID:   "orm",
				Type: "string",
				When: `transport == "grpc"`,
			},
			ctx:     dsl.Context{"transport": "http"},
			wantAsk: false,
		},
		{
			name: "when empty",
			input: dsl.Input{
				ID:   "orm",
				Type: "string",
				When: "",
			},
			ctx:     dsl.Context{},
			wantAsk: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ShouldAskInput(tt.input, tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error state: err=%v wantErr=%v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if got != tt.wantAsk {
				t.Fatalf("got %v want %v", got, tt.wantAsk)
			}
		})
	}
}

func TestApplyDefault(t *testing.T) {
	tests := []struct {
		name    string
		input   dsl.Input
		ctx     dsl.Context
		want    string
		wantErr bool
	}{
		{
			name: "default present",
			input: dsl.Input{
				ID:      "module_path",
				Type:    "string",
				Default: "github.com/user/app",
			},
			ctx:     dsl.Context{},
			want:    "github.com/user/app",
			wantErr: false,
		},
		{
			name: "default missing",
			input: dsl.Input{
				ID:   "project_name",
				Type: "string",
			},
			ctx:     dsl.Context{},
			want:    "",
			wantErr: false,
		},
		{
			name: "default interpolated",
			input: dsl.Input{
				ID:      "module_path",
				Type:    "string",
				Default: "github.com/{{ project_name | kebab_case }}",
			},
			ctx:     dsl.Context{"project_name": "My Project API"},
			want:    "github.com/my-project-api",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ApplyDefault(tt.input, tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error state: err=%v wantErr=%v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}

func TestValidateInputValue(t *testing.T) {
	tests := []struct {
		name    string
		input   dsl.Input
		value   string
		wantErr bool
	}{
		{
			name: "regex valid",
			input: dsl.Input{
				ID:       "project_name",
				Type:     "string",
				Validate: "^[a-z]+$",
			},
			value:   "abc",
			wantErr: false,
		},
		{
			name: "regex invalid",
			input: dsl.Input{
				ID:       "project_name",
				Type:     "string",
				Validate: "^[a-z]+$",
			},
			value:   "ABC",
			wantErr: true,
		},
		{
			name: "no regex always valid",
			input: dsl.Input{
				ID:   "project_name",
				Type: "string",
			},
			value:   "anything",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateInputValue(tt.input, tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error state: err=%v wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

func TestBuildContext(t *testing.T) {
	t.Run("complete inputs", func(t *testing.T) {
		inputs := []dsl.Input{
			{
				ID:       "project_name",
				Type:     "string",
				Required: true,
				// no validate; keep it simple
			},
			{
				ID:      "module_path",
				Type:    "string",
				Default: "github.com/{{ project_name | kebab_case }}",
			},
		}

		ctx, err := BuildContext(inputs, map[string]string{
			"project_name": "My Project API",
		})
		if err != nil {
			t.Fatalf("BuildContext returned error: %v", err)
		}

		if got, ok := ctx["module_path"].(string); !ok || got != "github.com/my-project-api" {
			t.Fatalf("unexpected module_path: %#v", ctx["module_path"])
		}
	})

	t.Run("when false ignores provided answer", func(t *testing.T) {
		inputs := []dsl.Input{
			{
				ID:   "transport",
				Type: "string",
			},
			{
				ID:      "orm",
				Type:    "string",
				Default: "none",
				When:    `transport == "grpc"`,
			},
		}

		ctx, err := BuildContext(inputs, map[string]string{
			"transport": "http",
			"orm":       "gorm",
		})
		if err != nil {
			t.Fatalf("BuildContext returned error: %v", err)
		}

		if got, ok := ctx["orm"].(string); !ok || got != "none" {
			t.Fatalf("expected orm default none, got %#v", ctx["orm"])
		}
	})
}

func TestRunInputsWithInitial_NoPrompt(t *testing.T) {
	t.Run("when true validated from initial, when false uses default", func(t *testing.T) {
		inputs := []dsl.Input{
			{
				ID:   "transport",
				Type: "string",
			},
			{
				ID:      "orm",
				Type:    "string",
				Default: "none",
				When:    `transport == "grpc"`,
			},
		}

		ctx, err := RunInputsWithInitial(inputs, dsl.Context{
			"transport": "http",
		})
		if err != nil {
			t.Fatalf("RunInputsWithInitial returned error: %v", err)
		}

		if got := ctx["transport"]; got != "http" {
			t.Fatalf("expected transport=http, got %#v", got)
		}
		if got := ctx["orm"]; got != "none" {
			t.Fatalf("expected orm=none, got %#v", got)
		}
	})

	t.Run("bool input validated and parsed from initial", func(t *testing.T) {
		inputs := []dsl.Input{
			{
				ID:   "flag",
				Type: "bool",
			},
		}

		ctx, err := RunInputsWithInitial(inputs, dsl.Context{
			"flag": true,
		})
		if err != nil {
			t.Fatalf("RunInputsWithInitial returned error: %v", err)
		}

		if got := ctx["flag"]; got != true {
			t.Fatalf("expected flag=true, got %#v", got)
		}
	})
}

