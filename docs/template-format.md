## Structify Template Format

### What is a Structify template repo?

A Structify template repo is a public GitHub repository that contains a `scaffold.yaml` manifest at its root and a `template/` directory with the files that Structify will scaffold into new projects. Users can install it with:

```bash
structify template add github.com/<user>/<repo>[@ref]
```

Once installed, the template becomes available to `structify new`.

---

### Required folder structure

At minimum, a compatible template repository must look like this:

```text
my-template-repo/
├── scaffold.yaml        # Template manifest (required)
├── template/            # Source files to generate (required)
│   ├── cmd/
│   │   └── main.go.tmpl
│   ├── internal/
│   │   └── app/
│   │       └── app.go.tmpl
│   └── ...
└── README.md            # Recommended documentation
```

- `scaffold.yaml` **must** live in the repository root.
- `template/` **must** exist and contain at least one file; Structify copies and renders from this directory.

---

### Required fields in `scaffold.yaml`

The `scaffold.yaml` file is defined by Structify's DSL (see `tasks/SKILL-dsl.md`). A minimal manifest looks like:

```yaml
name: "clean-architecture-go"          # Required, non-empty
version: "1.0.0"                       # Required, SemVer (X.Y.Z)
author: "your-github-username"        # Required
language: "go"                        # Required (go | typescript | rust | csharp | python)
architecture: "clean"                 # Required (clean | vertical-slice | hexagonal | mvc | monorepo)
description: "Clean Architecture in Go with HTTP support"  # Required
tags: ["go", "clean", "api"]          # Required (non-empty slice)

inputs: []                            # Required (can be empty)
files: []                             # Required (can be empty)
steps: []                             # Required (can be empty)
```

#### Inputs

Inputs define variables that Structify will ask the user for:

```yaml
inputs:
  - id: project_name
    prompt: "Project name?"
    type: string              # string | enum | bool
    required: true
    default: "my-project"
    validate: "^[a-z][a-z0-9-]+$"
```

#### Files rules

File rules control which paths from `template/` are included or excluded:

```yaml
files:
  - include: "internal/transport/http/**"
    when: transport == "http"

  - exclude: "internal/db/**"
    when: orm == "none"
```

Rules:

- Each rule must have either `include` **or** `exclude`, but not both.
- Optional `when:` expressions must be valid according to the DSL (see `SKILL-dsl.md`).
- When multiple rules apply, the **last matching rule wins**.

#### Steps

Steps are commands executed after files are generated:

```yaml
steps:
  - name: "Init go module"
    run: "go mod init {{ project_name }}"

  - name: "Tidy"
    run: "go mod tidy"
```

Each step requires:

- `name`: human friendly label.
- `run`: the command to execute (may contain `{{ }}` interpolations).
- Optional `when:` to run conditionally.

---

### Optional fields and recommendations

- `description`: although required, you should make it descriptive and helpful.
- `tags`: used for discovery; prefer short, lowercased tokens (`go`, `rest`, `microservices`).
- Additional inputs, file rules, and steps can be added freely as long as they pass `structify template validate`.

---

### `.tmpl` file naming conventions

Inside `template/`:

- Files with `.tmpl` extension are **rendered**:
  - `main.go.tmpl` → written as `main.go` with all `{{ }}` expressions interpolated.
- Files **without** `.tmpl` are copied as-is:
  - `LICENSE`, `.gitignore`, `Dockerfile`, etc.

Interpolation uses the DSL context:

```go
package main

func main() {
    println("Hello, {{ project_name | pascal_case }}!")
}
```

Rules:

- `{{ variable_id }}` pulls from `inputs[].id`.
- `{{ variable_id | filter }}` applies a filter (`snake_case`, `pascal_case`, `camel_case`, `kebab_case`, `upper`, `lower`).
- Unknown variables or filters are errors.

---

### Testing your template locally

Before publishing, always validate your template locally:

```bash
# From the root of your template repo
structify template validate .

# Or explicitly point to scaffold.yaml
structify template validate scaffold.yaml
```

Examples:

- **Valid template**:

```text
✓ Template is valid
Inputs: 2, Steps: 3, File rules: 4
```

- **Invalid template** (non-zero exit code):

```text
✗ Template is invalid:
- inputs[0].id: must be non-empty
- files[1].when: invalid expression at position 12
```

Use `structify template validate --json` if you need machine-readable output for CI.

---

### Publishing checklist

Before publishing your template repository on GitHub, run:

```bash
structify template publish
```

This command runs a checklist similar to:

```text
[✓] scaffold.yaml exists
[✓] scaffold.yaml is valid
[✓] template/ directory has files
[✗] README.md is missing — add documentation for your template
[✗] version field looks default — consider bumping before publishing
```

- Items marked `[✓]` are good.
- Items marked `[✗]` may be **critical errors** (affect exit code) or **warnings** (recommendations only).

Critical failures (e.g. missing or invalid `scaffold.yaml`, empty `template/`) will cause a non-zero exit code so CI can block publishing until fixed.

---

### Compatibility badge

You can advertise that your template is compatible with Structify by adding this badge to your `README.md`:

```markdown
[![Structify Template](https://img.shields.io/badge/Structify-Template-blue)](https://github.com/jamt29/structify)
```

And optionally include usage instructions:

```markdown
## Using this template with Structify

Install the template:

```bash
structify template add github.com/<your-user>/<your-repo>
```

Create a new project:

```bash
structify new --template <your-repo> --name my-app
```
```

