# V0.5.0 — Resultado (TUI: centrado, bugs templates, polish, GitHub, config)

## 1. Mapa de centrado antes / después

| Pantalla / estado | Antes | Después |
| --- | --- | --- |
| `screenMenu` | `centerContent` (H+V) | Igual; guard `width/height == 0` → `""`; contenido con `MaxWidthMenu` |
| `screenNew` / `stateProgress` | H solo | `AppCenteringMode` → `CenterHOnly`; `ApplyScreenCentering` |
| `screenNew` / resto de estados | H+V vía switch anidado | Misma política centralizada en `ApplyScreenCentering` + `MaxWidth` por estado |
| `screenTemplates` / `modeList`, `modeEdit` | Siempre H solo | `centeringMode()` distingue modos: list + edit → `CenterHOnly`; create, detail, delete → `CenterBoth` |
| `screenTemplates` / otros modos | (antes igual H solo) | `CenterBoth` para formularios/detalle |
| `screenGitHub`, `screenConfig` | H+V | Igual; contenido acotado (`MaxWidthGitHub` / `MaxWidthConfig`) |

**Nota (L030):** `structify new` sigue usando `App.View()` con **`ApplyScreenCentering`** y la misma tabla que `screenNew` en `RootModel`, sin duplicar `lipgloss.Place` a mano.

Archivos clave: [`internal/tui/layout.go`](internal/tui/layout.go) (`ApplyScreenCentering`, `AppCenteringMode`), [`internal/tui/root.go`](internal/tui/root.go) (`viewCurrentScreen`, `centeringMode`), [`internal/tui/app.go`](internal/tui/app.go) (`View` → `ApplyScreenCentering`).

## 2. Bugs corregidos (Grupo 2)

| Bug | Causa raíz | Fix |
| --- | --- | --- |
| `template "X" not found` al borrar / tras reload | `template.Remove` usa el **nombre de carpeta** bajo `~/.structify/templates/<dir>/`, no necesariamente `manifest.name`. | `templateStoreDirName(t)` = `filepath.Base(t.Path)` para `Remove`; `SelectByName` acepta coincidencia por manifiesto **o** por nombre de carpeta; tras guardar YAML, `reloadName` = carpeta en disco. |
| Nombre local en dos líneas | `lipgloss.Width` + texto largo forzaba wrap en la columna fija. | `localColWidth()` dinámico; filas con `MaxWidth`/`Inline` en nombre y metadatos en columna aparte. |
| Metadatos con `...` descuidado | `truncateStr` agresivo en ambos campos. | Truncado solo del nombre con lipgloss; arquitectura/idioma en línea de metadatos sin `...` artificial. |
| Barra de ayuda “flotando” | Ayuda pegada al bloque superior sin ocupar alto de terminal. | `padHelpBarToBottom` con `lipgloss.Height` y relleno de `\n` hasta `m.height`. |
| Header `stateInputs` desalineado | Bloque muy ancho vs terminal. | `viewMaxWidthForState` + `layoutWidthForInputs()` = `min(width, MaxWidthInputs)` para split y Huh. |

## 3. Polish aplicado (Grupo 3)

- **Selector (`stateSelectTemplate`):** título `Selecciona un template (N)`; delegate con borde izquierdo grueso (`ThickBorder`); badges de lenguaje con colores (go / typescript / rust / python); descripción ajustada; `spinner.Dot` + estilo azul para progreso.
- **`stateProgress`:** spinner en `colorActive`; steps skipped con `─` en gris (`colorMuted`).
- **`stateDone`:** bloque único alineado a la izquierda con padding y `MaxWidthDone`; “Próximos pasos” como líneas consistentes sin mezcla de centrados.
- **`screenTemplates`:** `styleMenuItemActive` con borde grueso en [`styles.go`](internal/tui/styles.go).
- **`screenGitHub`:** campo URL (`textinput`), Enter → `ParseGitHubURL` + `InstallFromGitHub` en goroutine con polling + spinner durante la instalación.

## 4. `screenGitHub` — input de URL

- Modelo en [`internal/tui/github_screen.go`](internal/tui/github_screen.go): `textinput`, `spinner.Dot`, canal `chan error` + mensaje `githubInstallPollMsg` para no bloquear el bucle de Bubble Tea.
- Instalación vía [`internal/template/install_github.go`](internal/template/install_github.go) (`InstallFromGitHub`), reutilizada por [`cmd/template/add.go`](cmd/template/add.go).

## 5. `screenConfig` — valores reales

- [`internal/tui/config_screen.go`](internal/tui/config_screen.go): `config.Load()`, rutas con `~` cuando aplica, versión desde [`internal/buildinfo`](internal/buildinfo/buildinfo.go), conteo de templates `Source == "local"`, `logLevel`, `nonInteractive`, dos cajas con borde redondeado.

## 6. Cobertura — `go test ./... -cover`

Salida representativa (todos los paquetes OK):

```
ok  	github.com/jamt29/structify/cmd		coverage: 63.5% of statements
ok  	github.com/jamt29/structify/cmd/template	coverage: 68.3% of statements
ok  	github.com/jamt29/structify/internal/config	coverage: 89.5% of statements
ok  	github.com/jamt29/structify/internal/dsl	coverage: 87.3% of statements
ok  	github.com/jamt29/structify/internal/engine	coverage: 62.9% of statements
ok  	github.com/jamt29/structify/internal/template	coverage: 68.2% of statements
ok  	github.com/jamt29/structify/internal/tui	coverage: 28.2% of statements
```

## 7. Estado final

- `go build ./...` — OK  
- `go test ./...` — OK  
- `go vet ./...` — OK  

## 8. Lecciones capturadas

- **Store vs manifiesto:** operaciones sobre `~/.structify/templates` deben usar el **nombre de directorio** (`filepath.Base(Path)`), no solo `manifest.name`, o `template.Remove` / re-selección fallan cuando difieren.
- **Un solo layout:** `ApplyScreenCentering` + `AppCenteringMode` evitan divergencias entre `RootModel` y `RunApp` (L030).
- **Versión en TUI:** `internal/buildinfo` evita import cycle con `cmd` y unifica welcome, `structify version` y pantalla de configuración.

## Post-fix UX (capturas)

- **Menú:** `WelcomeView` + ítems + ayuda en `lipgloss.JoinVertical(lipgloss.Center, …)`; `centerContent` usa padding vertical explícito + `PlaceHorizontal` (ver L034).
- **Selector / estados `centerBoth`:** mismo `centerContent` corregido.
- **Árbol:** `previewDisplayName` quita `.tmpl` en hojas (`internal/engine/preview.go`).

Verificación: `go build ./...` OK; `go test ./... -cover` OK.
