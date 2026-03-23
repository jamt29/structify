# V020-RESULTADO — v0.2.0 Templates + DSL

## 1) Feature 1 — `structify template import`

### Implementado
- Nuevo comando: `structify template import <source> [--name] [--yes]`.
- Soporta source local y GitHub URL.
- Analyzer nuevo en `internal/template/analyzer.go`:
  - Detección de lenguaje por archivos clave.
  - Detección de variables sugeridas (`project_name`, `module_path`) con enfoque pragmático.
  - Detección de includes/ignores por defaults.
- Generación de template importado:
  - Copia de archivos a `~/.structify/templates/<name>/template/`.
  - Reemplazo de variables detectadas por `{{ variable_id }}`.
  - Conversión a `.tmpl` solo cuando hubo reemplazos.
  - `scaffold.yaml` generado con inputs detectados y steps base por lenguaje.

### Verificación real (import local)
Comando:
```bash
./bin/structify template import /tmp/test-import --name imported-test --yes
```
Salida:
```text
✓ Template 'imported-test' creado en /home/develop/.structify/templates/imported-test/
Archivos: 2 incluidos, 0 ignorados
Variables: project_name, module_path
Inputs: 2 detectados
Para usarlo:
structify new --template imported-test
Para editarlo:
structify template edit imported-test
```

Validación y uso:
```bash
./bin/structify template validate ~/.structify/templates/imported-test/
./bin/structify new --template imported-test --name myapp --dry-run
```
Resultado:
```text
✓ Template is valid
Inputs: 2, Steps: 2, File rules: 0
...
Files that would be created:
go.mod
main.go
Steps that would run:
✓ go mod init github.com/test/test-import
✓ go mod tidy
```

### Verificación real (import GitHub)
Comando:
```bash
./bin/structify template import github.com/jamt29/structify --name structify-template --yes
```
Salida:
```text
✓ Template 'structify-template' creado en /home/develop/.structify/templates/structify-template/
Archivos: 170 incluidos, 1 ignorados
Variables: module_path
Inputs: 1 detectados
Para usarlo:
structify new --template structify-template
Para editarlo:
structify template edit structify-template
```

## 2) Feature 2 — `structify template edit <name>`

### Implementado
- Nuevo comando: `structify template edit <name>`.
- Abre `scaffold.yaml` con `$EDITOR` y fallback `vim`/`nano`/`vi`.
- Al salir del editor:
  - `dsl.LoadManifest`
  - `dsl.ValidateManifest`
- Si hay errores: flujo con opciones:
  1. Volver a editar
  2. Guardar de todas formas
  3. Descartar cambios

### Verificación real (scaffold válido)
Comando:
```bash
EDITOR=true ./bin/structify template edit imported-test
```
Salida:
```text
✓ scaffold.yaml actualizado y válido
```

### Verificación real (scaffold inválido)
Se forzó un `scaffold.yaml` inválido y se ejecutó `template edit`.
Salida:
```text
El scaffold.yaml tiene errores:
· name: name is required
· version: version is required
· language: language is required

¿Qué deseas hacer?
1) Volver a editar
2) Guardar de todas formas (no recomendado)
3) Descartar cambios
> Cambios descartados
```

## 3) Feature 3 — Variables avanzadas del DSL

### Implementado
- Nuevos tipos de input:
  - `multiselect`
  - `path` (con `must_exist`)
- Variables calculadas:
  - `computed` agregado al `Manifest`.
  - Resolución aplicada en `cmd/new.go` antes del uso final del contexto.
- Expresiones `when:` extendidas con `contains(...)`:
  - Lexer: token `,`.
  - Parser: `CallNode` y parseo de función.
  - Evaluator: `contains(string,string)` y `contains([]string,string)`.
- TUI:
  - Soporte para prompt de `multiselect` con `space` para toggle y `enter` para confirmar.

### Casos de test nuevos (`contains`)
Incluidos en `internal/dsl/evaluator_test.go`:
- `contains(features, "docker")` con `features=["logging","docker"]` => `true`
- `contains(features, "auth")` con `features=["logging","docker"]` => `false`
- `contains(features, "docker") && transport == "http"` => `true`

### Resultado de tests del evaluator
Comando:
```bash
go test ./internal/dsl/... -v -run TestEvaluate
```
Resultado: todos los casos pasaron (`PASS`).

## 4) Cobertura

Comando:
```bash
go test ./... -cover
```
Salida relevante:
```text
ok  	github.com/jamt29/structify/internal/dsl	coverage: 87.2% of statements
```

Estado: cobertura DSL >= 87% cumplida.

## 5) Estado final

Comandos:
```bash
go build ./...
go test ./...
```

Resultado:
- `go build ./...` => OK
- `go test ./...` => OK

## 6) Lecciones capturadas

Se añadieron lecciones en `tasks/lessons.md`:
- `L020` — Asegurar tests dirigidos para mantener cobertura DSL al agregar features.
- `L021` — Analyzer de import debe basarse en señales estructurales para reducir falsos positivos.
