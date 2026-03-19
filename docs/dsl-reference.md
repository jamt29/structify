# DSL Reference: `scaffold.yaml`

Structify templates are described by a `scaffold.yaml` file.

This document explains the user-facing DSL: `inputs`, `files` inclusion rules, `steps`, and the `{{ }}` interpolation system.

---

## `scaffold.yaml` structure

```yaml
name: "clean-architecture-go"          # Required: template identifier (non-empty)
version: "1.0.0"                       # Required: SemVer (X.Y.Z)
author: "your-github-username"        # Required: author text
language: "go"                        # Required: go | typescript | rust | csharp | python
architecture: "clean"                 # Recommended: clean | vertical-slice | hexagonal | mvc | monorepo
description: "..."                    # Required: one-line description
tags: ["go", "clean", "api"]          # Required: list of tokens for discovery

inputs: []                             # Optional: variables asked to the user
files: []                              # Optional: include/exclude rules for template files
steps: []                              # Optional: commands executed after files are generated
```

---

## Inputs

Each `inputs[]` entry declares a variable the template can use:

```yaml
inputs:
  - id: project_name
    prompt: "Project name?"
    type: string                # string | enum | bool
    required: true
    default: "my-project"      # Optional
    validate: "^[a-z][a-z0-9-]+$" # Optional regex validation
    when: transport != "grpc"  # Optional condition (DSL expression)

  - id: runtime
    prompt: "Runtime?"
    type: enum
    options: [express, fastify]
    default: express
```

Input types:

| type | Meaning | Value in context |
|---|---|---|
| `string` | free text | string |
| `enum` | one of a list | string (chosen option) |
| `bool` | true/false | boolean |

---

## `when:` expressions (conditions)

`when:` is used in `inputs[]`, `files[]`, and `steps[]`.

### Operators

| Operator | Meaning |
|---|---|
| `==` | equals |
| `!=` | not equals |
| `&&` | AND |
| `||` | OR |
| `!` | NOT (unary) |

### Parentheses

Use `(expr)` to control precedence.

### Literals and identifiers

- Strings must use double quotes, for example `"http"`.
- Booleans are `true` and `false`.
- Identifiers refer to input `id` values, for example `transport` or `use_prisma`.

### Examples

Valid:

```yaml
when: transport == "http"
when: orm != "none"
when: transport == "http" && orm != "none"
when: (transport == "http" || transport == "grpc") && orm != "none"
```

Invalid (common mistakes):

```yaml
when: transport = "http"       # ERROR: use == not =
when: transport == http        # ERROR: strings need double quotes
when: use_prisma == "true"    # ERROR: comparing bool with string
```

---

## `files:` rules (include/exclude)

`files[]` controls which paths inside `template/` are copied to the output directory.

Each rule must declare either `include` or `exclude` (not both).

Example:

```yaml
files:
  - include: "src/http/**"
    when: runtime == "express"

  - exclude: "src/experimental/**"
    when: use_prisma == false
```

Glob patterns:

- Use forward slashes (`/`) in paths.
- `**` matches zero or more path segments.
- `*` matches within a single path segment.

Rule precedence:

- When multiple rules match a file, the **last matching rule wins**.

---

## `steps:` (post-generation commands)

`steps[]` are executed after files are generated.

```yaml
steps:
  - name: "Init go module"
    run: "go mod init {{ module_path }}"

  - name: "Tidy"
    run: "go mod tidy"
```

Each step supports:
- `name`: human-friendly label (required)
- `run`: shell command string (required)
- `when`: optional condition using the DSL expression language

---

## Interpolation in `.tmpl` files: `{{ }}`

Files ending in `.tmpl` are rendered with the DSL context.

Interpolation syntax:

```text
{{ variable }}
{{ variable | filter }}
```

Supported filters:

| filter | Input example | Output example |
|---|---|---|
| `snake_case` | `MyProject` | `my_project` |
| `pascal_case` | `my-project` | `MyProject` |
| `camel_case` | `my-project` | `myProject` |
| `kebab_case` | `MyProject` | `my-project` |
| `upper` | `hello` | `HELLO` |
| `lower` | `HELLO` | `hello` |

Rules:

- Only one filter is supported per interpolation (no chaining).
- If a variable is not defined, rendering fails with a descriptive error.

Examples:

```ts
export interface {{ project_name | pascal_case }} {
  id: string;
}
```

---

## Common errors

1. `when:` uses wrong operator
   - Symptom: parse error or unexpected evaluation.
   - Fix: use `==`/`!=` (not `=`).

2. Missing double quotes for string literals
   - Symptom: lexer/parser error.
   - Fix: use `"http"`, not `http`.

3. Variables not defined
   - Symptom: `variable 'X' not defined in context`.
   - Fix: make sure the input `id: X` exists and is active for the current `when` path.

4. `filter chaining is not supported`
   - Symptom: interpolation error.
   - Fix: use only one filter: `{{ x | kebab_case }}`.

5. Unterminated interpolation (`{{ ...`)
   - Symptom: `unterminated interpolation starting at position ...`.
   - Fix: ensure every `{{` has a matching `}}`.

