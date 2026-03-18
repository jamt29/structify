## 1) Resumen ejecutivo

Se implementó el **DSL completo** en `internal/dsl/` incluyendo **lexer**, **parser** (precedencia + paréntesis), **evaluator** (type-safe + cortocircuito), **interpolación `{{ }}`** con filtros, **carga de `scaffold.yaml`**, **validator** (incluyendo detección de ciclos en `inputs[].when`) y una suite de **tests table-driven** con **cobertura > 80%**.

## 2) Componentes creados

- `internal/dsl/types.go`: Tipos compartidos del DSL (tokens, AST, `Context`) y structs del modelo `scaffold.yaml` (`Manifest`, `Input`, `FileRule`, `Step`).
- `internal/dsl/lexer.go`: Lexer para expresiones `when:` (strings con `"..."`, bool, ident, `== != && || ! ( )`, whitespace, `ILLEGAL` descriptivo).
- `internal/dsl/parser.go`: Parser recursivo descendente con precedencia `|| < && < ! < comparación` y errores con posición.
- `internal/dsl/evaluator.go`: Evaluación del AST contra `Context` con cortocircuito en `&&/||` y validación de tipos.
- `internal/dsl/interpolator.go`: Interpolación de `{{ variable }}` y `{{ variable | filtro }}` con validación (variable/filtro inexistente, no chaining).
- `internal/dsl/filters.go`: Implementación de filtros (`snake_case`, `pascal_case`, `camel_case`, `kebab_case`, `upper`, `lower`) soportando mezcla de separadores y CamelCase.
- `internal/dsl/manifest.go`: `LoadManifest(path)` con `gopkg.in/yaml.v3` y errores descriptivos.
- `internal/dsl/validator.go`: `ValidateManifest` acumulando todos los errores (metadata, inputs, files, steps, parseo de `when:` y ciclos en `inputs[].when`).
- `internal/dsl/*_test.go`: Tests exhaustivos (lexer/parser/evaluator/interpolator/filters/manifest/validator).

## 3) Decisiones de implementación

- **Precedencia del parser**: Implementada como `parseOr -> parseAnd -> parseNot -> parseCompare -> parsePrimary`, respetando `|| < && < ! < comparación` y soportando paréntesis.
- **Errores del lexer**: Para patrones comunes de usuario (`=` en vez de `==`, `&` en vez de `&&`, `|` en vez de `||`, comillas simples) se devuelven tokens `ILLEGAL` con mensaje accionable.
- **Word-splitting de filtros**: Se normalizan separadores (`-`, `_`, espacios) y luego se separa CamelCase manteniendo acrónimos (ej. `MyAPIClient` -> `My`, `API`, `Client`).
- **Interpolación**: Se limita a **un filtro** (v1) y se rechaza chaining con error descriptivo.
- **Ciclos en inputs**: Se parsean ASTs de `inputs[].when`, se extraen identificadores y se construye un grafo de dependencias para detectar ciclos (DFS), reportando errores en los campos `inputs[i].when`.

## 4) Cobertura de tests

- `go test ./internal/dsl/... -cover` → **coverage: 85.8% of statements**

## 5) Casos edge manejados

- **Comillas simples en `when:`** (`'http'`): Lexer devuelve `ILLEGAL` con mensaje “single quotes are not supported; use double quotes”.
- **Operador inválido `=`**: Lexer devuelve `ILLEGAL` con sugerencia “did you mean '=='?” y el parser propaga error con posición.
- **Interpolación sin cerrar** (`{{ ...`): error `unterminated interpolation starting at position N`.
- **Chaining de filtros** (`{{ x | a | b }}`): error `filter chaining is not supported`.
- **Ciclo en `inputs[].when`** (A depende de B y B de A): validator detecta y reporta el ciclo con ruta `inputs[i].when`.

## 6) Estado final

- `go build ./...` → **PASS**
- `go test ./internal/dsl/... -v` → **PASS** (todos los tests)

## 7) Lecciones capturadas

Se añadió/actualizó `tasks/lessons.md` con decisiones y casos edge del DSL (precedencia, split de palabras, ciclos de `when`, mensajes de error).

