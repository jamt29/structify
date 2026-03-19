package template

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestGitHubClient_ResolveDefaultBranch_Smoke(t *testing.T) {
	c := NewGitHubClient()

	c.httpClient = &http.Client{
		Timeout: 2 * time.Second,
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			if r == nil || r.URL == nil {
				return nil, nil
			}
			if !strings.Contains(r.URL.String(), "/repos/o/r") {
				t.Fatalf("unexpected URL: %s", r.URL.String())
			}
			body := `{"default_branch":"main"}`
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	ref := &GitHubRef{Owner: "o", Repo: "r", Ref: "main"}
	got, err := c.ResolveDefaultBranch(ref)
	if err != nil {
		t.Fatalf("ResolveDefaultBranch error: %v", err)
	}
	if got != "main" {
		t.Fatalf("expected main, got %q", got)
	}
}

func TestGitHubClient_ValidateTemplateRepo_Smoke(t *testing.T) {
	c := NewGitHubClient()
	clone := t.TempDir()

	// Minimal valid scaffold.yaml for dsl validator.
	manifestYAML := "" +
		"name: \"my-template\"\n" +
		"version: \"0.0.1\"\n" +
		"author: \"test\"\n" +
		"language: \"go\"\n" +
		"architecture: \"clean\"\n" +
		"description: \"test\"\n" +
		"tags: [\"test\"]\n" +
		"inputs: []\n" +
		"files: []\n" +
		"steps: []\n"

	if err := os.WriteFile(filepath.Join(clone, manifestFileName), []byte(manifestYAML), 0o644); err != nil {
		t.Fatalf("write scaffold.yaml: %v", err)
	}

	m, err := c.ValidateTemplateRepo(clone)
	if err != nil {
		t.Fatalf("ValidateTemplateRepo error: %v", err)
	}
	if m == nil {
		t.Fatalf("expected non-nil manifest")
	}
	if got := m.Name; got != "my-template" {
		t.Fatalf("expected manifest name my-template, got %q", got)
	}
}

