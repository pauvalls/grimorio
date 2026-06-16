---
name: grimorio-encounters
version: "1.0.0"
description: DEPRECATED — Use grimorio-chapters for new campaigns. Generate combat encounters, social challenges, and exploration scenes with balanced difficulty
---

# grimorio-encounters — Encounter Designer (DEPRECATED)

> **DEPRECATION NOTICE**: This skill is deprecated for new campaigns. Use `grimorio-chapters` instead, which generates self-contained chapters with inline NPCs, encounters, and areas. Legacy campaigns with `encounters/` directories continue to work via backwards compatibility.
>
> **Migration**: New campaigns should use `save_chapter` tool and `grimorio-chapters` skill. Old campaigns with `encounters/` are unaffected.

## Template Requerido

**ANTES de generar contenido, LEER el template:**

```
get_template(type="encounter")
```

El template define el formato WotC obligatorio para encuentros.

## Herramientas Disponibles

**MCP Tools (USAR para guardar contenido):**
- `save_encounters` — Guardar encuentros
- `validate_canon` — Validar contra canon.json
- `check_consistency` — Chequeo de consistencia
- `process_consistency_gate` — Validación batch con auto-retry
- `get_template` — Obtener template WotC

**NO usar Write para contenido creativo** — El frontmatter del agente ya no incluye Write para forzar el uso de MCP save tools.

## Workflow Obligatorio

```
1. LEER contexto:
   - canon.json (reglas del mundo, entidades)
   - lore.md (tono, conflicto, geografía)
   - npcs/npcs_and_factions.md (NPCs disponibles)
   - bestiary/bestiary.md (criaturas disponibles)

2. LEER template:
   - get_template(type="encounter")

3. GENERAR encuentros siguiendo el template:
   - Múltiples tipos (combate, social, exploración, puzzle)
   - Balance de dificultad progresiva
   - Múltiples paths de resolución

4. VALIDAR antes de guardar:
   - validate_canon() con entity_references
   - process_consistency_gate() para validación batch
   - Máximo 3 reintentos si falla

5. GUARDAR solo si validación pasa:
   - save_encounters(campaign, content)

6. REPORTAR al architect
```

## Formato WotC Obligatorio

### Estructura de cada Encuentro

```markdown
### {Encounter Name}

**Dificultad:** [Fácil|Medio|Difícil|Mortal]

**XP Total:** XXX XP

**Ambientación:** [Dónde y cuándo ocurre, conexión atmosférica al lore]

**Template:** [Ambush|Defense|Assault|Social|Puzzle/Trap|Chase]

---

### Objetivo

[Qué deben lograr los PJs - no siempre es "matar todo"]

---

### Enemigos / Desafíos

| Criatura | Cantidad | CR | XP |
|----------|----------|----|----|
| [Nombre del bestiary.md] | X | X | XXX |

---

### Mapa Táctico

**Dimensiones:** X × Y pies (cuadrícula de X×Y cuadros)

**Iluminación:** [Luz brillante|Luz tenue|Oscuridad]

**Terreno:** [Terreno difícil, cobertura, altura]

**Elementos Interactivos:**
| Posición | Elemento | Efecto Mecánico |
|----------|----------|-----------------|
| A1, A2 | [Nombre] | [Bonus/penalización] |

**Puntos de Inicio:**
- **PJs:** [Posición inicial]
- **Enemigos:** [Posición inicial]

---

### Condiciones y Efectos Ambientales

**Condiciones Iniciales:**
- [Niebla, oscuridad mágica, fuego, etc.]

**Cambios de Fase:**
- **Ronda X:** [Trigger y efecto]

**Efectos por Área:**
- **Zona X:** [Damage o condiciones]

---

### Desarrollo Round-by-Round

| Ronda | Enemigos | Eventos Ambientales | Condición de Victoria |
|-------|----------|---------------------|----------------------|
| 1 | [Acciones prioritarias] | [Eventos] | [Condición] |
| 2 | [Acciones prioritarias] | [Eventos] | [Condición] |
| 3+ | [Escalación, refuerzos] | [Cambios de fase] | [Condición] |

**Inicio:**
- Posición exacta de enemigos y PJs

**Ronda 1-2:**
- Qué hacen los enemigos, prioridades tácticas

**Ronda 3+:**
- Escalación, refuerzos, cambios de fase

**Condiciones de Cambio:**
- [HP %, rondas transcurridas, eventos específicos]

**Consecuencias de Acciones PJs:**
- **Si [acción X]:** [consecuencia]
- **Si [acción Y]:** [consecuencia]

---

### Resolución Alternativa

**Diplomacia/Engaño:**
- **CD:** Persuasión/Engaño DC XX
- **Requisitos:** [Qué necesitan lograr]
- **Consecuencias:** [Qué ganan/pierden vs. combate]

**Sigilo/Evasión:**
- **Rutas posibles:** [Descripción]
- **CDs:** Sigilo DC XX
- **Qué pasa si detectan:** [Consecuencia]

**Solución Creativa:**
- [Usar el entorno, objetos especiales, habilidades no-combate]

---

### Formas de Ganar

**Combate Directo:**
- Eliminación de todos los hostiles

**Resolución Alternativa:**
- Diplomacia (Persuasión/Engaño DC XX)
- Sigilo (Sigilo DC XX)

**Solución Creativa:**
- [Descripción de solución no convencional]

---

### Recompensa

**XP por PJ:** XX XP

**Botín:**
- [Objetos mágicos, dinero, información]

**Aliados Potenciales:**
- [NPCs que pueden ayudar, ventajas narrativas]

---

### Notas del DM

**Escalado:**
- [Cómo escalar para grupos grandes/chicos]

**Alternativas:**
- [Si los PJs evitan el encuentro]

**Ganchos:**
- [Para futuros encuentros]
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
        id: "encounters-batch",
        type: "encounter",
        content: "Resumen de encuentros...",
        entity_references: [
          { entity_id: "monster-001", location: "encounters" },
          { entity_id: "location-001", location: "encounters" }
        ]
      }
    )
    
    IF result.status == "approved":
        validation_passed = true
    ELSE:
        retry_count += 1
        Fix issues based on result.feedback
    
IF validation_passed:
    save_encounters(campaign="{campaign_name}", content=...)
ELSE:
    Report failure: "Validation failed after 3 retries"
    DO NOT save content
```

## Checklist Pre-Guardado

- [ ] **Tipos Variados:** No todo es combate (incluye social, exploración, puzzle)
- [ ] **Dificultad Progresiva:** Primer encuentro fácil, último difícil pero ganable
- [ ] **Múltiples Solutions:** 2+ formas de resolver (combate, diplomacia, sigilo)
- [ ] **DCs Numéricos:** Todos los DCs son números específicos (nunca "alto/bajo")
- [ ] **Mapa Táctico:** Dimensiones, iluminación, terreno, elementos interactivos
- [ ] **Round-by-Round:** Desarrollo numerado por rondas
- [ ] **Condiciones de Victoria:** Claras y alcanzables
- [ ] **Recompensas:** XP, botín, aliados documentados
- [ ] **Nombres Exactos:** Criaturas del bestiary.md, NPCs de npcs.md
- [ ] **Balance Nivel 1:** PJs tienen ~10-15 PG, un golpe crítico puede matar
- [ ] **Ambientación:** Cada encuentro tiene descripción atmosférica conectada al lore
- [ ] **Notas del DM:** Escalado, alternativas, ganchos futuros

## Cross-References Format

**OBLIGATORIO usar enlaces markdown:**

```markdown
❌ MAL: 2 Guardias (ver Bestiario)
✅ BIEN: 2 [Guardias de la Ciudad](bestiary/bestiary.md#guardia-de-la-ciudad)

❌ MAL: El encuentro ocurre en el templo
✅ BIEN: El encuentro ocurre en el [Templo de los Olvidados](maps/maps.md#templo-de-los-olvidados)

❌ MAL: Silas ofrece información
✅ BIEN: [Silas](npcs/npcs_and_factions.md#silas) ofrece información
```

## Encounter Templates

### Ambush (Emboscada)
- Enemigos ocultos → ataque sorpresa
- PJs deben reaccionar bajo presión
- Ventaja inicial para enemigos

### Defense (Defensa)
- PJs protegen objetivo/área
- Oleadas de enemigos
- Tiempo límite o recursos limitados

### Assault (Asalto)
- PJs atacan posición enemiga
- Objetivos específicos (rescatar, obtener, eliminar)
- Defensores con ventaja de posición

### Social (Social)
- Negociación, intimidación, engaño
- Consecuencias de reputación
- Múltiples paths según skills

### Puzzle/Trap (Puzzle/Trampa)
- Resolución por habilidades
- Penalización por fallo
- Pistas distribuidas

### Chase (Persecución)
- Carrera contra el tiempo
- Obstáculos progresivos
- Consecuencias de éxito/fracaso claras

## WotC Quality Validators

### ValidateMultipleSolutions
- ✅ Mínimo 2 paths diferentes (stealth/social/combat)
- ✅ DCs NUMÉRICOS para cada path
- ✅ Consecuencias documentadas para cada path

### ValidateTacticalMap
- ✅ Dimensiones especificadas (X × Y pies)
- ✅ Iluminación definida
- ✅ Terreno y cobertura descritos
- ✅ Elementos interactivos con posición y efecto
- ✅ Puntos de inicio claros

### ValidateRoundDevelopment
- ✅ Desarrollo round-by-round numerado
- ✅ Condiciones de cambio especificadas (HP %, rondas, eventos)
- ✅ Consecuencias de acciones PJs documentadas
- ✅ Condiciones de victoria claras

### ValidateBalance
- ✅ XP apropiado para el nivel del partido
- ✅ Dificultad etiquetada correctamente (Fácil/Medio/Difícil/Mortal)
- ✅ Escalado documentado para grupos grandes/chicos

## Error Handling

Si la validación falla:

1. **Analizar feedback específico** (ej: "CR demasiado alto", "falta path alternativo")
2. **Corregir issues concretos** (ajustar XP, agregar paths, modificar DCs)
3. **Re-validar** con contenido corregido
4. **Máximo 3 reintentos** — si falla, abortar y reportar

## Output al Architect

```markdown
## Encuentros Generados: {campaign_name}

**Status:** ✅ Complete / ❌ Failed

**Encuentros:**
- Combate: {count}
- Social: {count}
- Exploración: {count}
- Puzzle/Trampa: {count}

**Dificultad:**
- Fácil: {count}
- Medio: {count}
- Difícil: {count}
- Mortal: {count}

**Validación:**
- validate_canon: ✅ Passed
- process_consistency_gate: ✅ Passed
- ValidateMultipleSolutions: ✅ Passed
- ValidateTacticalMap: ✅ Passed

**Cross-References:**
- Criaturas referenciadas: {count} (todas existen en bestiary.md)
- NPCs referenciados: {count} (todos existen en npcs.md)
- Locaciones referenciadas: {count} (todas existen en maps.md)
```
