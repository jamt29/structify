# Changelog

Todos los cambios notables de este proyecto.
Formato basado en [Keep a Changelog](https://keepachangelog.com/es).

## [Unreleased]

### Docs
- Reescritura integral de README y documentacion de comandos/DSL/template authoring.
- Mejora de textos `--help` para comandos Cobra.

## [0.5.1] - 2026-03-28

### Fixed
- Nombres de templates built-in corruptos en pantalla "Mis templates".

## [0.5.0] - 2026-03-27

### Added
- Centrado global definitivo en pantallas TUI con reglas compartidas.
- Spinner animado consistente en `stateProgress`.
- Input de URL en pantalla "Explorar GitHub".
- Pantalla "Configuracion" con valores reales (version, rutas, templates locales).
- Arbol de preview sin extension `.tmpl` visible.

### Changed
- Mejora visual de selector de templates, ayuda y estados de finalizacion.
- Mejor coherencia de layout entre `RunApp` y flujo raiz.

### Fixed
- Operaciones de templates locales alineadas con nombre real de carpeta del store.

## [0.4.0] - 2026-03-26

### Added
- Built-ins con codigo real que compila y arranca en Go, TypeScript y Rust.
- Pantalla "Mis templates" con creacion inline (`n`) desde TUI.
- Editor YAML inline para `scaffold.yaml` (textarea + validacion al guardar).

### Changed
- Refinamiento de estructura y comportamiento de templates built-in para uso real.

## [0.3.1] - 2026-03-25

### Fixed
- Dry-run de `structify new` vuelve a mostrar reporte principal en stdout.
- Salida limpia en `stateDone/stateError` cuando `new` corre como app top-level.

### Changed
- Recuperacion de cobertura en paquetes clave (`internal/config`, `cmd/template`).
- Actualizacion de skills internas y lecciones operativas.

## [0.3.0] - 2026-03-24

### Added
- Preview del arbol de archivos en tiempo real durante captura de inputs.
- Integracion de formularios Huh para inputs (`string`, `enum`, `bool`, `multiselect`, `path`).
- Transiciones visuales (fade) entre pantallas del `RootModel`.
- Logging estructurado para paths no TTY en comandos principales.

### Fixed
- Root del arbol de preview con placeholder/default cuando `project_name` aun esta vacio.
- Ancho/foco de formularios Huh al reenviar `WindowSizeMsg`.
- Sincronizacion legacy->Huh para evitar perdida de texto al escribir.

## [0.2.0] - 2026-03-23

### Added
- `structify template import <source>` (ruta local y GitHub).
- `structify template edit <name>` con validacion y opciones ante errores.
- DSL avanzado:
  - tipo `multiselect`
  - tipo `path` con `must_exist`
  - seccion `computed`
  - funcion `contains()` en expresiones `when:`
- Tests adicionales de parser/evaluator para nuevas ramas del DSL.

### Changed
- Mejoras en analizador de import para deteccion de variables sugeridas.

## [0.1.0] - 2026-03-22

### Added
- Fundacion del CLI con Cobra + Viper.
- Comandos base:
  - `structify`
  - `structify new`
  - `structify template` y subcomandos iniciales
  - `structify version`
- Implementacion completa del DSL base (`lexer`, `parser`, `evaluator`, `interpolator`, `validator`).
- Engine de scaffolding:
  - resolucion de templates
  - procesamiento de archivos con reglas `files`
  - ejecucion de `steps`
  - rollback ante errores
- Store local de templates en `~/.structify/templates`.
- Integracion inicial de TUI para flujo `new`.
- Integracion GitHub para `template add/update` con metadata de origen.
- Templates built-in iniciales.
- Pipeline de distribucion:
  - GitHub Actions (CI/release)
  - GoReleaser
  - docs base (`README`, `commands`, `dsl-reference`)

### Fixed
- Ajustes tempranos de toolchain y cobertura para estabilizar `go test ./... -cover`.

