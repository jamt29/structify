package template

import "testing"

func TestGitHubClient_ErrorPaths(t *testing.T) {
	c := NewGitHubClient()
	if c == nil {
		t.Fatalf("expected non-nil client")
	}

	if err := c.Clone(nil, "dest"); err == nil {
		t.Fatalf("Clone(nil, ...) expected error")
	}

	if err := c.Clone(&GitHubRef{Owner: "o", Repo: "r", Ref: ""}, ""); err == nil {
		t.Fatalf("Clone(..., empty dest) expected error")
	}

	if _, err := c.ResolveDefaultBranch(nil); err == nil {
		t.Fatalf("ResolveDefaultBranch(nil) expected error")
	}

	if _, err := c.ValidateTemplateRepo(""); err == nil {
		t.Fatalf("ValidateTemplateRepo(empty) expected error")
	}
}

