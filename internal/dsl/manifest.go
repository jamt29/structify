package dsl

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ParseManifest parses scaffold YAML from bytes (no file I/O).
func ParseManifest(data []byte) (*Manifest, error) {
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing manifest: %w", err)
	}
	return &m, nil
}

func LoadManifest(path string) (*Manifest, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("loading manifest %s: %w", path, err)
	}

	m, err := ParseManifest(b)
	if err != nil {
		return nil, fmt.Errorf("parsing manifest %s: %w", path, err)
	}

	return m, nil
}
