package config

import (
	"bytes"
	"os"
	"testing"
)

func TestNewLogger_DefaultAndVerboseLevels(t *testing.T) {
	lInfo := NewLogger(false)
	if got := lInfo.GetLevel(); got.String() != "info" {
		t.Fatalf("expected info level, got %s", got.String())
	}

	lDbg := NewLogger(true)
	if got := lDbg.GetLevel(); got.String() != "debug" {
		t.Fatalf("expected debug level, got %s", got.String())
	}
}

func TestNewLogger_DebugEmissionDependsOnLevel(t *testing.T) {
	buf := &bytes.Buffer{}
	l := NewLogger(false)
	l.SetOutput(buf)
	l.Debug("hidden debug")
	if buf.Len() != 0 {
		t.Fatalf("expected no debug output in non-verbose mode")
	}

	l2 := NewLogger(true)
	l2.SetOutput(buf)
	l2.Debug("shown debug")
	if buf.Len() == 0 {
		t.Fatalf("expected debug output in verbose mode")
	}
}

func TestUseStructuredLogOut(t *testing.T) {
	// Non-*os.File writers (like test buffers) should keep fmt path.
	if UseStructuredLogOut(&bytes.Buffer{}) {
		t.Fatalf("expected false for bytes.Buffer writer")
	}

	// For regular files, function should return true (not terminal).
	f, err := os.CreateTemp("", "structify-log-writer-*")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	if !UseStructuredLogOut(f) {
		t.Fatalf("expected true for non-terminal *os.File writer")
	}
}
