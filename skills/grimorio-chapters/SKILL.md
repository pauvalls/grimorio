---
name: grimorio-chapters
version: "1.0.0"
description: Generate self-contained chapters with inline NPCs, encounters, and 10-15 WotC-format areas
---

# grimorio-chapters — Chapter Designer

## Template Requerido

**ANTES de generar contenido, LEER el template:**

```
get_template(type="chapter")
```

El template define el formato WotC obligatorio para capítulos auto-contenidos.

## Herramientas Disponibles

**MCP Tools (USAR para guardar contenido):**
- `save_chapter` — Guardar capítulo completo (inline NPCs + encuentros + áreas)
- `validate_canon` — Validar contra canon.json
- `get_template` — Obtener template WotC
- `check_consistency` — Chequeo de consistencia

**NO usar Write para contenido creativo** — El frontmatter del agente ya no incluye Write para forzar el uso de MCP save tools.

## Workflow Obligatorio

```
1. LEER contexto:
   - canon.json (hechos canónicos, entidades)
   - lore.md (tono, conflicto, setting)
   - chapters/chapter_{N-1}.md (si existe, para continuidad)
   - narrative_state.json (estado actual)

2. LEER template:
   - get_template(type="chapter")

3. GENERAR capítulo siguiendo el template:
   - Apertura narrativa (200-400 palabras)
   - NPCs inline (2-5, perfiles condensados 150-300w)
   - Encuentros (2-4, con dificultad y monstruos por nombre)
   - 10-15 áreas por capítulo (one-shot: 8-12)
   - Consecuencias y transición

4. VALIDAR antes de guardar:
   - validate_canon() con entity_references
   - Mínimo 10 áreas, máximo 15
   - Todos los monstruos referenciados por nombre (NO stat blocks inline)

5. GUARDAR solo si validación pasa:
   - save_chapter(campaign, chapter_number, title, content)

6. REPORTAR al architect
```

## Formato WotC Obligatorio

### Regla 1: Capítulo Auto-Contenido

Cada capítulo DEBE ser un único archivo markdown que el DM pueda ejecutar sin abrir otros archivos.

**Estructura obligatoria:**
```markdown
# Capítulo N: Título

## Apertura Narrativa

## NPCs en este Capítulo

## Encuentros

## Áreas

## Consecuencias y Transición
```

### Regla 2: NPCs Inline (Condensados)

- **150-300 palabras** por NPC
- NO incluir stat blocks completos inline
- Referenciar facciones por nombre
- Incluir: apariencia, motivación, relación con PJs, secreto

**Formato:**
```markdown
### Nombre del NPC
*Rol / Alignment*

Descripción condensada del NPC...

**Motivation:** ...
**Secret:** ...
```

### Regla 3: Encuentros Inline

- **2-4 encuentros** por capítulo
- Monstruos referenciados por nombre (ej: "3x Goblin", "1x Ogre")
- NO incluir stat blocks de monstruos
- Incluir: descripción, monstruos, recompensas, tácticas

**Formato:**
```markdown
### Encuentro N: Nombre
*Dificultad: Easy/Medium/Hard/Deadly*

Descripción del encuentro...

**Monstruos:**
- 3x Nombre del Monstruo
- 1x Nombre del Jefe

**Recompensas:**
- 100 XP
- 50 gold
```

### Regla 4: Áreas WotC

Cada área DEBE incluir:
- **Boxed text** (100-600 palabras, read-aloud)
- **DCs** para habilidades clave
- **Criaturas** referenciadas por nombre
- **Treasure** con valor en gp
- **Desarrollo** (qué pasa después)
- **Ganchos** (≥2 por área)

**Formato:**
```markdown
### Área N: Nombre

> Texto para leer en voz alta...

Descripción del área...

**Características:**
- **Nombre:** Descripción (DC XX)

**Treasure:**
- Nombre del item (valor gp)

**Desarrollo:** Qué pasa después...

**Ganchos:**
- Gancho 1
- Gancho 2
```

### Regla 5: Cross-Referencias Válidas

- Todos los monstruos referenciados DEBEN existir en bestiary/bestiary.md
- Todos los NPCs referenciados DEBEN existir en el capítulo o en appendices
- Cero cross-references rotas por construcción

## Notas de Deprecación

Este skill REEMPLAZA a grimorio-areas y grimorio-encounters para nuevas campañas.
Las campañas existentes con `areas/` siguen funcionando (backwards compatibility).
