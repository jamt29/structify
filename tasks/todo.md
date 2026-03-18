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

