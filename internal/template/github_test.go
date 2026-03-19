package template

import "testing"

func TestParseGitHubURL_ValidFormats(t *testing.T) {
	t.Helper()

	cases := []struct {
		in      string
		wantOwn string
		wantRep string
		wantRef string
	}{
		{"github.com/user/repo", "user", "repo", ""},
		{"github.com/user/repo@v1.2.0", "user", "repo", "v1.2.0"},
		{"https://github.com/user/repo", "user", "repo", ""},
		{"https://github.com/user/repo.git", "user", "repo", ""},
		{"git@github.com:user/repo.git", "user", "repo", ""},
	}

	for _, tc := range cases {
		got, err := ParseGitHubURL(tc.in)
		if err != nil {
			t.Fatalf("ParseGitHubURL(%q) unexpected error: %v", tc.in, err)
		}
		if got.Owner != tc.wantOwn || got.Repo != tc.wantRep || got.Ref != tc.wantRef {
			t.Fatalf("ParseGitHubURL(%q) = (%q,%q,%q), want (%q,%q,%q)", tc.in, got.Owner, got.Repo, got.Ref, tc.wantOwn, tc.wantRep, tc.wantRef)
		}
	}
}

func TestParseGitHubURL_Errors(t *testing.T) {
	t.Helper()

	cases := []struct {
		in   string
		want string
	}{
		{"", "invalid GitHub URL: missing owner or repository"},
		{"gitlab.com/user/repo", "unsupported host"},
		{"github.com/user", "invalid GitHub URL: missing owner or repository"},
		{"https://github.com/user/repo/extra", "invalid GitHub URL: unexpected path segments after repository name"},
	}

	for _, tc := range cases {
		_, err := ParseGitHubURL(tc.in)
		if err == nil {
			t.Fatalf("ParseGitHubURL(%q) expected error, got nil", tc.in)
		}
		if tc.want != "" && !containsSubstring(err.Error(), tc.want) {
			t.Fatalf("ParseGitHubURL(%q) error=%q, want substring %q", tc.in, err.Error(), tc.want)
		}
	}
}

func containsSubstring(s, substr string) bool {
	return len(substr) == 0 || (len(substr) > 0 && (len(s) >= len(substr) && (func() bool {
		for i := 0; i+len(substr) <= len(s); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	})()))
}

