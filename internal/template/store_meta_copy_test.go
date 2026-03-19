package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteTemplateMeta_Smoke(t *testing.T) {
	dir := t.TempDir()

	if err := WriteTemplateMeta(dir, nil); err != nil {
		t.Fatalf("WriteTemplateMeta(nil) error: %v", err)
	}

	meta := &TemplateMeta{
		SourceURL:   "github.com/user/repo",
		SourceRef:   "v1.2.3",
		InstalledAt: "2026-01-01T00:00:00Z",
	}
	if err := WriteTemplateMeta(dir, meta); err != nil {
		t.Fatalf("WriteTemplateMeta(meta) error: %v", err)
	}

	b, err := os.ReadFile(filepath.Join(dir, ".structify-meta.yaml"))
	if err != nil {
		t.Fatalf("read meta file: %v", err)
	}
	s := string(b)
	if !containsAll(s, []string{"source_url", "source_ref", "installed_at"}) {
		t.Fatalf("unexpected meta yaml content: %s", s)
	}
}

func TestCopyDirForTest_Smoke(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	// Create a nested structure + symlink to exercise copyDir.
	sub := filepath.Join(src, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatalf("mkdir sub: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sub, "file.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := os.Symlink(filepath.Join("sub", "file.txt"), filepath.Join(src, "link.txt")); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	if err := CopyDirForTest(src, dst); err != nil {
		t.Fatalf("CopyDirForTest error: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dst, "sub", "file.txt"))
	if err != nil {
		t.Fatalf("read copied file: %v", err)
	}
	if string(got) != "hello" {
		t.Fatalf("copied file content mismatch: %q", string(got))
	}

	// Ensure symlink exists (we don't verify target, but it should be a symlink).
	if fi, err := os.Lstat(filepath.Join(dst, "link.txt")); err != nil {
		t.Fatalf("lstat copied symlink: %v", err)
	} else if fi.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("expected copied link.txt to be symlink")
	}
}

func containsAll(s string, subs []string) bool {
	for _, sub := range subs {
		if !strings.Contains(s, sub) {
			return false
		}
	}
	return true
}

