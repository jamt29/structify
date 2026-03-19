### 1. Correcciones aplicadas
- Corrección 1 — `covdata`: renombrar archivos `.go` a `.go.tmpl` dentro de `templates/*/template/**`.
  - Cambio exacto: renombrado `templates/minimal-go/template/internal/app/app.go` -> `app.go.tmpl`.
  - Antes / Después del comportamiento:
    - Antes: `go test ./... -cover` fallaba con `go: no such tool "covdata"` debido a que los archivos `.go` dentro del árbol de templates eran interpretados por el tooling de coverage como código compilable.
    - Después: los templates se procesan como templates (`*.tmpl`), evitando el intento de compilación/coverage sobre esos archivos; el error `covdata` desaparece.

- Corrección 2 — `scaffold.yaml` de `clean-architecture-go`: condición ORM mezclada en steps.
  - Cambio exacto: en `templates/clean-architecture-go/scaffold.yaml` ajustados los steps:
    - `go get gorm.io/gorm` -> `when: orm == "gorm"`
    - `go get github.com/jmoiron/sqlx` -> `when: orm == "sqlx"`
  - Antes / Después del comportamiento:
    - Antes: los steps de ORM estaban condicionados por lógica mezclada: `transport != "grpc" && orm == ...` (y en `--dry-run` aparecían como “skipped” con esa razón).
    - Después: los steps de ORM dependen únicamente de `orm` (en `--dry-run` las razones de “skipped” muestran `orm == "gorm"` / `orm == "sqlx"`).

- Corrección 3 — TUI: separar lógica de negocio del render (Bubbletea).
  - Cambio exacto:
    - Añadido `internal/tui/logic.go` con funciones puras:
      - `ShouldAskInput`
      - `ApplyDefault`
      - `ValidateInputValue`
      - `BuildContext`
    - Refactor en `internal/tui/inputs.go` para delegar la construcción/evaluación de contexto a `BuildContext` y usar las funciones puras en lugar de lógica inline.
    - Tests:
      - `internal/tui/logic_test.go` (table-driven para las 4 funciones puras + tests de `RunInputsWithInitial` sin Bubbletea).
      - `internal/tui/models_coverage_test.go` y `internal/tui/prompt_models_coverage_test.go` para cubrir modelos TUI unitariamente.
  - Antes / Después del comportamiento:
    - Antes: `internal/tui` tenía ~7% de cobertura porque la lógica no era testeable unitariamente.
    - Después: `internal/tui` subió a `44.2%` de cobertura (y `go test ./... -cover` pasa sin errores).

### 2. Cobertura final
Resultado de `go test ./... -cover`:

```bash
ok  	github.com/jamt29/structify	(cached)	coverage: 100.0% of statements
ok  	github.com/jamt29/structify/cmd	(cached)	coverage: 66.4% of statements
ok  	github.com/jamt29/structify/cmd/structify	(cached)	coverage: 100.0% of statements
ok  	github.com/jamt29/structify/cmd/template	(cached)	coverage: 67.3% of statements
ok  	github.com/jamt29/structify/internal/config	(cached)	coverage: 81.8% of statements
ok  	github.com/jamt29/structify/internal/dsl	(cached)	coverage: 87.7% of statements
ok  	github.com/jamt29/structify/internal/engine	(cached)	coverage: 73.6% of statements
ok  	github.com/jamt29/structify/internal/template	(cached)	coverage: 73.7% of statements
ok  	github.com/jamt29/structify/internal/tui	(cached)	coverage: 44.2% of statements
```

### 3. Dry-run de los cinco built-ins
Output completo de cada dry-run.

```bash
=== clean-architecture-go ===
Dry run — no files will be written.
Template : clean-architecture-go
Output   : ./testproject
Variables: project_name=testproject, module_path=github.com/user/testproject, orm=none, transport=http
Files that would be created:
.gitignore
Makefile
README.md
cmd/main.go
internal/domain/entity/.gitkeep
internal/domain/repository/interface.go
internal/domain/repository/none/repository_impl.go
internal/domain/usecase/.gitkeep
internal/transport/http/handler.go
internal/transport/http/router.go
pkg/.gitkeep
Steps that would run:
✓ go mod init github.com/user/testproject
─ go get google.golang.org/grpc  (skipped: transport == "grpc")
─ go get gorm.io/gorm  (skipped: orm == "gorm")
─ go get github.com/jmoiron/sqlx  (skipped: orm == "sqlx")
✓ go mod tidy
No files were written.
=== vertical-slice-go ===
Dry run — no files will be written.
Template : vertical-slice-go
Output   : ./testproject
Variables: project_name=testproject, include_tests=true, module_path=github.com/user/testproject
Files that would be created:
.gitignore
Makefile
README.md
cmd/main.go
config/config.go
features/health/handler.go
features/health/handler_test.go
shared/middleware/.gitkeep
Steps that would run:
✓ go mod init github.com/user/testproject
✓ go mod tidy
No files were written.
=== clean-architecture-ts ===
Dry run — no files will be written.
Template : clean-architecture-ts
Output   : ./testproject
Variables: project_name=testproject, runtime=express, use_prisma=false
Files that would be created:
.gitignore
README.md
package.json
src/application/use-cases/.gitkeep
src/domain/entities/.gitkeep
src/domain/repositories/IExampleRepository.ts
src/http/routes.ts
src/http/server.ts
src/index.ts
src/infrastructure/repositories/.gitkeep
tsconfig.json
Steps that would run:
✓ npm init -y
✓ npm install express
✓ npm install -D typescript @types/node
─ npm install prisma @prisma/client  (skipped: use_prisma == true)
No files were written.
=== vertical-slice-ts ===
Dry run — no files will be written.
Template : vertical-slice-ts
Output   : ./testproject
Variables: project_name=testproject, runtime=express
Files that would be created:
.gitignore
README.md
package.json
src/features/health/health.handler.ts
src/features/health/health.routes.ts
src/index.ts
tsconfig.json
Steps that would run:
✓ npm install express
✓ npm install -D typescript @types/node
No files were written.
=== clean-architecture-rust ===
Dry run — no files will be written.
Template : clean-architecture-rust
Output   : ./testproject
Variables: project_name=testproject, transport=axum
Files that would be created:
.gitignore
Cargo.toml
README.md
src/application/mod.rs
src/domain/mod.rs
src/infrastructure/mod.rs
src/main.rs
Steps that would run:
✓ cargo build
No files were written.
```

Además (verificación extra `clean-architecture-go`):

```bash
=== clean-architecture-go orm=gorm ===
Dry run — no files will be written.
Template : clean-architecture-go
Output   : ./testproject
Variables: project_name=testproject, module_path=github.com/user/testproject, orm=gorm, transport=http
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
✓ go mod init github.com/user/testproject
─ go get google.golang.org/grpc  (skipped: transport == "grpc")
✓ go get gorm.io/gorm
─ go get github.com/jmoiron/sqlx  (skipped: orm == "sqlx")
✓ go mod tidy
No files were written.
=== clean-architecture-go orm=none ===
Dry run — no files will be written.
Template : clean-architecture-go
Output   : ./testproject
Variables: project_name=testproject, module_path=github.com/user/testproject, orm=none, transport=http
Files that would be created:
.gitignore
Makefile
README.md
cmd/main.go
internal/domain/entity/.gitkeep
internal/domain/repository/interface.go
internal/domain/repository/none/repository_impl.go
internal/domain/usecase/.gitkeep
internal/transport/http/handler.go
internal/transport/http/router.go
pkg/.gitkeep
Steps that would run:
✓ go mod init github.com/user/testproject
─ go get google.golang.org/grpc  (skipped: transport == "grpc")
─ go get gorm.io/gorm  (skipped: orm == "gorm")
─ go get github.com/jmoiron/sqlx  (skipped: orm == "sqlx")
✓ go mod tidy
No files were written.
```

### 4. Checklist de release v0.1.0
□ go build ./... → PASS
□ go test ./... -cover → PASS sin covdata error
□ internal/tui cobertura >= 40%
□ Dry-run clean-architecture-go con orm=gorm → steps correctos
□ Dry-run clean-architecture-go con orm=none → steps skipped
□ Dry-run vertical-slice-go → sin errores
□ Dry-run clean-architecture-ts → sin errores
□ Dry-run vertical-slice-ts → sin errores
□ Dry-run clean-architecture-rust → sin errores
□ .goreleaser.yaml presente
□ .github/workflows/ci.yml presente
□ .github/workflows/release.yml presente
□ README.md completo
□ docs/commands.md presente
□ docs/dsl-reference.md presente
□ docs/template-format.md presente

### 5. Comando de release
Los comandos exactos para hacer el primer release:
  git tag v0.1.0
  git push origin v0.1.0

### 6. Lecciones capturadas en esta limpieza
- `go test ./... -cover` cubre también paquetes dentro de `templates/` cuando existen archivos `.go` allí; por eso renombrarlos a `*.go.tmpl` evita errores de tooling (`covdata`) por confusión entre templates y código compilable.
- Para recompilar el binario del CLI en este repo, `go build -o ./bin/structify ./` no sirve si el paquete raíz no es `main`; en su lugar usar `go build -o ./bin/structify ./cmd/structify` (el binario correcto es `cmd/structify/main.go`).

