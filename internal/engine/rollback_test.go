package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRollbackManager_RemovesTracked(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "dir")
	file := filepath.Join(dir, "file.txt")

	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	rb := NewRollbackManager(false)
	rb.Track(file)
	rb.TrackDir(dir)

	if err := rb.Rollback(); err != nil {
		t.Fatalf("Rollback() error: %v", err)
	}

	if _, err := os.Stat(file); !os.IsNotExist(err) {
		t.Fatalf("expected file removed, stat err=%v", err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("expected dir removed, stat err=%v", err)
	}
}

