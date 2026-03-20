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
