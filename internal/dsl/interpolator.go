package dsl

import (
	"fmt"
	"sort"
	"strings"
)

func Interpolate(template string, ctx Context) (string, error) {
	var out strings.Builder

	i := 0
	for {
		start := strings.Index(template[i:], "{{")
		if start == -1 {
			out.WriteString(template[i:])
			return out.String(), nil
		}

		start += i
		out.WriteString(template[i:start])

		end := strings.Index(template[start+2:], "}}")
		if end == -1 {
			return "", fmt.Errorf("unterminated interpolation starting at position %d", start)
		}
		end = start + 2 + end

		raw := strings.TrimSpace(template[start+2 : end])
		if raw == "" {
			return "", fmt.Errorf("empty interpolation at position %d", start)
		}

		val, err := evalInterpolation(raw, ctx)
		if err != nil {
			return "", err
		}

		out.WriteString(val)
		i = end + 2
	}
}

func InterpolateFile(content []byte, ctx Context) ([]byte, error) {
	s, err := Interpolate(string(content), ctx)
	if err != nil {
		return nil, err
	}
	return []byte(s), nil
}

func evalInterpolation(expr string, ctx Context) (string, error) {
	parts := strings.Split(expr, "|")
	if len(parts) > 2 {
		return "", fmt.Errorf("filter chaining is not supported (got %d filters)", len(parts)-1)
	}

	varName := strings.TrimSpace(parts[0])
	if varName == "" {
		return "", fmt.Errorf("missing variable name in interpolation")
	}

	v, ok := ctx[varName]
	if !ok {
		return "", fmt.Errorf("variable '%s' not defined in context", varName)
	}

	strVal := fmt.Sprint(v)

	if len(parts) == 1 {
		return strVal, nil
	}

	filter := strings.TrimSpace(parts[1])
	if filter == "" {
		return "", fmt.Errorf("missing filter name in interpolation")
	}

	out, err := ApplyFilter(filter, strVal)
	if err != nil {
		return "", err
	}
	return out, nil
}

func availableFilters() []string {
	names := make([]string, 0, len(filterFns))
	for k := range filterFns {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

func joinAvailableFilters() string {
	return strings.Join(availableFilters(), ", ")
}
