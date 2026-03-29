package engine

import "testing"

func TestPreviewDisplayName_stripsTmplSuffix(t *testing.T) {
	if got := previewDisplayName("routes.ts.tmpl", false); got != "routes.ts" {
		t.Fatalf("got %q want routes.ts", got)
	}
	if got := previewDisplayName("handler.go.tmpl", false); got != "handler.go" {
		t.Fatalf("got %q want handler.go", got)
	}
	if got := previewDisplayName("src", true); got != "src" {
		t.Fatalf("dir unchanged: got %q", got)
	}
}
