package engine

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
)

// RollbackManager tracks created files/dirs so we can roll them back on failure.
type RollbackManager struct {
	dryRun bool

	mu    sync.Mutex
	files []string
	dirs  []string
	done  bool
}

func NewRollbackManager(dryRun bool) *RollbackManager {
	return &RollbackManager{dryRun: dryRun}
}

// Track registers a file to be removed if rollback occurs.
func (r *RollbackManager) Track(path string) {
	if r == nil || r.dryRun {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.done {
		return
	}
	r.files = append(r.files, path)
}

// TrackDir registers a directory to be removed if rollback occurs.
func (r *RollbackManager) TrackDir(path string) {
	if r == nil || r.dryRun {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.done {
		return
	}
	r.dirs = append(r.dirs, path)
}

// Commit marks the operation as successful; rollback becomes a no-op.
func (r *RollbackManager) Commit() {
	if r == nil || r.dryRun {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.done = true
}

// Rollback removes tracked files and dirs in reverse order.
// It returns a combined error if any deletion fails.
func (r *RollbackManager) Rollback() error {
	if r == nil || r.dryRun {
		return nil
	}

	r.mu.Lock()
	if r.done {
		r.mu.Unlock()
		return nil
	}
	files := append([]string(nil), r.files...)
	dirs := append([]string(nil), r.dirs...)
	r.mu.Unlock()

	var errs []error

	for i := len(files) - 1; i >= 0; i-- {
		p := files[i]
		if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
			errs = append(errs, fmt.Errorf("rollback remove file %s: %w", p, err))
		}
	}
	for i := len(dirs) - 1; i >= 0; i-- {
		p := dirs[i]
		if err := os.RemoveAll(p); err != nil && !errors.Is(err, os.ErrNotExist) {
			errs = append(errs, fmt.Errorf("rollback remove dir %s: %w", p, err))
		}
	}

	return joinErrors(errs)
}

func joinErrors(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	if len(errs) == 1 {
		return errs[0]
	}
	var b strings.Builder
	b.WriteString("rollback completed with warnings:\n")
	for _, err := range errs {
		b.WriteString("- ")
		b.WriteString(err.Error())
		b.WriteString("\n")
	}
	return fmt.Errorf(strings.TrimRight(b.String(), "\n"))
}

