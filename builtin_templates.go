package structify

import "embed"

//go:embed templates/**
var builtinTemplates embed.FS

// BuiltinTemplatesFS returns the embedded built-in templates filesystem.
func BuiltinTemplatesFS() embed.FS {
	return builtinTemplates
}

