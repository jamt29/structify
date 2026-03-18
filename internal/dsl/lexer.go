package dsl

import (
	"fmt"
	"unicode"
)

type Lexer struct {
	input string
	pos   int  // current position (points to current char)
	read  int  // next read position (after current char)
	ch    byte // current char under examination; 0 means EOF
}

func NewLexer(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	startPos := l.pos

	switch l.ch {
	case 0:
		return Token{Type: TOKEN_EOF, Literal: "", Pos: l.pos}
	case '(':
		l.readChar()
		return Token{Type: TOKEN_LPAREN, Literal: "(", Pos: startPos}
	case ')':
		l.readChar()
		return Token{Type: TOKEN_RPAREN, Literal: ")", Pos: startPos}
	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			return Token{Type: TOKEN_NEQ, Literal: "!=", Pos: startPos}
		}
		l.readChar()
		return Token{Type: TOKEN_NOT, Literal: "!", Pos: startPos}
	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			return Token{Type: TOKEN_EQ, Literal: "==", Pos: startPos}
		}
		il := l.illegalToken(startPos, "invalid operator '=', did you mean '=='?")
		l.readChar()
		return il
	case '&':
		if l.peekChar() == '&' {
			l.readChar()
			l.readChar()
			return Token{Type: TOKEN_AND, Literal: "&&", Pos: startPos}
		}
		il := l.illegalToken(startPos, "invalid operator '&', did you mean '&&'?")
		l.readChar()
		return il
	case '|':
		if l.peekChar() == '|' {
			l.readChar()
			l.readChar()
			return Token{Type: TOKEN_OR, Literal: "||", Pos: startPos}
		}
		il := l.illegalToken(startPos, "invalid operator '|', did you mean '||'?")
		l.readChar()
		return il
	case '"':
		lit, err := l.readString()
		if err != nil {
			return Token{Type: TOKEN_ILLEGAL, Literal: err.Error(), Pos: startPos}
		}
		return Token{Type: TOKEN_STRING, Literal: lit, Pos: startPos}
	case '\'':
		il := l.illegalToken(startPos, "single quotes are not supported; use double quotes")
		l.readChar()
		return il
	default:
		if isIdentStart(l.ch) {
			lit := l.readIdent()
			if lit == "true" || lit == "false" {
				return Token{Type: TOKEN_BOOL, Literal: lit, Pos: startPos}
			}
			return Token{Type: TOKEN_IDENT, Literal: lit, Pos: startPos}
		}
		il := l.illegalToken(startPos, fmt.Sprintf("unrecognized character %q", l.ch))
		l.readChar()
		return il
	}
}

func (l *Lexer) readChar() {
	if l.read >= len(l.input) {
		l.ch = 0
		l.pos = l.read
		return
	}
	l.ch = l.input[l.read]
	l.pos = l.read
	l.read++
}

func (l *Lexer) peekChar() byte {
	if l.read >= len(l.input) {
		return 0
	}
	return l.input[l.read]
}

func (l *Lexer) skipWhitespace() {
	for l.ch != 0 && (l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r') {
		l.readChar()
	}
}

func (l *Lexer) readIdent() string {
	start := l.pos
	for l.ch != 0 && isIdentPart(l.ch) {
		l.readChar()
	}
	return l.input[start:l.pos]
}

func (l *Lexer) readString() (string, error) {
	// current char is '"'
	l.readChar()
	start := l.pos
	for {
		if l.ch == 0 {
			return "", fmt.Errorf("unterminated string literal")
		}
		if l.ch == '"' {
			lit := l.input[start:l.pos]
			l.readChar()
			return lit, nil
		}
		l.readChar()
	}
}

func (l *Lexer) illegalToken(pos int, msg string) Token {
	return Token{Type: TOKEN_ILLEGAL, Literal: msg, Pos: pos}
}

func isIdentStart(ch byte) bool {
	r := rune(ch)
	return ch == '_' || unicode.IsLetter(r)
}

func isIdentPart(ch byte) bool {
	r := rune(ch)
	return ch == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}
