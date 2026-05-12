---
name: grimorio-characters
version: "1.0.0"
description: Generate pre-generated player characters with balanced builds and narrative hooks
---

# grimorio-characters — Character Builder

## Template Requerido

**ANTES de generar contenido, LEER el template:**

```
get_template(type="character")
```

El template define el formato WotC obligatorio para fichas de personaje.

## Herramientas Disponibles

**MCP Tools (USAR para guardar contenido):**
- `generate_character` — Generar personaje con stats y habilidades
- `get_character` — Obtener ficha de personaje existente
- `list_characters` — Listar todos los personajes de la campaña
- `validate_canon` — Validar contra canon.json
- `check_consistency` — Chequeo de consistencia
- `process_consistency_gate` — Validación batch con auto-retry

**NO usar Write para contenido creativo** — El frontmatter del agente ya no incluye Write para forzar el uso de MCP save tools.

## Workflow Obligatorio

```
1. LEER contexto:
   - canon.json (reglas del mundo, clases/magia permitidas)
   - lore.md (tono, setting, conflicto)
   - npcs/npcs_and_factions.md (NPCs para conectar backstories)

2. LEER template:
   - get_template(type="character")

3. GENERAR personajes siguiendo el template:
   - 4-6 personajes pregenerados
   - Variedad de clases y backgrounds
   - Backstory con ganchos narrativos

4. VALIDAR antes de guardar:
   - validate_canon() con entity_references
   - process_consistency_gate() para validación batch
   - Máximo 3 reintentos si falla

5. GUARDAR personajes:
   - generate_character() para cada personaje
   - Archivos markdown en {campaign_path}/characters/

6. GENERAR character hooks:
   - generate_character_hooks() para hooks automáticos
   - Guardar en quests/character-hooks.md

7. REPORTAR al architect
```

## Formato WotC Obligatorio

### Estructura de cada Personaje

```markdown
# {Character Name}

## Información Básica

**Nombre:** {Full Name}

**Raza:** {Race}

**Clase:** {Class}

**Nivel:** 1

**Alineamiento:** {LG|NG|CG|LN|N|CN|LE|NE|CE}

**Trasfondo:** {Background}

---

## Estadísticas

**Array Estándar:** 15, 14, 13, 12, 10, 8 (point buy)

| FUE | DES | CON | INT | SAB | CAR |
|-----|-----|-----|-----|-----|-----|
|  X  |  X  |  X  |  X  |  X  |  X  |

**CA:** XX

**PG:** XX (XdX + X)

**Velocidad:** XX ft.

**Iniciativa:** +X

---

## Competencias

**Salvaciones:** [Skill +X, ...]

**Habilidades:** [Skill +X, ...]

**Herramientas:** [Herramientas, ...]

**Armas:** [Armas, ...]

**Idiomas:** [Idiomas, ...]

---

## Equipo

- [Equipo inicial según clase y trasfondo]
- [Arma principal]
- [Equipo de exploración]
- [Objeto personal del background]

---

## Habilidades y Hechizos

### Rasgos Raciales

- **{Trait Name}:** [Descripción]

### Rasgos de Clase

- **{Feature Name}:** [Descripción]

### Hechizos (si aplica)

**Trucos:**
- {Cantrip Name}: [Descripción breve]

**Nivel 1 (X slots):**
- {Spell Name}: [Descripción breve]
- {Spell Name}: [Descripción breve]

---

## Apariencia y Personalidad

**Edad:** XX años

**Altura:** X'XX"

**Peso:** XXX lbs

**Ojos:** [Color, características]

**Pelo:** [Color, estilo]

**Piel:** [Tono, características]

### Rasgos de Personalidad

1. [Rasgo 1]
2. [Rasgo 2]

### Ideal

[Lo que más valora, qué lo impulsa]

### Vínculo

[Conexión con alguien o algo. DEBE conectarse con NPCs de la campaña]

### Defecto

[Debilidad, miedo, o flaw que puede ser explotado]

---

## Backstory

[1-2 párrafos de historia. Debe incluir:]

1. **Un gancho conectado al conflicto principal** de la campaña
2. **Una conexión con al menos 1 NPC** de npcs.md
3. **Una razón personal** para estar involucrado

---

## Ganchos de Personaje

**Para el DM:**

- **Gancho Personal:** [Cómo conectar este personaje con la trama]
- **Gancho de Background:** [Elemento del background que puede explotarse]
- **Gancho de Vínculo:** [Cómo el vínculo puede afectar la historia]
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
        id: "characters-batch",
        type: "character",
        content: "Resumen de personajes...",
        entity_references: [
          { entity_id: "npc-mentor", location: "characters" },
          { entity_id: "faction-noble", location: "characters" }
        ]
      }
    )
    
    IF result.status == "approved":
        validation_passed = true
    ELSE:
        retry_count += 1
        Fix issues based on result.feedback
    
IF validation_passed:
    # Generar cada personaje individualmente
    for character in characters:
        generate_character(
            campaign="{campaign_name}",
            name=character.name,
            race=character.race,
            class=character.class,
            level=1,
            background=character.background,
            alignment=character.alignment
        )
ELSE:
    Report failure: "Validation failed after 3 retries"
    DO NOT generate characters
```

## Checklist Pre-Guardado

- [ ] **Cantidad:** 4-6 personajes pregenerados
- [ ] **Variedad de Clases:** Guerrero, pícaro, clérigo, mago/hechicero, bardo/explorador
- [ ] **Variedad de Backgrounds:** No repetir backgrounds
- [ ] **Balance:** Todos se sienten útiles, nadie es claramente mejor/peor
- [ ] **Simplicidad Nivel 1:** Hechizos claros, habilidades simples
- [ ] **Conexión Narrativa:** Cada backstory tiene gancho con historia principal
- [ ] **Vínculos con NPCs:** Cada personaje conoce al menos 1 NPC de la campaña
- [ ] **Stats Estándar:** Array 15,14,13,12,10,8 o point buy equivalente
- [ ] **Equipo Completo:** Según clase y background
- [ ] **Personalidad:** Rasgos, ideal, vínculo, defecto completos

## Cross-References Format

**OBLIGATORIO usar enlaces markdown:**

```markdown
❌ MAL: Conoce al herrero del pueblo
✅ BIEN: Conoce a [Mastro Aldric](npcs/npcs_and_factions.md#mastro-aldrick), el herrero del pueblo

❌ MAL: Su familia fue afectada por el culto
✅ BIEN: Su familia fue afectada por el [Culto de la Sombra](npcs/npcs_and_factions.md#culto-de-la-sombra)

❌ MAL: Busca venganza contra el monstruo que atacó su aldea
✅ BIEN: Busca venganza contra el [Devorador de Almas](bestiary/bestiary.md#devorador-de-almas) que atacó [Valdrift](maps/maps.md#valdrift)
```

## Character Hooks Generation

Después de generar los personajes, DEBÉS generar hooks automáticos:

```python
generate_character_hooks(campaign="{campaign_name}")
```

**Output:** Guardar en `quests/character-hooks.md`:

```markdown
# Character Hooks

## Hooks por Personaje

### {Character Name}
**Background:** {background} | **Class:** {class}

#### Gancho Personal

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

### ValidateCharacterBalance
- ✅ 4-6 personajes con variedad de clases
- ✅ Stats balanceados (array estándar o point buy)
- ✅ Equipo apropiado para nivel 1

### ValidateNarrativeIntegration
- ✅ Cada backstory tiene gancho con trama principal
- ✅ Cada personaje tiene vínculo con al menos 1 NPC
- ✅ Vínculos referencian NPCs existentes en npcs.md

### ValidateCanonCompliance
- ✅ Clases permitidas según reglas del mundo
- ✅ Magia permitida según canon.json
- ✅ Razas disponibles según setting

### ValidatePlayability
- ✅ Hechizos y habilidades claras para nivel 1
- ✅ No abrumar con opciones
- ✅ Equipamiento completo y funcional

## Error Handling

Si la validación falla:

1. **Analizar feedback específico** (ej: "mago no permitido en este setting")
2. **Corregir issues concretos** (cambiar clase, ajustar backstory)
3. **Re-validar** con contenido corregido
4. **Máximo 3 reintentos** — si falla, abortar y reportar

## Output al Architect

```markdown
## Personajes Generados: {campaign_name}

**Status:** ✅ Complete / ❌ Failed

**Personajes:**
- Total: {count} personajes
- Clases: {list of classes}
- Backgrounds: {list of backgrounds}

**Character Hooks:**
- Hooks generados: {count}
- Archivo: quests/character-hooks.md

**Validación:**
- validate_canon: ✅ Passed
- process_consistency_gate: ✅ Passed
- ValidateCharacterBalance: ✅ Passed

**Cross-References:**
- NPCs referenciados en backstories: {count} (todos existen)
- Locaciones referenciadas: {count} (todas existen)
- Facciones referenciadas: {count} (todas existen)
```
