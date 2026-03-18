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

