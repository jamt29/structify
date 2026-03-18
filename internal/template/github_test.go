package template

import "testing"

func TestParseGitSource_BasicForms(t *testing.T) {
	cases := []struct {
		in       string
		wantURL  string
		wantRef  string
		wantFail bool
	}{
		{"github.com/user/repo", "github.com/user/repo", "", false},
		{"https://github.com/user/repo", "github.com/user/repo", "", false},
		{"github.com/user/repo@v1.2.0", "github.com/user/repo", "v1.2.0", false},
		{"github.com/user/repo@main", "github.com/user/repo", "main", false},
		{"http://github.com/user/repo@dev", "github.com/user/repo", "dev", false},
		{"gitlab.com/user/repo", "", "", true},
		{"github.com/user", "", "", true},
		{"", "", "", true},
	}

	for _, tc := range cases {
		got, err := ParseGitSource(tc.in)
		if tc.wantFail {
			if err == nil {
				t.Fatalf("ParseGitSource(%q) expected error, got nil", tc.in)
			}
			continue
		}
		if err != nil {
			t.Fatalf("ParseGitSource(%q) unexpected error: %v", tc.in, err)
		}
		if got.SourceURL != tc.wantURL {
			t.Fatalf("ParseGitSource(%q) SourceURL=%q, want %q", tc.in, got.SourceURL, tc.wantURL)
		}
		if got.Ref != tc.wantRef {
			t.Fatalf("ParseGitSource(%q) Ref=%q, want %q", tc.in, got.Ref, tc.wantRef)
		}
	}
}

