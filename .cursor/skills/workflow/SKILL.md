---
name: structify-workflow
description: Metodología de trabajo obligatoria para este proyecto. Leer al inicio de CADA sesión. Define cómo planificar, ejecutar, verificar y documentar. No opcional.
---

# Structify — Metodología de Trabajo

## Regla de Oro
**Leer tasks/todo.md y tasks/lessons.md al inicio de cada sesión.**
No asumir el estado del proyecto. Verificar siempre.

---

## 1. Orquestación del Flujo de Trabajo

### Modo Planificación (obligatorio para tareas no triviales)
- Entrar en modo planificación para CUALQUIER tarea de 3+ pasos o con decisiones arquitectónicas
- Escribir el plan en `tasks/todo.md` con elementos marcables antes de ejecutar
- Si algo sale mal → PARAR y replanificar de inmediato, no seguir forzando
- Usar planificación también para pasos de verificación, no solo construcción
- Escribir especificaciones detalladas por adelantado para reducir ambigüedad

### Estrategia de Subagentes
- Usar subagentes para mantener limpia la ventana de contexto principal
- Delegar investigación, exploración y análisis paralelo a subagentes
- Para problemas complejos → asignar más cómputo vía subagentes
- Un enfoque por subagente para ejecución enfocada

---

## 2. Gestión de Tareas

### Flujo estándar por tarea
```
1. Leer todo.md → identificar siguiente tarea pendiente
2. Leer lessons.md → verificar lecciones aplicables
3. Escribir plan de implementación (si la tarea es no trivial)
4. Implementar
5. Verificar que funciona (tests, compilación, comportamiento)
6. Marcar tarea como completada en todo.md
7. Si hubo corrección → actualizar lessons.md
```

### Formato de todo.md
```markdown
- [x] Tarea completada
- [ ] Tarea pendiente
- [~] Tarea en progreso
```

### Reglas de documentación
- Explicar cambios con resumen de alto nivel en cada paso
- Añadir sección de revisión en todo.md al completar una fase
- Capturar lecciones en lessons.md después de CUALQUIER corrección

---

## 3. Ciclo de Automejora

Después de CUALQUIER corrección del usuario:
1. Identificar el patrón del error (no solo el síntoma)
2. Escribir una lección en `tasks/lessons.md` con el formato:
```markdown
### LXXX — Título corto descriptivo
- **Contexto:** Cuándo/por qué surgió el error
- **Lección:** Regla concreta a seguir para evitarlo
- **Aplicar en:** Fase o archivo específico
```
3. Revisar si hay tareas futuras que podrían repetir el mismo error
4. Ajustar el plan si es necesario

---

## 4. Verificación Antes de Finalizar

**Nunca marcar una tarea como completada sin demostrar que funciona.**

Checklist de verificación:
```
□ ¿Compila sin errores? (go build ./...)
□ ¿Pasan los tests? (go test ./...)
□ ¿El comportamiento es el esperado?
□ ¿Aprobaría esto un engineer senior?
□ ¿Se cubrieron los casos edge documentados en SKILL-dsl.md?
```

Para cambios en el DSL específicamente:
```
□ ¿Todos los casos de test de la tabla en SKILL-dsl.md pasan?
□ ¿Los mensajes de error son descriptivos y accionables?
□ ¿Se manejan los casos edge listados?
```

---

## 5. Exigir Elegancia (Equilibrada)

Para cambios no triviales → pausar y preguntar: *"¿hay una forma más elegante?"*

Si un arreglo se siente apresurado:
> "Sabiendo todo lo que sé ahora, ¿cuál es la solución elegante?"

**Omitir esto para arreglos simples y obvios.** No sobre-diseñar.

Señales de que algo NO es elegante:
- Función de más de 50 líneas sin justificación
- Más de 3 niveles de indentación
- Lógica duplicada
- Nombre de variable o función que no dice lo que hace
- Comentario que explica código confuso en vez de simplificar el código

---

## 6. Corrección Autónoma de Errores

Cuando hay un error: **simplemente arreglarlo.** No pedir guía paso a paso.

Proceso:
1. Leer el error completo
2. Identificar la causa raíz (no el síntoma)
3. Revisar el archivo afectado
4. Aplicar el fix mínimo necesario
5. Verificar que el fix no rompe otra cosa
6. Documentar la lección si aplica

**Cero cambio de contexto por parte del usuario para bugs.**

---

## 7. Principios Fundamentales

### Simplicidad Primero
Hacer que cada cambio sea lo más simple posible.
Impactar el código mínimo necesario.

### Sin Pereza
Encontrar causas raíz. Sin arreglos temporales.
Estándares de desarrollador senior.

### Impacto Mínimo
Los cambios solo deben tocar lo necesario.
Evitar introducir errores en código que funcionaba.

---

## 8. Orden de Fases — No Negociable

```
F1 (Fundación) → F2 (DSL) → F3 (Engine) → F4 (new cmd) → F5 (templates local) → F6 (GitHub) → F7 (built-ins) → F8 (distribución)
```

**F2 es el corazón.** Todo depende del DSL.
No empezar F3 sin F2 con tests pasando.
No empezar F4 sin F3 con rollback implementado.

---

## 9. Convenciones Específicas del Proyecto

### Go
- Errores: siempre wrappear → `fmt.Errorf("contexto: %w", err)`
- No usar `panic()` fuera de `main()`
- Tests: tabla de casos (table-driven) como estándar
- Archivos: `snake_case.go`
- Packages: cortos, sin underscores (`dsl`, `engine`, `template`)

### Git
- Commits atómicos: un commit por tarea completada
- Mensaje: `feat(dsl): implement lexer for when expressions`
- Formato: `tipo(scope): descripción` en minúsculas

### Archivos de Skills
- Los archivos SKILL-*.md son **documentación viva**, actualizarlos si cambia el diseño
- Si el DSL cambia de spec → actualizar SKILL-dsl.md antes o junto al código
- Si la arquitectura de carpetas cambia → actualizar SKILL-structify.md

---

## 10. Referencia Rápida de Archivos Clave

| Archivo | Propósito | Leer cuando... |
|---|---|---|
| `tasks/todo.md` | Plan maestro + estado | Al inicio de cada sesión |
| `tasks/lessons.md` | Lecciones aprendidas | Al inicio + después de errores |
| `tasks/SKILL-structify.md` | Arquitectura completa del proyecto | Al tocar cmd/, internal/, templates/ |
| `tasks/SKILL-dsl.md` | Spec del DSL | Al tocar internal/dsl/ |
| `tasks/SKILL-workflow.md` | Este archivo — metodología | Al inicio de cada sesión |
