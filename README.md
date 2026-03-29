# Structify

[![CI](https://github.com/jamt29/structify/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/jamt29/structify/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Version](https://img.shields.io/badge/version-v0.5.1-8a2be2.svg)](https://github.com/jamt29/structify/releases)

> Scaffold opinionated projects in seconds.
> CLI multilenguaje para crear proyectos con arquitecturas
> bien definidas - desde Clean Architecture hasta
> Vertical Slice, en Go, TypeScript y Rust.

## Que es Structify?

Arrancar un proyecto nuevo suele implicar repetir siempre el mismo trabajo: crear carpetas, preparar estructura de capas, cablear un `main`, dejar comandos de build, y agregar archivos base para que todo compile. Ese tiempo no aporta valor directo al producto, pero se repite una y otra vez en cada repo.

Structify resuelve ese problema con templates versionables. En lugar de copiar boilerplates manualmente, eliges un template y generas la base completa con variables (nombre de proyecto, transporte, ORM, etc.), reglas condicionales y pasos post-generacion. Esto permite estandarizar equipos y reducir errores de setup.

El flujo combina modo interactivo TUI (`structify`) y modo no interactivo (`structify new --template ...`) para scripts y CI. Ademas de usar templates built-in, puedes crear los tuyos, importarlos desde proyectos existentes, instalarlos desde GitHub y validarlos con el DSL de `scaffold.yaml`.

## Demo

Si no tienes un GIF, este es el flujo real paso a paso en TUI:

```text
1) Ejecutas: structify
2) Aparece el menu principal (Nuevo proyecto, Mis templates, GitHub, Configuracion)
3) En "Nuevo proyecto" seleccionas un template
4) El formulario pide inputs (string, enum, bool, multiselect, path)
5) Ves preview del arbol de archivos en tiempo real
6) Confirmas y se ejecutan los steps (con spinner + estado)
7) Pantalla final con resumen y siguientes pasos
```

Ejemplo no interactivo:

```bash
structify new --template clean-architecture-go \
  --name my-api --var transport=http
```

## Instalacion

### Go install (recomendado)

```bash
go install github.com/jamt29/structify@latest
```

### Binary releases

Descarga binarios precompilados desde [GitHub Releases](https://github.com/jamt29/structify/releases).

### Homebrew (proximamente)

```bash
brew install structify  # coming soon
```

## Quick Start

```bash
structify
structify new --template clean-architecture-go \
  --name my-api --var transport=http
```

## Templates built-in

| Nombre | Lenguaje | Arquitectura | Descripcion |
|--------|----------|--------------|-------------|
| `clean-architecture-go` | Go | Clean | Clean Architecture para APIs Go con variantes por transporte y persistencia |
| `vertical-slice-go` | Go | Vertical Slice | Vertical Slice en Go con estructura por feature y endpoint de salud |
| `clean-architecture-ts` | TypeScript | Clean | Base Clean en TypeScript para proyectos Node con capas separadas |
| `vertical-slice-ts` | TypeScript | Vertical Slice | Vertical Slice en TypeScript orientado a features |
| `clean-architecture-rust` | Rust | Clean | Base Clean en Rust con estructura de dominio, aplicacion e infraestructura |

## Comandos principales

| Comando | Descripcion |
|---|---|
| `structify` | Lanza el TUI interactivo principal |
| `structify new` | Crea un proyecto desde un template (interactivo o por flags) |
| `structify template list` | Lista templates locales y built-in |
| `structify template add <source>` | Instala un template desde ruta local o GitHub |
| `structify template import <source>` | Crea un template a partir de un proyecto existente |
| `structify template edit <name>` | Abre y valida `scaffold.yaml` de un template local |
| `structify template validate <path>` | Valida estructura y DSL de un template |
| `structify template publish [path]` | Ejecuta checklist de publicacion de template |
| `structify version` | Muestra version, commit y fecha de build |

## Gestion de templates

### Crear tu propio template

1. Genera base minima:
   ```bash
   structify template create
   ```
2. Agrega archivos dentro de `template/` y define variables en `scaffold.yaml`.
3. Valida:
   ```bash
   structify template validate ~/.structify/templates/<tu-template>
   ```
4. Prueba generacion:
   ```bash
   structify new --template <tu-template> --name demo --dry-run
   ```

### Importar desde un proyecto existente

```bash
structify template import ./mi-proyecto
structify template import github.com/user/repo
```

### Instalar desde GitHub

```bash
structify template add github.com/user/repo
```

## Crear un template compatible

Revisa la guia de formato en [docs/template-authoring.md](docs/template-authoring.md) y la referencia formal del DSL.

## Documentacion

- [Referencia de comandos](docs/commands.md)
- [DSL Reference](docs/dsl-reference.md)
- [Template Authoring](docs/template-authoring.md)
- [Template Format](docs/template-format.md)

## Contributing

1. Fork y clone del repo.
2. Crea una rama para tu cambio.
3. Ejecuta:
   ```bash
   go build ./...
   go test ./... -cover
   ```
4. Abre un Pull Request con contexto, motivacion y evidencia de pruebas.

## License

MIT

