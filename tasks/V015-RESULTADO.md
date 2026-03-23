## V015-RESULTADO (RootModel unificado + pantallas TUI del menú)

### 1) Arquitectura del RootModel

```mermaid
flowchart TD
    menu[screenMenu: MenuModel] -->|enter Nuevo proyecto| new[screenNew: App]
    menu -->|enter Mis templates| templates[screenTemplates: TemplatesModel]
    menu -->|enter Explorar GitHub| github[screenGitHub: GitHubModel]
    menu -->|enter Configuración| config[screenConfig: ConfigModel]

    new -->|App.Done() al presionar tecla en Done/Error| menu
    templates -->|esc o Done()| menu
    github -->|esc| menu
    config -->|esc| menu

    menu -->|q/ctrl+c| exit[Salir (tea.Quit)]
    new -->|ctrl+c| exit
```

Archivo clave: `internal/tui/root.go`.

### 2) Solución al flash (causa raíz y por qué desaparece)

**Causa raíz:** antes se ejecutaban dos programas Bubbletea separados:
1) `RunMenu()` (AltScreen) termina y el terminal vuelve al modo normal  
2) `RunApp()` vuelve a crear AltScreen  

Entre ambas ejecuciones se ve un frame intermedio (“flash”).

**Solución:** se reemplaza `RunMenu()+RunApp()` por una sola sesión:
- `tui.Run(templates, eng)` crea un **único** `tea.Program` con `RootModel`
- `RootModel` mantiene un `screen` y delega el `Update()/View()` al modelo activo
- los cambios de “pantalla” son transiciones internas, no salidas del programa

Por eso no se alterna AltScreen entre menú y `new`, y el frame intermedio desaparece.

### 3) Pantallas implementadas

- `screenMenu`:
  - `internal/tui/menu.go`: navegación `↑/↓` (y `j/k`), `enter` selecciona, `q/ctrl+c` sale.
  - Se ajustó `MenuModel` para que `enter` **no** llame `tea.Quit` cuando es usado por `RootModel` (evita fin del programa).

- `screenNew`:
  - `internal/tui/app.go`: flujo completo de `structify new` sin cambios en arquitectura del estado machine.

- `screenTemplates`:
  - `internal/tui/templates_screen.go`
  - Muestra plantillas agrupadas en “Local” y “Built-in”.
  - `↑/↓` navega, `enter` abre detalle inline (manifest), `n` transiciona a `new` preseleccionando template, `esc` vuelve.

- `screenGitHub`:
  - `internal/tui/github_screen.go`
  - Pantalla informativa con el comando `structify template add github.com/<user>/<repo>` y recursos.
  - `esc` vuelve, `q/ctrl+c` sale.

- `screenConfig`:
  - `internal/tui/config_screen.go`
  - Renderiza valores reales de `config.Load()` y conteo de templates locales.
  - `esc` vuelve, `q/ctrl+c` sale.

### 4) Done() en App — detección del fin del flujo new

Cambios en `internal/tui/app.go`:
- Se añadió campo `done bool`
- Se añadió método `Done() bool`
- En `stateDone/stateError`:
  - cualquier `tea.KeyMsg` (excepto `ctrl+c`) pone `a.done = true` y **no** hace `tea.Quit`

Luego `RootModel` hace:
- delega `Update()` al `App`
- si `r.app.Done()` => `screenMenu` y se recrea el menú sin salir de AltScreen

### 5) Cobertura (`go test ./... -cover`)

Ejecutado recientemente:

```text
ok  	github.com/jamt29/structify/cmd	0.073s	coverage: 69.4% of statements
ok  	github.com/jamt29/structify/internal/tui	0.037s	coverage: 34.1% of statements
```

El resto de paquetes se mantuvo en verde (verificación completa en sección final).

### 6) Estado final

- `go build ./...`: OK
- `go test ./...`: OK
- `go test ./... -cover`: OK

### 7) Lecciones capturadas

Se recomienda/está documentado en `tasks/lessons.md` con el update previo:
- se evitó salida del `tea.Program` entre pantallas para prevenir “flash”
- para señalización de fin de flujo se usa `done`/`Done()` en vez de `tea.Quit`

