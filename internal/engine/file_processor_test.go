package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jamt29/structify/internal/dsl"
	"github.com/jamt29/structify/internal/template"
)

func TestMatchGlob(t *testing.T) {
	cases := []struct {
		pattern string
		path    string
		want    bool
	}{
		{"internal/transport/http/**", "internal/transport/http/handler.go", true},
		{"internal/transport/http/**", "internal/transport/http/v1/handler.go", true},
		{"internal/transport/http/**", "internal/transport/grpc/handler.go", false},
		{"docker/*", "docker/Dockerfile", true},
		{"docker/*", "docker/nested/Dockerfile", false},
		{"plain.txt", "plain.txt", true},
		{"plain.txt", "plain.tx", false},
	}

	for _, tc := range cases {
		got, err := matchGlob(tc.pattern, tc.path)
		if err != nil {
			t.Fatalf("matchGlob(%q,%q) error: %v", tc.pattern, tc.path, err)
		}
		if got != tc.want {
			t.Fatalf("matchGlob(%q,%q)=%v, want %v", tc.pattern, tc.path, got, tc.want)
		}
	}
}

func TestProcessFiles_TmplInterpolatedAndStripped(t *testing.T) {
	tmp := t.TempDir()
	tplDir := filepath.Join(tmp, "tpl")
	srcTemplateDir := filepath.Join(tplDir, "template")
	if err := os.MkdirAll(srcTemplateDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(srcTemplateDir, "hello.txt.tmpl"), []byte("Hi {{ name }}"), 0o644); err != nil {
		t.Fatalf("write tmpl: %v", err)
	}
	m := &dsl.Manifest{
		Name:         "x",
		Version:      "0.0.1",
		Author:       "x",
		Language:     "go",
		Architecture: "clean",
		Description:  "x",
		Tags:         []string{"x"},
	}

	outDir := filepath.Join(tmp, "out")
	req := &template.ScaffoldRequest{
		Template:  &template.Template{Manifest: m, Path: tplDir, Source: "local"},
		OutputDir: outDir,
		Variables: dsl.Context{"name": "Ada"},
	}

	created, skipped, err := ProcessFiles(req)
	if err != nil {
		t.Fatalf("ProcessFiles() error: %v", err)
	}
	if len(skipped) != 0 {
		t.Fatalf("skipped=%v, want empty", skipped)
	}
	if len(created) != 1 || created[0] != "hello.txt" {
		t.Fatalf("created=%v, want [hello.txt]", created)
	}

	b, err := os.ReadFile(filepath.Join(outDir, "hello.txt"))
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if string(b) != "Hi Ada" {
		t.Fatalf("output=%q, want %q", string(b), "Hi Ada")
	}
}

func TestProcessFiles_NonTmplCopied(t *testing.T) {
	tmp := t.TempDir()
	tplDir := filepath.Join(tmp, "tpl")
	srcTemplateDir := filepath.Join(tplDir, "template")
	if err := os.MkdirAll(srcTemplateDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(srcTemplateDir, "plain.txt"), []byte("PLAIN"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	m := &dsl.Manifest{
		Name:         "x",
		Version:      "0.0.1",
		Author:       "x",
		Language:     "go",
		Architecture: "clean",
		Description:  "x",
		Tags:         []string{"x"},
	}
	outDir := filepath.Join(tmp, "out")
	req := &template.ScaffoldRequest{
		Template:  &template.Template{Manifest: m, Path: tplDir, Source: "local"},
		OutputDir: outDir,
		Variables: dsl.Context{},
	}

	created, _, err := ProcessFiles(req)
	if err != nil {
		t.Fatalf("ProcessFiles() error: %v", err)
	}
	if len(created) != 1 || created[0] != "plain.txt" {
		t.Fatalf("created=%v, want [plain.txt]", created)
	}
	b, err := os.ReadFile(filepath.Join(outDir, "plain.txt"))
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if string(b) != "PLAIN" {
		t.Fatalf("output=%q, want %q", string(b), "PLAIN")
	}
}

func TestProcessFiles_DryRunDoesNotWrite(t *testing.T) {
	tmp := t.TempDir()
	tplDir := filepath.Join(tmp, "tpl")
	srcTemplateDir := filepath.Join(tplDir, "template")
	if err := os.MkdirAll(srcTemplateDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcTemplateDir, "plain.txt"), []byte("PLAIN"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	m := &dsl.Manifest{
		Name:         "x",
		Version:      "0.0.1",
		Author:       "x",
		Language:     "go",
		Architecture: "clean",
		Description:  "x",
		Tags:         []string{"x"},
	}

	outDir := filepath.Join(tmp, "out")
	req := &template.ScaffoldRequest{
		Template:  &template.Template{Manifest: m, Path: tplDir, Source: "local"},
		OutputDir: outDir,
		Variables: dsl.Context{},
		DryRun:    true,
	}

	created, skipped, err := ProcessFiles(req)
	if err != nil {
		t.Fatalf("ProcessFiles() error: %v", err)
	}
	if len(skipped) != 0 {
		t.Fatalf("skipped=%v, want empty", skipped)
	}
	if len(created) != 1 || created[0] != "plain.txt" {
		t.Fatalf("created=%v, want [plain.txt]", created)
	}
	if _, err := os.Stat(outDir); err == nil {
		t.Fatalf("expected outDir not to exist in dry run")
	}
}

func TestProcessFiles_FileRulesIncludeWhenFalseSkips(t *testing.T) {
	tmp := t.TempDir()
	tplDir := filepath.Join(tmp, "tpl")
	srcTemplateDir := filepath.Join(tplDir, "template")
	if err := os.MkdirAll(filepath.Join(srcTemplateDir, "docker"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcTemplateDir, "docker", "Dockerfile"), []byte("X"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	m := &dsl.Manifest{
		Name:         "x",
		Version:      "0.0.1",
		Author:       "x",
		Language:     "go",
		Architecture: "clean",
		Description:  "x",
		Tags:         []string{"x"},
		Files: []dsl.FileRule{
			{Include: "docker/**", When: "use_docker == true"},
		},
	}

	outDir := filepath.Join(tmp, "out")
	req := &template.ScaffoldRequest{
		Template:  &template.Template{Manifest: m, Path: tplDir, Source: "local"},
		OutputDir: outDir,
		Variables: dsl.Context{"use_docker": false},
	}

	created, skipped, err := ProcessFiles(req)
	if err != nil {
		t.Fatalf("ProcessFiles() error: %v", err)
	}
	if len(created) != 0 {
		t.Fatalf("created=%v, want empty", created)
	}
	if len(skipped) != 1 || skipped[0] != "docker" {
		t.Fatalf("skipped=%v, want [docker]", skipped)
	}
}

func TestProcessFiles_InvalidWhenReturnsError(t *testing.T) {
	tmp := t.TempDir()
	tplDir := filepath.Join(tmp, "tpl")
	srcTemplateDir := filepath.Join(tplDir, "template")
	if err := os.MkdirAll(srcTemplateDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcTemplateDir, "plain.txt"), []byte("PLAIN"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	m := &dsl.Manifest{
		Name:         "x",
		Version:      "0.0.1",
		Author:       "x",
		Language:     "go",
		Architecture: "clean",
		Description:  "x",
		Tags:         []string{"x"},
		Files: []dsl.FileRule{
			{Exclude: "plain.txt", When: "a = \"x\""},
		},
	}

	outDir := filepath.Join(tmp, "out")
	req := &template.ScaffoldRequest{
		Template:  &template.Template{Manifest: m, Path: tplDir, Source: "local"},
		OutputDir: outDir,
		Variables: dsl.Context{"a": "x"},
	}

	if _, _, err := ProcessFiles(req); err == nil {
		t.Fatalf("expected ProcessFiles() to error on invalid when")
	}
}

