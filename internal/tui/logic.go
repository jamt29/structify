package tui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jamt29/structify/internal/dsl"
)

// ShouldAskInput decides whether a given input should be prompted,
// based on its `when:` expression and the accumulated context.
func ShouldAskInput(input dsl.Input, ctx dsl.Context) (bool, error) {
	when := strings.TrimSpace(input.When)
	if when == "" {
		return true, nil
	}
	ast, err := dsl.NewParser(when).Parse()
	if err != nil {
		return false, err
	}
	return dsl.Evaluate(ast, ctx)
}

// ApplyDefault returns the input's default value as a string.
// If the default is a string containing `{{ }}` interpolations, it will be interpolated.
func ApplyDefault(input dsl.Input, ctx dsl.Context) (string, error) {
	if input.Default == nil {
		switch strings.ToLower(strings.TrimSpace(input.Type)) {
		case "string", "enum":
			return "", nil
		case "bool":
			return "false", nil
		default:
			return "", nil
		}
	}

	switch def := input.Default.(type) {
	case string:
		out, err := dsl.Interpolate(def, ctx)
		if err != nil {
			return "", fmt.Errorf("interpolating default for %q: %w", strings.TrimSpace(input.ID), err)
		}
		return out, nil
	default:
		return fmt.Sprint(def), nil
	}
}

// ValidateInputValue validates a user-provided (string) value against the input definition.
// This includes `validate` regex when present.
func ValidateInputValue(input dsl.Input, value string) error {
	id := strings.TrimSpace(input.ID)
	typ := strings.ToLower(strings.TrimSpace(input.Type))
	v := strings.TrimSpace(value)

	// Handle "empty" uniformly.
	if v == "" {
		if input.Default == nil && input.Required {
			return fmt.Errorf("value is required")
		}
		// If Default exists, empty is allowed (caller may later replace it).
		return nil
	}

	switch typ {
	case "string":
		if strings.TrimSpace(input.Validate) != "" {
			re, err := regexp.Compile(input.Validate)
			if err != nil {
				return fmt.Errorf("invalid validate regex for %q: %w", id, err)
			}
			if !re.MatchString(v) {
				return fmt.Errorf("does not match %s", input.Validate)
			}
		}
		return nil

	case "enum":
		for _, opt := range input.Options {
			if v == opt {
				return nil
			}
		}
		return fmt.Errorf("must be one of: %s", strings.Join(input.Options, ", "))

	case "bool":
		if _, ok := parseBoolString(v); ok {
			return nil
		}
		return fmt.Errorf("expected y/n (or true/false)")

	default:
		return fmt.Errorf("unsupported input type %q", input.Type)
	}
}

func coerceInputValue(input dsl.Input, value string) (any, error) {
	typ := strings.ToLower(strings.TrimSpace(input.Type))
	v := strings.TrimSpace(value)

	switch typ {
	case "string", "enum":
		return v, nil
	case "bool":
		b, ok := parseBoolString(v)
		if !ok {
			return nil, fmt.Errorf("expected y/n (or true/false)")
		}
		return b, nil
	default:
		// Best-effort: keep it as string so interpolation/templates can still use it.
		return v, nil
	}
}

// BuildContext constructs a dsl.Context from inputs + an answers map.
// It evaluates each input `when:` in order, applies defaults when answers are missing,
// and applies `when: false` defaults regardless of provided answers.
func BuildContext(inputs []dsl.Input, answers map[string]string) (dsl.Context, error) {
	ctx := dsl.Context{}

	for _, in := range inputs {
		id := strings.TrimSpace(in.ID)
		if id == "" {
			continue
		}

		ask, err := ShouldAskInput(in, ctx)
		if err != nil {
			return nil, fmt.Errorf("evaluating when for input %q: %w", id, err)
		}

		if !ask {
			defStr, err := ApplyDefault(in, ctx)
			if err != nil {
				return nil, fmt.Errorf("applying default for input %q: %w", id, err)
			}
			coerced, err := coerceInputValue(in, defStr)
			if err != nil {
				return nil, fmt.Errorf("coercing default for input %q: %w", id, err)
			}
			ctx[id] = coerced
			continue
		}

		// ask == true
		if ans, ok := answers[id]; ok {
			if err := ValidateInputValue(in, ans); err != nil {
				return nil, fmt.Errorf("invalid value for %q: %w", id, err)
			}
			coerced, err := coerceInputValue(in, ans)
			if err != nil {
				return nil, fmt.Errorf("coercing value for input %q: %w", id, err)
			}
			ctx[id] = coerced
			continue
		}

		defStr, err := ApplyDefault(in, ctx)
		if err != nil {
			return nil, fmt.Errorf("applying default for input %q: %w", id, err)
		}
		if err := ValidateInputValue(in, defStr); err != nil {
			return nil, fmt.Errorf("invalid default for %q: %w", id, err)
		}
		coerced, err := coerceInputValue(in, defStr)
		if err != nil {
			return nil, fmt.Errorf("coercing default for input %q: %w", id, err)
		}
		ctx[id] = coerced
	}

	return ctx, nil
}

