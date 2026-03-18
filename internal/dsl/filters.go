package dsl

import (
	"fmt"
	"strings"
	"unicode"
)

var filterFns = map[string]func(string) string{
	"snake_case":  toSnake,
	"pascal_case": toPascal,
	"camel_case":  toCamel,
	"kebab_case":  toKebab,
	"upper":       strings.ToUpper,
	"lower":       strings.ToLower,
}

func ApplyFilter(filter string, value string) (string, error) {
	fn, ok := filterFns[filter]
	if !ok {
		return "", fmt.Errorf("unknown filter '%s' (available: %s)", filter, joinAvailableFilters())
	}
	return fn(value), nil
}

func toSnake(s string) string {
	words := splitWords(s)
	for i := range words {
		words[i] = strings.ToLower(words[i])
	}
	return strings.Join(words, "_")
}

func toKebab(s string) string {
	words := splitWords(s)
	for i := range words {
		words[i] = strings.ToLower(words[i])
	}
	return strings.Join(words, "-")
}

func toPascal(s string) string {
	words := splitWords(s)
	for i := range words {
		words[i] = titleWord(words[i])
	}
	return strings.Join(words, "")
}

func toCamel(s string) string {
	words := splitWords(s)
	if len(words) == 0 {
		return ""
	}
	for i := range words {
		words[i] = titleWord(words[i])
	}
	words[0] = lowerFirst(words[0])
	return strings.Join(words, "")
}

func splitWords(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	// Normalize separators to spaces.
	var norm strings.Builder
	norm.Grow(len(s))
	for _, r := range s {
		switch r {
		case '-', '_', ' ', '\t', '\n', '\r':
			norm.WriteRune(' ')
		default:
			norm.WriteRune(r)
		}
	}

	parts := strings.Fields(norm.String())
	var out []string
	for _, part := range parts {
		out = append(out, splitCamel(part)...)
	}
	return out
}

// splitCamel splits a token on CamelCase boundaries, keeping acronyms together.
// Examples:
// - "MyProject" -> ["My", "Project"]
// - "myProject" -> ["my", "Project"]
// - "MyAPIClient" -> ["My", "API", "Client"]
func splitCamel(s string) []string {
	rs := []rune(s)
	if len(rs) == 0 {
		return nil
	}

	start := 0
	var words []string

	for i := 1; i < len(rs); i++ {
		prev := rs[i-1]
		cur := rs[i]

		prevIsLower := unicode.IsLower(prev)
		curIsUpper := unicode.IsUpper(cur)
		curIsLower := unicode.IsLower(cur)
		prevIsUpper := unicode.IsUpper(prev)

		// boundary: lower -> upper (myProject)
		if prevIsLower && curIsUpper {
			words = append(words, string(rs[start:i]))
			start = i
			continue
		}

		// boundary: acronym end before normal word (APIClient -> API + Client)
		if prevIsUpper && curIsLower && i-start > 1 {
			words = append(words, string(rs[start:i-1]))
			start = i - 1
			continue
		}
	}

	words = append(words, string(rs[start:]))
	return words
}

func titleWord(w string) string {
	if w == "" {
		return ""
	}
	rs := []rune(w)
	rs[0] = unicode.ToUpper(rs[0])
	for i := 1; i < len(rs); i++ {
		rs[i] = unicode.ToLower(rs[i])
	}
	return string(rs)
}

func lowerFirst(w string) string {
	if w == "" {
		return ""
	}
	rs := []rune(w)
	rs[0] = unicode.ToLower(rs[0])
	return string(rs)
}
