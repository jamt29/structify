# Como crear templates para Structify

## Introduccion

Un template de Structify es un proyecto base parametrizable:

- define variables de entrada en `scaffold.yaml`
- incluye archivos plantilla bajo `template/`
- ejecuta pasos opcionales despues de generar

Te conviene crear tu propio template cuando:

- repites la misma estructura de carpetas en varios repos
- quieres estandarizar convenciones de equipo
- necesitas bootstrap rapido para nuevos servicios
- quieres compartir un starter interno o publico en GitHub

## Opcion 1: Crear desde cero (structify template create)

Comando:

```bash
structify template create
```

Que hace:

1. Abre un asistente para pedir nombre, descripcion, lenguaje y arquitectura.
2. Genera un directorio con estructura minima.
3. Crea un `scaffold.yaml` base listo para editar.
4. Crea carpeta `template/` donde pondras los archivos fuente.

Tambien puedes elegir carpeta de salida:

```bash
structify template create --output ./templates-locales
```

Despues de crearlo:

1. entra al directorio
2. agrega tus archivos dentro de `template/`
3. edita `scaffold.yaml` para declarar variables
4. valida con `structify template validate`

## Opcion 2: Importar un proyecto existente

Si ya tienes un repo funcionando, puedes partir desde ahi:

```bash
structify template import ./mi-proyecto
```

Tambien acepta GitHub:

```bash
structify template import github.com/user/proyecto-real
```

Que detecta automaticamente:

- lenguaje principal del proyecto
- posibles variables frecuentes (por ejemplo nombre/modulo)
- archivos comunes para incluir o ignorar

Recomendacion: tras importar, abre y revisa `scaffold.yaml` para limpiar defaults y mejorar prompts.

## Estructura de un template

Estructura tipica:

```text
mi-template/
├── scaffold.yaml
├── template/
│   ├── README.md.tmpl
│   ├── cmd/
│   │   └── main.go.tmpl
│   ├── internal/
│   │   └── ...
│   └── .gitignore
└── (opcional) README.md
```

Que hace cada parte:

- `scaffold.yaml`: metadata, inputs, reglas de archivos y steps.
- `template/`: archivos que se copiaran/renderizaran al proyecto final.
- `.tmpl`: archivos que usan interpolacion `{{ }}`.
- archivos sin `.tmpl`: se copian tal cual.

## scaffold.yaml explicado

Ejemplo completo y practico:

```yaml
name: "api-go-base"
version: "1.0.0"
author: "acme-team"
language: "go"
architecture: "clean"
description: "Template base para APIs Go"
tags: ["go", "api", "clean"]

inputs:
  - id: project_name
    prompt: "Nombre del proyecto?"
    type: string
    required: true
    default: "my-api"
    validate: "^[a-z][a-z0-9-]+$"

  - id: transport
    prompt: "Tipo de transporte"
    type: enum
    options: [http, grpc]
    default: http

  - id: features
    prompt: "Features opcionales"
    type: multiselect
    options: [auth, metrics]
    default: [auth]

computed:
  - id: module_path
    value: "github.com/acme/{{ project_name | kebab_case }}"

files:
  - include: "internal/transport/http/**"
    when: transport == "http"
  - include: "internal/transport/grpc/**"
    when: transport == "grpc"
  - include: "internal/features/auth/**"
    when: contains(features, "auth")

steps:
  - name: "Init module"
    run: "go mod init {{ module_path }}"
  - name: "Tidy"
    run: "go mod tidy"
```

Lectura rapida:

- `inputs`: preguntas que vera el usuario.
- `computed`: valores derivados, utiles para no pedir datos repetidos.
- `files`: que carpetas entran o no segun respuestas.
- `steps`: comandos finales de setup.

## Anadir archivos al template

Regla simple:

- si un archivo tiene variables -> usa `.tmpl`
- si no tiene variables -> dejalo sin `.tmpl`

Ejemplo `template/README.md.tmpl`:

```md
# {{ project_name | pascal_case }}

Proyecto generado con Structify.
```

Ejemplo `template/.gitignore` (sin templating):

```text
bin/
dist/
.env
```

## Variables y condiciones

Uso de variables:

- `{{ project_name }}`
- `{{ project_name | kebab_case }}`
- `{{ module_path }}`

Condiciones con `when:`:

```yaml
when: transport == "http"
when: transport == "grpc"
when: contains(features, "auth")
```

Casos practicos:

- mostrar archivos HTTP solo si eligieron HTTP
- incluir carpeta de autenticacion solo si seleccionaron `auth`
- ejecutar `go get` solo para ORM elegido

## Probar tu template

Valida estructura y DSL:

```bash
structify template validate ./mi-template
```

Prueba generacion sin escribir archivos:

```bash
structify new --template mi-template --name prueba --dry-run
```

Prueba generacion real:

```bash
structify new --template mi-template --name prueba-real --output /tmp/prueba-real
```

Checklist sugerido:

1. `validate` sin errores
2. `dry-run` muestra archivos y steps esperados
3. proyecto generado compila/arranca
4. casos condicionales (`when`) funcionan

## Publicar en GitHub

Flujo recomendado:

1. Subir template a un repo publico.
2. Ejecutar checklist local:

   ```bash
   structify template publish ./mi-template
   ```

3. Compartir URL para instalacion:

   ```bash
   structify template add github.com/user/mi-template
   ```

Opcionalmente puedes fijar version/ref:

```bash
structify template add github.com/user/mi-template@v1.0.0
```

## Ejemplos completos

Mini ejemplo real (3-4 inputs) para una API TypeScript:

```yaml
name: "api-ts-slice"
version: "1.0.0"
author: "acme-team"
language: "typescript"
architecture: "vertical-slice"
description: "Vertical Slice TS para APIs"
tags: ["typescript", "api"]

inputs:
  - id: project_name
    prompt: "Nombre del proyecto?"
    type: string
    required: true
    default: "my-service"

  - id: runtime
    prompt: "Runtime HTTP"
    type: enum
    options: [express, fastify]
    default: express

  - id: features
    prompt: "Features"
    type: multiselect
    options: [auth, metrics, tracing]
    default: [metrics]

  - id: output_path
    prompt: "Ruta de salida"
    type: path
    must_exist: false

files:
  - include: "src/features/auth/**"
    when: contains(features, "auth")

steps:
  - name: "Install deps"
    run: "npm install"
```

Con esto ya tienes una base util, facil de versionar y facil de compartir.
