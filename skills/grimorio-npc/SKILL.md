---
name: grimorio-npc
version: "1.0.0"
description: Generate NPCs, factions, and social relationships with WotC-enhanced descriptions. Split mode for chapter-sequential campaigns.
---

# grimorio-npc — NPC Designer

## Modo Split (Chapter-Sequential)

Para campañas que usan `grimorio-chapters` y `save_chapter`:

**Fase de Capítulos (Inline):**
- Generar perfiles **condensados** (150-300 palabras) dentro de cada capítulo
- Incluir solo: apariencia, motivación, relación con PJs, secreto
- NO incluir stat blocks completos inline
- Rol: `chapter-inline`

**Fase de Appendices (Consolidado):**
- Generar perfiles **completos** (500-800 palabras) para todos los NPCs de la campaña
- Incluir stat blocks, equipamiento, historia completa
- Guardar en `npcs/npcs_and_factions.md` vía `save_npcs`

**Cross-References:**
- Los nombres de NPCs en capítulos deben coincidir exactamente con los de appendices
- `validate_canon` asegura consistencia de nombres entre capítulos

## Template Requerido

**ANTES de generar contenido, LEER el template:**

```
get_template(type="npc")
```

El template define el formato WotC obligatorio para NPCs y facciones.

## Herramientas Disponibles

**MCP Tools (USAR para guardar contenido):**
- `save_npcs` — Guardar NPCs y facciones
- `validate_canon` — Validar contra canon.json
- `check_consistency` — Chequeo de consistencia
- `process_consistency_gate` — Validación batch con auto-retry
- `update_faction_reputation` — Actualizar reputación de facciones

**NO usar Write para contenido creativo** — El frontmatter del agente ya no incluye Write para forzar el uso de MCP save tools.

## Workflow Obligatorio

```
1. LEER contexto:
   - canon.json (hechos canónicos, entidades)
   - lore.md (tono, conflicto, setting, geografía)

2. LEER template:
   - get_template(type="npc")

3. GENERAR NPCs y facciones siguiendo el template:
   - NPCs principales: 500-800 palabras cada uno
   - NPCs secundarios: 200-400 palabras cada uno
   - Facciones: 2-4 facciones con goals claros

4. VALIDAR antes de guardar:
   - validate_canon() con entity_references
   - process_consistency_gate() para validación batch
   - Máximo 3 reintentos si falla

5. GUARDAR solo si validación pasa:
   - save_npcs(campaign, content)

6. GENERAR stat blocks en bestiary.md para NPCs con combate

7. REPORTAR al architect
```

## Formato WotC Obligatorio

### NPCs Principales (500-800 palabras)

```markdown
### {NPC Name}

*{Alignment} {Race} {Class}*

**Ubicación:** Área X

**Estadísticas de Combate:** CA XX, PG XX (XdX+X), Attack +X (XdX+X)

**Rol en la historia:** [Función narrativa]

### Appearance Física

**Altura y Complexión:** [2-3 oraciones detalladas]

**Rostro:** [Ojos, nariz, boca, expresión - 2-3 oraciones]

**Cabello:** [Color, estilo, textura - 1-2 oraciones]

**Vestimenta:** [Ropa típica, accesorios, símbolos - 2-3 oraciones]

**Características Distintivas:** [Cicatrices, tatuajes, postura - 1-2 oraciones]

### Personality

**Mannerisms:** [Gestos, tics, hábitos - 2-3 oraciones]

**Speech Patterns:** [Cómo habla, vocabulario, tono - 2-3 oraciones]

**Motivations:** [Qué lo impulsa, metas, miedos - 2-3 oraciones]

### Voz

**Tono:** [Grave, agudo, ronco, suave]

**Accent:** [Regional, social, educativo]

**Catchphrases:** [Frases típicas, muletillas - 1-2 ejemplos]

### Secrets

**Secret Trivial:** [Algo curioso sin impacto en trama]

**Secret Importante #1:** [Información que puede iniciar quest]

**Secret Importante #2:** [Otra información relevante]

**Secret de Campaña:** [Información que cambia la trama principal]

### Ganchos de Trama

**Hook #1:** [Por qué interactúa con los PJs]

**Hook #2:** [Cómo puede ayudar u obstaculizar]

**Hook #3:** [Conexión con trama principal]

### Diálogo para Leer

*"[Línea 1 - saludo o apertura]"*

*"[Línea 2 - información o reacción]"*

*"[Línea 3 - cierre o llamada a la acción]"*

**Stat Block:** Ver [bestiary/bestiary.md#{npc-name}](bestiary/bestiary.md#{npc-name})
```

### Facciones (2-4)

```markdown
### {Faction Name}

**Tipo:** [Supervivientes|Culto|Gremio|Nobleza|Criminal|Militar|Religioso]

**Alignment:** [Ethos general]

**Líder:** [Nombre del NPC líder]

**Objetivo:** [Qué quiere la facción como grupo]

**Miembros:** [Quién se une y por qué]

**Recursos:** [Qué tienen - dinero, conexiones, ejércitos, información]

**Relación con jugadores:** [Amigable|Neutral|Hostil|Engañosa]

**Secret:** [Qué ocultan]

**Reputación inicial:** [0-100, donde 50 es neutral]
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
        id: "npc-batch",
        type: "npc",
        content: "Resumen de NPCs generados...",
        entity_references: [
          { entity_id: "npc-001", location: "npcs_and_factions" },
          { entity_id: "npc-002", location: "npcs_and_factions" },
          { entity_id: "faction-001", location: "npcs_and_factions" }
        ]
      }
    )
    
    IF result.status == "approved":
        validation_passed = true
    ELSE:
        retry_count += 1
        Fix issues based on result.feedback
    
IF validation_passed:
    save_npcs(campaign="{campaign_name}", content=...)
ELSE:
    Report failure: "Validation failed after 3 retries"
    DO NOT save content
```

## Checklist Pre-Guardado (v2.5 WotC Enhanced)

- [ ] **Appearance Física:** 3-5 párrafos detallados (150-250 palabras)
- [ ] **Personality:** 2-3 párrafos (mannerisms, speech, motivations)
- [ ] **Voz:** 1 párrafo (tono, accent, catchphrases)
- [ ] **Secrets:** 3-5 secretos (1 trivial, 2 importantes, 1 de campaña)
- [ ] **Plot Hooks:** 2-3 hooks explicando interacción con PJs
- [ ] **Diálogo:** 3-5 líneas para read-aloud
- [ ] **Stat Block:** En bestiary.md con formato completo
- [ ] **Referencia Cruzada:** "Ver [bestiary.md](bestiary/bestiary.md#{npc-name})" en npcs.md
- [ ] **Balance:** 1-2 aliados claros, 2-3 neutrales útiles, 1 hostil encubierto, 1 villano
- [ ] **Diversidad:** Mezcla de edades, géneros, razas, personalidades
- [ ] **Connections:** Cada NPC conectado con al menos 1 otro NPC/facción/quest
- [ ] **Word Count:** NPCs principales 500-800 palabras, secundarios 200-400 palabras

## Cross-References Format

**OBLIGATORIO usar enlaces markdown:**

```markdown
❌ MAL: El líder de los Contrabandistas (ver NPCs)
✅ BIEN: El líder de los [Contrabandistas](npcs/npcs_and_factions.md#contrabandistas) es [Silas](npcs/npcs_and_factions.md#silas)

❌ MAL: La facción de la Guardia
✅ BIEN: La [Guardia de la Ciudad](npcs/npcs_and_factions.md#guardia-de-la-ciudad)

❌ MAL: Como se menciona en la quest principal
✅ BIEN: Como se menciona en [Quest: El Secret del Herrero](quests/quests.md#el-secreto-del-herrero)
```

## NPC Stat Block Requirements

Cada NPC mencionado DEBE tener stat block en bestiary.md:

```markdown
### {NPC Name}

*"{Alignment} {Race} {Class}"*

**AC** XX | **HP** XX | **Speed** XX ft.

| STR | DEX | CON | INT | WIS | CHA |
|-----|-----|-----|-----|-----|-----|
| +X  | +X  | +X  | +X  | +X  | +X  |

**Skills** [Skill +X, ...]
**Senses** [darkvision 60 ft., passive Perception XX]
**Languages** [idiomas]
**Challenge** X (XXX XP)

**Actions**
**[Attack Name].** *Melee Weapon Attack:* +X to hit, reach/range X ft., one target. *Hit:* X (XdX + X) damage type.

**[Special Action].** [Description with mechanics]
```

## WotC Quality Validators

### ValidateNPCStatLinks
- ✅ Cada NPC en npcs.md tiene referencia en bestiary.md
- ✅ Stat block con formato completo (AC, HP, abilities, actions)
- ✅ Referencia cruzada bidireccional

### ValidateNPCWordCount
- ✅ NPCs principales: 500-800 palabras
- ✅ NPCs secundarios: 200-400 palabras
- ✅ Appearance: 150-250 palabras
- ✅ Personality: 100-150 palabras
- ✅ Secrets: 150-200 palabras

### ValidateNPCDepth
- ✅ 5+ párrafos de descripción total
- ✅ 3-5 secretos por NPC clave
- ✅ 3-5 líneas de diálogo para NPCs importantes

## Faction Reputation System

Usar `update_faction_reputation` para cambios de reputación:

```python
update_faction_reputation(
    campaign_id="{campaign_name}",
    faction_id="faction-id",
    party_id="party-1",
    delta=+10,  # -100 a +100
    reason="Los PJs ayudaron a resolver el asesinato del magistrado"
)
```

**Tiers de Reputación:**
- **Rank 1 (1-30):** Beneficios básicos
- **Rank 2 (31-70):** Acceso a recursos, descuentos
- **Rank 3 (71-100):** Alianza formal, apoyo militar

## Error Handling

Si la validación falla:

1. **Analizar feedback específico** (ej: "NPC X no existe en canon")
2. **Corregir issues concretos** (usar nombres de canon.json)
3. **Re-validar** con contenido corregido
4. **Máximo 3 reintentos** — si falla, abortar y reportar

## Output al Architect

```markdown
## NPCs Generados: {campaign_name}

**Status:** ✅ Complete / ❌ Failed

**NPCs:**
- Principales: {count} ({word_count} palabras promedio)
- Secundarios: {count} ({word_count} palabras promedio)

**Facciones:**
- {count} facciones con reputación inicial configurada

**Stat Blocks:**
- {count} NPCs con stat block en bestiary.md

**Validación:**
- validate_canon: ✅ Passed
- process_consistency_gate: ✅ Passed
- ValidateNPCStatLinks: ✅ Passed

**Cross-References:**
- NPCs referenciados en acts: {count} (todos existen)
- Facciones referenciadas: {count} (todas existen)
```
