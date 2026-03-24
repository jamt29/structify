package template

import "testing"

func TestCmdIsDefined(t *testing.T) {
	if Cmd == nil {
		t.Fatalf("expected Cmd to be defined")
	}
}
