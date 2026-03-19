package cmd

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func captureStdoutRaw(t *testing.T, fn func()) string {
	t.Helper()
	var buf bytes.Buffer
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()

	fn()

	_ = w.Close()
	<-done
	os.Stdout = orig
	return buf.String()
}

func TestRootExecute_Version_Smoke(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	os.Args = []string{"structify", "version"}
	_ = captureStdoutRaw(t, func() {
		// Execute should run successfully and must not os.Exit(1).
		Execute()
	})
}

