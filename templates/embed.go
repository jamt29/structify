package templates

import "embed"

// Primary templates are embedded by directory.
// We also embed dotfiles explicitly because go:embed patterns may skip files
// whose last path segment starts with '.'.
//go:embed clean-architecture-go vertical-slice-go clean-architecture-ts vertical-slice-ts clean-architecture-rust minimal-go
//go:embed clean-architecture-go/template/.gitignore
//go:embed vertical-slice-go/template/.gitignore
//go:embed clean-architecture-ts/template/.gitignore
//go:embed vertical-slice-ts/template/.gitignore
//go:embed clean-architecture-rust/template/.gitignore
//go:embed clean-architecture-go/template/internal/domain/entity/.gitkeep
//go:embed clean-architecture-go/template/internal/domain/usecase/.gitkeep
//go:embed clean-architecture-go/template/pkg/.gitkeep
//go:embed vertical-slice-go/template/shared/middleware/.gitkeep
//go:embed clean-architecture-ts/template/src/domain/entities/.gitkeep
//go:embed clean-architecture-ts/template/src/application/use-cases/.gitkeep
//go:embed clean-architecture-ts/template/src/infrastructure/repositories/.gitkeep
var builtinTemplates embed.FS

// BuiltinTemplatesFS returns the embedded built-in templates filesystem.
func BuiltinTemplatesFS() embed.FS {
	return builtinTemplates
}

