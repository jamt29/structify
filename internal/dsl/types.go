package dsl

// TokenType represents the type of a token produced by the lexer.
type TokenType string

// Token types for the DSL lexer.
const (
	// Literals
	TOKEN_STRING TokenType = "STRING" // "http"
	TOKEN_BOOL   TokenType = "BOOL"   // true | false
	TOKEN_IDENT  TokenType = "IDENT"  // transport, orm, project_name

	// Operators
	TOKEN_EQ     TokenType = "==" // ==
	TOKEN_NEQ    TokenType = "!=" // !=
	TOKEN_AND    TokenType = "&&" // &&
	TOKEN_OR     TokenType = "||" // ||
	TOKEN_NOT    TokenType = "!"  // !
	TOKEN_LPAREN TokenType = "("
	TOKEN_RPAREN TokenType = ")"
	TOKEN_COMMA  TokenType = ","

	// Control
	TOKEN_EOF     TokenType = "EOF"
	TOKEN_ILLEGAL TokenType = "ILLEGAL"
)

// Token represents a single lexical token with its literal value and position.
type Token struct {
	Type    TokenType
	Literal string
	Pos     int // position in the original input string
}

// Node is the base interface for all AST nodes.
type Node interface {
	nodeType() string
}

// IdentNode represents an identifier (variable reference).
type IdentNode struct {
	Name string
}

func (n *IdentNode) nodeType() string { return "Ident" }

// StringLiteralNode represents a string literal.
type StringLiteralNode struct {
	Value string
}

func (n *StringLiteralNode) nodeType() string { return "StringLiteral" }

// BoolLiteralNode represents a boolean literal.
type BoolLiteralNode struct {
	Value bool
}

func (n *BoolLiteralNode) nodeType() string { return "BoolLiteral" }

// CompareNode represents a comparison expression: left == right | left != right.
type CompareNode struct {
	Left     Node
	Operator string // "==" | "!="
	Right    Node
}

func (n *CompareNode) nodeType() string { return "Compare" }

// BinaryNode represents a logical binary expression: left && right | left || right.
type BinaryNode struct {
	Left     Node
	Operator string // "&&" | "||"
	Right    Node
}

func (n *BinaryNode) nodeType() string { return "Binary" }

// NotNode represents a logical NOT expression: !expr.
type NotNode struct {
	Expr Node
}

func (n *NotNode) nodeType() string { return "Not" }

// CallNode represents function calls like contains(features, "docker").
type CallNode struct {
	FuncName string
	Args     []Node
}

func (n *CallNode) nodeType() string { return "Call" }

// Context contains the runtime values for all variables referenced in expressions
// and used during interpolation.
type Context map[string]interface{}

// Manifest models the full scaffold.yaml manifest.
type Manifest struct {
	Name         string     `yaml:"name"`
	Version      string     `yaml:"version"`
	Author       string     `yaml:"author"`
	Language     string     `yaml:"language"`
	Architecture string     `yaml:"architecture"`
	Description  string     `yaml:"description"`
	Tags         []string   `yaml:"tags"`
	Inputs       []Input    `yaml:"inputs"`
	Computed     []Computed `yaml:"computed,omitempty"`
	Files        []FileRule `yaml:"files"`
	Steps        []Step     `yaml:"steps"`
}

// Input describes a single user-provided value in scaffold.yaml.
type Input struct {
	ID       string   `yaml:"id"`
	Prompt   string   `yaml:"prompt"`
	Type     string   `yaml:"type"`
	Required bool     `yaml:"required"`
	Default  any      `yaml:"default,omitempty"`
	Validate string   `yaml:"validate,omitempty"`
	Options  []string `yaml:"options,omitempty"`
	When     string   `yaml:"when,omitempty"`
	MustExist bool    `yaml:"must_exist,omitempty"`
}

type Computed struct {
	ID    string `yaml:"id"`
	Value string `yaml:"value"`
}

// FileRule controls which files or directories are included or excluded.
type FileRule struct {
	Include string `yaml:"include,omitempty"`
	Exclude string `yaml:"exclude,omitempty"`
	When    string `yaml:"when,omitempty"`
}

// Step represents a post-generation command to execute.
type Step struct {
	Name string `yaml:"name"`
	Run  string `yaml:"run"`
	When string `yaml:"when,omitempty"`
}
