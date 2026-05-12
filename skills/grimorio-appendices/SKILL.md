---
name: grimorio-appendices
version: "1.0.0"
description: Consolidate campaign reference material — magic items, stat blocks, handouts, maps, tables
---

# grimorio-appendices — Appendices Master

## Template Requerido

**ANTES de generar contenido, LEER el template:**

```
get_template(type="appendix")
```

El template define el formato WotC obligatorio para apéndices de campaña.

## Herramientas Disponibles

**MCP Tools (USAR para guardar contenido):**
- `save_appendices` — Guardar apéndices consolidados
- `validate_canon` — Validar contra canon.json
- `check_consistency` — Chequeo de consistencia
- `process_consistency_gate` — Validación batch con auto-retry

**NO usar Write para contenido creativo** — El frontmatter del agente ya no incluye Write para forzar el uso de MCP save_appendices tool.

## Workflow Obligatorio

```
1. LEER TODOS los archivos de referencia:
   - canon.json (hechos canónicos, entidades)
   - bestiary/bestiary.md (criaturas para stat blocks)
   - npcs/npcs_and_factions.md (NPCs para stat blocks)
   - handouts/handouts.md (handouts disponibles)
   - acts/ (encounters y treasure分布)

2. LEER template:
   - get_template(type="appendix")

3. CONSOLIDAR apéndices siguiendo el template:
   - Appendix A: Magic Items
   - Appendix B: NPCs and Monsters (stat blocks)
   - Appendix C: Handouts
   - Appendix D: Maps
   - Appendix E: Reference Tables

4. VALIDAR antes de guardar:
   - validate_canon() con entity_references
   - process_consistency_gate() para validación batch
   - Máximo 3 reintentos si falla

5. GUARDAR solo si validación pasa:
   - save_appendices(campaign, content)

6. REPORTAR al architect
```

## Formato WotC Obligatorio

```markdown
# Appendices: {Campaign Name}

---

## Appendix A: Magic Items

*Magic items found in this adventure. Items marked with  are unique to this campaign.*

### {Item Name}

*{Rarity}, {Type}*

{2-4 sentence description. What it does, how it works, what it looks like.}

**Activation:** {How to use it — command word, attunement, etc.}

---

## Appendix B: NPCs and Monsters

*Stat blocks for every NPC and monster that appears in this adventure.*

### NPCs

#### {NPC Name}

*{Alignment} {Race} {Class}*

**AC** {Number} | **HP** {Number} | **Speed** {Speed}

| STR | DEX | CON | INT | WIS | CHA |
|-----|-----|-----|-----|-----|-----|
| {10} (+0) | {10} (+0) | {10} (+0) | {10} (+0) | {10} (+0) | {10} (+0) |

**Skills** {Skills} | **Senses** {Senses} | **Languages** {Languages}

**Challenge** {CR} ({XP})

{Abilities, actions, and legendary actions as needed. Keep it concise — 10-20 lines total for a standard NPC.}

{If the NPC has special equipment, bonds, or secrets, describe them here.}

---

### Monsters

#### {Monster Name}

*{Size} {Type}, {Alignment}*

**AC** {Number} | **HP** {Number} | **Speed** {Speed}

| STR | DEX | CON | INT | WIS | CHA |
|-----|-----|-----|-----|-----|-----|
| {10} (+0) | {10} (+0) | {10} (+0) | {10} (+0) | {10} (+0) | {10} (+0) |

**Skills** {Skills} | **Senses** {Senses} | **Languages** {Languages}

**Challenge** {CR} ({XP})

**{Trait Name}.** {Effect}
{1-2 sentence description of the trait.}

**Actions**

**{Weapon/Spell Name}.** *Melee Weapon Attack:* +{to hit}, reach {5/10} ft., {target}. *Hit:* {damage} {type} damage.

---

## Appendix C: Handouts

*Player-facing materials — maps, clues, letters, and other documents.*

### Handout {Number}: {Name}

{What the players receive. A physical prop, a description to read aloud, or a handout to distribute.}

---

## Appendix D: Maps

*Key maps for the DM. Player versions are provided separately.*

### {Map Name}

{Description of what's shown. Scale, key features, points of interest.}

*[Map: {filename}-dm.png]*

---

## Appendix E: Reference Tables

### Random Encounters

| d{X} | Encounter | Location |
|------|-----------|----------|
| 1 | {Encounter description} | {Where} |
| 2 | {Encounter description} | {Where} |
| 3 | {Encounter description} | {Where} |
| 4 | {Encounter description} | {Where} |
| 5 | {Encounter description} | {Where} |
| 6 | {Encounter description} | {Where} |

### Treasure Generation

| CR | Gold Amount |
|----|-------------|
| 1-4 | {Amount} gp |
| 5-10 | {Amount} gp |
| 11-16 | {Amount} gp |
| 17+ | {Amount} gp |

---

*End of Appendices*
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
        id: "appendices-main",
        type: "lore",
        content: "Resumen de los appendices...",
        entity_references: [
          { entity_id: "npc-001", location: "appendices" },
          { entity_id: "monster-001", location: "appendices" },
          { entity_id: "item-001", location: "appendices" }
        ]
      }
    )
    
    IF result.status == "approved":
        validation_passed = true
    ELSE:
        retry_count += 1
        Fix issues based on result.feedback
    
IF validation_passed:
    save_appendices(campaign="{campaign_name}", content=...)
ELSE:
    Report failure: "Validation failed after 3 retries"
    DO NOT save content
```

## Checklist Pre-Guardado

- [ ] **Appendix A:** Magic items con rarity, type, description, activation
- [ ] **Appendix B:** NPCs con stat blocks completos (AC, HP, abilities, 10-20 líneas)
- [ ] **Appendix B:** Monsters con stat blocks completos (traits, actions)
- [ ] **Appendix C:** Handouts player-facing (sin spoilers)
- [ ] **Appendix D:** Maps con filename reference para compiler
- [ ] **Appendix E:** Random encounters table (d6)
- [ ] **Appendix E:** Treasure generation table por CR
- [ ] **Orden:** Items → NPCs → Monsters → Handouts → Maps → Tables
- [ ] **Concisión:** Stat blocks 10-20 líneas (no fluff)
- [ ] **Consistencia:** Solo contenido de la campaña (no todo el MM)

## Cross-References Format

**OBLIGATORIO usar enlaces markdown:**

```markdown
❌ MAL: Ver bestiary para stats
✅ BIEN: Ver [Appendix B: NPCs and Monsters](appendices/appendices.md#appendix-b-npcs-and-monsters)

❌ MAL: El mapa está en assets
✅ BIEN: *[Map: palacio-dm.png]* (el compiler busca este archivo)

❌ MAL: Como se menciona en el acto 2
✅ BIEN: Como se menciona en [Acto 2: La Ciudad](acts/chapter_02.md)
```

## Writing Standards

### Stat Blocks Concisos

**✅ BIEN (10-20 líneas):**
```markdown
#### Mastro Aldric

*LG male Chondathan human fighter*

**AC** 16 (chain mail) | **HP** 45 (6d8+18) | **Speed** 30 ft.

| STR | DEX | CON | INT | WIS | CHA |
|-----|-----|-----|-----|-----|-----|
| +4  | +1  | +4  | +0  | +2  | +1  |

**Skills** Athletics +6, Intimidation +3
**Senses** passive Perception 12 | **Languages** Common
**Challenge** 3 (700 XP)

**Actions**

**Longsword.** *Melee Weapon Attack:* +6 to hit, reach 5 ft., one target. *Hit:* 8 (1d8+4) slashing damage.

**Special Equipment:** Anillo de protección (+1 CA), carta de presentación de la Guardia.
```

### Magic Items con Activación Clara

**✅ BIEN:**
```markdown
### Amuleto de los Susurros

*Uncommon, Wondrous Item (requires attunement)*

Este amuleto de plata permite al usuario escuchar conversaciones a hasta 60 pies de distancia.

**Activation:** Como acción bonus, susurrá el nombre de la persona que querés escuchar. Si está dentro del rango, escuchás su voz claramente.
```

### Handouts Player-Facing

Los handouts DEBEN ser:
- ✅ Player-facing (sin spoilers de trama)
- ✅ Físicos o describibles (cartas, mapas, notas)
- ✅ Útiles para la inmersión

**✅ BIEN:**
```markdown
### Handout 1: Carta de Rescate

*Una carta arrugada con el sello de la familia Noble.*

"Querido hermano, si estás leyendo esto, he sido capturado. Me tienen en los sótanos de la villa. Buscad la llave bajo..."
```

## WotC Quality Validators

### ValidateAppendixStructure
- ✅ 5 apéndices presentes (A-E)
- ✅ Orden correcto (Items → NPCs → Monsters → Handouts → Maps → Tables)
- ✅ Cada apéndice tiene introducción contextual

### ValidateStatBlockConciseness
- ✅ NPCs: 10-20 líneas por stat block
- ✅ Monsters: 10-15 líneas por stat block
- ✅ No fluff, solo mecánicas relevantes

### ValidateItemClarity
- ✅ Rarity y type especificados
- ✅ Activación clara (command word, attunement, action)
- ✅ Efecto mecánico preciso

### ValidateHandoutSafety
- ✅ Handouts son player-facing (sin spoilers)
- ✅ Handouts son físicos o describibles
- ✅ Handouts tienen propósito narrativo

## Error Handling

Si la validación falla:

1. **Analizar feedback específico** (ej: "stat block incompleto")
2. **Corregir issues concretos** (agregar campos faltantes)
3. **Re-validar** con contenido corregido
4. **Máximo 3 reintentos** — si falla, abortar y reportar

## Output al Architect

```markdown
## Appendices Generados: {campaign_name}

**Status:** ✅ Complete / ❌ Failed

**Contenido:**
- Appendix A (Magic Items): {count} items
- Appendix B (NPCs): {count} stat blocks
- Appendix B (Monsters): {count} stat blocks
- Appendix C (Handouts): {count} handouts
- Appendix D (Maps): {count} maps
- Appendix E (Tables): {count} tables

**Validación:**
- validate_canon: ✅ Passed
- process_consistency_gate: ✅ Passed
- ValidateAppendixStructure: ✅ Passed

**Consistencia:**
- NPCs de npcs.md incluidos: {count}
- Monsters de bestiary.md incluidos: {count}
- Items de acts referenciados: {count}
```
