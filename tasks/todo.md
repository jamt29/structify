## Fase 1 — Fundación del Proyecto

- [x] 1.1 — Estructura de carpetas
- [x] 1.2 — Dependencias (go.mod + go mod tidy)
- [x] 1.3 — Setup Cobra (root, version, new, template + subcomandos)
- [x] 1.4 — Setup Viper (internal/config)
- [x] 1.5 — Makefile
- [x] 1.6 — Verificación binario y comandos
- [x] 1.7 — Cierre de fase y actualización de tareas

## Fase 2 — DSL completo

- [x] 2.1 — Spec y tipos base (`internal/dsl/types.go`)
- [x] 2.2 — Lexer (`internal/dsl/lexer.go`)
- [x] 2.3 — Parser (`internal/dsl/parser.go`)
- [x] 2.4 — Evaluator (`internal/dsl/evaluator.go`)
- [x] 2.5 — Interpolador (`internal/dsl/interpolator.go`)
- [x] 2.6 — Filtros (`internal/dsl/filters.go`)
- [x] 2.7 — Parser de scaffold.yaml (`internal/dsl/manifest.go`)
- [x] 2.8 — Validator (`internal/dsl/validator.go`)
- [x] 2.9 — Tests (`internal/dsl/`)

## Fase 3 — Engine de Scaffolding

- [x] 3.1 — Modelo de Template (`internal/template/model.go`)
- [x] 3.2 — Store de Templates (`internal/template/store.go`)
- [x] 3.3 — Loader de Templates (`internal/template/loader.go`)
- [x] 3.4 — Resolver (`internal/engine/resolver.go`)
- [x] 3.5 — Procesador de Archivos (`internal/engine/file_processor.go`)
- [x] 3.6 — Rollback (`internal/engine/rollback.go`)
- [x] 3.7 — Ejecutor de Steps (`internal/engine/executor.go`)
- [x] 3.8 — Engine Principal (`internal/engine/engine.go`)
- [x] 3.9 — Tests de Integración (`internal/engine/engine_test.go`, `internal/template/store_test.go`, `internal/engine/file_processor_test.go`)

## Fase 4 — Comando `structify new` end-to-end

- [x] 4.1 — TUI: Selector de Template (`internal/tui/selector.go`)
- [x] 4.2 — TUI: Formulario de Inputs (`internal/tui/inputs.go`)
- [x] 4.3 — TUI: Progreso de Generación (`internal/tui/progress.go`)
- [x] 4.4 — TUI: Resumen Final (`internal/tui/summary.go`)
- [x] 4.5 — Modo Flags (no-interactivo) (`cmd/new.go`)
- [x] 4.6 — Integración completa (`cmd/new.go`)

## Fase 5 — Sistema de Templates Local (`structify template *`)

- [x] 5.1 — `template list` (agrupación local/built-in, `--json`, tests)
- [x] 5.2 — `template info` (detalle con lipgloss, inputs/steps, tests)
- [x] 5.3 — `template validate` (directorio/archivo, resumen, `--json`, exit codes, tests)
- [x] 5.4 — Metadata de templates (`TemplateMeta`, `.structify-meta.yaml`, store + tests)
- [x] 5.5 — `template add` (GitHub + go-git, validación, metadata, `--force`, tests)
- [x] 5.6 — `template remove` (solo locales, confirmación interactiva, `--yes`, tests)
- [x] 5.7 — `template create` (wizard interactivo, `--output`, estructura base, tests de lógica)
- [x] 5.8 — `template update` (origen GitHub vía metadata, actualización versión, tests)
- [x] 5.9 — `template publish` (checklist interactivo, exit codes críticos, tests)
- [x] 5.10 — Verificación global (`go build ./...`, `go test ./cmd/template/... -v`, `go test ./... -cover` con cobertura ≥ ~70% en paquetes clave)

## Fase 6 — Sharing vía GitHub

- [x] 6.1 — Parser de URLs GitHub (`internal/template/github.go`)
- [x] 6.2 — Cliente GitHub (`internal/template/github.go`)
- [x] 6.3 — Comando `structify template add` (`cmd/template/add.go`)
- [x] 6.4 — Comando `structify template update` (`cmd/template/update.go`)
- [x] 6.5 — Formato estándar de repos de templates (`docs/template-format.md`)
- [x] 6.6 — Tests, verificación global y resultado de fase (`tasks/FASE6-RESULTADO.md`)

