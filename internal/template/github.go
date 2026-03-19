package template

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/jamt29/structify/internal/dsl"
)

// GitHubRef represents a normalized GitHub repository reference.
type GitHubRef struct {
	Owner  string
	Repo   string
	Ref    string
	RawURL string
}

// ParseGitHubURL parses and normalizes supported GitHub URL formats:
//   - github.com/user/repo
//   - github.com/user/repo@v1.2.0
//   - https://github.com/user/repo
//   - https://github.com/user/repo.git
//   - git@github.com:user/repo.git
func ParseGitHubURL(raw string) (*GitHubRef, error) {
	input := strings.TrimSpace(raw)
	if input == "" {
		return nil, fmt.Errorf("invalid GitHub URL: missing owner or repository")
	}

	var ref string
	if idx := strings.LastIndex(input, "@"); idx != -1 && !strings.HasPrefix(input, "git@") && !strings.Contains(input[:idx], "://") {
		ref = strings.TrimSpace(input[idx+1:])
		input = strings.TrimSpace(input[:idx])
	}

	host, owner, repo, err := parseGitHubLocation(input)
	if err != nil {
		return nil, err
	}

	if host != "github.com" {
		return nil, fmt.Errorf("unsupported host %q: only github.com is supported", host)
	}

	if owner == "" || repo == "" {
		return nil, fmt.Errorf("invalid GitHub URL: missing owner or repository")
	}

	return &GitHubRef{
		Owner:  owner,
		Repo:   repo,
		Ref:    ref,
		RawURL: raw,
	}, nil
}

func parseGitHubLocation(input string) (host, owner, repo string, _ error) {
	if strings.HasPrefix(input, "git@") {
		// SSH form: git@github.com:user/repo.git
		parts := strings.SplitN(input, ":", 2)
		if len(parts) != 2 {
			return "", "", "", fmt.Errorf("invalid GitHub URL: missing owner or repository")
		}
		hostPart := strings.TrimPrefix(parts[0], "git@")
		host = strings.TrimSpace(hostPart)
		pathPart := strings.TrimSpace(parts[1])
		pathPart = strings.TrimSuffix(pathPart, ".git")
		segs := strings.Split(pathPart, "/")
		if len(segs) != 2 {
			if len(segs) < 2 {
				return "", "", "", fmt.Errorf("invalid GitHub URL: missing owner or repository")
			}
			return "", "", "", fmt.Errorf("invalid GitHub URL: unexpected path segments after repository name")
		}
		return host, segs[0], segs[1], nil
	}

	trimmed := strings.TrimPrefix(input, "https://")
	trimmed = strings.TrimPrefix(trimmed, "http://")

	segs := strings.Split(trimmed, "/")
	if len(segs) < 3 {
		return "", "", "", fmt.Errorf("invalid GitHub URL: missing owner or repository")
	}

	host = segs[0]
	owner = segs[1]
	repo = strings.TrimSuffix(segs[2], ".git")

	if len(segs) > 3 {
		return "", "", "", fmt.Errorf("invalid GitHub URL: unexpected path segments after repository name")
	}

	return host, owner, repo, nil
}

// GitHubClient provides operations over GitHub repositories required by Structify.
type GitHubClient struct {
	httpClient *http.Client
}

// NewGitHubClient returns a GitHubClient with sane defaults.
func NewGitHubClient() *GitHubClient {
	return &GitHubClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Clone clones the given GitHub repository into destDir.
// If ref.Ref is empty, Clone will use the repository's default branch.
// For now, Clone performs a shallow clone (depth=1) regardless of ref shape.
func (c *GitHubClient) Clone(ref *GitHubRef, destDir string) error {
	if ref == nil {
		return fmt.Errorf("github ref is nil")
	}
	if strings.TrimSpace(destDir) == "" {
		return fmt.Errorf("destination directory is empty")
	}

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("creating destination dir %s: %w", destDir, err)
	}

	url := "https://github.com/" + ref.Owner + "/" + ref.Repo + ".git"

	cloneOpts := &git.CloneOptions{
		URL:   url,
		Depth: 1,
	}

	if _, err := git.PlainClone(destDir, false, cloneOpts); err != nil {
		return fmt.Errorf("cloning %s: %w", url, err)
	}

	return nil
}

// ResolveDefaultBranch returns the default branch name for the given repository
// using the public GitHub API.
func (c *GitHubClient) ResolveDefaultBranch(ref *GitHubRef) (string, error) {
	if ref == nil {
		return "", fmt.Errorf("github ref is nil")
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", ref.Owner, ref.Repo)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("creating request to resolve default branch: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("requesting default branch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("requesting default branch: unexpected status code %d", resp.StatusCode)
	}

	var payload struct {
		DefaultBranch string `json:"default_branch"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("decoding default branch response: %w", err)
	}

	if strings.TrimSpace(payload.DefaultBranch) == "" {
		return "", fmt.Errorf("default branch not found in GitHub response")
	}

	return payload.DefaultBranch, nil
}

// ValidateTemplateRepo verifies that clonedPath contains a valid Structify template.
// It expects scaffold.yaml at the repository root.
func (c *GitHubClient) ValidateTemplateRepo(clonedPath string) (*dsl.Manifest, error) {
	if strings.TrimSpace(clonedPath) == "" {
		return nil, fmt.Errorf("cloned path is empty")
	}

	manifestPath := filepath.Join(clonedPath, manifestFileName)
	m, err := dsl.LoadManifest(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("loading scaffold.yaml from cloned repo: %w", err)
	}
	if verrs := dsl.ValidateManifest(m); len(verrs) > 0 {
		var b strings.Builder
		b.WriteString("manifest validation failed:\n")
		for _, ve := range verrs {
			b.WriteString("- ")
			b.WriteString(ve.Field)
			b.WriteString(": ")
			b.WriteString(ve.Message)
			b.WriteString("\n")
		}
		return nil, fmt.Errorf("%s", strings.TrimRight(b.String(), "\n"))
	}

	return m, nil
}

