package dsl

import (
	"fmt"
)

type Parser struct {
	l    *Lexer
	cur  Token
	peek Token
}

func NewParser(input string) *Parser {
	l := NewLexer(input)
	p := &Parser{l: l}
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) Parse() (Node, error) {
	if p.cur.Type == TOKEN_EOF {
		return nil, fmt.Errorf("parse error at position %d: empty expression", p.cur.Pos)
	}

	expr, err := p.parseOr()
	if err != nil {
		return nil, err
	}

	if p.cur.Type != TOKEN_EOF {
		return nil, p.errf(p.cur, "unexpected token %q", p.cur.Literal)
	}

	return expr, nil
}

func (p *Parser) nextToken() {
	p.cur = p.peek
	p.peek = p.l.NextToken()
}

func (p *Parser) parseOr() (Node, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}

	for p.cur.Type == TOKEN_OR {
		op := p.cur.Literal
		p.nextToken()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &BinaryNode{Left: left, Operator: op, Right: right}
	}
	return left, nil
}

func (p *Parser) parseAnd() (Node, error) {
	left, err := p.parseNot()
	if err != nil {
		return nil, err
	}

	for p.cur.Type == TOKEN_AND {
		op := p.cur.Literal
		p.nextToken()
		right, err := p.parseNot()
		if err != nil {
			return nil, err
		}
		left = &BinaryNode{Left: left, Operator: op, Right: right}
	}
	return left, nil
}

func (p *Parser) parseNot() (Node, error) {
	if p.cur.Type == TOKEN_NOT {
		tok := p.cur
		p.nextToken()
		expr, err := p.parseNot()
		if err != nil {
			return nil, err
		}
		_ = tok
		return &NotNode{Expr: expr}, nil
	}
	return p.parseCompare()
}

func (p *Parser) parseCompare() (Node, error) {
	left, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	if p.cur.Type == TOKEN_EQ || p.cur.Type == TOKEN_NEQ {
		op := p.cur.Literal
		p.nextToken()
		right, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		return &CompareNode{Left: left, Operator: op, Right: right}, nil
	}

	if p.cur.Type == TOKEN_ILLEGAL {
		return nil, p.errf(p.cur, "%s", p.cur.Literal)
	}

	return left, nil
}

func (p *Parser) parsePrimary() (Node, error) {
	tok := p.cur

	switch tok.Type {
	case TOKEN_IDENT:
		p.nextToken()
		return &IdentNode{Name: tok.Literal}, nil
	case TOKEN_STRING:
		p.nextToken()
		return &StringLiteralNode{Value: tok.Literal}, nil
	case TOKEN_BOOL:
		p.nextToken()
		if tok.Literal == "true" {
			return &BoolLiteralNode{Value: true}, nil
		}
		if tok.Literal == "false" {
			return &BoolLiteralNode{Value: false}, nil
		}
		return nil, p.errf(tok, "invalid boolean literal %q", tok.Literal)
	case TOKEN_LPAREN:
		p.nextToken()
		expr, err := p.parseOr()
		if err != nil {
			return nil, err
		}
		if p.cur.Type != TOKEN_RPAREN {
			return nil, p.errf(p.cur, "expected ')', got %q", p.cur.Literal)
		}
		p.nextToken()
		return expr, nil
	case TOKEN_ILLEGAL:
		return nil, p.errf(tok, "%s", tok.Literal)
	case TOKEN_EOF:
		return nil, p.errf(tok, "unexpected end of input")
	default:
		return nil, p.errf(tok, "expected string, boolean, identifier, or '(', got %q", tok.Literal)
	}
}

func (p *Parser) errf(tok Token, format string, args ...any) error {
	msg := fmt.Sprintf(format, args...)
	return fmt.Errorf("parse error at position %d: %s", tok.Pos, msg)
}
