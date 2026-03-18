## 1) Resumen ejecutivo

Se implementó el **Engine de Scaffolding** completo (Fase 3) en `internal/engine/` y `internal/template/`: modelo de request/resultado, store local en `~/.structify/templates/`, loader con built-ins embebidos, resolver (local > built-in), procesador de archivos con `when:` + glob `**` + interpolación `{{ }}` (nombres y `.tmpl`), rollback ante fallas y ejecución de steps con `when:` e interpolación. Se agregaron tests de integración y unit tests para validar el flujo end-to-end.

## 2) Componentes creados

- `internal/template/model.go`: Tipos `Template`, `ScaffoldRequest`, `ScaffoldResult`, `StepResult`.
- `internal/template/store.go`: CRUD del store local `~/.structify/templates/` (List/Get/Exists/Add/Remove) y copia recursiva.
- `internal/template/loader.go`: `LoadFromPath` (carga+validación) y `LoadBuiltins` (carga desde `embed.FS` materializando a temp dir).
- `builtin_templates.go`: `//go:embed templates/**` y helper `BuiltinTemplatesFS()` para built-ins.
- `templates/minimal-go/*`, `templates/minimal-ts/*`: built-ins mínimos válidos (con `scaffold.yaml`).
- `internal/engine/resolver.go`: `Resolve` y `ListAll` (prioridad locales > built-ins).
- `internal/engine/file_processor.go`: `ProcessFiles` con matching `**`, `when:`, interpolación de nombres y render `.tmpl`.
- `internal/engine/rollback.go`: `RollbackManager` (Track/TrackDir/Rollback/Commit).
- `internal/engine/executor.go`: `ExecuteSteps` (when + interpolación + ejecución shell / dry-run).
- `internal/engine/engine.go`: `Engine` + `Scaffold` orquestando generación+steps+rollback.
- `internal/engine/testdata/simple_template/*`: template real para tests de integración.
- Tests:
  - `internal/engine/engine_test.go`
  - `internal/engine/file_processor_test.go`
  - `internal/engine/executor_test.go`
  - `internal/engine/rollback_test.go`
  - `internal/engine/resolver_test.go`
  - `internal/template/store_test.go`
  - `internal/template/loader_test.go`

## 3) Decisiones de implementación

- **Glob matching con `**`**: matcher por segmentos con soporte explícito para `**` (cero o más segmentos), y `path.Match` a nivel de segmento.
- **FileRules “última regla gana”**: se recorre `files[]` en orden inverso y se aplica la primera que matchee.
- **Built-ins embebidos**: `go:embed` vive en `builtin_templates.go` (raíz del repo) porque Go no permite patrones con `..`. `LoadBuiltins` materializa el FS embebido en un directorio temporal y luego reutiliza `LoadFromPath`.
- **Rollback**: se trackea el `OutputDir` (y en rollback se borra), garantizando limpieza completa si falla un step.

## 4) Cobertura de tests

Resumen por paquete (según `go test ./... -cover`):
- `internal/dsl`: **85.8%**
- `internal/engine`: **75.4%**
- `internal/template`: **75.0%**

## 5) Casos edge manejados

- **OutputDir existente con contenido**: error descriptivo, no se sobreescribe.
- **OutputDir existente pero vacío**: permitido.
- **Rollback por fallo en step**: se elimina el directorio de salida y su contenido.
- **`.tmpl` en subdirectorios**: render de contenido + remoción de extensión `.tmpl`.
- **Nombres con `{{ }}`**: interpolación por segmento de path (archivo y carpetas).
- **FileRules con `when:` inválido**: se retorna error al procesar (no se ignora silenciosamente).

## 6) Estado final

- `go build ./...` → **PASS**
- `go test ./internal/engine/... -v` → **PASS**
- `go test ./internal/template/... -v` → **PASS**
- `go test ./... -cover` → **PASS**

## 7) Lecciones capturadas

Se añadieron:
- **L005** (glob `**` requiere matcher propio)
- **L006** (`go:embed` no permite `..` en patrones)

