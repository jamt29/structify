# V0.4.0 — Resultado (PARTE A, B, C)

## PARTE A — Built-ins

### Diagnóstico previo (antes de correcciones)


| Template                | Resultado inicial                                                                                                                                                                 |
| ----------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| clean-architecture-go   | `go build` / `go vet` OK; `main` no usaba `internal/transport/http` (handler/router huérfanos en disco)                                                                           |
| vertical-slice-go       | OK; `config` en raíz en lugar de `shared/config` y health sin campo `service`                                                                                                     |
| clean-architecture-ts   | `tsc` OK; `index.ts` duplicaba Express en lugar de usar `src/http/server.ts` + `routes.ts`; `npm init -y` en steps podía sustituir `package.json`; `tsconfig` con `strict: false` |
| vertical-slice-ts       | `tsc` OK; `strict: false`; rutas con `any`                                                                                                                                        |
| clean-architecture-rust | `cargo build` OK; **stub sin servidor HTTP** (solo `println`)                                                                                                                     |


### Cambios aplicados

- **clean-architecture-go**: `cmd/main.go` usa `httptransport.NewRouter()`; interfaz `Repository` con CRUD (incl. `Delete`); implementaciones GORM/sqlx/none alineadas.
- **vertical-slice-go**: `shared/config/config.go`; health JSON con `status` + `service`.
- **clean-architecture-ts**: `index.ts` arranca `./http/server` en Express; `server.ts`/`routes.ts` tipados; `tsconfig` `strict: true`; `package.json` con `dev`; eliminado `npm init` del scaffold; paso opcional `@types/express` si `runtime == express`.
- **vertical-slice-ts**: `strict: true`; `health.routes` + `package.json` `dev`; paso `@types/express` si Express.
- **clean-architecture-rust**: Axum + Tokio en `:3000`, `GET /health` → JSON `{"status":"ok"}`; `domain` con trait `Repository`; `Cargo.toml` con dependencias; `#[allow(dead_code)]` en el trait para evitar warning.

### Verificación final (generado en `/tmp/verify-*`)


| Template                     | Comando                            |
| ---------------------------- | ---------------------------------- |
| clean-architecture-go        | `go build ./...` + `go test ./...` |
| vertical-slice-go            | `go build ./...` + `go test ./...` |
| clean-architecture-ts        | `tsc --noEmit`                     |
| vertical-slice-ts            | `tsc --noEmit`                     |
| clean-architecture-rust      | `cargo build`                      |
| clean-architecture-go (gorm) | `go build ./...` (extra)           |


Estado: **PASS** en todos los casos anteriores.

---

## PARTE B — Mis templates (creación inline)

- `**n`**: abre formulario Huh (nombre, descripción, lenguaje, arquitectura) sin salir del TUI; autor desde `git config user.name` o fallback.
- **Persistencia**: `template.CreateMinimalLocalTemplate` en `~/.structify/templates/<name>/` (misma semántica que `structify template create`).
- **Recarga**: al crear (o eliminar), `RootModel` llama `engine.ListAll()` y reconstruye `TemplatesModel`, seleccionando el template creado por nombre.
- **UI**: dos columnas (Local / Built-in), fila `[+] Nuevo template` bajo locales, teclas `enter` detalle, `n` nuevo, `e` editar, `d` eliminar (solo local), `esc` volver.
- **Eliminado**: transición `Templates → structify new` con template preseleccionado (`rootPendingTemplatesToNew`); el flujo “nuevo proyecto” sigue desde el menú principal.

### Flujo visual (texto)

1. Lista: `structify · Mis templates` con columnas Local | Built-in.
2. Tras `n`: cabecera `· Mis templates · Nuevo template` + formulario Huh.
3. Tras crear: toast de confirmación y lista actualizada con el nuevo template seleccionado.

---

## PARTE C — Editor YAML inline

- Archivo: `internal/tui/yaml_editor.go` — `textarea` (bubbles), `ctrl+s` → `dsl.ParseManifest` + `dsl.ValidateManifest` → escritura de `scaffold.yaml`.
- **Esc**: si hay cambios respecto al contenido cargado, confirmación Huh “¿Descartar cambios sin guardar?”.
- **Guardado exitoso**: `done=true`, `PendingReload()` → `RootModel` recarga lista.
- **DSL**: `dsl.ParseManifest([]byte)` en `internal/dsl/manifest.go` (compartido con `LoadManifest`).

Validación en tiempo real con debounce: **no implementada** (solo al guardar).

---

## Cobertura

Salida de `go test ./... -cover` (referencia):

- `internal/dsl`: ~87%
- `internal/template`: ~73%
- `internal/engine`: ~63%
- `cmd/template`: ~68%
- `internal/tui`: ~29% (TUI interactivo)

---

## Estado final

- `go build ./...`: OK  
- `go test ./...`: OK  
- `go test ./... -cover`: OK

---

## Lecciones (L032)

**L032 — Built-ins: verificar el pipeline real de generación**

- **Contexto:** Los templates podían compilar pero el código no usaba capas generadas (p. ej. HTTP en `main` vs `internal/transport/http`) o el scaffold ejecutaba `npm init -y` tras escribir `package.json`.
- **Lección:** Tras cada cambio en `.tmpl`/`scaffold.yaml`, regenerar en `/tmp` y ejecutar el toolchain del lenguaje (`go`/`tsc`/`cargo`); para Node, revisar el orden de steps respecto a `package.json` embebido.
- **Aplicar en:** `templates/*/`, `templates/*/scaffold.yaml`

