package template

import "testing"

func TestCreateHelperFunctionsCovered(t *testing.T) {
	_ = hasTTY()
	_ = detectGitUserName()
}
