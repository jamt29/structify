# Structify — Contexto consolidado (v0.3.x)

## Estado del proyecto
- Fases F1–F8: completadas.
- V0.2.0 (import/edit + DSL avanzado): completada.
- V0.3.0a (preview + Huh): completada.
- V0.3.0b (fixes preview/help/log/transiciones): completada.

## Estructura clave (actualizada)
- `cmd/`
  - `cmd/new.go` (flujo interactivo/no-interactivo, dry-run, ejecución)
  - `cmd/template/*.go` (list/info/validate/add/remove/create/update/publish/import/edit)
- `internal/tui/`
  - `app.go` (modelo principal del flujo `new`)
  - `root.go` (RootModel multipantalla y transiciones)
  - `menu.go`, `welcome.go`, `layout.go`
  - `tree.go` (render del preview de árbol)
  - `huh_inputs.go` (formularios Huh)
  - `styles.go` (paleta y estilos)
  - `templates_screen.go`, `github_screen.go`, `config_screen.go`
- `internal/engine/preview.go` (preview de archivos/steps)
- `internal/template/analyzer.go` (análisis para import)
- `internal/config/logger.go` (logger + heurística structured output)

## Dependencias relevantes
- `github.com/charmbracelet/bubbletea`
- `github.com/charmbracelet/lipgloss`
- `github.com/charmbracelet/huh`
- `github.com/charmbracelet/log`
- `github.com/spf13/cobra`, `github.com/spf13/viper`

## Flujo TUI consolidado
- Entrada `structify` (sin subcomando) usa `tui.Run(...)` con `RootModel`.
- Pantallas RootModel: menú principal, `new`, templates, GitHub, config.
- Transiciones entre pantallas en `root.go` (fade con `tea.Tick`).
- En `screenNew`, el `App` gestiona estados internos:
  - selector de template
  - inputs (Huh + preview de árbol en tiempo real)
  - confirmación
  - progreso
  - done/error

## Flujo `structify new`
- Si hay TTY y modo interactivo: `tui.RunApp(...)`.
- Sin TTY / flags-only:
  - valida y completa contexto
  - `--dry-run` imprime reporte en **stdout**
  - ejecución real usa observador de steps y log estructurado para progreso

## Comandos CLI activos
- `structify new`
- `structify template list|info|validate|add|remove|create|update|publish`
- `structify template import <source>`
- `structify template edit <name>`

## Restricciones de mantenimiento
- No tocar `internal/dsl/`, `internal/engine/`, `internal/template/` salvo tarea explícita.
- Mantener compatibilidad de tests existentes.
- Preferir cambios de alcance mínimo por bugfix.

