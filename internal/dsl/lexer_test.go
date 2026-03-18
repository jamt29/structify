package dsl

import "testing"

func TestLexer_NextToken_BasicTokens(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []Token
	}{
		{
			name:  "comparison_and_bool",
			input: `transport == "http" && use_docker == true`,
			want: []Token{
				{Type: TOKEN_IDENT, Literal: "transport"},
				{Type: TOKEN_EQ, Literal: "=="},
				{Type: TOKEN_STRING, Literal: "http"},
				{Type: TOKEN_AND, Literal: "&&"},
				{Type: TOKEN_IDENT, Literal: "use_docker"},
				{Type: TOKEN_EQ, Literal: "=="},
				{Type: TOKEN_BOOL, Literal: "true"},
				{Type: TOKEN_EOF, Literal: ""},
			},
		},
		{
			name:  "paren_or_not",
			input: `!(a != "x" || b == "y")`,
			want: []Token{
				{Type: TOKEN_NOT, Literal: "!"},
				{Type: TOKEN_LPAREN, Literal: "("},
				{Type: TOKEN_IDENT, Literal: "a"},
				{Type: TOKEN_NEQ, Literal: "!="},
				{Type: TOKEN_STRING, Literal: "x"},
				{Type: TOKEN_OR, Literal: "||"},
				{Type: TOKEN_IDENT, Literal: "b"},
				{Type: TOKEN_EQ, Literal: "=="},
				{Type: TOKEN_STRING, Literal: "y"},
				{Type: TOKEN_RPAREN, Literal: ")"},
				{Type: TOKEN_EOF, Literal: ""},
			},
		},
		{
			name:  "whitespace_ignored",
			input: "  \n\tuse_docker\t \r\n",
			want: []Token{
				{Type: TOKEN_IDENT, Literal: "use_docker"},
				{Type: TOKEN_EOF, Literal: ""},
			},
		},
		{
			name:  "illegal_single_quote",
			input: `transport == 'http'`,
			want: []Token{
				{Type: TOKEN_IDENT, Literal: "transport"},
				{Type: TOKEN_EQ, Literal: "=="},
				{Type: TOKEN_ILLEGAL},
			},
		},
		{
			name:  "illegal_single_equals",
			input: `transport = "http"`,
			want: []Token{
				{Type: TOKEN_IDENT, Literal: "transport"},
				{Type: TOKEN_ILLEGAL},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.input)
			for i := 0; i < len(tt.want); i++ {
				got := l.NextToken()
				if got.Type != tt.want[i].Type {
					t.Fatalf("token[%d] type: got %q want %q (lit=%q pos=%d)", i, got.Type, tt.want[i].Type, got.Literal, got.Pos)
				}
				if tt.want[i].Literal != "" && got.Literal != tt.want[i].Literal {
					t.Fatalf("token[%d] literal: got %q want %q (type=%q)", i, got.Literal, tt.want[i].Literal, got.Type)
				}
				if got.Type == TOKEN_ILLEGAL && got.Literal == "" {
					t.Fatalf("token[%d] illegal token should have descriptive literal", i)
				}
			}
		})
	}
}
