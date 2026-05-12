---
name: grimorio-setting-guide
version: "1.0.0"
description: Generate DM-only campaign setting guide with spoilers, geography, history, and factions
---

# grimorio-setting-guide — Setting Guide Master

## Template Requerido

**ANTES de generar contenido, LEER el template:**

```
get_template(type="setting")
```

El template define el formato WotC obligatorio para la guía de setting (DM-only).

## Herramientas Disponibles

**MCP Tools (USAR para guardar contenido):**
- `save_setting_guide` — Guardar setting guide
- `validate_canon` — Validar contra canon.json
- `check_consistency` — Chequeo de consistencia
- `process_consistency_gate` — Validación batch con auto-retry

**NO usar Write para contenido creativo** — El frontmatter del agente ya no incluye Write para forzar el uso de MCP save_setting_guide tool.

## Workflow Obligatorio

```
1. LEER contexto:
   - canon.json (hechos canónicos, entidades, reglas del mundo)
   - lore.md (conflicto central, tono, geografía)

2. LEER template:
   - get_template(type="setting")

3. GENERAR setting guide siguiendo el template:
   - Geography (major locations, landmarks)
   - History (recent events, ancient history)
   - Culture and Society (structure, religion, customs, crime)
   - Factions (2-4 facciones con goals y secrets)
   - NPCs Who Matter
   - Secrets and Lies (3 capas)
   - Environment and Weather
   - Economy

4. VALIDAR antes de guardar:
   - validate_canon() con entity_references
   - process_consistency_gate() para validación batch
   - Máximo 3 reintentos si falla

5. GUARDAR solo si validación pasa:
   - save_setting_guide(campaign, content)

6. REPORTAR al architect
```

## Formato WotC Obligatorio

```markdown
# Setting Guide: {Campaign Name}

> **DM-ONLY:** Este documento CONTIENE SPOILERS. No compartir con jugadores.

---

## Geography

### Major Locations

#### {Location 1}

**Type:** [City|Town|Dungeon|Wilderness|Village|Fortress|Other]

**Population:** [Approximate population]

**Government:** [Who rules and how]

**Description:**
[2-3 párrafos sobre el aspecto, sensación, y carácter del lugar. Qué notan los visitantes primero.]

#### {Location 2}

[Repetir estructura]

### Key Landmarks

| Landmark | Location | Description |
|----------|----------|-------------|
| {Name} | {Where} | {What it is and why it matters} |
| {Name} | {Where} | {What it is and why it matters} |

---

## History

### Recent Events ({Last 10-50 years})

1. **{Event}** — {When} — {What happened and why it matters}
2. **{Event}** — {When} — {What happened and why it matters}
3. **{Event}** — {When} — {What happened and why it matters}

### Ancient History

[Eventos de más tiempo atrás que dan forma al setting actual. Mitos de creación, imperios caídos, males antiguos. 2-3 párrafos.]

---

## Culture and Society

### Social Structure

[Cómo está organizada la sociedad. Clases, castas, o guilds. Quién tiene poder y quién no.]

### Religion and Beliefs

[Deidades principales, cults locales, prácticas religiosas. Qué creen las personas y cómo afecta la vida diaria.]

### Customs and Traditions

- **{Custom 1}:** {Description}
- **{Custom 2}:** {Description}
- **{Custom 3}:** {Description}

### Crime and Punishment

[Leyes,执法者, qué pasa con los criminales. Cómo funciona el sistema legal.]

---

## Factions

### {Faction 1}

**Alignment:** [Their overall ethos]

**Leader:** [Who runs them]

**Members:** [Who joins and why]

**Goals:** [What they want to achieve]

**Resources:** [What they have — money, connections, armies]

**Secret:** [What they're hiding]

**Reputation Tiers:**
- **Rank 1 (1-30):** [Beneficios básicos]
- **Rank 2 (31-70):** [Acceso a recursos, descuentos]
- **Rank 3 (71-100):** [Alianza formal, apoyo militar]

### {Faction 2}

[Repetir estructura]

---

## The NPCs Who Matter

> *Full stat blocks in Appendix B.*

| NPC | Location | Role | Stat Block |
|-----|----------|------|------------|
| **{Name}** | {Where} | {What they do} | Appendix B |
| **{Name}** | {Where} | {What they do} | Appendix B |

---

## Secrets and Lies

### What Everyone Knows

[Public knowledge — what's commonly believed]

### What Some Know

[Restricted information — only certain groups know this]

### What Nobody Knows

[Secrets — plot twists, hidden truths, things even the DM should keep in reserve]

---

## Environment and Weather

### Typical Climate

[Weather patterns, seasons, temperature ranges.]

### Random Weather Table

| d6 | Weather |
|----|---------|
| 1 | {Weather} |
| 2 | {Weather} |
| 3 | {Weather} |
| 4 | {Weather} |
| 5 | {Weather} |
| 6 | {Weather} |

---

## Economy

### Standard Exchange

- **1 gp** = {What it buys}

**Notable prices:**
- {Item}: {Cost}
- {Item}: {Cost}

### Trade Goods

[What the region produces, imports, and exports. Major trade routes.]
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
        id: "setting-guide-main",
        type: "lore",
        content: "Resumen del setting guide...",
        entity_references: [
          { entity_id: "location-001", location: "setting-guide" },
          { entity_id: "faction-001", location: "setting-guide" }
        ]
      }
    )
    
    IF result.status == "approved":
        validation_passed = true
    ELSE:
        retry_count += 1
        Fix issues based on result.feedback
    
IF validation_passed:
    save_setting_guide(campaign="{campaign_name}", content=...)
ELSE:
    Report failure: "Validation failed after 3 retries"
    DO NOT save content
```

## Checklist Pre-Guardado

- [ ] **DM-ONLY Warning:** Advertencia de spoilers al inicio
- [ ] **Geography:** Major locations con type, population, government, description
- [ ] **Landmarks:** Tabla de landmarks con ubicación y descripción
- [ ] **History:** Recent events (3+) y ancient history (2-3 párrafos)
- [ ] **Social Structure:** Cómo está organizada la sociedad
- [ ] **Religion:** Deidades, cults, prácticas religiosas
- [ ] **Customs:** 3+ costumbres y tradiciones
- [ ] **Crime:** Sistema legal y castigos
- [ ] **Factions:** 2-4 facciones con alignment, leader, members, goals, resources, secret
- [ ] **Reputation Tiers:** 3 tiers de reputación por facción
- [ ] **NPCs Who Matter:** Tabla con NPCs importantes
- [ ] **Secrets:** 3 capas (Everyone/Some/Nobody Knows)
- [ ] **Weather:** Climate description + random weather table (d6)
- [ ] **Economy:** Standard exchange, notable prices, trade goods
- [ ] **Consistencia:** No contradice lore.md ni canon.json

## Cross-References Format

**OBLIGATORIO usar enlaces markdown:**

```markdown
❌ MAL: El líder de la Guardia
✅ BIEN: El líder de la [Guardia de la Ciudad](npcs/npcs_and_factions.md#guardia-de-la-ciudad)

❌ MAL: El templo en la ciudad
✅ BIEN: El [Templo de los Olvidados](maps/maps.md#templo-de-los-olvidados)

❌ MAL: Ver apéndice para stats
✅ BIEN: Ver [Appendix B: NPCs and Monsters](appendices/appendices.md#appendix-b-npcs-and-monsters)
```

## Writing Standards

### DM-ONLY Content

Este documento es **exclusivamente para el DM**. Debe contener:
- ✅ Spoilers y plot twists
- ✅ Información oculta que los jugadores no saben
- ✅ Secretos de facciones
- ✅ Motivaciones reales de NPCs (pueden diferir de lo que muestran)
- ✅ Consecuencias no obvias de decisiones

### Secrets en Capas

**What Everyone Knows:**
> "El Archimago Valdris desapareció hace 50 años."

**What Some Know:**
> "Valdris fue visto por última vez entrando en su torre con un extraño encapuchado."

**What Nobody Knows:**
> "Valdris no desapareció — fue asesinado por su propio aprendiz, quien ahora ocupa su lugar usando un hechizo de ilusión."

### Faction Depth

Cada facción debe tener:
- **Alignment:** Ethos general (no necesariamente D&D alignment)
- **Leader:** Nombre del líder (con referencia a NPCs)
- **Members:** Quién se une y por qué (motivación)
- **Goals:** Qué quieren lograr (concreto)
- **Resources:** Qué tienen (dinero, conexiones, ejércitos, información)
- **Secret:** Qué ocultan (algo que, si se revela, cambiaría la percepción)

## WotC Quality Validators

### ValidateSettingDepth
- ✅ Geography con locations detalladas (type, population, government)
- ✅ History con recent events y ancient history
- ✅ Culture con structure, religion, customs, crime

### ValidateFactionComplexity
- ✅ 2-4 facciones con goals claros
- ✅ Cada facción tiene secret oculto
- ✅ Reputation tiers definidos (3 niveles)

### ValidateSecretLayers
- ✅ 3 capas de secretos (Everyone/Some/Nobody)
- ✅ Secrets son significativos (afectan trama)
- ✅ Nobody Knows incluye plot twists

### ValidateDMUtility
- ✅ Economía con precios notables para improvisación
- ✅ Weather table para random encounters
- ✅ NPCs Who Matter con referencias a Appendix B

## Error Handling

Si la validación falla:

1. **Analizar feedback específico** (ej: "facción contradice lore.md")
2. **Corregir issues concretos** (ajustar goals para respetar canon)
3. **Re-validar** con contenido corregido
4. **Máximo 3 reintentos** — si falla, abortar y reportar

## Output al Architect

```markdown
## Setting Guide Generado: {campaign_name}

**Status:** ✅ Complete / ❌ Failed

**Contenido:**
- Locations: {count} major locations
- Landmarks: {count} landmarks
- Factions: {count} facciones
- NPCs Who Matter: {count} NPCs
- Secrets: {count} en cada capa

**Validación:**
- validate_canon: ✅ Passed
- process_consistency_gate: ✅ Passed
- ValidateSettingDepth: ✅ Passed

**Consistencia:**
- Locations de lore.md respetadas: {count}
- Facciones de npcs.md alineadas: {count}
- NPCs referenciados existen: {count}
```
