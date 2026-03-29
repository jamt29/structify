# DSL Reference - scaffold.yaml

Esta referencia describe el DSL de templates de Structify (v0.5.1) para autores de `scaffold.yaml`.

## Estructura completa

```yaml
# Metadata
name: "clean-architecture-go"
version: "1.0.0"
author: "jamt29"
language: "go"
architecture: "clean"
description: "Clean Architecture para APIs en Go"
tags: ["go", "clean", "api"]

# Inputs que se preguntan al usuario
inputs:
  - id: project_name
    prompt: "Nombre del proyecto?"
    type: string
    required: true
    default: "my-api"
    validate: "^[a-z][a-z0-9-]+$"

  - id: transport
    prompt: "Transporte?"
    type: enum
    options: [http, grpc]
    default: http

  - id: orm
    prompt: "Persistencia?"
    type: enum
    options: [gorm, sqlx, none]
    default: none
    when: transport == "http"

  - id: include_docker
    prompt: "Incluir Docker?"
    type: bool
    default: true

  - id: features
    prompt: "Features opcionales"
    type: multiselect
    options: [auth, metrics, tracing]
    default: [auth]

  - id: project_path
    prompt: "Ruta destino"
    type: path
    must_exist: false

# Variables calculadas
computed:
  - id: module_path
    value: "github.com/acme/{{ project_name | kebab_case }}"

# Reglas de archivos
files:
  - include: "internal/transport/http/**"
    when: transport == "http"

  - include: "internal/transport/grpc/**"
    when: transport == "grpc"

  - include: "deploy/docker/**"
    when: include_docker == true

  - include: "internal/feature/auth/**"
    when: contains(features, "auth")

  - include: "internal/feature/metrics/**"
    when: contains(features, "metrics")

  - exclude: "internal/db/**"
    when: orm == "none"

# Comandos post-generacion
steps:
  - name: "Init module"
    run: "go mod init {{ module_path }}"

  - name: "Add gorm"
    run: "go get gorm.io/gorm"
    when: orm == "gorm"

  - name: "Add sqlx"
    run: "go get github.com/jmoiron/sqlx"
    when: orm == "sqlx"

  - name: "Tidy"
    run: "go mod tidy"
```

## Metadata

Campos principales de metadata:

- `name` (requerido): identificador de template.
- `version` (requerido): formato `X.Y.Z`.
- `author` (requerido): autor o equipo.
- `language` (requerido): `go`, `typescript`, `rust`, `csharp`, `python`.
- `architecture` (opcional): descripcion corta del estilo arquitectonico.
- `description` (opcional): texto libre.
- `tags` (opcional): lista de etiquetas.

## Inputs

Cada item de `inputs` define una variable que luego se usa en:

- reglas `when:`
- comandos `steps[].run`
- interpolacion de archivos `.tmpl`

Campos comunes por input:

- `id` (requerido): nombre interno, solo `^[a-z_]+$`.
- `prompt` (recomendado): texto mostrado en TUI.
- `type` (requerido): tipo del input.
- `required` (opcional): si es obligatorio cuando el input esta activo.
- `default` (opcional): valor por defecto.
- `when` (opcional): condicion para mostrar/pedir ese input.
- `validate` (opcional): regex para `type: string`.

### Tipo string

Texto libre, con opcion de regex:

```yaml
- id: project_name
  prompt: "Nombre del proyecto?"
  type: string
  required: true
  validate: "^[a-z][a-z0-9-]+$"
```

### Tipo enum

Seleccion de una sola opcion:

```yaml
- id: transport
  prompt: "Transporte?"
  type: enum
  options: [http, grpc, cli]
  default: http
```

### Tipo bool

Valor booleano:

```yaml
- id: include_ci
  prompt: "Incluir CI?"
  type: bool
  default: true
```

### Tipo multiselect (nuevo)

Seleccion de multiples opciones.

Internamente el valor queda como `[]string` para usar en `contains(...)`.

```yaml
- id: features
  prompt: "Features opcionales"
  type: multiselect
  options: [auth, metrics, tracing]
  default: [auth]
```

### Tipo path (nuevo)

Ruta de archivo o directorio.

`must_exist` controla validacion de existencia.

```yaml
- id: output_path
  prompt: "Ruta de salida"
  type: path
  must_exist: true
```

Comportamiento:

- `must_exist: true` -> falla si la ruta no existe.
- `must_exist: false` (o ausente) -> no exige existencia.

## Computed variables (nuevo)

`computed` permite crear variables derivadas a partir de otras:

```yaml
computed:
  - id: module_path
    value: "github.com/acme/{{ project_name | kebab_case }}"
```

Reglas:

- `id` sigue la misma convension de nombres que `inputs[].id`.
- `value` usa interpolacion `{{ }}`.
- las computed quedan disponibles para `steps`, `files` y `.tmpl`.

## File rules

`files` controla que partes de `template/` se incluyen o excluyen.

Cada regla debe tener exactamente una de estas claves:

- `include`
- `exclude`

Ejemplo:

```yaml
files:
  - include: "src/http/**"
    when: transport == "http"
  - exclude: "src/legacy/**"
    when: include_legacy == false
```

Reglas importantes:

- Se usan globs con `/`.
- `**` matchea cero o mas segmentos.
- Si varias reglas aplican al mismo path, la ultima regla gana.

## Steps

`steps` define comandos post-generacion.

```yaml
steps:
  - name: "Install deps"
    run: "npm install"
  - name: "Enable prisma"
    run: "npm install prisma @prisma/client"
    when: contains(features, "prisma")
```

Campos:

- `name` (requerido): etiqueta visible.
- `run` (requerido): comando shell.
- `when` (opcional): condicion DSL.

## Expresiones when:

Las expresiones `when:` se usan en:

- `inputs[].when`
- `files[].when`
- `steps[].when`

### Operadores

Comparacion:

- `==`
- `!=`

Logicos:

- `&&`
- `||`
- `!`

Agrupacion:

- `( ... )`

Literales:

- strings con comillas dobles (`"http"`)
- booleanos (`true`, `false`)

### Funcion contains() (nuevo)

La funcion `contains()` devuelve booleano y admite 2 formas:

1. `contains(<string>, <substring>)`
2. `contains(<[]string>, <item>)` (ideal para `multiselect`)

Ejemplos validos:

```yaml
when: contains(project_name, "api")
when: contains(features, "auth")
when: contains(features, "metrics") && transport == "http"
```

Errores comunes:

- `contains(features)` -> faltan argumentos
- `contains(features, true)` -> segundo argumento debe ser string
- `contains(include_ci, "x")` -> primer argumento debe ser string o []string

### Ejemplos

```yaml
when: transport == "http"
when: orm != "none"
when: !include_docker
when: transport == "http" && orm == "gorm"
when: (transport == "http" || transport == "grpc") && include_ci == true
when: contains(features, "auth")
```

## Interpolacion {{ }}

Se usa en:

- archivos `.tmpl`
- strings de `steps[].run`
- strings de `computed[].value`

Sintaxis:

```text
{{ variable }}
{{ variable | filtro }}
```

Reglas:

- delimitadores `{{` y `}}`
- solo un filtro por expresion (sin chaining)
- si la variable no existe, la renderizacion falla

### Filtros disponibles (tabla)

| Filtro | Entrada ejemplo | Salida ejemplo | Uso |
|---|---|---|---|
| `snake_case` | `MyProject` | `my_project` | `{{ project_name \| snake_case }}` |
| `pascal_case` | `my-project` | `MyProject` | `{{ project_name \| pascal_case }}` |
| `camel_case` | `my-project` | `myProject` | `{{ project_name \| camel_case }}` |
| `kebab_case` | `MyProject` | `my-project` | `{{ project_name \| kebab_case }}` |
| `upper` | `hello` | `HELLO` | `{{ env \| upper }}` |
| `lower` | `HELLO` | `hello` | `{{ env \| lower }}` |

## Errores comunes y como resolverlos

1. Operador incorrecto en `when:`
   - Error tipico: `transport = "http"`
   - Solucion: usar `==` o `!=`.

2. Strings sin comillas dobles
   - Error tipico: `transport == http`
   - Solucion: `transport == "http"`.

3. `enum` o `multiselect` sin `options`
   - Error tipico: falta de opciones.
   - Solucion: definir lista `options: [...]`.

4. Variable no definida en interpolacion
   - Error tipico: `{{ module }}` cuando no existe `module`.
   - Solucion: revisar `inputs`/`computed` disponibles.

5. Chaining de filtros
   - Error tipico: `{{ name | snake_case | upper }}`
   - Solucion: usar solo un filtro por expresion.

6. `path` con `must_exist: true` apuntando a ruta inexistente
   - Error tipico: validacion falla en input.
   - Solucion: crear la ruta antes o usar `must_exist: false`.

7. Ciclos entre `inputs[].when`
   - Error tipico: A depende de B y B depende de A.
   - Solucion: romper la dependencia circular y dejar un orden evaluable.

8. `contains()` con tipos incorrectos
   - Error tipico: usar bool como primer argumento.
   - Solucion: primer arg string o `[]string`, segundo arg string.

