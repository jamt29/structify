## 1) Resumen ejecutivo

Se implementó `structify new` **end-to-end** con modo **interactivo (TUI wizard)** y modo **no-interactivo por flags** (CI-friendly), incluyendo **dry-run**, **modo mixto flags+TUI**, progreso en tiempo real y pantalla de resumen final. Además se agregaron tests de lógica (sin testear rendering) y se ajustó el Makefile para producir `./bin/structify` desde `./cmd/structify`.

## 2) Componentes creados

- `internal/tui/selector.go`: selector de templates con `bubbles/list` + búsqueda incremental y edge cases (0/1 templates).
- `internal/tui/inputs.go`: wizard de inputs (`when`, defaults, required, regex, enum/bool) + soporte de contexto inicial para modo mixto.
- `internal/tui/progress.go`: spinner + progreso de steps en tiempo real mientras corre el scaffold.
- `internal/tui/summary.go`: resumen final con lipgloss y “Next steps” por lenguaje.
- `cmd/new.go`: implementación completa del comando `structify new` (TTY/no-TTY, flags, mixto, dry-run).
- `internal/engine/executor.go`: observer opcional para reportar progreso de steps sin romper compatibilidad.

## 3) Decisiones de implementación

- **Detección de TTY**: se usa `golang.org/x/term` para decidir si se puede correr TUI; sin TTY se exige modo flags completo.
- **Modo mixto (flags + TUI)**: `--name` precarga `project_name` y `--var key=value` precarga variables; si faltan inputs requeridos activos según `when:`, se completa lo faltante con TUI (`RunInputsWithInitial`).
- **Progreso en tiempo real**: se extendió el ejecutor de steps con un `StepObserver` (`ExecuteStepsWithObserver`) para poder renderizar `✓/─/✗` durante la ejecución.

## 4) Demo de uso

### `./bin/structify new --help`

```
Create a new project from a Structify template.
Use flags to select the template, project name, additional variables, and output directory.

Usage:
  structify new [flags]

Flags:
      --dry-run           show what would be generated without writing files
  -h, --help              help for new
      --name string       name of the project to create
      --output string     output directory for the generated project
      --template string   template name or path to use
      --var stringArray   additional variables in key=value form (repeatable)

Global Flags:
      --config string   config file (default is $HOME/.structify/config.yaml)
      --verbose         enable verbose output
```

### `./bin/structify new --template minimal-go --name myapp --dry-run`

```
Dry run — no files will be written.
Template : minimal-go
Output   : ./myapp
Variables: project_name=myapp, orm=none
Files that would be created:
cmd/main.go
go.mod
internal/app/app.go
Steps that would run:
✓ go mod tidy
─ go get gorm.io/gorm  (skipped: orm == "gorm")
No files were written.
```

## 5) Cobertura de tests

Salida de `go test ./... -cover` (por paquete):

- `github.com/jamt29/structify`: 100.0%
- `github.com/jamt29/structify/cmd`: 27.1%
- `github.com/jamt29/structify/cmd/structify`: 0.0% (smoke test para que `-cover` no falle)
- `github.com/jamt29/structify/cmd/template`: 50.0%
- `github.com/jamt29/structify/internal/config`: 81.8%
- `github.com/jamt29/structify/internal/dsl`: 85.8%
- `github.com/jamt29/structify/internal/engine`: 73.6%
- `github.com/jamt29/structify/internal/template`: 75.0%
- `github.com/jamt29/structify/internal/tui`: 7.0% (tests de lógica, no rendering)

## 6) Estado final

- `go build ./...` → PASS
- `go test ./...` → PASS
- `go test ./... -cover` → PASS

## 7) Lecciones capturadas

Se añadió **L007** a `tasks/lessons.md` sobre:
- evitar `fmt.Errorf(msg)` (usar `errors.New` o `fmt.Errorf("%s", msg)`) bajo chequeos printf/vet,
- y garantizar que `go test ./... -cover` no falle en paquetes sin tests agregando smoke tests mínimos.

