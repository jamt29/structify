package structify

import "embed"

// Primary templates (non-dot files) are embedded with `templates/**`.
// We also embed dotfiles explicitly because go:embed patterns may skip files
// whose last path segment starts with '.'.
//go:embed templates/**
//go:embed templates/clean-architecture-go/template/.gitignore
//go:embed templates/vertical-slice-go/template/.gitignore
//go:embed templates/clean-architecture-ts/template/.gitignore
//go:embed templates/vertical-slice-ts/template/.gitignore
//go:embed templates/clean-architecture-rust/template/.gitignore
//go:embed templates/clean-architecture-go/template/internal/domain/entity/.gitkeep
//go:embed templates/clean-architecture-go/template/internal/domain/usecase/.gitkeep
//go:embed templates/clean-architecture-go/template/pkg/.gitkeep
//go:embed templates/vertical-slice-go/template/shared/middleware/.gitkeep
//go:embed templates/clean-architecture-ts/template/src/domain/entities/.gitkeep
//go:embed templates/clean-architecture-ts/template/src/application/use-cases/.gitkeep
//go:embed templates/clean-architecture-ts/template/src/infrastructure/repositories/.gitkeep
var builtinTemplates embed.FS

// BuiltinTemplatesFS returns the embedded built-in templates filesystem.
func BuiltinTemplatesFS() embed.FS {
	return builtinTemplates
}

