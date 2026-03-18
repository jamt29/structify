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

