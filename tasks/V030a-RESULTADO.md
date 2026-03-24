# V030a-RESULTADO — v0.3.0a Preview en tiempo real + Huh

## 1) Preview del árbol

### Implementado
- Engine:
  - Nuevo `PreviewFiles(req *template.ScaffoldRequest) (*FileTree, error)` en `internal/engine/preview.go`.
  - Reutiliza `ProcessFiles` y `ExecuteSteps` con `DryRun=true` (sin escritura en disco).
  - Estructuras nuevas:
    - `FileTree { Root, Children, Total, Steps }`
    - `TreeNode { Name, IsDir, Children, Skipped }`
- TUI:
  - Nuevo renderer `RenderTree(tree, width, maxLines)` en `internal/tui/tree.go`.
  - Colores:
    - Directorios: `colorPrimary` + bold
    - Archivos: `colorText`
    - Skipped: `colorMuted` + strikethrough
    - Conectores: `colorBorder`
    - Footer: `colorMuted`
  - `stateInputs` usa split layout:
    - `>=120`: 50/50
    - `>=80`: 55/45
    - `<80`: solo panel izquierdo
  - `stateConfirm` ahora muestra también árbol de preview (máx 10 líneas con `(+ N más)`).

### Salida real usada para validación funcional (dry-run)
Comando:
```bash
go run . new --template clean-architecture-go --name my-api --var transport=http --var orm=gorm --dry-run
```

Salida relevante:
```text
Files that would be created:
.gitignore
Makefile
README.md
cmd/main.go
internal/domain/entity/.gitkeep
internal/domain/repository/gorm/repository_impl.go
internal/domain/repository/interface.go
internal/domain/usecase/.gitkeep
internal/transport/http/handler.go
internal/transport/http/router.go
pkg/.gitkeep
Steps that would run:
✓ go mod init github.com/user/my-api
─ go get google.golang.org/grpc  (skipped: transport == "grpc")
✓ go get gorm.io/gorm
─ go get github.com/jmoiron/sqlx  (skipped: orm == "sqlx")
✓ go mod tidy
```

### Ejemplo de render esperado en TUI (stateInputs/stateConfirm)
```text
my-api/
├── cmd/
│   └── main.go
├── internal/
│   ├── domain/
│   │   └── repository/
│   │       └── interface.go
│   └── transport/
│       ├── http/
│       │   ├── handler.go
│       │   └── router.go
│       └── grpc/        (skipped)
│           └── server.go (skipped)
└── go.mod
11 archivos · 3 steps
```

## 2) Huh integrado

### Dependencia
- Instalada: `github.com/charmbracelet/huh v1.0.0`.

### Mapeo de tipos DSL → campos Huh
- `string` / `path` → `huh.NewInput()`
- `enum` → `huh.NewSelect[string]()`
- `bool` → `huh.NewConfirm()`
- `multiselect` → `huh.NewMultiSelect[string]()`

### Implementación
- Nuevo archivo: `internal/tui/huh_inputs.go`.
- API principal:
  - `BuildHuhForm(inputs []dsl.Input, ctx dsl.Context) (*huh.Form, error)`
- Integración real en `App`:
  - `huhForm *huh.Form`
  - mapas `huhString`, `huhBool`, `huhMulti`
  - `stateInputs` delega en `a.huhForm.Update(msg)` y usa `a.huhForm.View()`.

### Tema aplicado
- `structifyHuhTheme()` sobre `huh.ThemeBase()`:
  - título focused con `colorText` + bold
  - base focused con borde redondeado y `colorActive`
  - select/multiselect/prompt alineados con `colorPrimary`
  - blurred con borde oculto y tonos muted

## 3) `when:` condicional

- Estrategia usada: `WithHideFunc` por grupo (`huh.Group`) para cada input.
- Evaluación dinámica:
  - `ShouldAskInput(input, currentCtx)`
  - `currentCtx` se deriva del contexto parcial actual (`buildPartialRequest`).
- Justificación:
  - evita reconstrucción completa del form en cada tecla
  - mantiene UX estable con un único modelo Bubble Tea.

## 4) Decisiones técnicas de integración

- No se ejecuta `huh.NewForm().Run()` en programa separado.
- La integración queda embebida en `App` (modelo unificado):
  - `Update` de `App` delega a `huhForm.Update`
  - transición a `stateConfirm` cuando el contexto completo es válido
  - preview en tiempo real y navegación mantienen un solo ciclo de Bubble Tea.
- Se preservó compatibilidad con tests existentes con `syncLegacyInputsToHuh`.

## 5) Cobertura

Comando:
```bash
go test ./... -cover
```

Salida relevante:
```text
ok  	github.com/jamt29/structify/internal/dsl	coverage: 87.2% of statements
ok  	github.com/jamt29/structify/internal/engine	coverage: 62.9% of statements
ok  	github.com/jamt29/structify/internal/tui	coverage: 33.8% of statements
```

## 6) Estado final

Comandos:
```bash
go build ./...
go test ./... -cover
```

Resultado:
- `go build ./...` => OK
- `go test ./... -cover` => OK

## 7) Lecciones capturadas

Se añadió:
- `L022` en `tasks/lessons.md` sobre compatibilidad de tests al migrar UI a nuevos componentes.
