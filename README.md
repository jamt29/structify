# Structify

Multilenguaje project scaffolding CLI basado en arquitecturas de software.

[![CI](https://github.com/jamt29/structify/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/jamt29/structify/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## What is Structify

Structify ayuda a desarrolladores a arrancar proyectos con una estructura consistente y lista para extender, reduciendo el tiempo dedicado a copiar/pegar boilerplate.

En lugar de comenzar desde cero, eliges una arquitectura (por ejemplo Clean Architecture o Vertical Slice) y un lenguaje (Go, TypeScript, Rust). Structify genera la estructura base, incluyendo endpoints de ejemplo, capas y (cuando aplica) stubs por transporte u ORM.

## Quick Start

```bash
go install github.com/jamt29/structify@latest
structify new --template clean-architecture-go --name my-app --output ./my-app
```

## Installation

### Homebrew (coming soon)

```bash
brew install structify # coming soon
```

### Script (placeholder)

```bash
curl -fsSL https://structify.dev/install.sh | sh # coming soon
```

### Go install

```bash
go install github.com/jamt29/structify@latest
```

### Binary releases

Descarga el binario desde [GitHub Releases](https://github.com/jamt29/structify/releases).

## Built-in Templates

| Name | Language | Architecture | Description |
|---|---|---|---|
| `clean-architecture-go` | Go | clean | Clean Architecture en Go con stubs HTTP/gRPC |
| `vertical-slice-go` | Go | vertical-slice | Vertical Slice en Go con `GET /health` |
| `clean-architecture-ts` | TypeScript | clean | Clean Architecture en TS con runtime express/fastify (stubs) |
| `vertical-slice-ts` | TypeScript | vertical-slice | Vertical Slice en TS con `GET /health` |
| `clean-architecture-rust` | Rust | clean | Clean Architecture en Rust con transport axum/actix (stubs) |

## Usage

### Crear un proyecto (modo interactivo)

```bash
structify new
```

### Crear un proyecto (modo no interactivo)

```bash
structify new --template clean-architecture-go --name testproject --output /tmp/testproj --dry-run
```

Output esperado (ejemplo):

```text
Dry run — no files will be written.
Template : clean-architecture-go
Files that would be created:
  cmd/main.go
  internal/transport/http/router.go
Steps that would run:
  ✓ go mod init ...
  ✓ go mod tidy
```

### Gestionar templates

```bash
structify template list
structify template info clean-architecture-go
structify template validate ./path-to-template
```

### Instalar templates desde GitHub

```bash
structify template add github.com/<user>/<repo>
```

## Creating Templates

Formato del repo de templates: [docs/template-format.md](docs/template-format.md).

Referencia del DSL `scaffold.yaml`: [docs/dsl-reference.md](docs/dsl-reference.md).

## Contributing

1. Clona el repo.
2. Ejecuta tests: `go test ./...`
3. Verifica el build: `go build ./...`

### Release con GoReleaser

Si no tienes GoReleaser instalado:

```bash
go install github.com/goreleaser/goreleaser@latest
```

Luego valida la config:

```bash
goreleaser check
```

## License

MIT

