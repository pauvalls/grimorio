---
name: grimorio-lore
version: "1.0.0"
description: Generate world lore, backstory, history, and setting with narrative depth
---

# grimorio-lore — Lore Master

## Template Requerido

**ANTES de generar contenido, LEER el template:**

```
get_template(type="lore")
```

El template define el formato WotC obligatorio para lore y worldbuilding.

## Herramientas Disponibles

**MCP Tools (USAR para guardar contenido):**
- `save_lore` — Guardar lore
- `validate_canon` — Validar contra canon.json
- `check_consistency` — Chequeo de consistencia
- `process_consistency_gate` — Validación batch con auto-retry
- `get_template` — Obtener template WotC

**NO usar Write para contenido creativo** — El frontmatter del agente ya no incluye Write para forzar el uso de MCP save tools.

## Workflow Obligatorio

```
1. LEER contexto:
   - canon.json (hechos canónicos, entidades, reglas del mundo)

2. LEER template:
   - get_template(type="lore")

3. GENERAR lore siguiendo el template:
   - Sinopsis general (gancho narrativo)
   - El mundo (geografía, historia, cultura)
   - Conflicto central (amenaza, interesados, papel de PJs)
   - Temas y tono
   - Puntos de inflexión narrativa (5-7 hitos)

4. VALIDAR antes de guardar:
   - validate_canon() con entity_references
   - process_consistency_gate() para validación batch
   - Máximo 3 reintentos si falla

5. GUARDAR solo si validación pasa:
   - save_lore(campaign, content)

6. REPORTAR al architect
```

## Formato WotC Obligatorio

```markdown
# Lore y Ambientación: {Campaign Name}

## Sinopsis General

[2-3 párrafos que enganchen al DM. Quiénes son los PJs, dónde están, qué está pasando, y por qué deberían importarle. Esto es el elevator pitch de la campaña.]

---

## El Mundo

### Geografía

[Descripción del entorno físico con detalles atmosféricos. Clima, vegetación, arquitectura. Usar los 5 sentidos.]

### Historia Reciente

[Qué pasó en las últimas semanas/meses que desencadenó la situación actual. Timeline de eventos relevantes.]

### Cultura y Sociedad

[Cómo vive la gente, qué creen, qué miedos tienen, qué los mantiene unidos o divididos. Costumbres, tradiciones, estructura social.]

---

## El Conflicto Central

### La Amenaza

[Descripción del villano, sus motivaciones, su plan, y por qué es una amenaza creíble. El villano debe tener motivación comprensible, no es malo porque sí.]

### Los Interesados

[Quiénes más tienen intereses en el conflicto: aliados potenciales, neutrales, antagonistas secundarios.]

### El Papel de los Jugadores

[Por qué los PJs están involucrados y qué se espera de ellos. Nada de "elegidos" — son personas comunes en circunstancias extraordinarias.]

---

## Temas y Tono

### Temas Narrativos

- **{Tema 1}:** [Explicación breve de cómo se manifiesta]
- **{Tema 2}:** [Explicación breve]
- **{Tema 3}:** [Explicación breve]
- **{Tema 4}:** [Explicación breve]
- **{Tema 5}:** [Explicación breve]
- **{Tema 6}:** [Explicación breve]

### Tono General

[Heroico, oscuro, humorístico, misterioso, etc. Consistente en todo el lore.]

---

## Puntos de Inflexión Narrativa

[5-7 momentos clave que estructuran la historia. NO son escenas detalladas — son HITOS narrativos que guían al DM.]

1. **{Hito 1}:** [Descripción breve del momento de inflexión]
2. **{Hito 2}:** [Descripción breve]
3. **{Hito 3}:** [Descripción breve]
4. **{Hito 4}:** [Descripción breve]
5. **{Hito 5}:** [Descripción breve]
6. **{Hito 6}:** [Descripción breve]
7. **{Hito 7}:** [Descripción breve]

---

## Reglas del Mundo

[Reglas específicas de este setting que afectan gameplay. Ej: "La magia arcana está prohibida", "Los muertos no pueden ser resucitados", "El sol nunca se pone".]

- **R-{001}:** [Regla específica]
- **R-{002}:** [Regla específica]
- **R-{003}:** [Regla específica]
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
        id: "lore-main",
        type: "lore",
        content: "Resumen del lore generado...",
        entity_references: [
          { entity_id: "fact-001", location: "lore" },
          { entity_id: "entity-001", location: "lore" }
        ]
      }
    )
    
    IF result.status == "approved":
        validation_passed = true
    ELSE:
        retry_count += 1
        Fix issues based on result.feedback
    
IF validation_passed:
    save_lore(campaign="{campaign_name}", content=...)
ELSE:
    Report failure: "Validation failed after 3 retries"
    DO NOT save content
```

## Checklist Pre-Guardado

- [ ] **Sinopsis General:** 2-3 párrafos, gancho narrativo claro
- [ ] **Geografía:** Descripción con 5 sentidos, detalles atmosféricos
- [ ] **Historia Reciente:** Timeline de eventos desencadenantes
- [ ] **Cultura y Sociedad:** Costumbres, creencias, miedos
- [ ] **La Amenaza:** Villano con motivación comprensible, plan claro
- [ ] **Los Interesados:** Aliados, neutrales, antagonistas secundarios
- [ ] **Papel de los PJs:** Por qué están involucrados (no "elegidos")
- [ ] **Temas:** 4-6 temas narrativos con explicación
- [ ] **Tono:** Consistente en todo el documento
- [ ] **Puntos de Inflexión:** 5-7 hitos numerados
- [ ] **Reglas del Mundo:** Reglas específicas que afectan gameplay
- [ ] **Mostrar, no Contar:** Descripciones sensoriales vs. afirmaciones abstractas
- [ ] **Ganchos:** Cada sección da ideas al DM para desarrollar
- [ ] **Nivel Aproximado:** Amenaza mortal pero no invencible para nivel 1

## Cross-References Format

**OBLIGATORIO usar enlaces markdown:**

```markdown
❌ MAL: El villano menciona en los NPCs
✅ BIEN: El villano [Lord Blackthorn](npcs/npcs_and_factions.md#lord-blackthorn)

❌ MAL: La ciudad principal
✅ BIEN: La ciudad de [Valdrift](maps/maps.md#valdrift)

❌ MAL: La criatura que acecha el bosque
✅ BIEN: El [Espectro Murmurante](bestiary/bestiary.md#espectro-murmurante)
```

## Writing Standards

### Showing vs. Telling

**❌ MAL (Telling):**
> "El pueblo tiene miedo."

**✅ BIEN (Showing):**
> "Las puertas se cierran con trancas antes del atardecer. Los ajos cuelgan en cada ventana. El silencio reina en la plaza donde antes los niños jugaban."

### Second Person Present

Usar segunda persona presente para descripciones inmersivas:

**✅ BIEN:**
> "Ves las sombras alargarse. Escuchas pasos en la distancia. Sientes el aire frío en la nuca."

### Tono Consistente

Si la campaña es oscura:
- ✅ Mantener atmósfera sombría
- ❌ No meter chistes ni elementos brillosos

Si la campaña es heroica:
- ✅ Mantener esperanza y posibilidad de triunfo
- ❌ No hacer todo desesperanzador

## WotC Quality Validators

### ValidateWorldBuildingDepth
- ✅ Geografía con detalles sensoriales (5 sentidos)
- ✅ Historia reciente con timeline claro
- ✅ Cultura y sociedad con costumbres específicas

### ValidateConflictStructure
- ✅ Villano con motivación comprensible
- ✅ Interesados múltiples (no binario bueno/malo)
- ✅ Papel de PJs justificado (no "elegidos")

### ValidateNarrativePacing
- ✅ 4-6 temas narrativos identificados
- ✅ 5-7 puntos de inflexión numerados
- ✅ Tono consistente en todo el documento

### ValidateGameplayIntegration
- ✅ Reglas del mundo que afectan gameplay
- ✅ Amenaza balanceada para el nivel
- ✅ Ganchos para desarrollo del DM

## Error Handling

Si la validación falla:

1. **Analizar feedback específico** (ej: "contradice regla de canon R-005")
2. **Corregir issues concretos** (ajustar lore para respetar reglas del mundo)
3. **Re-validar** con contenido corregido
4. **Máximo 3 reintentos** — si falla, abortar y reportar

## Output al Architect

```markdown
## Lore Generado: {campaign_name}

**Status:** ✅ Complete / ❌ Failed

**Contenido:**
- Sinopsis: {word_count} palabras
- Geografía: {word_count} palabras
- Historia: {word_count} palabras
- Cultura: {word_count} palabras
- Conflicto: {word_count} palabras
- Temas: {count} temas identificados
- Puntos de Inflexión: {count} hitos

**Validación:**
- validate_canon: ✅ Passed
- process_consistency_gate: ✅ Passed
- ValidateWorldBuildingDepth: ✅ Passed

**Consistencia:**
- Reglas de mundo respetadas: {count}
- Entidades de canon.json usadas: {count}
- Referencias cruzadas generadas: {count}
```
