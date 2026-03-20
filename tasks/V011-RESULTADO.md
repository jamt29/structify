## V011-RESULTADO (iteración v0.1.1)

### 1) Causa raíz encontrada de cada problema

#### Problema 1 — TUI incompleta (selector sí, pero el resto cae a prompts planos)
- **Archivo:** `cmd/template/create.go`
- **Causa raíz:** el wizard de `structify template create` estaba implementado con `bufio.Reader`/`fmt.Fprint` (prompts planos), en vez de reutilizar los componentes Bubbletea ya existentes para formularios (`internal/tui/inputs.go`), dejando al comando con UX “terminal plana” aunque `structify new` ya usaba Bubbletea en sus pasos TUI.

#### Problema 2 — Validación de `project_name` demasiado restrictiva
- **Archivo(s):** `templates/*/scaffold.yaml`
- **Causa raíz:** el campo `inputs[].validate` de `project_name` estaba con regex vieja `^[a-z][a-z0-9-]+$`, lo que rechazaba casos que el usuario debía poder usar (p.ej. `_` y letras mayúsculas).
- **Archivo:** `cmd/new.go`
- **Causa raíz adicional:** aunque la validación de `validate` existía en `internal/tui/logic.go`, el modo “flags-only”/`--dry-run` no aplicaba la validación al `dsl.Context` final. Eso hacía que la UX fuese inconsistente entre TUI y flags-only.

#### Problema 3 — `template create` no crea plantillas desde cero (scaffold inválido para `structify new`)
- **Archivo:** `cmd/template/create.go`
- **Causa raíz:** `writeScaffoldYAML` generaba un `scaffold.yaml` con `inputs: []` (sin `project_name`), por lo que `structify new` no podía pedir/obtener `project_name` en modo interactivo y fallaba con `project_name is required`.

---

### 2) Cambios realizados por problema (archivos modificados)

#### Problema 1
- `cmd/template/create.go`
  - Reemplazado wizard plano (bufio/fmt) por wizard basado en `tui.RunInputs` (Bubbletea) cuando hay TTY.
  - Se mantuvo fallback no-TTY con `bufio.Reader` para permitir automatización por stdin (solo para entornos sin TTY).

#### Problema 2
- `cmd/new.go`
  - Agregada validación explícita de `inputs[].validate` sobre el `dsl.Context` final (`validateManifestInputs`), aplicada también en modo flags-only.
- `templates/*/scaffold.yaml`
  - Actualizada regex de `project_name` a: `^[a-zA-Z][a-zA-Z0-9_-]*$`
  - `templates/minimal-ts/scaffold.yaml`: añadido `inputs.project_name` con `validate` (para consistencia con el contrato de `structify new`).

#### Problema 3
- `cmd/template/create.go`
  - `writeScaffoldYAML` ahora genera scaffold mínimo usable:
    - `inputs.project_name` con `validate: ^[a-zA-Z][a-zA-Z0-9_-]*$`
    - `steps: []`
  - Estructura base corregida:
    - crea `template/.gitkeep` (y no `README.md.tmpl`)
- `cmd/template/create_test.go`
  - Actualizado para reflejar el nuevo contenido esperado de `scaffold.yaml`.

---

### 3) Antes / Después (outputs reales del CLI)

#### Problema 1 — `template create` (prompt plano antes vs Bubbletea después)

**Antes (código viejo: prompts planos con bufio/fmt):**
```text
Template name: Description: Language (go/typescript/rust/...): Architecture (clean/vertical-slice/...): Author [jamt29]: Template created at /tmp/structify-v011-before-templates/v011-before-template
You can now add files under the template/ directory.
Then run: structify new --template v011-before-template --dry-run
```

**Después (bubbletea: pantalla del wizard en pseudo-TTY):**
```text
Template name?
>  

(enter to confirm, esc to cancel)
```

Evidencia adicional de que el flujo `new` renderiza progreso TUI (spinner + checkmarks) en pseudo-TTY usando un template temporal con un step seguro:
```text
Creating project

✓ Done
✓ Project created successfully

Path : /tmp/v011-output-progress-1774027006
Files : 1

Steps
  ✓ say hello

Next steps
  cd my-progress
  go run ./cmd/...
```

#### Problema 2 — regex vieja vs regex nueva (`project_name`)

**Antes (simulación con un template temporal usando regex vieja `^[a-z][a-z0-9-]+$`):**
```text
Error: invalid value for "project_name": does not match ^[a-z][a-z0-9-]+$
invalid value for "project_name": does not match ^[a-z][a-z0-9-]+$
```

**Después (built-in `clean-architecture-go` con regex nueva `^[a-zA-Z][a-zA-Z0-9_-]*$`):**
- `./bin/structify new --template clean-architecture-go --name my_project --dry-run` (OK)
  ```text
  Dry run — no files will be written.
  Template : clean-architecture-go
  Output   : ./my_project
  Variables: project_name=my_project, module_path=github.com/user/my-project, orm=none, transport=http
  Files that would be created:
  .gitignore
  Makefile
  README.md
  cmd/main.go
  internal/domain/entity/.gitkeep
  internal/domain/repository/interface.go
  internal/domain/repository/none/repository_impl.go
  internal/domain/usecase/.gitkeep
  internal/transport/http/handler.go
  internal/transport/http/router.go
  pkg/.gitkeep
  Steps that would run:
  ✓ go mod init github.com/user/my-project
  ─ go get google.golang.org/grpc  (skipped: transport == "grpc")
  ─ go get gorm.io/gorm  (skipped: orm == "gorm")
  ─ go get github.com/jmoiron/sqlx  (skipped: orm == "sqlx")
  ✓ go mod tidy
  No files were written.
  ```
- `./bin/structify new --template clean-architecture-go --name myApp --dry-run` (OK)
  ```text
  Dry run — no files will be written.
  Template : clean-architecture-go
  Output   : ./myApp
  Variables: project_name=myApp, module_path=github.com/user/my-app, orm=none, transport=http
  Files that would be created:
  .gitignore
  Makefile
  README.md
  cmd/main.go
  internal/domain/entity/.gitkeep
  internal/domain/repository/interface.go
  internal/domain/repository/none/repository_impl.go
  internal/domain/usecase/.gitkeep
  internal/transport/http/handler.go
  internal/transport/http/router.go
  pkg/.gitkeep
  Steps that would run:
  ✓ go mod init github.com/user/my-app
  ─ go get google.golang.org/grpc  (skipped: transport == "grpc")
  ─ go get gorm.io/gorm  (skipped: orm == "gorm")
  ─ go get github.com/jmoiron/sqlx  (skipped: orm == "sqlx")
  ✓ go mod tidy
  No files were written.
  ```
- Casos inválidos (fallan con mensaje de regex):
  - `--name "my project" --dry-run`:
    ```text
    Error: invalid value for "project_name": does not match ^[a-zA-Z][a-zA-Z0-9_-]*$
    invalid value for "project_name": does not match ^[a-zA-Z][a-zA-Z0-9_-]*$
    ```
  - `--name 123project --dry-run`:
    ```text
    Error: invalid value for "project_name": does not match ^[a-zA-Z][a-zA-Z0-9_-]*$
    invalid value for "project_name": does not match ^[a-zA-Z][a-zA-Z0-9_-]*$
    ```
  - `--name -project --dry-run`:
    ```text
    Error: invalid value for "project_name": does not match ^[a-zA-Z][a-zA-Z0-9_-]*$
    invalid value for "project_name": does not match ^[a-zA-Z][a-zA-Z0-9_-]*$
    ```

#### Problema 3 — `template create` scaffold usable para `structify new`

**Antes (template creado “viejo” sin inputs en scaffold -> `new` falla):**
```text
Error: project_name is required
project_name is required
```

**Después (nuevo `structify template create` end-to-end):**
`template create` (crea `v011-test-template`):
```text
✓ Template 'v011-test-template' created at /home/develop/.structify/templates/v011-test-template

Next steps
  1. Add your files to: /home/develop/.structify/templates/v011-test-template/template/
  2. Edit scaffold.yaml to add inputs and steps
  3. Test it: structify template validate /home/develop/.structify/templates/v011-test-template/
  4. Use it:  structify new --template v011-test-template
```

`template list` contiene el template:
```text
Local templates:
Name                Language  Architecture  Description
v011-test-template  go        clean         A test template
```

`template validate` OK:
```text
✓ Template is valid
Inputs: 1, Steps: 0, File rules: 0
```

`new --template v011-test-template --name myapp --dry-run` OK:
```text
Dry run — no files will be written.
Template : v011-test-template
Variables: project_name=myapp
Files that would be created:
.gitkeep
```

---

### 4) Cobertura
Comando:
```bash
go test ./... -cover
```
Resultado: **OK** (sin errores). Resumen de output:
```text
ok  	github.com/jamt29/structify	(cached)	coverage: 0.0% of statements
...
ok  	github.com/jamt29/structify/templates	(cached)	coverage: 100.0% of statements
```

---

### 5) Verificación de los tres problemas con output real

#### Problema 1 (TUI)
- `structify new` renderiza progreso Bubbletea (spinner + checkmarks) en pseudo-TTY con template temporal:
  - Output incluye `Creating project`, `✓ Done`, `✓ Project created successfully` y `Steps / ✓ say hello`.
- Evidencia parcial del paso inputs (Bubbletea prompt del `textinput`):
  - `Project name?`, `> my-progress`, `(enter to confirm, esc to cancel)`.

Nota: la automatización del “enter” para cerrar prompts Bubbletea en este entorno no siempre avanza; por eso se validó el paso 3/4 con `--name` (saltando inputs) para asegurar que la progresión sí es TUI.

#### Problema 2 (regex)
Válidos (OK):
- `./bin/structify new --template clean-architecture-go --name my-project --dry-run` (OK)
- `./bin/structify new --template clean-architecture-go --name my_project --dry-run` (OK)
- `./bin/structify new --template clean-architecture-go --name myApp --dry-run` (OK)
- `./bin/structify new --template clean-architecture-go --name api-v2 --dry-run` (OK)
- `./bin/structify new --template clean-architecture-go --name project123 --dry-run` (OK)

Inválidos (fallan con regex nueva):
- `./bin/structify new --template clean-architecture-go --name "my project" --dry-run` -> `does not match ^[a-zA-Z][a-zA-Z0-9_-]*$`
- `./bin/structify new --template clean-architecture-go --name 123project --dry-run` -> `does not match ^[a-zA-Z][a-zA-Z0-9_-]*$`
- `./bin/structify new --template clean-architecture-go --name -project --dry-run` -> `does not match ^[a-zA-Z][a-zA-Z0-9_-]*$`

#### Problema 3 (`template create`)
- `./bin/structify template create` -> creó `v011-test-template` y generó scaffold usable.
- `./bin/structify template list` muestra `v011-test-template`.
- `./bin/structify template validate ~/.structify/templates/v011-test-template/` -> OK.
- `./bin/structify new --template v011-test-template --name myapp --dry-run` -> OK (crearía `.gitkeep`).

---

### 6) Lecciones capturadas
Se añadieron lecciones a `tasks/lessons.md`:
- `L015 — Validar inputs[].validate también en modo flags`
- `L016 — template create debe generar scaffold mínimamente usable + UX consistente`

