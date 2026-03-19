package dsl

import "testing"

func TestInterpolateFile_Smoke(t *testing.T) {
	ctx := Context{
		"project_name": "MyProject",
	}
	out, err := InterpolateFile([]byte("Hello {{ project_name }}"), ctx)
	if err != nil {
		t.Fatalf("InterpolateFile returned error: %v", err)
	}
	if got := string(out); got != "Hello MyProject" {
		t.Fatalf("got %q want %q", got, "Hello MyProject")
	}
}

