package dsl

import "testing"

func TestEvalCallBool_InternalBranches(t *testing.T) {
	_, err := evalCallBool(&CallNode{FuncName: "contains", Args: []Node{&IdentNode{Name: "a"}}}, Context{"a": "x"})
	if err == nil {
		t.Fatalf("expected arg count error")
	}

	_, err = evalCallBool(&CallNode{FuncName: "unknown", Args: []Node{}}, Context{})
	if err == nil {
		t.Fatalf("expected unknown function error")
	}
}

func TestEvalValue_CallNodeReturnsBoolType(t *testing.T) {
	v, typ, err := evalValue(&CallNode{
		FuncName: "contains",
		Args:     []Node{&StringLiteralNode{Value: "abc"}, &StringLiteralNode{Value: "b"}},
	}, Context{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if typ != "bool" {
		t.Fatalf("expected bool type, got %q", typ)
	}
	if vb, ok := v.(bool); !ok || !vb {
		t.Fatalf("expected true bool value, got %#v", v)
	}
}
