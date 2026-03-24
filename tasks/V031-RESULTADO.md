# V031-RESULTADO — Consolidación v0.3.1

## 1) Bug A — dry-run sin output visible

### Causa raíz confirmada
`runDryRun()` había migrado a logging estructurado en stderr (v0.3.0b). En flujos donde el usuario inspecciona stdout (o redirige stdout), parecía “sin salida”.

### Fix aplicado
- `cmd/new.go`:
  - `runDryRun(req, eng)` vuelve a imprimir TODO el informe con `fmt.Fprint*` a `os.Stdout`.
  - Se mantiene `charm/log` para otros mensajes no-interactivos (progreso/errores) fuera del reporte principal dry-run.

### Output real (`/tmp/out.txt`)
Comando:
```bash
go run ./cmd/structify new --template clean-architecture-go --name testproject --dry-run > /tmp/out.txt
```

Extracto real:
```text
Dry run — no files will be written.
Template : clean-architecture-go
Output   : ./testproject
Variables: project_name=testproject, module_path=github.com/user/testproject, orm=none, transport=http
Files that would be created:
.gitignore
Makefile
README.md
...
Steps that would run:
✓ go mod init github.com/user/testproject
...
No files were written.
```

## 2) Bug C — stateDone no sale con tecla

### Elección aplicada
**Opción A** (mínima y correcta para `RunApp`).

### Confirmación de path
`cmd/new.go` con TTY llama `tui.RunApp(...)` (no `tui.Run(...)`), por lo que `App` corre como top-level program.

### Fix aplicado
- `internal/tui/app.go`:
  - Nuevo campo `quitOnDoneKey bool`.
  - `RunApp` activa `m.quitOnDoneKey = true`.
  - En `stateDone/stateError`, al recibir tecla:
    - siempre marca `a.done = true`
    - si `quitOnDoneKey == true`, retorna `tea.Quit`.

Esto preserva el comportamiento embebido con `RootModel` (transición interna) y resuelve salida limpia en `RunApp`.

## 3) Cobertura — antes/después

| Paquete | Antes | Después |
|---|---:|---:|
| `internal/config` | 47.4% | 89.5% |
| `cmd/template` | 45.3% | 68.1% |
| `internal/dsl` | 87.2% | 87.2% |
| `internal/engine` | 62.9% | 62.9% |

### Tests añadidos para recuperar cobertura
- `internal/config/logger_test.go`
- `cmd/template/import_helpers_test.go`
- `cmd/template/create_helpers_test.go`
- `cmd/template/command_paths_test.go`
- `internal/tui/app_test.go` (caso `stateDone`/`quitOnDoneKey`)

## 4) SKILLs actualizados

- `tasks/SKILL-structify.md`
  - estructura de carpetas/archivos actual v0.3.x
  - dependencias (`huh`, `log`)
  - flujo TUI con RootModel + App + preview
  - comandos `template import` y `template edit`
  - estado de fases F1–F8 + v0.2.0 + v0.3.0 completadas

- `tasks/SKILL-dsl.md`
  - nuevos inputs: `multiselect`, `path` (`must_exist`)
  - sección `computed`
  - `contains()` en expresiones `when:`
  - casos obligatorios de tests para `contains()`

- `tasks/SKILL-workflow.md`
  - referencia rápida de archivos clave
  - flujo operativo consolidado
  - resumen de lecciones L015–L030

## 5) Estado final

- `go build ./...` → OK
- `go test ./... -cover` → OK
- Objetivos de cobertura cumplidos:
  - `internal/config >= 75` (89.5)
  - `cmd/template >= 60` (68.1)

## 6) Lecciones nuevas

- `L031` añadida en `tasks/lessons.md`:
  - política explícita de salida por entrypoint (`RunApp` vs `RootModel`) en `stateDone/stateError`.

