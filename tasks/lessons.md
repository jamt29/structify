## Lecciones Aprendidas

### L002 — Precedencia del DSL (when:)
- **Contexto:** Las expresiones `when:` combinan `||`, `&&`, `!` y comparaciones.
- **Lección:** Implementar el parser como `OR -> AND -> NOT -> COMPARE -> PRIMARY` para respetar `|| < && < ! < comparación`, y asegurar que paréntesis sobrescriben precedencia.
- **Aplicar en:** `internal/dsl/parser.go`, `internal/dsl/parser_test.go`

### L003 — Filtros deben soportar separadores mixtos
- **Contexto:** Los nombres de proyecto y variables pueden venir como `MyProject`, `my_project`, `my-project`, o con espacios.
- **Lección:** Normalizar separadores (`-`, `_`, espacios) y luego partir CamelCase manteniendo acrónimos para aplicar `snake_case/pascal_case/camel_case/kebab_case` de forma consistente.
- **Aplicar en:** `internal/dsl/filters.go`, `internal/dsl/filters_test.go`

### L004 — Validar y reportar ciclos en inputs[].when
- **Contexto:** Inputs condicionales pueden depender entre sí y crear ciclos difíciles de diagnosticar.
- **Lección:** Extraer identificadores desde el AST de `when:` y detectar ciclos con DFS; reportar el error en `inputs[i].when` y no parar en el primer error (acumular todos).
- **Aplicar en:** `internal/dsl/validator.go`, `internal/dsl/validator_test.go`

### L005 — `**` en globs requiere matcher propio
- **Contexto:** Las reglas `files[].include/exclude` usan patrones tipo `internal/transport/http/**`.
- **Lección:** `path.Match` no soporta `**` (solo `*` por segmento); implementar matching por segmentos con soporte explícito para `**` (cero o más segmentos) y que la “última regla gana” según orden del YAML.
- **Aplicar en:** `internal/engine/file_processor.go`, `internal/engine/file_processor_test.go`

### L006 — `go:embed` no permite `..` en patrones
- **Contexto:** Los templates built-in viven en `templates/` en la raíz del repo, pero el loader está en `internal/template/`.
- **Lección:** Los patrones de `//go:embed` no pueden contener `..`; para embebir `templates/**` el archivo con el embed debe vivir en (o por encima de) `templates/`. Solución: exponer un `embed.FS` desde el paquete raíz y consumirlo desde el loader.
- **Aplicar en:** `builtin_templates.go`, `internal/template/loader.go`

### L007 — Errores de printf/vet y cobertura en toolchain nuevo
- **Contexto:** Con el toolchain actualizado, `go test` puede fallar por `fmt.Errorf(<string variable>)` (chequeo printf/vet) y `go test ./... -cover` puede fallar en paquetes “sin tests”.
- **Lección:** Si el error es “non-constant format string”, usar `errors.New(msg)` o `fmt.Errorf("%s", msg)`; y para que `-cover` no falle en paquetes sin tests, agregar tests mínimos (smoke) o al menos un `*_test.go` que compile.
- **Aplicar en:** `internal/template/*`, `internal/engine/rollback.go`, `builtin_templates_test.go`, `cmd/structify/main_test.go`

### L008 — Metadata de origen y operaciones GitHub deben ser explícitas
- **Contexto:** Para poder actualizar templates instalados desde GitHub (`template update`), es necesario conocer de forma fiable el origen (URL y ref) usado en `template add`.
- **Lección:** Guardar siempre metadata de origen en un archivo dedicado (`.structify-meta.yaml`) con campos claros (`source_url`, `source_ref`, `installed_at`) y usar ese archivo como única fuente de verdad para operaciones subsecuentes (`update`). Si falta metadata, fallar de forma explícita con un mensaje accionable.
- **Aplicar en:** `internal/template/store.go`, `internal/template/github.go`, `cmd/template/add.go`, `cmd/template/update.go`

### L009 — Checklists de CLI deben separar errores críticos de advertencias
- **Contexto:** `template publish` mezcla requisitos “duros” (manifiesto válido, archivos en `template/`) con recomendaciones (README, versión razonable).
- **Lección:** Distinguir entre ítems críticos (que afectan exit code) e ítems no críticos (solo advertencias), pero mostrar todos en un checklist unificado con marcas `[✓]/[✗]`. Esto mantiene la UX clara para humanos y scripts, sin romper pipelines por detalles no esenciales.
- **Aplicar en:** `cmd/template/publish.go`, futuros comandos de validación/checklist.

### L010 — Mockear clientes externos en tests de CLI
- **Contexto:** La integración con GitHub introduce dependencias de red (`go-git`, API HTTP) en los comandos `template add/update`, que no deben ejecutarse en tests unitarios.
- **Lección:** Definir interfaces pequeñas (por ejemplo `githubClient` y factorías como `newGitHubClientFn`) para inyectar implementaciones fake en tests. Así se prueban flujos completos de CLI sin red real, simulando errores y estados de metadata de forma determinista.
- **Aplicar en:** `internal/template/github.go`, `cmd/template/add.go`, `cmd/template/update.go`, `cmd/template/add_test.go`, `cmd/template/update_test.go`.

### L011 — Dotfiles en `go:embed`
- **Contexto:** Los built-in templates incluían `.gitignore` y `.gitkeep`, pero el patrón `//go:embed templates/**` omitía archivos cuyo último segmento de path empieza con `.`.
- **Lección:** Validar explícitamente la presencia de dotfiles en built-ins y embebirlos mediante patrones explícitos o inclusión controlada, evitando asumir que `**` los cubre.
- **Aplicar en:** `builtin_templates.go` y validación end-to-end con `new --dry-run`.

### L012 — Interpolación anidada en defaults de inputs
- **Contexto:** Algunos templates requieren defaults que usan interpolación basada en otros inputs, por ejemplo `module_path` con `{{ project_name | kebab_case }}`.
- **Lección:** Resolver interpolaciones anidadas dentro de valores string del contexto una vez que ya están definidas las variables base.
- **Aplicar en:** `cmd/new.go` (función `resolveContextInterpolations`).

### L013 — `.go` dentro de templates rompe `-cover`
- **Contexto:** `go test ./... -cover` falló con `go: no such tool "covdata"` porque archivos con extensión `.go` dentro del árbol de templates fueron interpretados como código Go compilable.
- **Lección:** Asegurar que los archivos “de plantilla” no tengan extensión `.go` (renombrar a `.go.tmpl`) para que el tooling no los trate como paquetes Go.
- **Aplicar en:** directorios built-in bajo `templates/*/template/**` antes de release.

### L014 — Construcción del binario: usar el package `main`
- **Contexto:** `go build -o ./bin/structify ./` produjo un `ar archive` (no un ejecutable) porque el package raíz no era `main`.
- **Lección:** Para producir el binario del CLI, compilar desde `./cmd/structify` (donde vive `package main`), no desde `./`.
- **Aplicar en:** desarrollo local y preparación de checks/dry-runs del release.

### L015 — Validar `inputs[].validate` también en modo flags
- **Contexto:** El modo TUI valida valores del usuario contra `inputs[].validate` (regex), pero el modo “flags-only” (`--name`, `--dry-run`) podía saltarse esa validación, dejando que valores inválidos pasaran sin error.
- **Lección:** En `cmd/new.go`, aplicar explícitamente `inputs[].validate` sobre el `dsl.Context` final antes de construir el request (para que TUI y flags tengan el mismo contrato).
- **Aplicar en:** `cmd/new.go` (función `validateManifestInputs`).

### L016 — `template create` debe generar scaffold mínimamente usable + UX consistente
- **Contexto:** `template create` generaba `scaffold.yaml` con `inputs: []`, por lo que `structify new` no podía pedir/obtener `project_name` en modo interactivo. Además, el wizard se hacía con `bufio.Reader` en lugar de reutilizar los componentes Bubbletea existentes.
- **Lección:** Generar siempre `inputs` mínimos (como `project_name` con `validate`) y una carpeta base `template/.gitkeep`, y reutilizar `tui.RunInputs` para que el wizard se sienta igual que `new` (y respete ESC/cancel).
- **Aplicar en:** `cmd/template/create.go` (`writeScaffoldYAML` + wizard).

### L017 — Nunca serializar valores nil a textinput
- **Contexto:** En el TUI unificado apareció `"<nil>"` como valor visible al construir inputs desde `fmt.Sprint(ctx[id])` cuando no existía respuesta previa.
- **Lección:** Para `textinput`, usar helper que convierta `nil` a string vacío y setear placeholder explícito (`default` o `""`), evitando exponer valores técnicos al usuario.
- **Aplicar en:** `internal/tui/app.go` (`prepareInputs`, helpers de respuestas).

### L018 — Menús TUI en root necesitan salida explícita para tests/no-TTY
- **Contexto:** Al mover `structify` (sin subcomando) a TUI, en entornos sin `/dev/tty` el menú falla al abrir terminal interactiva y puede romper flujos de prueba.
- **Lección:** Encapsular `RunMenu` detrás de función inyectable y manejar un error sentinel de salida (`ErrMenuExit`) para tener salida limpia y pruebas deterministas.
- **Aplicar en:** `cmd/root.go`, `internal/tui/menu.go`, tests de comando raíz.

### L019 — Evitar `tea.Quit` entre pantallas para prevenir flash AltScreen
- **Contexto:** Si `RunMenu()` y `RunApp()` ejecutan `tea.WithAltScreen()` como programas separados, el terminal alterna modo y se ve un frame intermedio.
- **Lección:** Mantener un único `tea.Program` (RootModel) y transicionar por `screen` interno. Para señalar fin de sub-flujos, usar `done`/`Done()` y no salir con `tea.Quit`.
- **Aplicar en:** `internal/tui/root.go`, `internal/tui/app.go`, `internal/tui/menu.go`.

### L020 — Features de DSL deben venir con tests dirigidos a cobertura
- **Contexto:** Al extender el DSL con `contains()` y nuevos tipos, la cobertura de `internal/dsl` cayó por debajo del umbral mínimo.
- **Lección:** Cada rama nueva del evaluator/parser/validator debe tener tests explícitos (incluyendo errores y ramas internas) antes de cerrar la feature.
- **Aplicar en:** `internal/dsl/evaluator_test.go`, `internal/dsl/parser_test.go`, `internal/dsl/validator_test.go`, `internal/dsl/*_internal_test.go`.

### L021 — Analyzer de import debe priorizar señales estructurales
- **Contexto:** Buscar nombre de proyecto sin restricciones genera falsos positivos cuando el nombre es muy genérico.
- **Lección:** Detectar variables sugeridas solo en contextos estructurales (module/import/name de manifests) y no en texto libre/comentarios.
- **Aplicar en:** `internal/template/analyzer.go`, futuros detectores automáticos de variables.

### L022 — Migraciones de UI deben conservar compatibilidad de tests internos
- **Contexto:** Al migrar `stateInputs` de componentes manuales a `huh.Form`, tests existentes de `internal/tui/app_test.go` seguían inyectando valores sobre `app.inputs`.
- **Lección:** Mantener una capa de sincronización de compatibilidad durante la migración (`syncLegacyInputsToHuh`) evita regressions y permite evolución gradual sin reescribir toda la suite al mismo tiempo.
- **Aplicar en:** `internal/tui/app.go` durante futuras migraciones de componentes visuales.

### L023 — El size inicial debe propagarse a pantallas creadas después
- **Contexto:** Aunque Bubble Tea envía `WindowSizeMsg` al inicio, si `App` se crea después (desde menú) puede no recibir ese mensaje inicial y renderizar con defaults desalineados.
- **Lección:** Mantener fallback seguro (`80x24`) y, al transicionar de pantalla, inyectar explícitamente el tamaño conocido (`tea.WindowSizeMsg`) al sub-modelo recién creado.
- **Aplicar en:** `internal/tui/root.go` al crear `App` desde `screenMenu`/`screenTemplates`.

### L024 — Huh `Value()` requiere punteros con lifetime correcto
- **Contexto:** `Value(&localVar)` en Huh escribe en esa variable; usar `&val` dentro de un `for range` o `&s` en el cuerpo del loop hace que varios campos compartan la misma dirección o apunten a memoria inválida. El formulario parece aceptar input pero al confirmar los valores llegan vacíos o incorrectos.
- **Lección:** Los punteros pasados a `Value()` deben vivir en el heap (`new(T)`) o en campos del modelo `App` que persistan mientras exista el form, nunca a variables locales efímeras del builder.
- **Aplicar en:** `internal/tui/huh_inputs.go` siempre que se use `huh.NewInput` / `NewSelect` / `NewConfirm` / `NewMultiSelect` con `.Value(...)`.

### L025 — Huh embebido debe recibir `WindowSizeMsg` del modelo raíz
- **Contexto:** `App.Update` manejaba `tea.WindowSizeMsg` al inicio y retornaba sin pasarlo a `huh.Form`. El `Init()` de Huh encola `tea.WindowSize()` para dimensionar grupos; ese mensaje nunca llegaba al formulario si el padre lo absorbía.
- **Lección:** Cuando Huh vive dentro de otro modelo Bubble Tea, reenviar `WindowSizeMsg` a `form.Update` (y alinear `WithWidth` con el panel real si hay split/lipgloss) evita campos con ancho 0 y el síntoma de “no puedo escribir / no hay foco”.
- **Aplicar en:** `internal/tui/app.go` (`Update` + `applyHuhFormWidth` / `inputsFormWidth`).

### L027 — Transiciones en Bubble Tea: no capturar el modelo en closures retardados
- **Contexto:** Al animar cambios de pantalla en `RootModel` con `tea.Tick`, un `pendingSwitch func() tea.Cmd` que cerraba sobre el `RootModel` recibido en `Update` aplicaba mutaciones sobre una copia obsoleta del estado (el runtime ya había sustituido el modelo).
- **Lección:** Describir el cambio pendiente con datos (`enum` + campos auxiliares como `pendingSelTemplate`) y ejecutar `applyPendingTransition()` en el tick usando el receptor **actual** devuelto por el framework.
- **Aplicar en:** `internal/tui/root.go`, futuros modelos con animaciones o pasos diferidos.

### L028 — Log estructurado en subcomandos sin rom tests que capturan `SetOut(buf)`
- **Contexto:** `UseStructuredLogOut` basado solo en `os.Stdout` o en “no TTY” rompe tests que redirigen `cmd.SetOut(&bytes.Buffer{})` o fuerzan salida no terminal.
- **Lección:** Considerar estructurado solo si `OutOrStdout()` es `*os.File` y `!term.IsTerminal(fd)`; los `io.Writer` que no son `*os.File` mantienen `fmt` hacia el writer del comando.
- **Aplicar en:** `internal/config/logger.go`, `cmd/template/*`, otros CLIs con cobertura sobre buffers.

### L029 — Informe principal del CLI vs charmbracelet/log
- **Contexto:** Tras enrutar el dry-run de `new` por `charmbracelet/log` a stderr, algunas terminales o integraciones mostraban la sesión “en blanco” porque el usuario miraba solo stdout.
- **Lección:** El **cuerpo redirigible y visible por defecto** del comando (listas dry-run, tablas pensadas para pipe) debe ir a **stdout** con `fmt`; reservar log con niveles a stderr para progreso/errores secundarios o mantener coherencia con herramientas Unix.
- **Aplicar en:** `cmd/new.go` (`runDryRun`), futuros comandos con salida máquina+humano.

### L030 — `RunApp` no pasa por `RootModel`: el centrado debe vivir también en `App.View`
- **Contexto:** `structify new` en TTY usa `tui.RunApp`, que dibuja `*App` sin envolver `RootModel.View`; el comentario decía que RootModel centraba, pero ese código no se ejecutaba en esta ruta, dejando `stateProgress` arriba-izquierda.
- **Lección:** Si hay dos entradas TUI (`Run` vs `RunApp`), duplicar en `App.View()` las reglas de layout mínimas (p. ej. `centerContent` / `centerContentHorizontal` por estado) o factorizar un helper compartido.
- **Aplicar en:** `internal/tui/app.go`, `internal/tui/root.go`.

### L031 — `stateDone` necesita política de salida explícita según entrypoint
- **Contexto:** `stateDone/stateError` marcaba `done=true` pero no siempre emitía `tea.Quit`; en `RunApp` (App top-level) eso dejaba la UI abierta esperando Ctrl+C, mientras en `RootModel` sí conviene no salir y volver al menú.
- **Lección:** Definir un flag de comportamiento de salida (`quitOnDoneKey`) para separar el path top-level (`RunApp` => `tea.Quit`) del path embebido (`RootModel` => transición interna).
- **Aplicar en:** `internal/tui/app.go`, futuros modelos reutilizados en más de un entrypoint.

### L026 — No sincronizar widgets legacy → Huh en cada tick si el legacy está desactualizado
- **Contexto:** Tras migrar a Huh se mantuvieron `app.inputs` con `textinput` para tests. `syncLegacyInputsToHuh()` copiaba esos valores a `huhString` antes de cada `huhForm.Update`. El teclado alimenta solo Huh, no el `textinput`, así que el legacy seguía vacío y **pisaba** el texto tecleado; si además se reconstruía el form al detectar cambio, el foco parpadeaba y parecía imposible escribir.
- **Lección:** Con `huhForm != nil`, no volcar legacy vacío sobre maps que ya reflejan Huh; no reconstruir el form en bucle por ese “cambio”. En Enter, volcar primero `syncFromHuhForm()` y luego fusionar solo lo que aporte el legacy (p. ej. tests con `ti.SetValue`).
- **Aplicar en:** `internal/tui/app.go` (`updateInputs`, `syncLegacyInputsToHuh`).

### L033 — TUI: clave del store local vs nombre del manifiesto
- **Contexto:** Los templates locales viven en `~/.structify/templates/<carpeta>/`; `template.Get`/`Remove` usan ese segmento, mientras que la UI muestra `manifest.name`, que puede no coincidir.
- **Lección:** Para operaciones de almacén y re-selección tras reload, usar `filepath.Base(template.Path)` como clave estable además de emparejar por `manifest.name` cuando aplique.
- **Aplicar en:** `internal/tui/templates_screen.go`, `internal/tui/yaml_editor.go`, cualquier llamada a `template.Remove`/`Get` desde el TUI.

### L034 — Centrado TUI: bloque interno vs `lipgloss.Place` y padding vertical
- **Contexto:** El menú mostraba ASCII art y opciones alineadas a la izquierda mientras la ayuda parecía centrada; en `stateSelectTemplate` el contenido quedaba arriba pese a `centerBoth`.
- **Lección:** (1) Unir art/tagline/menú con `lipgloss.JoinVertical(lipgloss.Center, …)` para que el bloque sea una unidad antes del `MaxWidth`/`RootModel`. (2) `lipgloss.Place` vertical a veces no coincide con la expectativa; un padding superior explícito `(height - lipgloss.Height(content)) / 2` como `\n` + `PlaceHorizontal` replica el centrado V de forma predecible.
- **Aplicar en:** `internal/tui/menu.go`, `internal/tui/welcome.go`, `internal/tui/layout.go`.

### L035 — Preview de árbol: nombres de archivo fuente `.tmpl`
- **Contexto:** El preview listaba `routes.ts.tmpl` porque los segmentos del path venían del template en disco.
- **Lección:** En el árbol de preview, mostrar el nombre generado (`TrimSuffix(..., ".tmpl")` en nodos hoja), no el nombre del archivo plantilla.
- **Aplicar en:** `internal/engine/preview.go` (`findOrCreateNode` / `previewDisplayName`).

### L036 — `lipgloss.MaxWidth` sin `Align(Left)` en bloques TUI
- **Contexto:** En `stateDone` y `stateProgress` algunas líneas cortas (“Steps”, ayuda) parecían centradas dentro del bloque mientras el resto quedaba a la izquierda; el bloque seguía sin sentirse centrado en horizontal respecto a la terminal.
- **Lección:** Tras `MaxWidth`, fijar `Align(lipgloss.Left)` en el estilo que envuelve todo el `ViewContent()` del `App`, para que `Place`/`PlaceHorizontal` centren **un bloque alineado a la izquierda**, no líneas sueltas centradas dentro del ancho máximo. Evitar doble `MaxWidth` en `renderDone` y en el wrapper.
- **Aplicar en:** `internal/tui/app.go` (`ViewContent`, `renderDone`).

### L037 — Spinner en `stateProgress` y cola de mensajes
- **Contexto:** Tras cada `msgStepStart` / `msgFilesDone` solo se encolaba `waitProgressMsg` sin `spinner.Tick`, así el spinner podía dejar de animarse entre mensajes del canal.
- **Lección:** Hacer `tea.Batch(a.spin.Tick, waitProgressMsg(ch))` (y lo mismo al recibir `msgProgressReady`) para mantener ticks mientras se espera el siguiente paso.
- **Aplicar en:** `internal/tui/app.go` (`Update` / `updateProgress`).

### L038 — Pantalla Mis templates: ancho del nombre por columna (Local vs Built-in)
- **Contexto:** `rowLineTwoCol` aplica `lipgloss.MaxWidth(nameW)` al nombre; `nameW` se calculaba como `leftW - metaColWidth() - 2` para **ambas** columnas. La columna Built-in es más ancha (`rightW`), pero el nombre seguía limitado por el ancho local (~9 celdas), y `ansi.Truncate` cortaba strings largos (`clean-architecture-go`, etc.) de forma que parecían nombres corruptos.
- **Lección:** Calcular `nameWLocal` desde `leftW` y `nameWBuiltin` desde `rightW` por separado; no reutilizar el ancho de nombre pensado solo para la columna izquierda.
- **Aplicar en:** `internal/tui/templates_screen.go` (`viewContentInner`, `rowLineTwoCol`).

### L032 — Built-ins: verificar el pipeline real de generación
- **Contexto:** Los templates podían compilar pero el código no usaba capas generadas (p. ej. HTTP en `main` vs `internal/transport/http`) o el scaffold ejecutaba `npm init -y` tras escribir `package.json`.
- **Lección:** Tras cada cambio en `.tmpl`/`scaffold.yaml`, regenerar en `/tmp` y ejecutar el toolchain del lenguaje (`go`/`tsc`/`cargo`); para Node, revisar el orden de steps respecto a `package.json` embebido.
- **Aplicar en:** `templates/*/`, `templates/*/scaffold.yaml`

### L039 — Verificación de CLI en `./bin`: reconstruir binario antes de pruebas manuales
- **Contexto:** `go test`/`go build ./...` validan el código fuente, pero las pruebas manuales de comandos (`./bin/structify ...`) pueden ejecutar un binario desactualizado.
- **Lección:** Cuando la verificación incluye `./bin/structify`, reconstruir explícitamente con `go build -o ./bin/structify ./cmd/structify` justo antes de correr comandos manuales.
- **Aplicar en:** cierres de iteración con validación manual en `tasks/Vxxx-RESULTADO.md`.

### L040 — Huh: no sobrescribir defaults con respuestas vacías
- **Contexto:** En `new` vía TUI, `buildContextFromHuh` enviaba `""` para inputs string/enum no tocados; eso anulaba defaults del DSL (incluyendo interpolados como `module_path`) y podía romper `steps` con `exit status 1`.
- **Lección:** Al construir contexto desde Huh, solo incluir respuestas string/enum/path cuando tengan contenido; si están vacías, omitirlas para que `BuildContext` aplique defaults y validaciones correctamente.
- **Aplicar en:** `internal/tui/app.go` (`buildContextFromHuh`) y tests de regresión en `internal/tui/app_test.go`.
