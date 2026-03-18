package engine

import (
	"strings"
	"testing"
)

func TestResolve_Builtin(t *testing.T) {
	tpl, err := Resolve("minimal-go")
	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}
	if tpl == nil || tpl.Manifest == nil {
		t.Fatalf("Resolve() returned nil template/manifest")
	}
	if tpl.Source != "builtin" {
		t.Fatalf("Source=%q, want builtin", tpl.Source)
	}
}

func TestResolve_MissingListsAvailable(t *testing.T) {
	_, err := Resolve("does-not-exist")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "available") {
		t.Fatalf("expected error to include available templates, got: %v", err)
	}
}

