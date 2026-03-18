package dsl

import (
	"strings"
	"testing"
)

func TestParser_Parse_ValidExpressions(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple_compare",
			input: `transport == "http"`,
			want:  `(== transport "http")`,
		},
		{
			name:  "not_ident",
			input: `!use_docker`,
			want:  `(! use_docker)`,
		},
		{
			name:  "and_precedence_over_or",
			input: `a == "x" || b == "y" && c == "z"`,
			// && binds tighter than ||
			want: `(|| (== a "x") (&& (== b "y") (== c "z")))`,
		},
		{
			name:  "paren_overrides",
			input: `(a == "x" || b == "y") && c != "z"`,
			want:  `(&& (|| (== a "x") (== b "y")) (!= c "z"))`,
		},
		{
			name:  "not_applies_to_expression",
			input: `!use_docker == true`,
			want:  `(! (== use_docker true))`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, err := NewParser(tt.input).Parse()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := nodeString(n)
			if got != tt.want {
				t.Fatalf("ast:\n got: %s\nwant: %s", got, tt.want)
			}
		})
	}
}

func TestParser_Parse_Errors(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantSubstrs []string
	}{
		{
			name:        "single_equals_is_illegal",
			input:       `transport = "http"`,
			wantSubstrs: []string{"parse error at position", "invalid operator '='"},
		},
		{
			name:        "missing_rhs",
			input:       `transport ==`,
			wantSubstrs: []string{"parse error at position", "unexpected end of input"},
		},
		{
			name:        "unclosed_paren",
			input:       `(transport == "http"`,
			wantSubstrs: []string{"parse error at position", "expected ')'"}, // message includes token context
		},
		{
			name:        "single_quotes_not_supported",
			input:       `transport == 'http'`,
			wantSubstrs: []string{"parse error at position", "single quotes are not supported"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewParser(tt.input).Parse()
			if err == nil {
				t.Fatalf("expected error")
			}
			for _, sub := range tt.wantSubstrs {
				if !strings.Contains(err.Error(), sub) {
					t.Fatalf("error mismatch:\n got: %q\nwant to contain: %q", err.Error(), sub)
				}
			}
		})
	}
}

func nodeString(n Node) string {
	switch t := n.(type) {
	case *IdentNode:
		return t.Name
	case *StringLiteralNode:
		return `"` + t.Value + `"`
	case *BoolLiteralNode:
		if t.Value {
			return "true"
		}
		return "false"
	case *NotNode:
		return "(! " + nodeString(t.Expr) + ")"
	case *CompareNode:
		return "(" + t.Operator + " " + nodeString(t.Left) + " " + nodeString(t.Right) + ")"
	case *BinaryNode:
		return "(" + t.Operator + " " + nodeString(t.Left) + " " + nodeString(t.Right) + ")"
	default:
		return "<unknown>"
	}
}
