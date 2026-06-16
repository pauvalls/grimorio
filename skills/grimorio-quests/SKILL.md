---
name: grimorio-quests
version: "1.0.0"
description: Generate personal quests, side missions, and character-specific narrative hooks
---

# grimorio-quests — Quest Designer

## Template Requerido

**ANTES de generar contenido, LEER el template:**

```
get_template(type="quest")
```

El template define el formato WotC obligatorio para quests y misiones.

## Herramientas Disponibles

**MCP Tools (USAR para guardar contenido):**
- `create_personal_quest` — Crear quest personal para personaje
- `list_quests` — Listar todas las quests de la campaña
- `update_quest_status` — Actualizar estado de quest (active/completed/failed/on_hold)
- `validate_canon` — Validar contra canon.json
- `check_consistency` — Chequeo de consistencia
- `process_consistency_gate` — Validación batch con auto-retry

**NO usar Write para contenido creativo** — El frontmatter del agente ya no incluye Write para forzar el uso de MCP save tools.

## Workflow Obligatorio

```
1. LEER contexto:
   - canon.json (hechos canónicos, entidades)
   - lore.md (mundo y conflicto)
   - npcs/npcs_and_factions.md (NPCs que pueden dar misiones)
   - narrative_state.json (estado actual de quests)

2. LEER template:
   - get_template(type="quest")

3. GENERAR quests siguiendo el template:
   - Tipos: redencion|venganza|descubrimiento|proteccion
   - 1 quest por PJ o 2-3 quests generales para one-shot
   - Recompensas significativas (no solo oro)

4. VALIDAR antes de guardar:
   - validate_canon() con entity_references
   - process_consistency_gate() para validación batch
   - Máximo 3 reintentos si falla

5. GUARDAR quests:
   - create_personal_quest() para CADA quest

6. GENERAR character hooks:
   - generate_character_hooks() para hooks automáticos
   - Guardar en quests/character-hooks.md

7. REPORTAR al architect
```

## Formato WotC Obligatorio

### Estructura de cada Quest

```markdown
# {Quest Title}

**Tipo:** [redencion|venganza|descubrimiento|proteccion]

**Para:** {Character Name} o "Cualquier PJ"

**Estado:** active

**NPC que da la quest:** [Nombre del NPC]

---

## Hook

[1-2 párrafos. Cómo se introduce la misión. Qué NPC la da, qué la desencadena.]

---

## Stakes

[Qué se pierde si la misión falla. Debería ser significativo a nivel narrativo o mecánico.]

---

## Objetivos

### Principal
- [Objetivo claro y específico]

### Secundarios (opcionales)
- [Objetivo opcional 1]
- [Objetivo opcional 2]

---

## Desarrollo

### Acto 1: Introducción

[Cómo los PJs se enteran de la quest. Primer paso.]

### Acto 2: Desarrollo

[Obstáculos, encuentros, revelaciones. El núcleo de la quest.]

### Acto 3: Resolución

[Confrontación final o momento de decisión.]

---

## Recompensa

**Principal:**
- [Objeto mágico, información clave, aliado, desarrollo de personaje]

**Secundaria:**
- [Oro, reputación, acceso a recursos]

**Narrativa:**
- [Cómo esta quest afecta el arco del personaje]

---

## Consecuencias

### Si completan la quest:
- [Consecuencias positivas]

### Si fallan la quest:
- [Consecuencias negativas]

### Si ignoran la quest:
- [Consecuencias diferidas o alternativas]

---

## Connections

**Conecta con:**
- [Quest principal de la campaña]
- [Otras side quests]
- [NPCs relevantes]

**Requiere:**
- [Prerrequisitos si aplica]

**Desbloquea:**
- [Quests o contenido futuro]
```

## Validación de Canon (CRÍTICO)

```python
max_retries = 3
retry_count = 0
validation_passed = false

WHILE retry_count < max_retries AND NOT validation_passed:
    result = validate_canon(
      campaign_id="{campaign_name}",
      proposal={
        id: "quest-{name}",
        type: "quest",
        content: "Resumen de la misión...",
        entity_references: [
          { entity_id: "npc-giver", location: "quest" },
          { entity_id: "location-target", location: "quest" }
        ]
      }
    )
    
    IF result.status == "approved":
        validation_passed = true
    ELSE:
        retry_count += 1
        Fix issues based on result.feedback
    
IF validation_passed:
    create_personal_quest(
        campaign="{campaign_name}",
        quest_title="{Quest Title}",
        quest_type="{type}",
        character_name="{character}",
        hook="{hook text}",
        stakes="{stakes text}",
        reward="{reward text}"
    )
ELSE:
    Report failure: "Validation failed after 3 retries"
    DO NOT create quest
```

## Tipos de Quest

| Tipo | Descripción | Ejemplo |
|------|-------------|---------|
| **redencion** | El personaje busca perdonarse por algo | "Redimir el honor familiar perdido" |
| **venganza** | Alguien le hizo daño y quiere cobrar | "Encontrar al asesino de su mentor" |
| **descubrimiento** | Busca la verdad sobre algo del pasado | "Descubrir el secreto de su nacimiento" |
| **proteccion** | Defiende a alguien o algo que ama | "Proteger el pueblo de su infancia" |

## Checklist Pre-Guardado

- [ ] **Tipo de Quest:** redencion|venganza|descubrimiento|proteccion
- [ ] **Hook Claro:** 1-2 párrafos explicando cómo se introduce
- [ ] **Stakes Significativos:** Qué se pierde si falla (no solo oro)
- [ ] **Objetivos Claros:** Se sabe cuándo se completó la quest
- [ ] **Recompensas Variadas:** Objeto, información, aliado, desarrollo
- [ ] **Consecuencias:** Éxito, fracaso, e ignorar documentados
- [ ] **Connections:** Conecta con trama principal u otras quests
- [ ] **NPCs Existentes:** NPC que da la quest existe en npcs.md
- [ ] **Locaciones Existentes:** Lugares referenciados existen en maps.md
- [ ] **Cantidad Aproximada:** 1 por PJ o 2-3 generales para one-shot

## Cross-References Format

**OBLIGATORIO usar enlaces markdown:**

```markdown
❌ MAL: Hablá con el herrero
✅ BIEN: Hablá con [Mastro Aldric](npcs/npcs_and_factions.md#mastro-aldrick)

❌ MAL: La quest ocurre en el bosque
✅ BIEN: La quest ocurre en el [Bosque de los Susurros](maps/maps.md#bosque-de-los-susurros)

❌ MAL: Encontrá la llave del tesoro
✅ BIEN: Encontrá la [Llave de Platino](bestiary/bestiary.md#llave-de-platino) que abre la cámara de [Lord Blackthorn](npcs/npcs_and_factions.md#lord-blackthorn)
```

## Character Hooks Generation

Después de generar las quests, DEBÉS generar hooks automáticos:

```python
generate_character_hooks(campaign="{campaign_name}")
```

**Output:** Guardar en `quests/character-hooks.md`:

```markdown
# Character Hooks

## Hooks por Personaje

### {Character Name}
**Background:** {background} | **Class:** {class}

### Gancho Personal

[Hook text - 2-3 oraciones conectando el personaje con la trama principal]

**Conexión con la Trama:** [Cómo se conecta con el plot principal]

---

## Hooks por Área (para incluir en cada área)

| Área | Personaje | Background | Gancho |
|------|-----------|------------|--------|
| Área 1 | {Name} | {Background} | {Hook truncated} |
| Área 2 | {Name} | {Background} | {Hook truncated} |

---

## Instrucciones para el DM

1. **Antes de la Sesión Cero:** Generá estos hooks y compartilos individualmente con cada jugador
2. **Durante el Juego:** Incluí referencias a estos hooks en las áreas correspondientes
3. **Evolución:** Actualizá los hooks según las decisiones de los personajes
4. **Recompensas:** Los hooks bien interpretados pueden dar ventaja en tiradas sociales
```

## WotC Quality Validators

### ValidateQuestStructure
- ✅ Hook claro (1-2 párrafos)
- ✅ Stakes significativos
- ✅ Objetivos claros y específicos
- ✅ Recompensas variadas (no solo oro)

### ValidateNarrativeIntegration
- ✅ Conecta con trama principal
- ✅ NPCs referenciados existen en npcs.md
- ✅ Locaciones referenciadas existen en maps.md

### ValidateQuestBalance
- ✅ 1 quest por PJ o 2-3 generales para one-shot
- ✅ No satura con contenido descartable
- ✅ Recompensas significativas (información, aliados, desarrollo)

### ValidateEmotionalHooks
- ✅ Ganchos emocionales (no solo pago de oro)
- ✅ El PJ QUIERE hacer la quest (motivación intrínseca)
- ✅ Conecta con backstory del personaje

## Quest Status Management

Usar `update_quest_status` para actualizar el estado:

```python
update_quest_status(
    campaign="{campaign_name}",
    quest_id="quest-id",
    status="active",  # active|completed|failed|on_hold
    notes="Progreso: Los PJs encontraron la primera pista"
)
```

## Error Handling

Si la validación falla:

1. **Analizar feedback específico** (ej: "NPC no existe en canon")
2. **Corregir issues concretos** (usar NPC existente o agregarlo)
3. **Re-validar** con contenido corregido
4. **Máximo 3 reintentos** — si falla, abortar y reportar

## Output al Architect

```markdown
## Quests Generadas: {campaign_name}

**Status:** ✅ Complete / ❌ Failed

**Quests:**
- Personales: {count} (1 por PJ)
- Generales: {count} (opcionales)

**Tipos:**
- Redención: {count}
- Venganza: {count}
- Descubrimiento: {count}
- Protección: {count}

**Character Hooks:**
- Hooks generados: {count}
- Archivo: quests/character-hooks.md

**Validación:**
- validate_canon: ✅ Passed
- process_consistency_gate: ✅ Passed

**Cross-References:**
- NPCs referenciados: {count} (todos existen)
- Locaciones referenciadas: {count} (todas existen)
- Objetos referenciados: {count} (todos existen)
```
