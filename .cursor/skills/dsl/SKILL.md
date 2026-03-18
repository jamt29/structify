---
name: structify-dsl
description: Spec completa del DSL de Structify. Leer SIEMPRE antes de tocar cualquier archivo en internal/dsl/. Contiene sintaxis, tokens, AST, reglas del evaluador y casos edge.
---

# Structify DSL — Especificación Completa

## Filosofía
El DSL de Structify vive en `scaffold.yaml`. Su propósito es:
1. Declarar variables que se le preguntan al usuario (`inputs`)
2. Controlar qué archivos se generan (`files`) mediante expresiones `when:`
3. Definir comandos post-generación (`steps`) con ejecución condicional

Los archivos de contenido (`.tmpl`) usan interpolación `{{ variable | filtro }}`.
Son dos sistemas separados: el DSL del YAML y el interpolador de archivos.

---

## Estructura Completa de scaffold.yaml

```yaml
# ── METADATA ──────────────────────────────────────────────
name: "clean-architecture-go"
version: "1.0.0"
author: "username"
language: "go"                  # go | typescript | rust | csharp | python
architecture: "clean"           # clean | vertical-slice | hexagonal | mvc | monorepo
description: "Clean Architecture en Go con soporte HTTP/gRPC"
tags: ["go", "clean", "api"]    # Para búsqueda en el registry

# ── INPUTS ────────────────────────────────────────────────
# Variables que se preguntan al usuario antes de generar
inputs:
  - id: project_name            # Identificador único (usar en {{ }} y en when:)
    prompt: "Project name?"     # Texto mostrado al usuario
    type: string                # string | enum | bool
    required: true              # Si es false, puede quedar vacío
    default: "my-project"       # Valor por defecto (opcional)
    validate: "^[a-z][a-z0-9-]+$"  # Regex de validación (opcional)

  - id: transport
    prompt: "Transport layer?"
    type: enum
    options: [http, grpc, cli]
    default: http

  - id: orm
    prompt: "ORM?"
    type: enum
    options: [gorm, sqlx, none]
    default: gorm
    when: transport != "cli"    # Solo se pregunta si esta condición es true

  - id: use_docker
    prompt: "Include Docker support?"
    type: bool
    default: true

# ── FILES ─────────────────────────────────────────────────
# Control de qué archivos/carpetas incluir o excluir
files:
  - include: "internal/transport/http/**"
    when: transport == "http"

  - include: "internal/transport/grpc/**"
    when: transport == "grpc"

  - exclude: "internal/db/**"
    when: orm == "none"

  - include: "docker/**"
    when: use_docker == true

# ── STEPS ─────────────────────────────────────────────────
# Comandos a ejecutar después de generar los archivos
steps:
  - name: "Init go module"
    run: "go mod init {{ project_name }}"
    # Sin when: → siempre se ejecuta

  - name: "Install GORM"
    run: "go get gorm.io/gorm"
    when: orm == "gorm"

  - name: "Install sqlx"
    run: "go get github.com/jmoiron/sqlx"
    when: orm == "sqlx"

  - name: "Install Gin"
    run: "go get github.com/gin-gonic/gin"
    when: transport == "http"

  - name: "Tidy"
    run: "go mod tidy"
```

---

## Tipos de Input

| Tipo | Descripción | Valor resultante |
|---|---|---|
| `string` | Texto libre | `string` |
| `enum` | Una opción de una lista | `string` (el valor elegido) |
| `bool` | Sí/No | `bool` |

---

## Sintaxis de Expresiones `when:`

### Operadores de comparación
```
==    igual
!=    distinto
```

### Operadores lógicos
```
&&    AND
||    OR
!     NOT (unario, solo para bool)
```

### Agrupación
```
(expr)    paréntesis para controlar precedencia
```

### Valores literales
```
"http"     string (comillas dobles obligatorias)
true       booleano
false      booleano
```

### Identificadores
```
project_name    referencia a un input por su id
transport
orm
use_docker
```

### Ejemplos válidos
```yaml
when: transport == "http"
when: orm != "none"
when: use_docker == true
when: !use_docker
when: transport == "http" && orm != "none"
when: transport == "http" || transport == "grpc"
when: (transport == "http" || transport == "grpc") && orm != "none"
```

### Ejemplos inválidos (errores de usuario)
```yaml
when: transport = "http"       # ERROR: = no es operador válido
when: transport == http        # ERROR: string sin comillas
when: transport == 'http'      # ERROR: comillas simples no soportadas
when: nombre > 5               # ERROR: > no es operador válido
```

---

## Sintaxis de Interpolación `{{ }}`

Usada en:
- Valores de `run:` en steps
- Archivos `.tmpl` de contenido

### Sintaxis básica
```
{{ variable_id }}
{{ project_name }}
```

### Con filtro
```
{{ variable_id | filtro }}
{{ project_name | snake_case }}
```

### Filtros disponibles

| Filtro | Entrada | Salida | Ejemplo |
|---|---|---|---|
| `snake_case` | `MyProject` | `my_project` | `{{ name \| snake_case }}` |
| `pascal_case` | `my-project` | `MyProject` | `{{ name \| pascal_case }}` |
| `camel_case` | `my-project` | `myProject` | `{{ name \| camel_case }}` |
| `kebab_case` | `MyProject` | `my-project` | `{{ name \| kebab_case }}` |
| `upper` | `hello` | `HELLO` | `{{ name \| upper }}` |
| `lower` | `HELLO` | `hello` | `{{ name \| lower }}` |

### Reglas de interpolación
- Los delimitadores son `{{` y `}}` con espacios opcionales
- Solo se soporta un filtro por expresión (no chaining: `{{ x | a | b }}` no es válido en v1)
- Si la variable no existe → error descriptivo, no silencio
- Los archivos `.tmpl` se procesan antes de escribir al destino

---

## Implementación: Componentes del DSL

### 1. Lexer (`internal/dsl/lexer.go`)

Tokens a reconocer:

```go
type TokenType string

const (
    // Literales
    TOKEN_STRING     TokenType = "STRING"      // "http"
    TOKEN_BOOL       TokenType = "BOOL"        // true | false
    TOKEN_IDENT      TokenType = "IDENT"       // transport, orm, project_name

    // Operadores
    TOKEN_EQ         TokenType = "=="
    TOKEN_NEQ        TokenType = "!="
    TOKEN_AND        TokenType = "&&"
    TOKEN_OR         TokenType = "||"
    TOKEN_NOT        TokenType = "!"

    // Agrupación
    TOKEN_LPAREN     TokenType = "("
    TOKEN_RPAREN     TokenType = ")"

    // Control
    TOKEN_EOF        TokenType = "EOF"
    TOKEN_ILLEGAL    TokenType = "ILLEGAL"
)

type Token struct {
    Type    TokenType
    Literal string
    Pos     int  // posición en el string original (para mensajes de error)
}
```

### 2. Parser (`internal/dsl/parser.go`)

AST nodes:

```go
type Node interface {
    nodeType() string
}

// Nodo hoja: variable o literal
type IdentNode struct {
    Name string
}

type StringLiteralNode struct {
    Value string
}

type BoolLiteralNode struct {
    Value bool
}

// Nodo de comparación: left == right | left != right
type CompareNode struct {
    Left     Node
    Operator string  // "==" | "!="
    Right    Node
}

// Nodo lógico binario: left && right | left || right
type BinaryNode struct {
    Left     Node
    Operator string  // "&&" | "||"
    Right    Node
}

// Nodo NOT unario: !expr
type NotNode struct {
    Expr Node
}
```

Precedencia (menor a mayor):
```
||    (menor precedencia)
&&
!     (mayor precedencia, unario)
==, !=
(expr)
```

### 3. Evaluator (`internal/dsl/evaluator.go`)

```go
// Context contiene los valores de todas las variables del usuario
type Context map[string]interface{}

// Evaluate evalúa un nodo AST contra un contexto
// Retorna (bool, error)
func Evaluate(node Node, ctx Context) (bool, error)
```

Reglas:
- `IdentNode` → buscar en Context, error si no existe
- `StringLiteralNode` → valor directo
- `BoolLiteralNode` → valor directo
- `CompareNode` → evaluar ambos lados y comparar (type-safe)
- `BinaryNode &&` → cortocircuito: si left=false, no evaluar right
- `BinaryNode ||` → cortocircuito: si left=true, no evaluar right
- `NotNode` → evaluar expr, negar

### 4. Interpolador (`internal/dsl/interpolator.go`)

```go
// Interpolate reemplaza {{ var }} y {{ var | filter }} en un string
func Interpolate(template string, ctx Context) (string, error)

// InterpolateFile procesa un archivo completo línea a línea
func InterpolateFile(content []byte, ctx Context) ([]byte, error)
```

Algoritmo:
1. Buscar `{{` en el string
2. Encontrar el `}}` correspondiente
3. Extraer el contenido entre delimitadores, trimear espacios
4. Si contiene `|`: separar en `variable` y `filtro`
5. Buscar variable en Context
6. Aplicar filtro si existe
7. Reemplazar `{{ ... }}` por el valor resultante
8. Repetir hasta no quedar más `{{`

### 5. Validator (`internal/dsl/validator.go`)

Validaciones al cargar un `scaffold.yaml`:

```
- name: requerido, no vacío
- version: requerido, formato semver (X.Y.Z)
- language: requerido, valor en lista conocida
- inputs[].id: requerido, único, solo [a-z_]
- inputs[].type: requerido, valor en [string, enum, bool]
- inputs[].options: requerido si type == enum
- inputs[].when: si existe, debe parsear sin error
- files[].include XOR files[].exclude: no pueden tener ambos
- files[].when: si existe, debe parsear sin error
- steps[].name: requerido
- steps[].run: requerido
- steps[].when: si existe, debe parsear sin error
```

---

## Mensajes de Error del DSL

Los errores deben ser accionables. Ejemplos:

```
Error en scaffold.yaml:
  → steps[2].when: expresión inválida en posición 12
    "transport = 'http'"
                ^
  Sugerencia: usar == en lugar de =, y comillas dobles "http"

Error en interpolación de cmd/main.go.tmpl, línea 5:
  → Variable 'project' no definida
    Disponibles: project_name, transport, orm
```

---

## Tests Obligatorios (Fase 2.9)

Tabla de casos para el evaluator:

```go
// Casos que DEBEN pasar
{"transport == 'http'",  ctx{"transport": "http"},  true}
{"transport != 'grpc'",  ctx{"transport": "http"},  true}
{"use_docker == true",   ctx{"use_docker": true},    true}
{"!use_docker",          ctx{"use_docker": false},   true}
{"a == 'x' && b == 'y'", ctx{"a":"x","b":"y"},       true}
{"a == 'x' && b == 'y'", ctx{"a":"x","b":"z"},       false}
{"a == 'x' || b == 'y'", ctx{"a":"z","b":"y"},       true}
{"(a == 'x' || b == 'y') && c != 'z'", ctx{...},    true}

// Casos que DEBEN fallar con error descriptivo
{"transport = 'http'",   ctx{},   error: "operador inválido '='"}
{"undeclared == 'x'",    ctx{},   error: "variable 'undeclared' no definida"}
```

---

## Casos Edge Importantes

1. **String vacío como input**: `project_name == ""` es válido
2. **Bool como string**: Si el usuario tipea `"true"` en vez de `true` → error descriptivo
3. **Input con `when:` circular**: Input A con `when: B == "x"` y B con `when: A == "y"` → detectar y error
4. **Archivo `.tmpl` con `{{` sin cerrar**: Error con número de línea
5. **Steps con `run:` que contiene interpolación**: Evaluar antes de ejecutar
6. **`when:` vacío**: Tratar como ausente (siempre ejecutar), no como error
