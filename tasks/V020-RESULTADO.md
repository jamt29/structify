# V020-RESULTADO (Centrado global TUI + Built-ins compilables)

## 1) Causa raíz del centrado (qué estaba fallando)

En el TUI, el centrado no estaba centralizado en un único lugar:

- Cada sub-modelo hacía su propio “placement” (menú, `App`/inputs/confirm/progress/done, selector de templates, GitHub, Config). Eso hacía que las reglas de centrado fueran inconsistentes entre pantallas y estados.
- `RootModel` no propagaba correctamente el `tea.WindowSizeMsg` a los modelos que quedaban “guardados” mientras se cambiaba de `screen`. Como consecuencia, al volver a una pantalla, su `width/height` internos podían quedar desactualizados respecto al tamaño real del terminal, causando desalineaciones (“pegado” arriba/izquierda o centrados parciales).
- Bug adicional: el texto `"(presiona cualquier tecla para salir)"` aparecía duplicado en el flujo `stateDone` porque se imprimía **inline** dentro de `App.renderDone()` y además se mostraba también en la barra de ayuda inferior mediante `App.helpText()` + `styleHelpBar`.

## 2) Solución aplicada (qué cambió)

### Centrados

- `internal/tui/root.go`: `RootModel.View()` quedó como ÚNICO responsable del alineado/centrado.
  - `screenMenu` y estados `stateSelectTemplate/stateInputs/stateConfirm/stateDone/stateError` (en `screenNew`) usan centrado **H+V**.
  - `stateProgress` (en `screenNew`) usa centrado **solo horizontal (H-only)**.
  - `screenTemplates` usa centrado **solo horizontal**.
  - `screenGitHub` y `screenConfig` usan centrado **H+V**.
- `internal/tui/layout.go`: se añadió `centerContentHorizontal(width, content)` para el caso H-only.
- Sub-modelos: ahora devuelven contenido “crudo” vía `ViewContent()` (sin `lipgloss.Place` ni `centerContent` propios):
  - `internal/tui/menu.go`
  - `internal/tui/app.go`
  - `internal/tui/templates_screen.go`
  - `internal/tui/github_screen.go`
  - `internal/tui/config_screen.go`

### Propagación de dimensiones

- `internal/tui/root.go`: al recibir `tea.WindowSizeMsg`, ahora se actualizan todos los sub-modelos activos (acumulando `tea.Cmd` con `tea.Batch`), incluyendo `menu`, y los punteros no-nil: `app`, `templatesScreen`, `githubScreen`, `configScreen`.

### Bug: “cualquier tecla para salir” duplicado

- `internal/tui/app.go`: se eliminó `"(presiona cualquier tecla para salir)"` de `renderDone()`.
- El texto queda **solo** en la barra de ayuda a través de `helpText()` + `styleHelpBar`.

## 3) Descripción de pantallas después del fix

Nota: no se capturaron screenshots (no había interacción visual disponible desde este entorno), pero el comportamiento está verificado por:
1) `go test ./...` y asserts sobre duplicación del texto en `internal/tui/app_test.go`.
2) La lógica de centrado/estado aplicada de forma determinística en `RootModel.View()`.

### Pantalla 1 — Menú principal (`screenMenu`)
- ASCII art + tagline/version + lista de opciones se renderizan como contenido crudo y `RootModel` aplica centrado **H+V**.
- La tagline/version ya no se re-centran dentro de `WelcomeView` (se elimina el placement interno), evitando desalineación interna del bloque.
- La barra de ayuda inferior se muestra solo una vez.

### Pantalla 2 — Selector de templates (`screenTemplates`)
- El contenido se centra **solo horizontalmente** (H-only). No se agrega “padding vertical” por centrado global, evitando inconsistencias cuando la lista crece.

### Pantalla 3 — Inputs (`screenNew/stateInputs`)
- `RootModel` aplica centrado **H+V** sobre el contenido (header/bloque de inputs).

### Pantalla 4 — Confirmación (`screenNew/stateConfirm`)
- `RootModel` aplica centrado **H+V** sobre el resumen y la caja de variables.

### Pantalla 5 — Done/resultado (`screenNew/stateDone`)
- `RootModel` aplica centrado **H+V**.
- Se elimina la duplicación: `"(presiona cualquier tecla para salir)"` ya no aparece dentro del body; el texto aparece solo en la barra de ayuda inferior.

Además:
- `stateProgress`: centrado **H-only**, evitando salto/“jump” vertical a medida que crece el contenido.

## 4) Diagnóstico de templates (built-ins) — errores exactos

Ejecuté los built-ins en `/tmp` con el binario `./bin/structify` y luego compilé/checqué como en la solicitud. Resultado: **no hubo errores de compilación** en los comandos verificados (exit 0).

### clean-architecture-go

Comando `structify new`:
```
./bin/structify new --template clean-architecture-go --name testapp --var module_path=github.com/test/testapp --var transport=http --var orm=none --output /tmp/test-clean-go
```
Salida:
```
  → Creating project...
  ✓ go mod init
  ─ install gRPC deps (skipped)
  ─ install gorm deps (skipped)
  ─ install sqlx deps (skipped)
  ✓ go mod tidy
  ✓ Created 11 files
```
`go build ./...`:
- Archivo de log sin salida (`/tmp/structify_diag_clean_go_build.txt` quedó vacío) y exit 0.

### vertical-slice-go

Comando `structify new`:
```
./bin/structify new --template vertical-slice-go --name testapp --output /tmp/test-vslice-go
```
Salida:
```
  → Creating project...
  ✓ go mod init
  ✓ go mod tidy
  ✓ Created 8 files
```
`go build ./...`:
- Archivo de log sin salida (`/tmp/structify_diag_vslice_go_build.txt` quedó vacío) y exit 0.

### clean-architecture-ts

Comando `structify new`:
```
./bin/structify new --template clean-architecture-ts --name testapp --var runtime=express --var use_prisma=false --output /tmp/test-clean-ts
```
Salida:
```
  → Creating project...
  ✓ npm init
  ✓ install runtime
  ✓ install dev deps
  ─ install prisma (skipped)
  ✓ Created 11 files
```
`npm install && npx tsc --noEmit`:
Salida (no hubo errores de tsc; el log contiene solo el audit de npm):
```
up to date, audited 69 packages in 1s

22 packages are looking for funding
  run `npm fund` for details

found 0 vulnerabilities
```

### vertical-slice-ts

Comando `structify new`:
```
./bin/structify new --template vertical-slice-ts --name testapp --var runtime=express --output /tmp/test-vslice-ts
```
Salida:
```
  → Creating project...
  ✓ install runtime
  ✓ install dev deps
  ✓ Created 7 files
```
`npm install && npx tsc --noEmit`:
Salida (sin errores de tsc; solo audit npm):
```
up to date, audited 69 packages in 1s

22 packages are looking for funding
  run `npm fund` for details

found 0 vulnerabilities
```

### clean-architecture-rust

Comando `structify new`:
```
./bin/structify new --template clean-architecture-rust --name testapp --var transport=axum --output /tmp/test-clean-rust
```
Salida:
```
  → Creating project...
  ✓ cargo build
  ✓ Created 7 files
```
`cargo check`:
```
    Checking testapp v0.1.0 (/tmp/test-clean-rust)
    Finished `dev` profile [unoptimized + debuginfo] target(s) in 0.07s
```

## 5) Correcciones aplicadas por template (diff de `.tmpl`)

No se realizaron correcciones sobre `templates/<builtin>/...`:
- Los comandos verificados retornaron exit 0.
- No se requirió modificar ningún `.tmpl` para que compile/checqué.

## 6) Verificación de compilación — outputs por template (re-run)

Repetí la verificación por template y se mantuvo el resultado (exit 0). Resumen:
- `go build ./...` en `/tmp/test-clean-go` y `/tmp/test-vslice-go`: sin salida (logs vacíos).
- `npm install && npx tsc --noEmit` en `/tmp/test-clean-ts` y `/tmp/test-vslice-ts`: sin errores de tsc (solo audit npm).
- `cargo check` en `/tmp/test-clean-rust`: `Finished dev profile ...`.

## 7) Cobertura — `go test ./... -cover` (output exacto)

Salida:
```
ok  	github.com/jamt29/structify	0.020s	coverage: 0.0% of statements
ok  	github.com/jamt29/structify/cmd	0.059s	coverage: 69.4% of statements
ok  	github.com/jamt29/structify/cmd/structify	0.020s	coverage: 100.0% of statements
ok  	github.com/jamt29/structify/cmd/template	0.036s	coverage: 62.3% of statements
ok  	github.com/jamt29/structify/internal/config	(cached)	coverage: 81.8% of statements
ok  	github.com/jamt29/structify/internal/dsl	(cached)	coverage: 87.7% of statements
ok  	github.com/jamt29/structify/internal/engine	(cached)	coverage: 74.1% of statements
ok  	github.com/jamt29/structify/internal/template	(cached)	coverage: 73.7% of statements
ok  	github.com/jamt29/structify/internal/tui	0.017s	coverage: 33.7% of statements
ok  	github.com/jamt29/structify/templates	(cached)	coverage: 100.0% of statements
```

## 8) Lecciones capturadas

- El centrado en un TUI mult-pantalla debe ser un “contrato” único (en este caso `RootModel.View()`), no una responsabilidad distribuida en cada sub-modelo.
- Si varias pantallas comparten un mismo `tea.Program`, `WindowSizeMsg` debe propagarse a todos los modelos activos para evitar estados con `width/height` obsoletos.
- Evitar duplicaciones de UI: si una frase es “barra de ayuda”, no debe imprimirse también inline en el body (especialmente en `stateDone`/`stateError`).

