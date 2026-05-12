---
name: grimorio-bestiary
version: "1.0.0"
description: Generate monsters, creatures, and stat blocks with D&D 5e mechanics and tactical depth
---

# grimorio-bestiary — Bestiary Designer

## Template Requerido

**ANTES de generar contenido, LEER el template:**

```
get_template(type="monster")
```

El template define el formato WotC obligatorio para stat blocks de monstruos.

## Herramientas Disponibles

**MCP Tools (USAR para guardar contenido):**
- `save_bestiary` — Guardar bestiario
- `validate_canon` — Validar contra canon.json
- `check_consistency` — Chequeo de consistencia
- `process_consistency_gate` — Validación batch con auto-retry
- `get_template` — Obtener template WotC

**NO usar Write para contenido creativo** — El frontmatter del agente ya no incluye Write para forzar el uso de MCP save tools.

## Workflow Obligatorio

```
1. LEER contexto:
   - canon.json (reglas del mundo, entidades canónicas)
   - lore.md (tono, temática, geografía)

2. LEER template:
   - get_template(type="monster")

3. GENERAR criaturas siguiendo el template:
   - Stat blocks con formato D&D 5e oficial
   - Tácticas estructuradas round-by-round
   - Variantes y grupos de encuentro

4. VALIDAR antes de guardar:
   - validate_canon() con entity_references
   - process_consistency_gate() para validación batch
   - Máximo 3 reintentos si falla

5. GUARDAR solo si validación pasa:
   - save_bestiary(campaign, content)

6. REPORTAR al architect
```

## Formato WotC Obligatorio

### Estructura de cada Criatura

```markdown
### {Monster Name}

*{Size} {Type}, {Alignment}*

**Rol de combate:** [tank|skirmisher|controller|artillery|lurker|leader|brute|minion]

**Grupos de encuentro:** [ej: "2-3 con 1 líder", "Solitario", "Manada (1d6+2)"]

> {Descripción atmosférica de 1-3 oraciones}

**AC** XX (Armadura) | **HP** XX (XdX+X) | **Speed** XX ft.

| STR | DEX | CON | INT | WIS | CHA |
|-----|-----|-----|-----|-----|-----|
| +X  | +X  | +X  | +X  | +X  | +X  |

**Saving Throws** [Skill +X, ...]

**Skills** [Skill +X, ...]

**Damage Resistances** [tipo1, tipo2]

**Damage Immunities** [tipo1, tipo2]

**Condition Immunities** [condición1, condición2]

**Senses** [darkvision XX ft., passive Perception XX]

**Languages** [idiomas]

**Challenge** X (XXX XP)

---

### Habilidades Especiales

**{Nombre de Habilidad}.** [Descripción con mecánicas completas. Incluir DCs si aplica.]

**{Nombre de Habilidad 2}.** [Otra habilidad especial]

---

### Acciones

**{Attack Name}.** *Melee/Ranged Weapon Attack:* +X to hit, reach/range X ft., one target. *Hit:* X (XdX + X) [damage type] damage plus [effect].

**{Special Action}.** [Descripción de acción especial con mecánicas]

---

### Acciones Legendarias (si aplica)

{Legendary actions count}

**{Action Name}.** [Description]

---

### Tácticas Estructuradas

**Apertura (Rondas 1-2):**
- Posicionamiento inicial
- Primera acción prioritaria
- Uso de habilidades especiales

**Prioridades:**
1. [Primera prioridad - ej: "Separar al healer del grupo"]
2. [Segunda prioridad - ej: "Atacar al PJ con menos HP"]
3. [Tercera prioridad - ej: "Usar habilidad de área"]

**Sinergia con Aliados:**
- [Cómo interactúa con otras criaturas]
- [Buff/debuff que proporciona o recibe]

**Retirada:**
- [Condiciones de HP % para huir]
- [Condiciones de aliados caídos]
- [Objetivo conseguido]

**Variantes Tácticas:**
- **Con ventaja:** [Cómo cambia el comportamiento]
- **Con desventaja:** [Cómo cambia el comportamiento]
- **Terreno favorable:** [Aprovechamiento]
- **Terreno desfavorable:** [Adaptación]

---

### Variantes (si aplica)

#### {Variante Name}

[Diferencias con la versión base: HP, habilidades, tácticas]

---

### Lore y Ecología

**Hábitat:** [Dónde vive]

**Organización Social:** [Solitario, manada, jerarquía]

**Debilidad Explotable:** [Al menos una debilidad que los PJs puedan descubrir]

**Botín Típico:** [Qué dejan cuando son derrotados]
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
        id: "bestiary-batch",
        type: "bestiary",
        content: "Resumen del bestiario...",
        entity_references: [
          { entity_id: "monster-001", location: "bestiary" },
          { entity_id: "monster-002", location: "bestiary" }
        ]
      }
    )
    
    IF result.status == "approved":
        validation_passed = true
    ELSE:
        retry_count += 1
        Fix issues based on result.feedback
    
IF validation_passed:
    save_bestiary(campaign="{campaign_name}", content=...)
ELSE:
    Report failure: "Validation failed after 3 retries"
    DO NOT save content
```

## Checklist Pre-Guardado

- [ ] **CR Balanceado:** Criaturas apropiadas para el nivel del partido
- [ ] **Debilidades Claras:** Al menos 1 debilidad explotable por criatura
- [ ] **Acciones Variadas:** Mínimo 2 opciones en combate (no solo "ataca")
- [ ] **Tácticas Detalladas:** Apertura, prioridades, sinergia, retirada documentadas
- [ ] **Formato 5e:** AC, HP, abilities, skills, saves, senses, languages, CR, XP
- [ ] **Habilidades Especiales:** 2-4 habilidades únicas por criatura
- [ ] **Lore:** Descripción atmosférica + ecología + hábitat
- [ ] **Grupos de Encuentro:** Cómo se encuentran típicamente
- [ ] **Rol de Combate:** tank/skirmisher/controller/artillery/lurker/leader/brute/minion
- [ ] **Variantes:** Si aplica, variantes con diferencias mecánicas
- [ ] **Nombres Exactos:** Coinciden con referencias en acts/encounters.md

## Cross-References Format

**OBLIGATORIO usar enlaces markdown:**

```markdown
❌ MAL: 2 Espectros (ver Bestiario)
✅ BIEN: 2 [Espectros Murmurantes](bestiary/bestiary.md#espectro-murmurante)

❌ MAL: El boss final, un dragón
✅ BIEN: [Vorgathax el Corrupto](bestiary/bestiary.md#vorgathax-el-corrupto), dragón de sombra ancient

❌ MAL: Como se menciona en el encuentro 3
✅ BIEN: Como se menciona en [Encuentro: Emboscada en el Bosque](encounters/encounters.md#emboscada-en-el-bosque)
```

## CR Balance Guidelines

| Nivel del Partido | CR Aproximado | XP por Encuentro |
|------------------|---------------|------------------|
| Nivel 1 | CR 1/8 - 2 | 300-400 XP total |
| Nivel 2-3 | CR 2-5 | 600-900 XP total |
| Nivel 4-5 | CR 5-8 | 1200-1800 XP total |
| Nivel 6-8 | CR 8-12 | 2400-3600 XP total |
| Nivel 9-11 | CR 12-16 | 4800-7200 XP total |
| Nivel 12+ | CR 16+ | 9600+ XP total |

**Boss Final:** CR 2-3 niveles por encima del partido, con debilidades explotables y fases múltiples.

## WotC Quality Validators

### ValidateStatBlockFormat
- ✅ Todas las secciones requeridas presentes (AC, HP, abilities, actions)
- ✅ Formato de tabla de atributos correcto
- ✅ Saves y skills listados
- ✅ Senses y languages especificados

### ValidateTacticsDepth
- ✅ Apertura documentada (rondas 1-2)
- ✅ Prioridades de acción listadas (mínimo 3)
- ✅ Sinergia con aliados descrita
- ✅ Condiciones de retirada especificadas
- ✅ Variantes tácticas incluidas

### ValidateWeaknesses
- ✅ Al menos 1 debilidad explotable por criatura
- ✅ Debilidad es descubrible por los PJs
- ✅ Debilidad tiene impacto mecánico real

### ValidateEncounterGroups
- ✅ Grupos de encuentro especificados
- ✅ Rol de combate identificado
- ✅ Variantes documentadas si aplica

## Error Handling

Si la validación falla:

1. **Analizar feedback específico** (ej: "CR demasiado alto para nivel 1")
2. **Corregir issues concretos** (ajustar stats, HP, damage output)
3. **Re-validar** con contenido corregido
4. **Máximo 3 reintentos** — si falla, abortar y reportar

## Output al Architect

```markdown
## Bestiario Generado: {campaign_name}

**Status:** ✅ Complete / ❌ Failed

**Criaturas:**
- Total: {count} criaturas
- Únicas: {count} (custom para esta campaña)
- Del MM: {count} (referencia Monster Manual)

**Distribución por CR:**
- CR 1/8-2: {count} (minions, early encounters)
- CR 3-6: {count} (mid-tier threats)
- CR 7+: {count} (bosses, elite enemies)

**Validación:**
- validate_canon: ✅ Passed
- process_consistency_gate: ✅ Passed
- ValidateStatBlockFormat: ✅ Passed
- ValidateTacticsDepth: ✅ Passed

**Cross-References:**
- Criaturas referenciadas en acts: {count} (todas existen)
- Criaturas en encounters: {count} (todas existen)
```
