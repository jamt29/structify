package main

import "testing"

func TestMainSmoke(t *testing.T) {
	// Smoke test to ensure `go test ./... -cover` can run coverage tooling
	// even for the root package (which otherwise has no tests).
}

