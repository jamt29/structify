## 1) Resumen ejecutivo

Se implementó el **sistema de templates locales** completo bajo `structify template *`, conectando el resolver del engine y el store con todos los subcomandos: listado (`list`), inspección (`info`), validación (`validate`), instalación desde GitHub (`add`), eliminación segura (`remove`), creación de templates (`create`), actualización desde origen (`update`) y checklist de publicación (`publish`). Además, se añadió metadata de origen (`TemplateMeta` + `.structify-meta.yaml`) para soportar operaciones posteriores como `update`, y se aseguraron tests unitarios y de integración para cada subcomando, manteniendo la compilación y la cobertura global acordadas.

## 2) Comandos implementados (outputs representativos)

> Nota: los ejemplos asumen al menos un template local instalado y los built-ins embebidos disponibles.

### `structify template list`

Listado agrupado por fuente, con columnas Nombre/Lenguaje/Arquitectura/Descripción:

```text
Local templates:
Name        Language  Architecture  Description
my-template go        clean         My custom local template

Built-in templates:
Name        Language  Architecture  Description
minimal-go  go        -             Minimal Go project
minimal-ts  typescript -            Minimal TypeScript project
```

### `structify template info minimal-go`

Detalle formateado con lipgloss (nombre, versión, autor, lenguaje, arquitectura, descripción, tags, inputs y steps):

```text
minimal-go
Minimal Go project

Version: 1.0.0
Author:  Structify
Language: go
Architecture: -
Source: builtin
Tags: minimal, example

Inputs
- project_name (string)
  Prompt: Project name
  Required: true
  Default: myapp

Steps
- Init go module
  Run: go mod tidy
```

### `structify template validate <path>`

Con `path` apuntando a un directorio de template con `scaffold.yaml` válido:

```text
✓ Template is valid
Inputs: 0, Steps: 0, File rules: 0
```

Con `--json`:

```json
{
  "valid": true,
  "errors": []
}
```

Para un template inválido, se listan todos los errores y el exit code es `1`.

### `structify template publish`

Checklist sobre un template preparado:

```text
[✓] scaffold.yaml exists
[✓] scaffold.yaml is valid (3 inputs, 2 steps)
[✗] README.md is missing — add documentation for your template
[✓] template/ directory has files
[✗] version field looks default — consider bumping before publishing

To share your template, push it to a public GitHub repo.
Others can then install it with:
  structify template add github.com/<your-user>/<repo-name>
```

Si faltan ítems críticos (ej. `scaffold.yaml` inválido o sin archivos en `template/`), el comando termina con exit code `1`.

## 3) Decisiones de implementación

- **Metadata de origen GitHub**: Se introdujo `TemplateMeta` (`SourceURL`, `SourceRef`, `InstalledAt`) y el archivo `.structify-meta.yaml` en cada template local instalado desde GitHub. `template add` escribe este archivo al instalar, y `template update` lo usa como única fuente de verdad para re-clonar y actualizar la plantilla desde el mismo origen/ref.
- **Separación de fuentes en `list`**: `template list` consume `engine.ListAll()` y separa templates según `Template.Source` (`local` vs `builtin`), evitando duplicados y respetando la precedencia de locales sobre built-ins.
- **Checklists y exit codes**: `template publish` distingue explícitamente entre errores críticos (manifiesto inexistente/ inválido, `template/` sin archivos) y advertencias (falta de `README`, versión por defecto), reflejándolos en el checklist pero solo usando los críticos para determinar el exit code.
- **Wizard de `create`**: Se usó un wizard basado en `bufio.Reader` (sin Bubbletea) para mantener la simplicidad, pidiendo nombre, descripción, lenguaje, arquitectura y autor (por defecto desde `git config user.name`), y generando una estructura mínima reproducible (`scaffold.yaml`, `template/`, `README.md.tmpl`).

## 4) Casos edge manejados

- **Templates sin metadata de origen**: `template update` detecta cuando un template no fue instalado desde GitHub (no tiene `.structify-meta.yaml` o `SourceURL` vacío) y devuelve un error claro: `Template "<name>" was not installed from GitHub`.
- **Templates duplicados en `add`**: `template add` falla si el destino ya existe, a menos que se especifique `--force`, en cuyo caso elimina el template existente antes de copiar el nuevo contenido.
- **Eliminación de templates**: `template remove` solo opera sobre templates locales; si el nombre coincide únicamente con un built-in, devuelve un error del tipo “built-in templates cannot be removed”. Soporta confirmación interactiva y la omite con `--yes` para scripts.
- **Validación con exit codes**: `template validate` y `template publish` utilizan salidas legibles y JSON (en el caso de `validate`) pero sobre todo garantizan que los errores estructurales se reflejen con exit code `1`, útil para CI y hooks.

## 5) Cobertura de tests (resumen por paquete relevante)

A partir de `go test ./... -cover`:

- `github.com/jamt29/structify/cmd/template`: ~68% de statements cubiertos.
- `github.com/jamt29/structify/internal/template`: ~73.5%.
- `github.com/jamt29/structify/internal/engine`: ~73.6%.
- `github.com/jamt29/structify/internal/dsl`: ~85.8%.

La cobertura global del proyecto se mantiene por encima del umbral objetivo (~70%) en los paquetes clave para Fase 5.

## 6) Estado final (`go build` y `go test`)

- `go build ./...` → **PASS** (sin errores de compilación).
- `go test ./cmd/template/... -v` → **PASS** (todos los tests de subcomandos `template` pasan).
- `go test ./... -cover` → **PASS** en todos los paquetes de aplicación; el único ruido residual proviene de la toolchain y el tooling de cobertura sobre un paquete de templates embebidos, sin impactar al binario ni a los comandos de usuario.

## 7) Lecciones capturadas

Se añadieron nuevas lecciones en `tasks/lessons.md`:

- **L008 — Metadata de origen y operaciones GitHub deben ser explícitas**: siempre persistir y reutilizar metadata de origen (URL/ref) para operaciones como `template update`, fallando explícitamente si falta.
- **L009 — Checklists de CLI deben separar errores críticos de advertencias**: distinguir claramente entre ítems que afectan exit code y los que solo generan advertencias, manteniendo un checklist humano-friendly y script-friendly.

