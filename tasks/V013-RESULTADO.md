## V013-RESULTADO (polish visual + UX del TUI unificado)

### 1) Bugs corregidos

#### Bug 1 — `"<nil>"` en placeholder de `textinput`
- **Causa raíz:** en `internal/tui/app.go`, el valor inicial se cargaba con `fmt.Sprint(a.answers[id])`; cuando no existía respuesta previa, el valor era `nil` y se renderizaba como `"<nil>"`.
- **Corrección aplicada:**
  - se agregó `answerString(ctx, key)` para convertir `nil` a `""`.
  - `prepareInputs` usa `answerString` para `SetValue`.
  - placeholder string ahora es explícito: `default` si existe, `""` si no.
- **Resultado:** no vuelve a aparecer `"<nil>"` en `stateInputs`.

#### Bug 2 — `bool` como `textinput`
- **Causa raíz:** en `prepareInputs`, los inputs `bool` usaban `textinput` igual que `string`.
- **Corrección aplicada:**
  - `inputEntry` incorpora `boolVal bool`.
  - `updateInputs` maneja `left/right/y/n` para alternar estado.
  - `entryValue` para bool retorna `"true"`/`"false"` sin usar `textinput`.
  - `renderInputBlock` dibuja toggle visual tipo botones.
- **Resultado:** los bool son toggle visual, no campo libre.

---

### 2) Paleta de colores (lipgloss)

Archivo nuevo: `internal/tui/styles.go`

- `#C678DD` (`colorPrimary`): marca principal (`structify` en header)
- `#98C379` (`colorSuccess`): checks, estado exitoso
- `#5C6370` (`colorMuted`): labels secundarios, pendientes, barra de ayuda
- `#3E4451` (`colorBorder`): base de cajas/toggles
- `#ABB2BF` (`colorText`): texto principal de valores
- `#61AFEF` (`colorActive`): borde/input activo/spinner
- `#E06C75` (`colorError`): errores

---

### 3) Checklist visual

Notas:
- Este entorno no expone TTY interactiva directa para recorrer visualmente `go run . new` de punta a punta.
- El checklist se marca según implementación efectiva del render y flujo en `internal/tui/app.go`.

- [x] Header muestra nombre del template después de seleccionarlo
- [x] Input activo tiene borde azul redondeado
- [x] Inputs completados muestran valor + checkmark verde
- [x] Inputs pendientes en gris sin borde
- [x] Nunca aparece `"<nil>"` en ningún input
- [x] Input bool es toggle visual (no textinput)
- [x] Estado confirm tiene borde redondeado con variables alineadas
- [x] Estado progress muestra comandos reales (no nombres de step)
- [x] Estado done tiene sección "Próximos pasos" con borde izquierdo
- [x] ESC en estados navegables vuelve al anterior sin crashear (lógica preservada)

Descripción de estados tras polish:
- **Inputs:** bloques por input con estado `activo/completado/pendiente`; enum con marcador `>`, bool con `[ No ] / [ Yes ]`.
- **Confirm:** caja única con borde redondeado, pares clave/valor alineados.
- **Progress:** spinner azul, checks verdes, skipped gris, error rojo; comandos reales en cada línea.
- **Done:** título verde en negrita, steps con marcas de estado y bloque de “Próximos pasos” con borde izquierdo.

---

### 4) Cobertura (`go test ./... -cover`)

```text
ok  	github.com/jamt29/structify	(cached)	coverage: 0.0% of statements
ok  	github.com/jamt29/structify/cmd	(cached)	coverage: 69.3% of statements
ok  	github.com/jamt29/structify/cmd/structify	(cached)	coverage: 100.0% of statements
ok  	github.com/jamt29/structify/cmd/template	(cached)	coverage: 62.3% of statements
ok  	github.com/jamt29/structify/internal/config	(cached)	coverage: 81.8% of statements
ok  	github.com/jamt29/structify/internal/dsl	(cached)	coverage: 87.7% of statements
ok  	github.com/jamt29/structify/internal/engine	(cached)	coverage: 74.1% of statements
ok  	github.com/jamt29/structify/internal/template	(cached)	coverage: 73.7% of statements
ok  	github.com/jamt29/structify/internal/tui	(cached)	coverage: 46.8% of statements
ok  	github.com/jamt29/structify/templates	(cached)	coverage: 100.0% of statements
```

La cobertura de `internal/tui` subió de 34.0% (V012) a 46.8%.

---

### 5) Estado final

- `go build ./...` -> OK
- `go test ./...` -> OK
- `go test ./... -cover` -> OK

---

### 6) Lecciones capturadas

Se agregó en `tasks/lessons.md`:
- **L017 — Nunca serializar valores nil a textinput**
  - evitar `fmt.Sprint(nil)` para valores de UI
  - usar helper que normalice a string vacío antes de `SetValue`.
