package dsl

import (
	"fmt"
	"regexp"
	"strings"
)

type ValidationError struct {
	Field   string
	Message string
}

func ValidateManifest(m *Manifest) []ValidationError {
	if m == nil {
		return []ValidationError{{Field: "manifest", Message: "manifest is nil"}}
	}

	var errs []ValidationError

	trim := func(s string) string { return strings.TrimSpace(s) }

	if trim(m.Name) == "" {
		errs = append(errs, ValidationError{Field: "name", Message: "name is required"})
	}

	if trim(m.Version) == "" {
		errs = append(errs, ValidationError{Field: "version", Message: "version is required"})
	} else if !semverRe.MatchString(trim(m.Version)) {
		errs = append(errs, ValidationError{Field: "version", Message: "version must be in format X.Y.Z"})
	}

	if trim(m.Language) == "" {
		errs = append(errs, ValidationError{Field: "language", Message: "language is required"})
	} else if !allowedLanguage(trim(m.Language)) {
		errs = append(errs, ValidationError{Field: "language", Message: fmt.Sprintf("unsupported language %q", trim(m.Language))})
	}

	// Inputs
	ids := map[string]int{}
	for i := range m.Inputs {
		in := m.Inputs[i]
		fieldPrefix := fmt.Sprintf("inputs[%d]", i)

		if trim(in.ID) == "" {
			errs = append(errs, ValidationError{Field: fieldPrefix + ".id", Message: "id is required"})
		} else if !inputIDRe.MatchString(in.ID) {
			errs = append(errs, ValidationError{Field: fieldPrefix + ".id", Message: "id must match ^[a-z_]+$"})
		} else {
			if prev, ok := ids[in.ID]; ok {
				errs = append(errs, ValidationError{
					Field:   fieldPrefix + ".id",
					Message: fmt.Sprintf("duplicate input id %q (also defined at inputs[%d].id)", in.ID, prev),
				})
			} else {
				ids[in.ID] = i
			}
		}

		t := trim(in.Type)
		if t == "" {
			errs = append(errs, ValidationError{Field: fieldPrefix + ".type", Message: "type is required"})
		} else if t != "string" && t != "enum" && t != "bool" {
			errs = append(errs, ValidationError{Field: fieldPrefix + ".type", Message: fmt.Sprintf("unsupported input type %q", t)})
		}

		if t == "enum" && len(in.Options) == 0 {
			errs = append(errs, ValidationError{Field: fieldPrefix + ".options", Message: "options is required when type is enum"})
		}

		if trim(in.When) != "" {
			if _, err := NewParser(in.When).Parse(); err != nil {
				errs = append(errs, ValidationError{Field: fieldPrefix + ".when", Message: err.Error()})
			}
		}
	}

	// Circular input when dependencies (edge case).
	errs = append(errs, validateInputWhenCycles(m)...)

	// Files
	for i := range m.Files {
		r := m.Files[i]
		fieldPrefix := fmt.Sprintf("files[%d]", i)

		if trim(r.Include) != "" && trim(r.Exclude) != "" {
			errs = append(errs, ValidationError{Field: fieldPrefix, Message: "include and exclude are mutually exclusive"})
		}
		if trim(r.Include) == "" && trim(r.Exclude) == "" {
			errs = append(errs, ValidationError{Field: fieldPrefix, Message: "either include or exclude is required"})
		}
		if trim(r.When) != "" {
			if _, err := NewParser(r.When).Parse(); err != nil {
				errs = append(errs, ValidationError{Field: fieldPrefix + ".when", Message: err.Error()})
			}
		}
	}

	// Steps
	for i := range m.Steps {
		s := m.Steps[i]
		fieldPrefix := fmt.Sprintf("steps[%d]", i)

		if trim(s.Name) == "" {
			errs = append(errs, ValidationError{Field: fieldPrefix + ".name", Message: "name is required"})
		}
		if trim(s.Run) == "" {
			errs = append(errs, ValidationError{Field: fieldPrefix + ".run", Message: "run is required"})
		}
		if trim(s.When) != "" {
			if _, err := NewParser(s.When).Parse(); err != nil {
				errs = append(errs, ValidationError{Field: fieldPrefix + ".when", Message: err.Error()})
			}
		}
	}

	return errs
}

var (
	semverRe  = regexp.MustCompile(`^\d+\.\d+\.\d+$`)
	inputIDRe = regexp.MustCompile(`^[a-z_]+$`)
	languages = map[string]struct{}{"go": {}, "typescript": {}, "rust": {}, "csharp": {}, "python": {}}
)

func allowedLanguage(lang string) bool {
	_, ok := languages[lang]
	return ok
}

func validateInputWhenCycles(m *Manifest) []ValidationError {
	// Build dependency graph: input id -> referenced input ids inside its when.
	deps := map[string][]string{}
	for i := range m.Inputs {
		id := strings.TrimSpace(m.Inputs[i].ID)
		if id == "" {
			continue
		}
		when := strings.TrimSpace(m.Inputs[i].When)
		if when == "" {
			continue
		}
		ast, err := NewParser(when).Parse()
		if err != nil {
			continue // already reported elsewhere
		}
		refSet := map[string]struct{}{}
		collectIdents(ast, refSet)
		for ref := range refSet {
			if _, ok := findInputIndex(m, ref); ok && ref != id {
				deps[id] = append(deps[id], ref)
			}
		}
	}

	vis := map[string]int{} // 0 unvisited, 1 visiting, 2 done
	var stack []string
	var errs []ValidationError

	var dfs func(string)
	dfs = func(n string) {
		vis[n] = 1
		stack = append(stack, n)

		for _, nxt := range deps[n] {
			if vis[nxt] == 0 {
				dfs(nxt)
				continue
			}
			if vis[nxt] == 1 {
				// cycle found: from nxt to end of stack
				cycle := cycleFromStack(stack, nxt)
				msg := fmt.Sprintf("circular when dependency: %s", strings.Join(cycle, " -> "))
				for _, id := range cycle {
					if idx, ok := findInputIndex(m, id); ok {
						errs = append(errs, ValidationError{Field: fmt.Sprintf("inputs[%d].when", idx), Message: msg})
					}
				}
			}
		}

		stack = stack[:len(stack)-1]
		vis[n] = 2
	}

	for id := range deps {
		if vis[id] == 0 {
			dfs(id)
		}
	}

	return errs
}

func cycleFromStack(stack []string, start string) []string {
	var idx int
	for i := range stack {
		if stack[i] == start {
			idx = i
			break
		}
	}
	cycle := append([]string{}, stack[idx:]...)
	cycle = append(cycle, start)
	return cycle
}

func findInputIndex(m *Manifest, id string) (int, bool) {
	for i := range m.Inputs {
		if m.Inputs[i].ID == id {
			return i, true
		}
	}
	return -1, false
}

func collectIdents(n Node, out map[string]struct{}) {
	switch t := n.(type) {
	case *IdentNode:
		out[t.Name] = struct{}{}
	case *CompareNode:
		collectIdents(t.Left, out)
		collectIdents(t.Right, out)
	case *BinaryNode:
		collectIdents(t.Left, out)
		collectIdents(t.Right, out)
	case *NotNode:
		collectIdents(t.Expr, out)
	}
}
