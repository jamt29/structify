package engine

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jamt29/structify/internal/template"
)

// Resolve finds a template by name, preferring local templates over built-ins.
func Resolve(name string) (*template.Template, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("template name is empty")
	}

	if t, err := template.Get(name); err == nil {
		return t, nil
	}

	builtins, err := template.LoadBuiltins()
	if err != nil {
		return nil, fmt.Errorf("loading built-in templates: %w", err)
	}
	for _, t := range builtins {
		if t.Manifest != nil && t.Manifest.Name == name {
			return t, nil
		}
	}

	all, _ := ListAll()
	avail := make([]string, 0, len(all))
	for _, t := range all {
		if t.Manifest != nil && strings.TrimSpace(t.Manifest.Name) != "" {
			avail = append(avail, t.Manifest.Name)
		}
	}
	sort.Strings(avail)

	if len(avail) == 0 {
		return nil, fmt.Errorf("template %q not found (no templates available)", name)
	}
	return nil, fmt.Errorf("template %q not found (available: %s)", name, strings.Join(avail, ", "))
}

// ListAll returns all templates (local + built-in), without duplicates by name.
// Local templates take priority over built-ins when names collide.
func ListAll() ([]*template.Template, error) {
	locals, err := template.List()
	if err != nil {
		return nil, err
	}
	builtins, err := template.LoadBuiltins()
	if err != nil {
		return nil, err
	}

	byName := map[string]*template.Template{}
	for _, t := range builtins {
		if t == nil || t.Manifest == nil || strings.TrimSpace(t.Manifest.Name) == "" {
			continue
		}
		byName[t.Manifest.Name] = t
	}
	for _, t := range locals {
		if t == nil || t.Manifest == nil || strings.TrimSpace(t.Manifest.Name) == "" {
			continue
		}
		byName[t.Manifest.Name] = t
	}

	names := make([]string, 0, len(byName))
	for n := range byName {
		names = append(names, n)
	}
	sort.Strings(names)

	out := make([]*template.Template, 0, len(names))
	for _, n := range names {
		out = append(out, byName[n])
	}
	return out, nil
}

