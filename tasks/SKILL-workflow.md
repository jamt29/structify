# Structify — Workflow operativo (consolidación)

## Regla base por sesión
1. Leer `tasks/todo.md` y `tasks/lessons.md`.
2. Ejecutar tareas en orden, verificando cada una antes de avanzar.
3. Cerrar con build/tests y documento de resultado de la iteración.

## Referencia rápida de archivos clave
- Planificación y seguimiento:
  - `tasks/todo.md`
  - `tasks/lessons.md`
- Skills locales:
  - `tasks/SKILL-workflow.md`
  - `tasks/SKILL-structify.md`
  - `tasks/SKILL-dsl.md`
- Resultados históricos:
  - `tasks/V020-RESULTADO.md`
  - `tasks/V030a-RESULTADO.md`
  - `tasks/V030a-FIX-RESULTADO.md`
  - `tasks/V030b-RESULTADO.md`

## Secuencia recomendada por tarea
1. Identificar causa raíz.
2. Aplicar fix de impacto mínimo.
3. Verificar localmente (unit + integración básica).
4. Actualizar `todo.md` y, si aplica, `lessons.md`.
5. Registrar salida real en archivo `tasks/Vxxx-RESULTADO.md`.

## Checklist de verificación estándar
- `go build ./...`
- `go test ./... -cover`
- Verificación manual de UX cuando aplica (TUI/TTY).

## Lecciones prioritarias (L015–L030)
- L015: validar `inputs[].validate` también en modo flags.
- L016–L018: mantener UX consistente entre paths TUI y no-TTY.
- L019: evitar cortar pantalla con `tea.Quit` entre programas.
- L022–L026: migraciones a Huh sin romper focus/width/sync legacy.
- L027: no capturar modelos obsoletos en transiciones Bubble Tea.
- L028: structured logging sólo cuando el writer lo permite.
- L029: reportes principales redirigibles deben ir a stdout.
- L030: si existe `RunApp` y `Run`, el layout no puede depender solo de RootModel.

## Criterio de cambios
- Preferir parches pequeños y localizados.
- Evitar refactors amplios durante bugfixes.
- No alterar lógica de negocio fuera del scope pedido.

