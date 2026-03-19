# Structify Commands

## Global flags

- `--config string`: config file (default: `$HOME/.structify/config.yaml`)
- `--verbose`: enable verbose output

## Root

### `structify --help`

Show help for any command:

```bash
structify --help
structify new --help
structify template list --help
```

### `structify version`

Print the current version:

```bash
structify version
```

## Project Scaffolding

### `structify new`

Create a new project from a template.

Flags:
- `--template string`: template name or path to use
- `--name string`: name of the project to create
- `--var stringArray`: additional variables in `key=value` (repeatable)
- `--output string`: output directory for the generated project
- `--dry-run`: show what would be generated without writing files

Example (non-interactive dry-run):

```bash
structify new \
  --template clean-architecture-go \
  --name testproject \
  --var module_path=github.com/test/testproject \
  --output /tmp/testproj \
  --dry-run
```

Expected output starts with:

```text
Dry run — no files will be written.
```

## Template Management

### `structify template list [--json]`

List templates (built-in + local).

```bash
structify template list
structify template list --json
```

### `structify template info <name>`

Show detailed information about a template, including `inputs` and `steps`.

```bash
structify template info clean-architecture-go
```

### `structify template validate <path> [--json]`

Validate a template directory or a single `scaffold.yaml`.

```bash
structify template validate .
structify template validate scaffold.yaml --json
```

### `structify template add <source> [--force] [--name <localName>]`

Install a template from a local path or GitHub URL.

Flags:
- `--force`: overwrite existing template with the same name
- `--name`: local name to use for the installed template

```bash
structify template add github.com/<user>/<repo>
structify template add ./path/to/template
```

### `structify template update [name] [--dry-run]`

Update one or all templates installed from GitHub.

```bash
structify template update
structify template update my-template --dry-run
```

### `structify template remove <name> [-y, --yes]`

Remove a local template.

```bash
structify template remove my-template
structify template remove my-template --yes
```

### `structify template create [--output <dir>]`

Start a wizard to create a new template repo skeleton.

```bash
structify template create --output ./my-template
```

### `structify template publish [path]`

Run the publishing checklist.

```bash
structify template publish
structify template publish ./my-template
```

## Shell completion

### `structify completion <shell>`

Generate shell completion script.

