package main

import (
	"os"
	"testing"
)

func TestMainPackageExecute_Version_Smoke(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	os.Args = []string{"structify", "version"}
	main()
}

