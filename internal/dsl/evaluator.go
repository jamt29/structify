package dsl

import (
	"fmt"
)

func Evaluate(node Node, ctx Context) (bool, error) {
	return evalBool(node, ctx)
}

func evalBool(node Node, ctx Context) (bool, error) {
	switch n := node.(type) {
	case *BoolLiteralNode:
		return n.Value, nil
	case *IdentNode:
		v, ok := ctx[n.Name]
		if !ok {
			return false, fmt.Errorf("variable '%s' not defined in context", n.Name)
		}
		b, ok := v.(bool)
		if !ok {
			return false, fmt.Errorf("variable '%s' is %T, expected bool", n.Name, v)
		}
		return b, nil
	case *NotNode:
		v, err := evalBool(n.Expr, ctx)
		if err != nil {
			return false, err
		}
		return !v, nil
	case *BinaryNode:
		switch n.Operator {
		case "&&":
			left, err := evalBool(n.Left, ctx)
			if err != nil {
				return false, err
			}
			if !left {
				return false, nil
			}
			return evalBool(n.Right, ctx)
		case "||":
			left, err := evalBool(n.Left, ctx)
			if err != nil {
				return false, err
			}
			if left {
				return true, nil
			}
			return evalBool(n.Right, ctx)
		default:
			return false, fmt.Errorf("unknown binary operator %q", n.Operator)
		}
	case *CompareNode:
		lv, lt, err := evalValue(n.Left, ctx)
		if err != nil {
			return false, err
		}
		rv, rt, err := evalValue(n.Right, ctx)
		if err != nil {
			return false, err
		}
		if lt != rt {
			return false, fmt.Errorf("cannot compare %s with %s", lt, rt)
		}
		switch lt {
		case "string":
			ls := lv.(string)
			rs := rv.(string)
			if n.Operator == "==" {
				return ls == rs, nil
			}
			if n.Operator == "!=" {
				return ls != rs, nil
			}
		case "bool":
			lb := lv.(bool)
			rb := rv.(bool)
			if n.Operator == "==" {
				return lb == rb, nil
			}
			if n.Operator == "!=" {
				return lb != rb, nil
			}
		default:
			return false, fmt.Errorf("unsupported comparison type %s", lt)
		}
		return false, fmt.Errorf("unknown comparison operator %q", n.Operator)
	case *StringLiteralNode:
		return false, fmt.Errorf("string literal is not a boolean expression")
	default:
		return false, fmt.Errorf("unknown node type %T", node)
	}
}

func evalValue(node Node, ctx Context) (any, string, error) {
	switch n := node.(type) {
	case *StringLiteralNode:
		return n.Value, "string", nil
	case *BoolLiteralNode:
		return n.Value, "bool", nil
	case *IdentNode:
		v, ok := ctx[n.Name]
		if !ok {
			return nil, "", fmt.Errorf("variable '%s' not defined in context", n.Name)
		}
		switch vv := v.(type) {
		case string:
			return vv, "string", nil
		case bool:
			return vv, "bool", nil
		default:
			return nil, "", fmt.Errorf("variable '%s' is %T, expected string or bool", n.Name, v)
		}
	default:
		// For comparisons, logical operators, and not-nodes, require boolean evaluation.
		b, err := evalBool(node, ctx)
		if err != nil {
			return nil, "", err
		}
		return b, "bool", nil
	}
}
