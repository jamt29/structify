# V030a-FIX-RESULTADO — Fix de bugs reportados

## 1) Causa raíz exacta

### Bug 1 — Árbol desbordado en primer render
- **Archivo:** `internal/tui/root.go`
- **Problema real:** el `App` de `screenNew` podía crearse después de recibir el `WindowSizeMsg` inicial del programa, por lo que arrancaba con tamaño default y no con el tamaño real de terminal.
- **Líneas implicadas (antes del fix):**
  - creación de `App` en transición de menú (`ActionNew`) sin propagar size actual al sub-modelo.
  - transición de `Templates -> New` (`transitionToNew`) también sin propagar size actual.
- **Efecto observado:** primer frame en `stateInputs` calculaba split con dimensiones no reales; al redimensionar, se corregía (porque recién ahí llegaba otro `WindowSizeMsg`).

### Bug 2 — Input de Huh no respondía teclado (Init)
- **Archivo:** `internal/tui/app.go`
- **Problema real:** el formulario Huh se construía en `prepareInputs()` pero no se ejecutaba su `Init()` al entrar a `stateInputs`.
- **Evidencia:**
  - `prepareInputs()` creaba `a.huhForm`.
  - `updateSelect()` cambiaba a `stateInputs` y retornaba `nil` (sin `a.huhForm.Init()`).
  - `App.Init()` devolvía `nil` siempre.
- **Efecto observado:** `Project name?` renderizado pero sin estado de entrada activo (parecía congelado).

### Bug 3 — Nombre de proyecto / primer campo sin foco o sin escritura
- **Archivo:** `internal/tui/app.go`
- **Problema real:** aun con `huhForm.Init()`, `tea.WindowSizeMsg` se resolvía solo en `App` (resize de listas legacy) y **no** se reenviaba a `huh.Form`. Con `f.width == 0`, Huh dimensiona grupos en el handler de `WindowSizeMsg`; al no recibirlo, los inputs podían quedar mal dimensionados (p. ej. ancho 0) y parecer no seleccionables.
- **Complemento:** en vista partida, el formulario se dibuja en el panel izquierdo; se aplica `WithWidth` acorde a `renderInputsSplit` para alinear layout interno con el panel visible.

## 2) Fix aplicado (mínimo impacto)

### Bug 1
- `internal/tui/root.go`:
  - `NewRootModel`: fallback seguro inicial a `width=80`, `height=24`.
  - `Init()`: se añadió comando `initialWindowSizeCmd()` con `term.GetSize(...)` para emitir `tea.WindowSizeMsg` temprano.
  - Al crear `App` en transiciones:
    - `ActionNew`
    - `Templates -> New`
    - se inyecta explícitamente `tea.WindowSizeMsg{Width:r.width, Height:r.height}` al `App` recién creado.

### Bug 2
- `internal/tui/app.go`:
  - `App.Init()` ahora retorna `a.huhForm.Init()` cuando `stateInputs` está activo.
  - Al entrar desde selector (`updateSelect` con Enter) se retorna `a.enterStateInputs()` para inicializar el form.
  - Al volver desde confirm (`updateConfirm` con Esc/B) también se llama `a.enterStateInputs()`.
  - Se mantuvo compatibilidad mínima con tests legacy (`syncLegacyInputsToHuh` en Enter), sin refactor adicional.

### Bug 3
- `internal/tui/app.go`:
  - En el branch global de `tea.WindowSizeMsg`, si `state == stateInputs` y `huhForm != nil`: `applyHuhFormWidth()`, luego `huhForm.Update(msg)` y `syncFromHuhForm()`.
  - `inputsFormWidth()` + `applyHuhFormWidth()` tras crear el form en `prepareInputs`.

### Bug 4 — Foco del input parpadea / no se puede escribir el nombre
- **Causa:** `syncLegacyInputsToHuh()` corría antes de cada `huhForm.Update` y copiaba `textinput` legacy (siempre vacío mientras se escribe en Huh) sobre `huhString`, borrando el valor; además marcaba “cambio” y **reconstruía** el formulario, reseteando foco.
- **Fix:** Eliminar el rebuild en bucle; con `huhForm != nil` no pisar strings con legacy vacío ni sincronizar bool/multiselect desde legacy; en Enter ejecutar `syncFromHuhForm()` antes de fusionar legacy para tests (`ti.SetValue`).

## 3) Verificación

### Build y tests
Comandos ejecutados:
```bash
go build ./...
go test ./... -cover
```

Resultado:
```text
go build ./...                         => OK
go test ./... -cover                  => OK
internal/dsl coverage                 => 87.2%
internal/engine coverage              => 62.9%
internal/tui coverage                 => 33.6%
```

### Checklist de bugs
- [x] Build compila sin errores.
- [x] Tests existentes pasan (`go test ./... -cover`).
- [x] `stateInputs` inicializa `huhForm` vía `Init()` al entrar.
- [x] `ESC` en `stateInputs` sigue volviendo al selector.
- [x] Se propaga tamaño actual a `App` al crear pantalla `new`.
- [ ] Verificación visual manual en terminal real pendiente del usuario:
  - árbol completo desde primer frame sin redimensionar
  - escritura en `Project name?` visible
  - Enter avanza entre campos/confirm
  - actualización del árbol en cambios de inputs

## 4) Lecciones capturadas

Se añadieron en `tasks/lessons.md`:
- `L023`: fallback de tamaño seguro + propagación explícita del `WindowSizeMsg` a pantallas creadas tardíamente.
- `L025`: reenviar `WindowSizeMsg` al `huh.Form` embebido y alinear ancho con panel en split.
- `L026`: evitar sync legacy→Huh que borra texto y reconstruye el form en cada tick.
