package dsl

import "testing"

func TestNodeTypeMethods_Smoke(t *testing.T) {
	nodes := []Node{
		&IdentNode{Name: "a"},
		&StringLiteralNode{Value: "s"},
		&BoolLiteralNode{Value: true},
		&CompareNode{
			Left:     &StringLiteralNode{Value: "l"},
			Operator: "==",
			Right:    &StringLiteralNode{Value: "r"},
		},
		&BinaryNode{
			Left:     &BoolLiteralNode{Value: true},
			Operator: "&&",
			Right:    &BoolLiteralNode{Value: true},
		},
		&NotNode{
			Expr: &BoolLiteralNode{Value: true},
		},
	}

	for _, n := range nodes {
		if n.nodeType() == "" {
			t.Fatalf("expected non-empty nodeType for %T", n)
		}
	}
}

