package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jamt29/structify/internal/dsl"
	"github.com/jamt29/structify/internal/template"
)

func TestEngine_ScaffoldSuccess(t *testing.T) {
	tpl, err := template.LoadFromPath(filepath.Join("testdata", "simple_template"))
	if err != nil {
		t.Fatalf("LoadFromPath() error: %v", err)
	}

	out := filepath.Join(t.TempDir(), "proj")
	req := &template.ScaffoldRequest{
		Template:  tpl,
		OutputDir: out,
		Variables: dsl.Context{"project_name": "ok", "use_docker": true},
	}

	e := New()
	res, err := e.Scaffold(req)
	if err != nil {
		t.Fatalf("Scaffold() error: %v", err)
	}

	// file name interpolation + tmpl stripping
	if _, err := os.Stat(filepath.Join(out, "ok.txt")); err != nil {
		t.Fatalf("expected ok.txt: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(out, "ok.txt"))
	if err != nil {
		t.Fatalf("read ok.txt: %v", err)
	}
	if string(b) != "Name=ok\n" {
		t.Fatalf("ok.txt=%q, want %q", string(b), "Name=ok\n")
	}

	// non-tmpl copied
	if _, err := os.Stat(filepath.Join(out, "plain.txt")); err != nil {
		t.Fatalf("expected plain.txt: %v", err)
	}

	// docker included (use_docker==true)
	if _, err := os.Stat(filepath.Join(out, "docker", "Dockerfile")); err != nil {
		t.Fatalf("expected docker/Dockerfile: %v", err)
	}

	// step executed
	if _, err := os.Stat(filepath.Join(out, "marker.txt")); err != nil {
		t.Fatalf("expected marker.txt: %v", err)
	}

	if res == nil {
		t.Fatalf("result is nil")
	}
	if len(res.FilesCreated) == 0 {
		t.Fatalf("FilesCreated is empty")
	}
	if len(res.StepsExecuted) == 0 {
		t.Fatalf("StepsExecuted is empty")
	}
}

func TestEngine_DryRunDoesNotWrite(t *testing.T) {
	tpl, err := template.LoadFromPath(filepath.Join("testdata", "simple_template"))
	if err != nil {
		t.Fatalf("LoadFromPath() error: %v", err)
	}
	out := filepath.Join(t.TempDir(), "proj")

	req := &template.ScaffoldRequest{
		Template:  tpl,
		OutputDir: out,
		Variables: dsl.Context{"project_name": "ok", "use_docker": true},
		DryRun:    true,
	}

	e := New()
	_, err = e.Scaffold(req)
	if err != nil {
		t.Fatalf("Scaffold() error: %v", err)
	}
	if _, err := os.Stat(out); err == nil {
		t.Fatalf("expected output dir not to exist in dry run")
	}
}

func TestEngine_RollbackOnStepFailure(t *testing.T) {
	tpl, err := template.LoadFromPath(filepath.Join("testdata", "simple_template"))
	if err != nil {
		t.Fatalf("LoadFromPath() error: %v", err)
	}

	out := filepath.Join(t.TempDir(), "proj")
	req := &template.ScaffoldRequest{
		Template:  tpl,
		OutputDir: out,
		Variables: dsl.Context{"project_name": "fail", "use_docker": true},
	}

	e := New()
	_, err = e.Scaffold(req)
	if err == nil {
		t.Fatalf("expected scaffold to fail")
	}

	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatalf("expected output dir rolled back (missing), got stat err=%v", err)
	}
}

func TestEngine_OutputDirExistsWithContent(t *testing.T) {
	tpl, err := template.LoadFromPath(filepath.Join("testdata", "simple_template"))
	if err != nil {
		t.Fatalf("LoadFromPath() error: %v", err)
	}

	out := filepath.Join(t.TempDir(), "proj")
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatalf("mkdir out: %v", err)
	}
	if err := os.WriteFile(filepath.Join(out, "existing.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write existing: %v", err)
	}

	e := New()
	_, err = e.Scaffold(&template.ScaffoldRequest{
		Template:  tpl,
		OutputDir: out,
		Variables: dsl.Context{"project_name": "ok", "use_docker": true},
	})
	if err == nil {
		t.Fatalf("expected error due to non-empty output dir")
	}
}

func TestEngine_FileRulesExcludeDocker(t *testing.T) {
	tpl, err := template.LoadFromPath(filepath.Join("testdata", "simple_template"))
	if err != nil {
		t.Fatalf("LoadFromPath() error: %v", err)
	}

	out := filepath.Join(t.TempDir(), "proj")
	e := New()
	_, err = e.Scaffold(&template.ScaffoldRequest{
		Template:  tpl,
		OutputDir: out,
		Variables: dsl.Context{"project_name": "ok", "use_docker": false},
	})
	if err != nil {
		t.Fatalf("Scaffold() error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(out, "docker", "Dockerfile")); err == nil {
		t.Fatalf("expected docker/Dockerfile to be skipped")
	}
}

func TestEngine_StepWhenFalseIsSkipped(t *testing.T) {
	tpl, err := template.LoadFromPath(filepath.Join("testdata", "simple_template"))
	if err != nil {
		t.Fatalf("LoadFromPath() error: %v", err)
	}

	out := filepath.Join(t.TempDir(), "proj")
	e := New()
	res, err := e.Scaffold(&template.ScaffoldRequest{
		Template:  tpl,
		OutputDir: out,
		Variables: dsl.Context{"project_name": "ok", "use_docker": true},
	})
	if err != nil {
		t.Fatalf("Scaffold() error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(out, "skipped_step.txt")); err == nil {
		t.Fatalf("expected skipped_step.txt not to exist")
	}

	var foundSkipped bool
	for _, sr := range res.StepsExecuted {
		if sr.Name == "Skipped step" {
			foundSkipped = true
			if !sr.Skipped {
				t.Fatalf("expected step to be marked skipped")
			}
		}
	}
	if !foundSkipped {
		t.Fatalf("expected to find 'Skipped step' in results")
	}
}

