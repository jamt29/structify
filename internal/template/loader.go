package template

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamt29/structify/internal/dsl"
	templatesfs "github.com/jamt29/structify/templates"
)

const builtinsTempDirPrefix = "structify-builtins-"

// LoadFromPath loads a template from a local filesystem path.
func LoadFromPath(path string) (*Template, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("path is empty")
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("abs path: %w", err)
	}

	manifestPath := filepath.Join(abs, manifestFileName)
	if _, err := os.Stat(manifestPath); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("template at %s is missing %s", abs, manifestFileName)
		}
		return nil, fmt.Errorf("stat manifest %s: %w", manifestPath, err)
	}

	m, err := dsl.LoadManifest(manifestPath)
	if err != nil {
		return nil, err
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
		return nil, errors.New(strings.TrimRight(b.String(), "\n"))
	}

	return &Template{
		Manifest: m,
		Path:     abs,
		Source:   detectSource(abs),
	}, nil
}

// LoadBuiltins loads templates embedded in the binary.
func LoadBuiltins() ([]*Template, error) {
	bfs := templatesfs.BuiltinTemplatesFS()

	// Ensure built-in template root exists inside the embedded FS.
	if _, err := fs.Stat(bfs, "."); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return []*Template{}, nil
		}
		return nil, fmt.Errorf("stat embedded builtins root: %w", err)
	}

	entries, err := fs.ReadDir(bfs, ".")
	if err != nil {
		return nil, fmt.Errorf("readdir embedded builtins root: %w", err)
	}

	tmpRoot, err := os.MkdirTemp("", builtinsTempDirPrefix)
	if err != nil {
		return nil, fmt.Errorf("create temp dir for builtins: %w", err)
	}

	var out []*Template
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		name := e.Name()
		srcDir := filepath.ToSlash(name)
		dstDir := filepath.Join(tmpRoot, name)

		if err := materializeEmbeddedDir(bfs, srcDir, dstDir); err != nil {
			return nil, fmt.Errorf("materializing builtin %q: %w", name, err)
		}

		t, err := LoadFromPath(dstDir)
		if err != nil {
			return nil, fmt.Errorf("loading builtin %q: %w", name, err)
		}
		t.Source = "builtin"
		out = append(out, t)
	}

	return out, nil
}

func detectSource(absPath string) string {
	templatesRoot := TemplatesDir()
	templatesRootAbs, err := filepath.Abs(templatesRoot)
	if err == nil {
		root := filepath.Clean(templatesRootAbs) + string(os.PathSeparator)
		p := filepath.Clean(absPath) + string(os.PathSeparator)
		if strings.HasPrefix(p, root) {
			return "local"
		}
	}

	base := filepath.Base(filepath.Dir(absPath))
	if strings.HasPrefix(base, builtinsTempDirPrefix) || strings.HasPrefix(filepath.Base(absPath), builtinsTempDirPrefix) {
		return "builtin"
	}

	return "github"
}

func materializeEmbeddedDir(bfs embed.FS, embeddedDir string, dstDir string) error {
	return fs.WalkDir(bfs, embeddedDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(embeddedDir, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dstDir, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		b, err := fs.ReadFile(bfs, path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, b, 0o644)
	})
}

