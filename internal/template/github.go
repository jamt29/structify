package template

import (
	"fmt"
	"path"
	"strings"
)

// GitSource represents a parsed GitHub source.
type GitSource struct {
	SourceURL string
	Ref       string
}

// ParseGitSource parses inputs like:
// - github.com/user/repo
// - https://github.com/user/repo
// - github.com/user/repo@v1.2.0
// - github.com/user/repo@main
func ParseGitSource(input string) (*GitSource, error) {
	raw := strings.TrimSpace(input)
	if raw == "" {
		return nil, fmt.Errorf("source is empty")
	}

	var ref string
	if idx := strings.LastIndex(raw, "@"); idx != -1 {
		ref = strings.TrimSpace(raw[idx+1:])
		raw = strings.TrimSpace(raw[:idx])
	}

	raw = strings.TrimPrefix(raw, "https://")
	raw = strings.TrimPrefix(raw, "http://")

	if !strings.HasPrefix(raw, "github.com/") {
		return nil, fmt.Errorf("unsupported source %q: expected github.com/<user>/<repo>", input)
	}

	trimmed := strings.TrimPrefix(raw, "github.com/")
	parts := strings.Split(trimmed, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid GitHub source %q: expected github.com/<user>/<repo>", input)
	}

	user := parts[0]
	repo := parts[1]
	if user == "" || repo == "" {
		return nil, fmt.Errorf("invalid GitHub source %q: empty user or repo", input)
	}

	sourceURL := path.Join("github.com", user, repo)

	return &GitSource{
		SourceURL: sourceURL,
		Ref:       ref,
	}, nil
}

