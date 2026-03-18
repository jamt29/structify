package template

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamt29/structify/internal/dsl"
	"gopkg.in/yaml.v3"
)

const (
	templatesRootDirName = ".structify"
	templatesDirName     = "templates"
	manifestFileName     = "scaffold.yaml"
)

// TemplatesDir returns the absolute path to ~/.structify/templates.
func TemplatesDir() string {
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		// Best-effort fallback: relative path (keeps function signature simple).
		return filepath.Join(templatesRootDirName, templatesDirName)
	}
	return filepath.Join(home, templatesRootDirName, templatesDirName)
}

// List returns all locally installed templates from ~/.structify/templates/.
// Folders without scaffold.yaml are ignored (not an error).
func List() ([]*Template, error) {
	root := TemplatesDir()
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Template{}, nil
		}
		return nil, fmt.Errorf("listing templates dir %s: %w", root, err)
	}

	var out []*Template
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		p := filepath.Join(root, e.Name())
		if !hasManifest(p) {
			continue
		}
		t, err := loadAndValidateTemplate(p, "local")
		if err != nil {
			return nil, fmt.Errorf("invalid template %s: %w", e.Name(), err)
		}
		out = append(out, t)
	}

	return out, nil
}

// Get returns a local template by directory name from ~/.structify/templates/<name>/.
func Get(name string) (*Template, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("template name is empty")
	}
	root := TemplatesDir()
	p := filepath.Join(root, name)
	if _, err := os.Stat(p); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("template %q not found", name)
		}
		return nil, fmt.Errorf("stat template %q: %w", name, err)
	}
	if !hasManifest(p) {
		return nil, fmt.Errorf("template %q does not contain %s", name, manifestFileName)
	}
	t, err := loadAndValidateTemplate(p, "local")
	if err != nil {
		return nil, fmt.Errorf("loading template %q: %w", name, err)
	}
	return t, nil
}

// Exists reports whether a local template directory exists in the store.
func Exists(name string) (bool, error) {
	if strings.TrimSpace(name) == "" {
		return false, fmt.Errorf("template name is empty")
	}
	p := filepath.Join(TemplatesDir(), name)
	_, err := os.Stat(p)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("stat template %q: %w", name, err)
}

// Add copies a local directory into the templates store.
// The destination name is the base name of sourcePath.
func Add(sourcePath string) error {
	if strings.TrimSpace(sourcePath) == "" {
		return fmt.Errorf("sourcePath is empty")
	}
	srcAbs, err := filepath.Abs(sourcePath)
	if err != nil {
		return fmt.Errorf("abs sourcePath: %w", err)
	}
	fi, err := os.Stat(srcAbs)
	if err != nil {
		return fmt.Errorf("stat sourcePath %s: %w", srcAbs, err)
	}
	if !fi.IsDir() {
		return fmt.Errorf("sourcePath %s is not a directory", srcAbs)
	}
	if !hasManifest(srcAbs) {
		return fmt.Errorf("sourcePath %s does not contain %s", srcAbs, manifestFileName)
	}
	if _, err := loadAndValidateTemplate(srcAbs, "local"); err != nil {
		return fmt.Errorf("source template is invalid: %w", err)
	}

	root := TemplatesDir()
	if err := os.MkdirAll(root, 0o755); err != nil {
		return fmt.Errorf("creating templates dir %s: %w", root, err)
	}

	name := filepath.Base(srcAbs)
	dest := filepath.Join(root, name)
	if _, err := os.Stat(dest); err == nil {
		return fmt.Errorf("template %q already exists", name)
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("stat destination %s: %w", dest, err)
	}

	if err := copyDir(srcAbs, dest); err != nil {
		return fmt.Errorf("copying template %q: %w", name, err)
	}
	return nil
}

// Remove deletes a local template from the store.
func Remove(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("template name is empty")
	}
	p := filepath.Join(TemplatesDir(), name)
	if _, err := os.Stat(p); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("template %q not found", name)
		}
		return fmt.Errorf("stat template %q: %w", name, err)
	}
	if err := os.RemoveAll(p); err != nil {
		return fmt.Errorf("removing template %q: %w", name, err)
	}
	return nil
}

func hasManifest(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, manifestFileName))
	return err == nil
}

func loadAndValidateTemplate(dir string, source string) (*Template, error) {
	m, err := dsl.LoadManifest(filepath.Join(dir, manifestFileName))
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
		return nil, fmt.Errorf("%s", strings.TrimRight(b.String(), "\n"))
	}

	meta, _ := loadTemplateMeta(dir)

	return &Template{
		Manifest: m,
		Path:     dir,
		Source:   source,
		Meta:     meta,
	}, nil
}

func loadTemplateMeta(dir string) (*TemplateMeta, error) {
	metaPath := filepath.Join(dir, ".structify-meta.yaml")
	b, err := os.ReadFile(metaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading template metadata %s: %w", metaPath, err)
	}

	var raw struct {
		SourceURL   string `yaml:"source_url"`
		SourceRef   string `yaml:"source_ref"`
		InstalledAt string `yaml:"installed_at"`
	}
	if err := yaml.Unmarshal(b, &raw); err != nil {
		return nil, fmt.Errorf("parsing template metadata %s: %w", metaPath, err)
	}

	if strings.TrimSpace(raw.SourceURL) == "" && strings.TrimSpace(raw.SourceRef) == "" && strings.TrimSpace(raw.InstalledAt) == "" {
		return nil, nil
	}

	return &TemplateMeta{
		SourceURL:   raw.SourceURL,
		SourceRef:   raw.SourceRef,
		InstalledAt: raw.InstalledAt,
	}, nil
}

// WriteTemplateMeta writes .structify-meta.yaml inside dir with the given metadata.
func WriteTemplateMeta(dir string, meta *TemplateMeta) error {
	if meta == nil {
		return nil
	}
	metaPath := filepath.Join(dir, ".structify-meta.yaml")
	raw := struct {
		SourceURL   string `yaml:"source_url"`
		SourceRef   string `yaml:"source_ref"`
		InstalledAt string `yaml:"installed_at"`
	}{
		SourceURL:   meta.SourceURL,
		SourceRef:   meta.SourceRef,
		InstalledAt: meta.InstalledAt,
	}
	b, err := yaml.Marshal(&raw)
	if err != nil {
		return fmt.Errorf("marshaling template metadata: %w", err)
	}
	if err := os.WriteFile(metaPath, b, 0o644); err != nil {
		return fmt.Errorf("writing template metadata %s: %w", metaPath, err)
	}
	return nil
}

// CopyDirForTest exposes copyDir for internal use in tests and other packages that
// need to reuse the same semantics. It is not part of the public API surface.
func CopyDirForTest(src, dst string) error {
	return copyDir(src, dst)
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		info, err := d.Info()
		if err != nil {
			return err
		}

		mode := info.Mode()
		if d.IsDir() {
			if rel == "." {
				return os.MkdirAll(dst, 0o755)
			}
			return os.MkdirAll(target, 0o755)
		}

		if mode&os.ModeSymlink != 0 {
			link, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(link, target)
		}

		return copyFile(path, target, mode)
	})
}

func copyFile(src, dst string, mode fs.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode.Perm())
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
	}()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

