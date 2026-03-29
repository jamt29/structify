# V061 — Resultado de iteracion

Fecha: 2026-03-29

## 1) `structify update` (Feature 1)

Implementado en:
- `cmd/update.go`
- `cmd/update_test.go`

Incluye:
- `structify update --check`
- `structify update --yes`
- comparacion semver via `isNewer(current, latest)`
- consulta de latest release en GitHub (`/releases/latest`)
- fallback cuando `go` no esta en PATH con instruccion a releases

Salida real (`--check`):

```text
Version actual:  desconocida
Ultima version:  v0.6.0

Hay una nueva version disponible.
Ejecuta: structify update
```

Salida real (`--yes`):

```text
Version actual:  desconocida
Ultima version:  v0.6.0

Actualizando...
✓ Structify actualizado a v0.6.0

Reinicia la terminal para usar la nueva version.
```

## 2) `template import` mejorado (Feature 2)

Cambios principales:
- Deteccion extendida en `internal/template/analyzer.go`:
  - Go: `module_path`, `go_version`, `package main`, conteo de binarios en `cmd/`, deps (`gorm`, `gin`, `echo`)
  - TypeScript/Node: `name`, `version`, runtime (`express`/`fastify`), flags de tests (`jest`/`vitest`), `strict`, `outDir`, `scripts.dev`
  - Rust: `name`, `version`, `workspace`, deps (`axum`, `actix-web`, `serde`)
- Nuevos campos en `AnalysisResult`:
  - `DetectedDeps`
  - `SuggestedInputs`
  - `Confidence`
- Output de revision enriquecido en `cmd/template/import.go`
- `scaffold.yaml` generado con inputs sugeridos adicionales

### Prueba manual solicitada

Comando ejecutado:

```bash
./bin/structify template import /tmp/test-import-go --name test-import --yes
```

Salida real:

```text
Template imported name=test-import path=/home/develop/.structify/templates/test-import/
files included=2 ignored=0
variables ids="module_path, go_version"
inputs detected count=7
suggested deps count=2
use template cmd="structify new --template test-import"
edit template cmd="structify template edit test-import"
```

`~/.structify/templates/test-import/scaffold.yaml` generado:

```yaml
name: test-import
version: 0.1.0
author: ""
language: go
architecture: unknown
description: Imported from test-import-go
tags: []
inputs:
    - id: module_path
      prompt: Go module path
      type: string
      required: true
      default: github.com/testuser/my-api
    - id: go_version
      prompt: Go version
      type: string
      required: true
      default: "1.21"
    - id: orm
      prompt: ORM?
      type: enum
      required: false
      default: gorm
      options: [gorm, sqlx, none]
    - id: transport
      prompt: Framework HTTP?
      type: enum
      required: false
      default: gin
      options: [gin, echo, fiber, none]
    - id: is_executable
      prompt: Es ejecutable (package main)?
      type: bool
      required: false
      default: true
    - id: binary_count
      prompt: Cantidad de binarios en /cmd?
      type: string
      required: false
      default: "1"
    - id: project_name
      prompt: Project name?
      type: string
      required: true
      validate: ^[a-zA-Z][a-zA-Z0-9_-]*$
files: []
steps:
    - name: Init go module
      run: go mod init {{ module_path }}
    - name: Tidy
      run: go mod tidy
```

## 3) Detecciones implementadas (tabla)

| Lenguaje | Detecciones |
|---|---|
| Go | `module_path`, `go_version`, `package main`, conteo `cmd/*`, deps `gorm`/`gin`/`echo`, sugerencias `orm`/`transport` |
| TypeScript/Node | `project_name`, `version`, `runtime` (`express`/`fastify`), `include_tests`, `strict_mode`, `out_dir`, `has_dev_script` |
| Rust | `project_name`, `version`, `workspace`, deps `axum`/`actix-web`/`serde`, sugerencia `transport` |

## 4) Cobertura y tests

Comando:

```bash
go test ./... -cover
```

Resultado (extracto):

```text
ok github.com/jamt29/structify/cmd              coverage: 64.4% of statements
ok github.com/jamt29/structify/cmd/template     coverage: 67.2% of statements
ok github.com/jamt29/structify/internal/dsl     coverage: 87.3% of statements
ok github.com/jamt29/structify/internal/template coverage: 64.4% of statements
```

Tests nuevos/relevantes:
- `TestIsNewer`
- `TestRunUpdate_GoNotFoundShowsManualInstructions`
- `TestAnalyzeProject_GoDependenciesSuggestInputs`

## 5) Estado final

Comandos ejecutados:

```bash
go build ./...
go test ./...
go test ./... -cover
```

Estado: OK.

## 6) Lecciones capturadas

- Se agrego `L039` en `tasks/lessons.md`:
  - reconstruir `./bin/structify` antes de verificaciones manuales de CLI.
