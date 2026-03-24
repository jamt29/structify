# V030b-RESULTADO — v0.3.0b cuatro fixes ordenados

## Fix 1 — Raíz del árbol con placeholder

**Antes:** En `stateInputs`, con `project_name` vacío, `buildPartialRequest` usaba `OutputDir: a.outputDir()`. Eso ignoraba el texto vivo de Huh y, al unir `cwd` con nombre vacío, `filepath.Base` del directorio de salida coincidía con el directorio actual (p. ej. primera línea del árbol `structify/`).

**Después:** Tras construir el contexto parcial, `previewEffectiveProjectName` rellena `project_name` con el default del manifiesto (`ApplyDefault`) o con `"<project>"`, actualiza `ctx["project_name"]` y el `OutputDir` del request parcial es `filepath.Join(cwd, nombreEfectivo)`. La raíz del árbol muestra `<project>/` o el default del template antes de escribir el nombre.

**Archivos:** `internal/tui/app.go` (`buildPartialRequest`, nueva `previewEffectiveProjectName`).

---

## Fix 2 — Barra de ayuda nativa de Huh

**Confirmación:** En `github.com/charmbracelet/huh@v1.0.0` se usa `(*Form).WithShowHelp(false)` en la cadena del formulario.

**Archivo:** `internal/tui/huh_inputs.go` — `huh.NewForm(...).WithTheme(...).WithShowHelp(false)`.

---

## Fix 3 — Charmbracelet/log en modo no interactivo

**Archivos modificados (log / detección de writer)**

- `go.mod` / `go.sum` — dependencia directa `github.com/charmbracelet/log`
- `internal/config/logger.go` — `NewLogger(verbose)`, `UseStructuredLogOut(w io.Writer)`
- `cmd/new.go` — rama sin TTY: `runDryRun`, `runNonInteractive`, `printStepObserver`, resumen final
- `cmd/new_coverage_smoke_test.go` — captura de **stderr** (el logger escribe ahí)
- `cmd/template/cli_log.go` — `tmplStructuredLog`, `tmplVerbose`
- `cmd/template/add.go`, `update.go`, `remove.go`, `list.go`, `validate.go`, `create.go`, `publish.go`, `import.go` — mensajes de progreso/resumen cuando `UseStructuredLogOut(cmd.OutOrStdout())` (stdout es `*os.File` y no TTY); tests con `bytes.Buffer` siguen la ruta `fmt` sin cambios.

**Output real** (`go run ./cmd/structify new --template clean-architecture-go --name test-log --dry-run` — stderr):

```text
 Dry run — no files will be written.
 summary template=clean-architecture-go output=./test-log variables="project_name=test-log, module_path=github.com/user/test-log, orm=none, transport=http"
 files that would be created
 .gitignore
 Makefile
 ...
 steps that would run
 ✓ go mod init github.com/user/test-log
 ─ go get google.golang.org/grpc  (skipped: transport == "grpc")
 ...
 No files were written.
```

**Con `--verbose`:** timestamps en cada línea y nivel Debug disponible para observadores de steps en ejecución real no dry-run.

---

## Fix 4 — Transiciones entre pantallas del RootModel

**Implementación:** Máquina de estados `transitionFadeOut` → aplicar cambio de pantalla → `transitionFadeIn`, con `tea.Tick(16ms)` y paso de opacidad `0.22` por tick (~5 ticks por fase ≈ 80 ms por mitad, ~160 ms ciclo completo en el peor caso).

**Detalle técnico:** No se usan closures que capturen `RootModel` al programar el cambio (evita estado obsoleto en Bubble Tea). Se usa `rootPendingKind` + `pendingSelTemplate` y `applyPendingTransition()` sobre el modelo actual.

**Vista:** `applyRootTransitionAlpha` — sin opacidad real en terminal: vacío casi transparente, medio tono gris `#5C6370`, o sin cambio cerca de 1.0.

**Limitaciones:** La transición no aplica a cambios internos de `App` (`stateInputs` → `confirm`, etc.). Durante fade se ignoran mensajes de teclado excepto `WindowSizeMsg` (ya manejado al inicio). Si `newApp` falla, se aborta el fade-in y se devuelve `tea.Quit` sin reanudar el tick de fade-in.

---

## Cobertura

Comando: `go test ./... -cover`

```text
ok  github.com/jamt29/structify/cmd                    coverage: 63.3%
ok  github.com/jamt29/structify/cmd/template           coverage: 45.3%
ok  github.com/jamt29/structify/internal/config       coverage: 47.4%
ok  github.com/jamt29/structify/internal/dsl           coverage: 87.2%
ok  github.com/jamt29/structify/internal/engine        coverage: 62.9%
ok  github.com/jamt29/structify/internal/template      coverage: 73.9%
ok  github.com/jamt29/structify/internal/tui           coverage: 33.7%
```

## Estado final

- `go build ./...` — OK  
- `go test ./...` — OK  

## Lecciones capturadas

- Ver `tasks/lessons.md` — L027 (transiciones RootModel / sin closures sobre modelo), L028 (`UseStructuredLogOut` + tests con buffer).
