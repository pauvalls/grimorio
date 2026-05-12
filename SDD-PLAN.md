# SDD Plan: Grimorio Refactor — Unificación, Skills e Integrator

> Plan para ejecutar como cambios SDD. Cada sección con `## Change:` es un cambio candidate a ejecutar via `/sdd-new`.

---

## Diagnóstico (pre-SDD)

### 16 agentes definidos — 1 zombie

| Agente | Estado |
|--------|--------|
| `grimorio-architect` | ✅ Orquestador, 919 líneas — la fuente más completa |
| `grimorio-artist` | ✅ Usado en Fase 6-8 |
| `grimorio-cartographer` | ✅ Batch 3 |
| `grimorio-lore` | ✅ Batch 2 |
| `grimorio-npc` | ✅ Batch 1 |
| `grimorio-bestiary` | ✅ Batch 1 |
| `grimorio-encounters` | ✅ Batch 2 |
| `grimorio-areas` | ✅ Batch 3 |
| `grimorio-quests` | ✅ Batch 2 |
| `grimorio-maps` | ✅ Batch 1 |
| `grimorio-characters` | ✅ Batch 2 |
| `grimorio-narrative-custodian` | ✅ Fase 11 |
| `grimorio-introduction` | ✅ Fase 3a |
| `grimorio-setting-guide` | ✅ Batch 2 |
| `grimorio-appendices` | ✅ Fase 5e |
| **`grimorio-integrator`** | ❌ **Definido pero NUNCA invocado** |

### Fuente de verdad duplicada

- **`agents/grimorio-architect.md`** — 919 líneas, workflow completo, prompts de delegación, batches, gates
- **`opencode.json → command.grimorio.template`** — versión truncada (~80 líneas efectivas), sin nombres de agentes, sin WotC gate, sin introduction/appendices explícitos, truncada a 2000 chars

### Skills: 10 directorios, 9 vacíos

```bash
skills/
├── dnd-5e-srd/SKILL.md           # 431 líneas — el único con contenido real
├── grimorio-appendices/           # VACÍO
├── grimorio-areas/                # VACÍO
├── grimorio-bestiary/             # VACÍO
├── grimorio-encounters/           # VACÍO
├── grimorio-introduction/         # VACÍO
├── grimorio-lore/                 # VACÍO
├── grimorio-maps/                 # VACÍO
├── grimorio-npc/                  # VACÍO
└── grimorio-setting-guide/        # VACÍO
```

Las instrucciones reales de cada agente están en `agents/grimorio-*.md`, NO en los skills.

### Write vs MCP tools

| Agente | Write? | MCP save_ tool | Problema |
|--------|--------|----------------|----------|
| `grimorio-areas` | ❌ Removido | `save_areas` | ✅ |
| `grimorio-quests` | ❌ Removido | `create_personal_quest` | ✅ |
| `grimorio-characters` | ❌ Removido | `generate_character` | ✅ |
| `grimorio-maps` | ❌ Removido | `save_maps` | ✅ |
| `grimorio-lore` | ❌ Removido | `save_lore` | ✅ |
| `grimorio-encounters` | ❌ Removido | `save_encounters` | ✅ |
| `grimorio-bestiary` | ❌ Removido | `save_bestiary` | ✅ |
| `grimorio-npc` | ❌ Removido | `save_npcs` | ✅ |
| `grimorio-appendices` | ✅ TIENE | `save_appendices` | ❌ Debería remover Write |
| `grimorio-setting-guide` | ✅ TIENE | `save_setting_guide` | ❌ Debería remover Write |
| `grimorio-introduction` | ✅ TIENE | `save_introduction` | ❌ Debería remover Write |
| `grimorio-narrative-custodian` | ✅ TIENE | validación tools | ✅ Lógico (reportes) |
| `grimorio-integrator` | ✅ TIENE | `save_areas/npcs/encounters` | ✅ Necesario (auto-fix) |
| `grimorio-artist` | ✅ TIENE | `generate_image/map/divider` | ✅ Necesario (batch-spec.json) |
| `grimorio-cartographer` | ✅ TIENE | `generate_map/divider/flowchart` | ✅ Necesario (SVG output) |

### WotC validation: solo al final

Actualmente `validate-campaign.sh` corre solo en Phase X (antes del PDF). Los sub-agentes NO validan inline antes de devolver.

---

## Changes propuestos

---

## Change: single-source-of-truth

### Problema
`command.grimorio.template` en `opencode.json` duplica (mal) el workflow de `agents/grimorio-architect.md`. El template está truncado a 2000 chars, sin nombres de agentes, sin WotC gate, sin introduction/appendices explícitos.

### Solución
1. Reemplazar el `template` del comando `/grimorio` con una versión mínima que diga:
   > "This is a convenience entry point. Delegate to `grimorio-architect` agent immediately. The full workflow is defined in `agents/grimorio-architect.md`."
2. El `grimorio-architect.md` queda como **única fuente de verdad** del workflow completo.

### Archivos afectados
- `~/.config/opencode/opencode.json` — modificar `command.grimorio.template`

### Riesgo
- Bajo. El comando `/grimorio` actualmente arranca la orchestrator en main thread; apuntar al architect cambia el entry point pero el resultado es el mismo (el architect delega igual).

---

## Change: activate-integrator

### Problema
`grimorio-integrator` existe como agente (334 líneas, 7 fases completas: cross-reference audit, technical standardization, balance audit, integration, handouts, auto-fix, final validation) pero NUNCA se invoca en el workflow.

### Solución
Agregar **Fase 5f: Integration** en el workflow del architect, DESPUÉS de Appendices (5e) y ANTES de Artist (6):

```
Phase 5e: Appendices → grimorio-appendices
Phase 5f: Integration → grimorio-integrator  ← NUEVA
Phase 6: Artist batch-spec → grimorio-artist
```

El integrator:
1. Lee TODOS los archivos de la campaña
2. Hace cross-reference audit (criaturas vs bestiary, NPCs vs npcs.md, encuentros vs encounters.md)
3. Convierte referencias en texto plano a **enlaces markdown clicables**: `(Ver NPCs: X)` → `[X](npcs.md#x)`
4. Genera `INDEX.md` con tabla de contenidos (todos los NPCs, áreas, quests con enlaces)
5. Agrega breadcrumbs de navegación en cada archivo
6. Estandariza formatos (tesoro, conexiones bidireccionales, CDs numéricos)
7. Audita balance (XP budget, dificultad, curva)
8. Auto-fix de issues comunes
9. Genera handouts para jugadores
10. Corre validación final (validate_canon + check_consistency + process_consistency_gate)

### Archivos afectados
- `agents/grimorio-architect.md` — agregar Fase 5f después de Appendices

### Riesgo
- Medio. El integrator modifica archivos existentes (auto-fix). Podría sobre-corregir. Mitigación: solo auto-fix si la corrección es OBVIA, si no → warning.
- El integrator DEBE correr en el momento correcto: después de que todo el contenido existe, antes de imágenes (para que las referencias estén correctas cuando el artista genere imágenes de NPCs/criaturas).

---

## Change: introduction-before-content

### Problema
La Fase 3a (Introduction → grimorio-introduction) ya existe en architect.md pero está en Batch 1. La introducción DEBERÍA generarse ANTES de que cualquier sub-agente empiece a crear contenido, porque establece el tono, el setting, los hooks y las reglas de campaña.

### Solución
Mover `grimorio-introduction` a **Fase 2c**, inmediatamente después de `generate_adventure_bible` (Fase 2b) y ANTES de Batch 1:

```
Phase 2: Create Campaign Structure
Phase 2b: Generate Adventure Bible (canon.json)
Phase 2c: Introduction → grimorio-introduction  ← MOVER AQUÍ
Phase 3: Batch 1 (NPCs, Bestiary, Maps)
```

La introduction.md generada sirve como input de contexto para todos los sub-agentes posteriores.

### Archivos afectados
- `agents/grimorio-architect.md` — reordenar fases

### Riesgo
- Bajo. Solo reordenar fases existentes.

---

## Change: wotc-validation-per-agent

### Problema
La validación WotC (`validate-campaign.sh`) solo corre al final (Phase X). Los sub-agentes pueden devolver contenido que no cumple formato WotC, y el error se descubre tarde.

### Solución
Cada sub-agente DEBE correr validación antes de devolver:
1. Generar contenido
2. Guardar con MCP save_* tool
3. Validar con `validate_canon` MCP
4. Validar con `process_consistency_gate` MCP
5. Si falla → corregir y re-validar (auto-retry)
6. Solo entonces devolver al architect

Esto se codifica en el SKILL.md de cada agente (ver change siguiente).

### Herramientas MCP disponibles para validación
- `get_template` — obtener el template de formato correcto
- `validate_canon` — validar contra canon.json
- `check_consistency` — chequeo de consistencia cross-artifact
- `process_consistency_gate` — consistency gate con auto-retry

### Archivos afectados
- `skills/grimorio-*/SKILL.md` (nuevos skills)

### Riesgo
- Medio. Más llamadas MCP por sub-agente = más tiempo de ejecución. Pero detecta errores temprano, evitando retrabajo masivo al final.

---

## Change: skills-for-all-agents

### Problema
9 de 10 directorios `skills/grimorio-*/` están vacíos. Las instrucciones viven SOLO en `agents/grimorio-*.md`. No hay un skill portable que un agente pueda cargar para entender su trabajo + templates + validación.

Además, **si los skills tienen instrucciones técnicas y los agent .md también, se duplica la fuente de verdad** — exactamente el problema que ya tenemos entre command template y architect.md.

### Solución: separación de responsabilidades

```
agents/grimorio-areas.md                    skills/grimorio-areas/SKILL.md
┌─────────────────────────────────┐         ┌──────────────────────────────────┐
│ YAML frontmatter                │         │ Template WotC obligatorio        │
│   - name, description, tools    │         │ Cómo usar save_areas             │
│   - grimorio_mcp tools list     │         │ Pasos de validación              │
│   - color, model                │         │ Formato de salida esperado       │
│                                 │         │ Reglas de word count             │
│ Rol: "Sos el diseñador de       │         │ Cross-references a generar       │
│ áreas. Instrucciones            │         │                                  │
│ detalladas en tu skill."        │         │  →  INSTRUCCIONES TÉCNICAS       │
│                                 │         │                                  │
│  →  QUÉ HACER + METADATA       │         └──────────────────────────────────┘
└─────────────────────────────────┘
```

| Capa | Qué contiene | Quién la mantiene |
|------|-------------|-------------------|
| `agents/*.md` | Frontmatter (tools, grimorio_mcp, model) + "Cargá tu skill" | Cambios de tooling/config |
| `skills/*/SKILL.md` | Cómo generar contenido, template, validación, formato, cross-refs | Cambios de reglas de contenido |
| Delegate prompt (architect.md) | Tarea concreta ("generá acto 1 para campaña X") | Cambios de workflow/orquestación |

**Las instrucciones técnicas HOY en `agents/*.md` se MIGRAN a `skills/*/SKILL.md`**. Los agent .md se quedan como metadata + "cargá tu skill". Sin duplicación.

### Plan de migración por agente

Para CADA agente de contenido:
1. Crear `skills/grimorio-<agent>/SKILL.md` con las instrucciones técnicas extraídas de `agents/grimorio-<agent>.md`
2. Podar `agents/grimorio-<agent>.md`: dejar solo frontmatter + "Cargá tu skill para instrucciones detalladas"
3. Actualizar `agents/grimorio-architect.md`: los delegate prompts mencionan "cargá tu skill antes de empezar"

### Estructura del SKILL.md

```markdown
---
name: grimorio-<agent>
version: "1.0.0"
description: <rol del agente>
---

# grimorio-<agent> — <Rol>

## Template Requerido
- Template: `internal/compiler/templates/<template>.md.tmpl`
- **LEER el template antes de generar contenido**
- El template define el formato WotC obligatorio

## Herramientas Disponibles
- MCP tools: [listar grimorio_mcp específicas]
- NO uses Write para guardar contenido creativo — usa la MCP save_* tool

## Workflow
1. Leer canon.json + introduction.md + lore.md (contexto)
2. Leer template correspondiente
3. Generar contenido siguiendo el template
4. Guardar usando MCP save_* tool
5. **Validar**: `validate_canon` + `process_consistency_gate`
6. Si validation falla → corregir y re-validar (max 2 retries)
7. Reportar resultado al architect

## Formato WotC Obligatorio y Cross-References
- Usar enlaces markdown `[NPC](npcs.md#id)` en vez de texto plano
- Ver [Change: cross-references-and-navigation](#change-cross-references-and-navigation)
```

### Archivos afectados
- `skills/grimorio-*/SKILL.md` (nuevos 15 archivos)
- `agents/grimorio-*.md` (podar 15 archivos — dejar solo frontmatter + "cargá tu skill")
- `agents/grimorio-architect.md` (delegate prompts: mencionar que el sub-agente cargue su skill)

### Riesgo
- Bajo. Los skills no se cargan automáticamente — se invocan explícitamente desde el prompt del agente. No hay breaking change.

---

## Change: remove-write-from-content-agents

### Problema
3 agentes de contenido tienen `tools: ["Read", "Write", ...]` cuando deberían usar solo MCP save_* tools, igual que los otros 8 agentes de contenido:

| Agente | Tiene Write | Debería usar |
|--------|-------------|-------------|
| `grimorio-appendices` | ✅ | `save_appendices` |
| `grimorio-setting-guide` | ✅ | `save_setting_guide` |
| `grimorio-introduction` | ✅ | `save_introduction` |

El resto (narrative-custodian, integrator, artist, cartographer) SÍ necesitan Write para reportes, auto-fix y archivos no-MCP.

### Solución
Para cada uno de los 3:
1. Remover `"Write"` de `tools` en el frontmatter YAML
2. Agregar comentario `# Write removed — forces MCP save tool`
3. Actualizar el prompt para que use la MCP save_* tool

### Archivos afectados
- `agents/grimorio-appendices.md`
- `agents/grimorio-setting-guide.md`
- `agents/grimorio-introduction.md`

### Riesgo
- Bajo. Ya es el patrón establecido en los otros 8 agentes.

---

## Change: cross-references-and-navigation

### Problema
Nuestras campañas tienen **0 cross-references** entre archivos. WotC tiene 43-157 por aventura. El DM salta entre 16 archivos sin navegación.

| Métrica | WotC (promedio) | Grimorio (promedio) |
|---------|:---------------:|:-------------------:|
| Archivos por campaña | 1 | 16 |
| Cross-references entre secciones | 43 | **0** |
| Branching condicionales | 182 | 24 |
| Cross-references a capítulos | 157 (CoS) | **0** |

**Flujo actual del DM:** Abre `chapter_01.md`. Ve "(Ver NPCs: Pipián)". Abre `npcs.md`. Busca "Pipián". Vuelve. Ve "(Bestiario: Matón)". Abre `bestiary.md`. Busca. Luego Quest S1 → abre `quests.md`. **3-5 saltos por área. 12 áreas = 36 saltos.**

### Solución — Dos opciones complementarias

#### Opción A: Cross-references en el generador de contenido (durante creación)
Cada agente que genera contenido DEBE usar enlaces clicables markdown en vez de texto plano:

```
❌ (Ver NPCs: Pipián)
✅ [Pipián](npcs/npcs_and_factions.md#pipián)

❌ (Bestiario: Matón)  
✅ [Matón](bestiary/bestiary.md#matón)

❌ → Quest S1: Lúpulo de Verdad
✅ → [Quest S1: Lúpulo de Verdad](quests/quests.md#s1)
```

Esto se codifica en los SKILL.md de cada agente (formato de salida obligatorio).

#### Opción B: Post-procesamiento con grimorio-integrator (durante integración)
El integrator parsea todos los archivos generados y:
1. Detecta referencias en texto plano a NPCs, criaturas, quests
2. Las convierte a enlaces markdown `[Nombre](archivo.md#id)`
3. Genera un `INDEX.md` con tabla de contenidos completa
4. Agrega breadcrumbs de navegación en cada archivo

El integrator YA hace cross-reference audit — esta es una extensión natural de su trabajo.

#### Opción C (largo plazo): Compilación a 1 archivo WotC
La MCP `compile_pdf` ya existe. Extenderla para que también genere `campaign-completa.md`:
- Table of Contents generado automáticamente
- Cross-references numéricas tipo `(see Chapter 2, Area 7, p. 34)`
- Chapters con áreas, quests y encuentros intercalados
- Appendices inline
- Index de NPCs, áreas y quests con página

### Checklist por agente (qué referencias debe generar)

| Agente | Genera referencias a |
|--------|---------------------|
| `grimorio-areas` | NPCs, bestiary, quests, encounters |
| `grimorio-encounters` | NPCs, bestiary |
| `grimorio-quests` | NPCs, areas |
| `grimorio-npc` | quests, factions |
| `grimorio-narrative-custodian` | Verifica que todas las refs existan |
| `grimorio-integrator` | Post-procesa refs rotas, genera INDEX.md |

### Archivos afectados
- `skills/grimorio-*/SKILL.md` — formato de salida con enlaces
- `agents/grimorio-integrator.md` — agregar post-procesamiento de refs + INDEX.md
- `internal/compiler/` (largo plazo) — compilación a 1 archivo

### Riesgo
- Bajo para Opción A (solo cambiar el formato de output de los agentes)
- Medio para Opción B (el integrator modifica archivos existentes)
- Alto para Opción C (cambia el compiler, podría afectar PDF)

---

## Resumen de orden de ejecución propuesto

```
1. single-source-of-truth             (opencode.json → architect.md)
2. remove-write-from-content-agents   (3 agentes)
3. introduction-before-content        (reordenar architect.md)
4. activate-integrator                (agregar Fase 5f en architect.md)
5. skills-for-all-agents              (15 SKILL.md)
6. wotc-validation-per-agent          (validación en skills)
7. cross-references-and-navigation    (enlaces + INDEX.md + 1-archivo)
```

### Dependencias
- `remove-write-from-content-agents` — independiente
- `single-source-of-truth` — independiente
- `skills-for-all-agents` — depende de `remove-write-from-content-agents` (los skills deben reflejar el tooling correcto)
- `introduction-before-content` — depende de `skills-for-all-agents` (introduction skill debe existir)
- `activate-integrator` — depende de `skills-for-all-agents` (integrator skill debe existir)
- `wotc-validation-per-agent` — depende de `skills-for-all-agents` (la validación se define en los skills)
- `cross-references-and-navigation` — depende de `activate-integrator` (el integrator hace post-procesamiento) + `skills-for-all-agents` (formato de output)

---

## Riesgos globales

| Riesgo | Probabilidad | Mitigación |
|--------|-------------|------------|
| El integrator modifica archivos y rompe referencias | Media | Solo auto-fix OBVIO; si hay duda → warning. Reporte detallado de cada cambio |
| Skills desactualizados vs agentes .md | Media | Los skills son la fuente de instrucciones; los agentes .md quedan como definición de frontmatter y metadata |
| La validación WotC por agente duplica el tiempo | Alta | Pero detecta errores temprano. Meta: < 30s extra por agente |
| El comando `/grimorio` cambia comportamiento | Baja | El template nuevo solo redirige al architect, que ya existe y funciona |
