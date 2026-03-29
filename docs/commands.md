# Referencia de comandos

Esta referencia describe el estado actual de comandos de Structify v0.5.1.
Incluye descripcion, flags y ejemplos ejecutables.

## structify (sin argumentos)

Lanza el TUI interactivo principal.

### Flags

- `--config string`: ruta de archivo de configuracion (default: `$HOME/.structify/config.yaml`)
- `--verbose`: habilita salida mas detallada

### Ejemplos

```bash
structify
structify --verbose
structify --config ~/.structify/config.yaml
```

## structify new

Crea un nuevo proyecto a partir de un template instalado o built-in.

En TTY entra al flujo interactivo; en no-TTY se usa modo por flags.

### Flags

- `--template string`: nombre o ruta del template a usar
- `--name string`: nombre del proyecto (`project_name`)
- `--var stringArray`: variables adicionales en formato `key=value` (repetible)
- `--output string`: directorio de salida del proyecto generado
- `--dry-run`: muestra que se generaria sin escribir archivos

### Ejemplos

```bash
structify new
```

```bash
structify new --template clean-architecture-go --name my-api
```

```bash
structify new \
  --template clean-architecture-go \
  --name my-api \
  --var transport=http \
  --var orm=none \
  --dry-run
```

```bash
structify new --template ~/.structify/templates/mi-template --name prueba-local
```

## structify template

Grupo de comandos para administrar templates locales, built-in y remotos.

### Ejemplo

```bash
structify template --help
```

### structify template list

Lista templates disponibles, agrupados por fuente.

#### Flags

- `--json`: imprime salida en JSON

#### Ejemplos

```bash
structify template list
structify template list --json
```

### structify template add

Instala un template desde una ruta local o desde GitHub.

Uso: `structify template add <source>`

`<source>` puede ser:
- ruta local (`./mi-template`)
- URL GitHub corta (`github.com/user/repo`)
- URL GitHub con ref (`github.com/user/repo@v1.2.0`)

#### Flags

- `--force`: sobrescribe template local existente con el mismo nombre
- `--name string`: alias local para guardar el template instalado

#### Ejemplos

```bash
structify template add ./mi-template
```

```bash
structify template add github.com/user/repo
```

```bash
structify template add github.com/user/repo@v1.2.0 --name mi-template
```

### structify template import

Importa un proyecto existente y lo convierte en template Structify.

Uso: `structify template import <source>`

`<source>` puede ser ruta local o repositorio GitHub.

#### Flags

- `--name string`: nombre local del template generado
- `--yes`: omite confirmacion interactiva

#### Ejemplos

```bash
structify template import ./mi-proyecto
```

```bash
structify template import github.com/user/proyecto-real --name mi-template --yes
```

### structify template edit

Abre `scaffold.yaml` de un template local en tu editor y valida al guardar.

Uso: `structify template edit <name>`

#### Flags

Sin flags propios.

#### Ejemplos

```bash
structify template edit mi-template
EDITOR=nano structify template edit mi-template
```

### structify template create

Crea la estructura base de un template nuevo mediante asistente.

#### Flags

- `--output string`: directorio donde crear el template (default: store local)

#### Ejemplos

```bash
structify template create
```

```bash
structify template create --output ./templates-locales
```

### structify template validate

Valida un directorio template o un archivo `scaffold.yaml`.

Uso: `structify template validate <path>`

#### Flags

- `--json`: salida en JSON para automatizaciones

#### Ejemplos

```bash
structify template validate ./mi-template
```

```bash
structify template validate ./mi-template/scaffold.yaml --json
```

### structify template remove

Elimina un template local instalado.

Uso: `structify template remove <name>`

#### Flags

- `-y, --yes`: confirma eliminacion sin prompt

#### Ejemplos

```bash
structify template remove mi-template
```

```bash
structify template remove mi-template --yes
```

### structify template info

Muestra metadata, inputs y steps de un template.

Uso: `structify template info <name>`

#### Flags

Sin flags propios.

#### Ejemplos

```bash
structify template info clean-architecture-go
structify template info mi-template
```

### structify template update

Actualiza uno o todos los templates instalados desde GitHub.

Uso: `structify template update [name]`

#### Flags

- `--dry-run`: muestra que templates se actualizarian sin modificar archivos

#### Ejemplos

```bash
structify template update
```

```bash
structify template update mi-template
```

```bash
structify template update mi-template --dry-run
```

### structify template publish

Ejecuta checklist de publicacion para un template.

Uso: `structify template publish [path]`

#### Flags

Sin flags propios.

#### Ejemplos

```bash
structify template publish
```

```bash
structify template publish ./mi-template
```

### structify version

Muestra version, commit y fecha de build del binario actual.

#### Flags

Sin flags propios.

#### Ejemplos

```bash
structify version
```

## Flags globales

Disponibles en todos los comandos:

- `--config string`: ruta del archivo de configuracion
- `--verbose`: salida de logs mas detallada

## Notas practicas

- `structify` sin argumentos inicia TUI.
- `structify new` en pipelines suele requerir `--template`, `--name` y variables con `--var`.
- `--dry-run` es recomendado para validar resultados antes de generar.
- `template update` solo actua sobre templates instalados desde GitHub (con metadata de origen).

