## V014-RESULTADO (welcome + menú principal TUI)

### 1) Estructura de archivos nuevos

- `internal/tui/menu.go`
  - `MenuModel` con navegación (`↑/↓`, `j/k`, `enter`, `q/ctrl+c`)
  - `menuItem` y `menuAction` (más aliases exportados `ActionNew`, `ActionTemplates`, `ActionGitHub`, `ActionConfig`)
  - `RunMenu()` para lanzar menú en `tea.WithAltScreen()`
  - `ErrMenuExit` para salida limpia al cancelar
- `internal/tui/welcome.go`
  - `WelcomeView(width int)` con ASCII art exacto + tagline + versión
- `internal/tui/layout.go`
  - helper `centerContent(width, height, content)` para centrar pantalla completa
- `internal/tui/styles.go` (extendido)
  - estilos de welcome y menú (activo/inactivo, flecha, label, tagline)

Archivos integrados:
- `cmd/root.go`
  - root ahora ejecuta `runInteractive()` cuando no hay subcomando
  - integración `RunMenu -> switch(action) -> RunApp/stubs`
  - stubs implementados para Templates/GitHub/Config
  - hooks inyectables para pruebas del flujo root

---

### 2) Flujo de navegación

1. `structify` (sin subcomando) -> `runInteractive()` -> `tui.RunMenu()`
2. En menú principal:
   - **Nuevo proyecto** -> `resolveAllTemplates()` -> `tui.RunApp(templates, engine.New())`
   - **Mis templates** -> stub textual (lista simple de templates disponibles)
   - **Explorar GitHub** -> stub textual (`structify template add ...`)
   - **Configuración** -> stub textual (ruta de config cargada)
3. `structify new` sigue entrando al flujo existente de `cmd/new.go`
4. `structify new --template ... --name ... --dry-run` se mantiene sin TUI (modo flags)

---

### 3) Output visual (descripción)

#### Pantalla welcome/menu (`structify`)
- ASCII art centrado horizontalmente, color `#C678DD`
- tagline centrado en color muted
- versión centrada en gris oscuro
- cuatro items de menú centrados
- item activo con:
  - fondo `#2d2f3f`
  - borde izquierdo en color primary
  - texto principal destacado
- items inactivos en `colorMuted`
- barra de ayuda inferior: `↑↓ navegar  enter seleccionar  q salir`

#### Pantalla `structify new`
- No se alteró arquitectura interna de `internal/tui/app.go`
- entra directo al selector/state machine existente

#### Limitación del entorno de ejecución usado
- En este entorno no hay TTY real (`/dev/tty`), por eso:
  - `go run .` falla con: `could not open a new TTY`
  - visual completo del menú solo es verificable en terminal interactiva real

---

### 4) Decisiones de integración (RunMenu + RunApp)

- Se mantuvo separación clara:
  - `RunMenu()` decide intención del usuario
  - `RunApp()` mantiene intacto el flujo unificado de `new`
- No se duplicó lógica de `new` dentro del menú; `ActionNew` llama directamente `RunApp`
- Se añadió error sentinel (`ErrMenuExit`) para que cancelar menú no genere error visible
- Se añadieron function hooks en `cmd/root.go` para testear rutas de navegación sin TTY real

---

### 5) Cobertura (`go test ./... -cover`)

```text
ok  	github.com/jamt29/structify	(cached)	coverage: 0.0% of statements
ok  	github.com/jamt29/structify/cmd	(cached)	coverage: 70.4% of statements
ok  	github.com/jamt29/structify/cmd/structify	(cached)	coverage: 100.0% of statements
ok  	github.com/jamt29/structify/cmd/template	(cached)	coverage: 62.3% of statements
ok  	github.com/jamt29/structify/internal/config	(cached)	coverage: 81.8% of statements
ok  	github.com/jamt29/structify/internal/dsl	(cached)	coverage: 87.7% of statements
ok  	github.com/jamt29/structify/internal/engine	(cached)	coverage: 74.1% of statements
ok  	github.com/jamt29/structify/internal/template	(cached)	coverage: 73.7% of statements
ok  	github.com/jamt29/structify/internal/tui	(cached)	coverage: 44.6% of statements
ok  	github.com/jamt29/structify/templates	(cached)	coverage: 100.0% of statements
```

Comparado con V013:
- `cmd`: sube de 69.3% a 70.4%
- se mantiene el criterio de no bajar cobertura global de referencia del paquete `cmd`

---

### 6) Estado final

- `go build ./...` -> OK
- `go test ./...` -> OK
- `go test ./... -cover` -> OK
- `go run . new --template clean-architecture-go --name my-api --dry-run` -> OK
- `go run .` -> requiere TTY real (en este entorno falla por ausencia de `/dev/tty`)

---

### 7) Lecciones capturadas

Se agregó en `tasks/lessons.md`:
- **L018 — Menús TUI en root necesitan salida explícita para tests/no-TTY**
