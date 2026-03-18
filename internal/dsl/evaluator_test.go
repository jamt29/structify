package dsl

import (
	"strings"
	"testing"
)

func TestEvaluate_TableFromSpecAndErrors(t *testing.T) {
	tests := []struct {
		name      string
		expr      string
		ctx       Context
		want      bool
		wantErr   bool
		errSubstr string
	}{
		// Must pass (adapted to double quotes)
		{name: "eq_string_true", expr: `transport == "http"`, ctx: Context{"transport": "http"}, want: true},
		{name: "neq_string_true", expr: `transport != "grpc"`, ctx: Context{"transport": "http"}, want: true},
		{name: "eq_bool_true", expr: `use_docker == true`, ctx: Context{"use_docker": true}, want: true},
		{name: "not_bool", expr: `!use_docker`, ctx: Context{"use_docker": false}, want: true},
		{name: "and_true", expr: `a == "x" && b == "y"`, ctx: Context{"a": "x", "b": "y"}, want: true},
		{name: "and_false", expr: `a == "x" && b == "y"`, ctx: Context{"a": "x", "b": "z"}, want: false},
		{name: "or_true", expr: `a == "x" || b == "y"`, ctx: Context{"a": "z", "b": "y"}, want: true},
		{name: "grouped_complex", expr: `(a == "x" || b == "y") && c != "z"`, ctx: Context{"a": "z", "b": "y", "c": "q"}, want: true},

		// Errors
		{name: "invalid_operator_equals", expr: `transport = "http"`, ctx: Context{}, wantErr: true, errSubstr: "invalid operator '='"},
		{name: "undeclared_variable", expr: `undeclared == "x"`, ctx: Context{}, wantErr: true, errSubstr: "variable 'undeclared' not defined in context"},
		{name: "type_mismatch_compare", expr: `use_docker == "true"`, ctx: Context{"use_docker": true}, wantErr: true, errSubstr: "cannot compare bool with string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := NewParser(tt.expr).Parse()
			if err != nil {
				if !tt.wantErr {
					t.Fatalf("unexpected parse error: %v", err)
				}
				if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Fatalf("parse error mismatch:\n got: %q\nwant contain: %q", err.Error(), tt.errSubstr)
				}
				return
			}

			got, err := Evaluate(ast, tt.ctx)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (value=%v)", got)
				}
				if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Fatalf("error mismatch:\n got: %q\nwant contain: %q", err.Error(), tt.errSubstr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
	}
}

func TestEvaluate_ShortCircuit(t *testing.T) {
	// Right side references unknown variable; should not be evaluated.
	ast, err := NewParser(`a == "x" || missing == "y"`).Parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	got, err := Evaluate(ast, Context{"a": "x"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != true {
		t.Fatalf("got %v want true", got)
	}

	ast2, err := NewParser(`a != "x" && missing == "y"`).Parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	got2, err := Evaluate(ast2, Context{"a": "x"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got2 != false {
		t.Fatalf("got %v want false", got2)
	}
}
