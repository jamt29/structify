package main

import "testing"

func TestMainPackageBuilds(t *testing.T) {
	// This test exists to ensure `go test ./... -cover` doesn't fail
	// on packages with no tests under the selected toolchain.
}

