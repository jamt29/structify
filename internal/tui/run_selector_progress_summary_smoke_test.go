package tui

import (
	"testing"

	"github.com/jamt29/structify/internal/dsl"
	"github.com/jamt29/structify/internal/template"
)

func TestRunSelector_Smoke_NoTUI(t *testing.T) {
	_, err := RunSelector(nil)
	if err == nil {
		t.Fatalf("expected error for empty templates slice")
	}

	tpl := &template.Template{
		Manifest: &dsl.Manifest{Name: "single"},
	}

	// Note: we can pass a minimal template instance; RunSelector returns immediately when len==1.
	got, err := RunSelector([]*template.Template{tpl})
	if err != nil {
		t.Fatalf("RunSelector returned error: %v", err)
	}
	if got != tpl {
		t.Fatalf("expected same template pointer")
	}
}

func TestRunProgress_Smoke_EarlyErrors(t *testing.T) {
	_, err := RunProgress(nil, nil)
	if err == nil {
		t.Fatalf("expected error for nil req")
	}

	_, err = RunProgress(&template.ScaffoldRequest{Template: nil}, nil)
	if err == nil {
		t.Fatalf("expected error for nil engine")
	}

	// Also cover nil template early checks.
	_, err = RunProgress(&template.ScaffoldRequest{
		Template:  nil,
		OutputDir: "x",
		Variables:  nil,
	}, nil)
	if err == nil {
		t.Fatalf("expected error for missing template/variables")
	}
}

func TestShowSummary_Smoke_EarlyReturn(t *testing.T) {
	// Should not panic.
	ShowSummary(nil, nil)
}

