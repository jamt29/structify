package template

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/jamt29/structify/internal/dsl"
	"github.com/spf13/cobra"
)

func TestHandleEditValidationErrors_DiscardRestores(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "scaffold.yaml")
	orig := []byte("name: old\n")
	if err := os.WriteFile(path, []byte("name: broken"), 0o644); err != nil {
		t.Fatal(err)
	}

	in := bytes.NewBufferString("3\n")
	out := &bytes.Buffer{}
	cmd := &cobra.Command{}
	cmd.SetIn(in)
	cmd.SetOut(out)

	cont := handleEditValidationErrors(cmd, path, orig, []dsl.ValidationError{{Field: "name", Message: "required"}})
	if cont {
		t.Fatalf("expected stop after discard")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != string(orig) {
		t.Fatalf("expected original restored")
	}
}
