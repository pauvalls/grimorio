---
name: grimorio-areas
version: "1.0.0"
description: DEPRECATED — Use grimorio-chapters for new campaigns. Generate WotC-style numbered playable areas (10-15 per act) with DCs, treasure, and mechanics
---

# grimorio-areas — Dungeon Master Designer (DEPRECATED)

> **DEPRECATION NOTICE**: This skill is deprecated for new campaigns. Use `grimorio-chapters` instead — see `skills/grimorio-chapters/SKILL.md` for the v2 rule set (21 reglas: 14 heredadas + 7 WotC nuevas). Legacy campaigns with `areas/` directories continue to work via backwards compatibility.
>
> **Migration**: see `skills/grimorio-chapters/SKILL.md` for the v2 rule set (21 reglas: 14 heredadas + 7 WotC nuevas). New campaigns should use the `save_chapter` tool and `grimorio-chapters` skill. Old campaigns with `areas/` are unaffected.

## Template Requerido

**ANTES de generar contenido, LEER el template:**

```
get_template(type="areas")
```

El template define el formato WotC obligatorio para áreas numeradas.

## Herramientas Disponibles

**MCP Tools (USAR para guardar contenido):**
- `save_areas` — Guardar capítulos de áreas
- `validate_canon` — Validar contra canon.json
- `get_template` — Obtener template WotC
- `check_consistency` — Chequeo de consistencia
- `process_consistency_gate` — Validación batch con auto-retry

**NO usar Write para contenido creativo** — El frontmatter del agente ya no incluye Write para forzar el uso de MCP save tools.

## Workflow Obligatorio

```
1. LEER contexto:
   - canon.json (hechos canónicos, entidades)
   - lore.md (tono, conflicto, setting)
   - npcs/npcs_and_factions.md (nombres exactos de NPCs)
   - bestiary/bestiary.md (nombres exactos de criaturas)
   - encounters/encounters.md (encuentros disponibles)
   - maps/maps.md (localizaciones)
   - narrative_state.json (estado actual)

2. LEER template:
   - get_template(type="areas")

3. GENERAR áreas siguiendo el template:
   - 10-15 áreas por acto (one-shot: 8-12)
   - 150-200 palabras por área
   - Formato WotC estricto

4. VALIDAR antes de guardar:
   - validate_canon() con entity_references
   - process_consistency_gate() para validación batch
   - Máximo 3 reintentos si falla

5. GUARDAR solo si validación pasa:
   - save_areas(campaign, content)

6. REPORTAR al architect
```

## Formato WotC Obligatorio

### Regla 13: Chapter Opener (Apertura del Capítulo)

Cada acto DEBE comenzar con una sección "Apertura del Capítulo" antes de "Adventure Background". Esta sección incluye:

#### 13.1 Modo de Juego

Selecciona el modo según la posición del acto en la campaña:

| Posición | Modo Recomendado | Justificación |
|----------|------------------|---------------|
| Acto 1 | dungeon_lineal o investigacion | Enganche rápido, primer éxito |
| Acto 2 | sandbox_urbano o downtime | Expansión, base building |
| Acto 3-N | variado | No repetir modo más de 2 veces consecutivas |
| Acto Final | confrontacion | Clímax combativo o dramático |

**Modos válidos** (usar exactamente así, lowercase con underscores):
- investigacion
- sandbox_urbano
- dungeon_lineal
- escape
- viaje
- intriga
- confrontacion
- downtime

**Modo híbrido**: Si el capítulo combina dos modos (ej: investigación + dungeon), declarar como "investigacion + dungeon_lineal". El primer modo es el dominante.

**Override**: Si la narrativa requiere 3+ actos consecutivos con el mismo modo (ej: mega-dungeon), justificar explícitamente en Running Guidance.

#### 13.2 Objetivos del Capítulo

Generar 2-3 objetivos que:
- Comienzan con verbo de acción (Rescatar, Obtener, Descubrir, Derrotar, Explorar)
- Son específicos (se sabe cuándo se completaron)
- Al menos uno conecta con el arco de campaña (no solo local)

**Ejemplo**:
```markdown
**Objetivos del Capítulo:**
- Rescatar a Floon Blagmar de los capos del crimen
- Obtener la recompensa de Renaer Neverember
- Descubrir la primera pista sobre la Piedra de Golorr
```

#### 13.3 Duración Estimada

Formato obligatorio:
- `1 sesión` (exactamente así, singular)
- `2-3 sesiones` (rango numérico)
- `4-6 sesiones`

**NO usar**: "varias sesiones", "largo", "2 horas"

#### 13.4 Running Guidance

Escribir 2-4 párrafos (150-400 palabras) que incluyan:

**Párrafo 1**: Estructura del capítulo (cómo se conectan las áreas, flujo)
**Párrafo 2**: Énfasis de tono (qué resaltar, beats emocionales)
**Párrafo 3**: Puntos de decisión clave (momentos de ramificación con consecuencias)
**Párrafo 4** (opcional): Transición al siguiente capítulo

**Ejemplo**:
```markdown
**Running this Chapter:**
Este capítulo combina investigación urbana (Calles de Waterdeep) con un dungeon lineal (Villa Gralhund). Comienza con el gancho en el Portal Bramante, luego permite a los PJs seguir pistas por la ciudad antes del asalto final.

Énfasis en el contraste: la elegancia de la villa noble vs. la brutalidad del sótano donde retienen a Floon. Los PJs deben sentir que están escalando de "aventureros buscando trabajo" a "jugadores en un conflicto de facciones".

Si los PJs matan a los Gralhund, registra que la familia está "eliminada" para consecuencias en actos posteriores.
```

#### 13.5 Asset Handoff

Cada acto DEBE producir un asset concreto que sea prerequisito del siguiente acto.

**Tipos de Asset** (usar exactamente así):

| Tipo | Ejemplo | Validación |
|------|---------|------------|
| **Objeto** | mapa, llave, carta de presentación, diario, arma mágica | Debe ser físico, concreto |
| **Información** | ubicación del villano, debilidad del boss, identidad del traidor | Debe ser conocimiento específico |
| **Aliado** | NPC que guía al siguiente capítulo, facción que ofrece apoyo | Debe nombrar al NPC/facción |
| **Base** | taverna, barco, refugio, guarida | Debe ser ubicación que los PJs controlan |

**Formato**:
```markdown
**Asset para el Siguiente Capítulo:**
- **Tipo:** Nombre del asset (propósito para el siguiente acto)
```

**Ejemplo**:
```markdown
**Asset para el Siguiente Capítulo:**
- **Objeto:** Escritura de la Taberna Trollskull (prerrequisito para Acto 2 sandbox)
- **Información:** Mención de "Xanathar" en documentos (gancho para Acto 3)
```

**NO usar**: "experiencia", "amistad", "confianza" (son vagos, no concretos)

#### 13.6 Variedad de Modos

**Regla**: NO generar más de 2 actos consecutivos con el mismo modo primario.

**Algoritmo de selección**:
```
SI Acto N-1.mode == Acto N-2.mode:
  ENTONCES Acto N.mode DEBE SER diferente
SINO:
  Acto N.mode PUEDE SER igual a Acto N-1.mode
```

**Excepción**: Campañas de mega-dungeon o viaje épico pueden tener 3+ consecutivos si se justifica en Running Guidance.

**Ejemplo de justificación**:
```markdown
**Running this Chapter:**
**Override de Variedad:** Este es el tercer acto consecutivo de tipo dungeon_lineal. Justificación: campaña de mega-dungeon (Undermountain-style) donde cada nivel del dungeon es un acto. La variedad viene de los tipos de encuentros (combate, puzzle, social con facciones del dungeon) no del modo de exploración.
```

### Regla 14: Edge Cases en Chapter Opener

#### 14.1 Capítulo sin NPCs nuevos

Si el capítulo es dungeon crawl o exploración sin NPCs nuevos:
- Running Guidance enfatiza **storytelling ambiental** en lugar de interacciones
- Asset handoff debe ser objeto o información (no aliado)
- Modo debe reflejar contenido (dungeon_lineal o exploracion)

#### 14.2 TPK (Total Party Kill)

Si hay riesgo de TPK en el capítulo:
- Running Guidance incluye **contingencia TPK** (cómo continuar)
- Opciones: resurrection quest, nueva party hereda misión, campaña termina
- Asset handoff puede transferirse a nueva party o devenir MacGuffin

#### 14.3 Capítulo Salteado (Sequence Breaking)

Si los PJs encuentran atajo:
- Running Guidance incluye **contingencia skip** (contenido mínimo que debe ocurrir)
- Asset handoff debe reubicarse (DM lo mueve al área final)
- Consecuencias de saltar el dungeon se reflejan en el mundo

#### 14.4 Campaña One-Shot (1 solo acto)

- Asset handoff se convierte en **epílogo** en lugar de puente
- Regla de variedad no aplica (solo 1 acto)
- Running Guidance incluye notas de cierre de campaña

#### 14.5 Campaña Reanuda después de Hiatus

- Running Guidance incluye **recap hook** (1 párrafo resumiendo acto anterior)
- Asset handoff reafirma explícitamente qué tienen los PJs

### Checklist de Validación de Chapter Opener

Antes de guardar cada acto, verificá:

- [ ] **Modo de Juego**: ¿Es uno de los 8 modos canónicos?
- [ ] **Variedad de Modos**: ¿Este acto + los 2 anteriores tienen modos diferentes (o hay override justificado)?
- [ ] **Objetivos**: ¿Son 2-3 objetivos con verbos de acción?
- [ ] **Duración**: ¿Formato correcto ("1 sesión" o "X-Y sesiones")?
- [ ] **Running Guidance**: ¿150-400 palabras? ¿2-4 párrafos?
- [ ] **Asset Handoff**: ¿Es concreto (objeto/información/aliado/base)?
- [ ] **Cadena de Assets**: ¿El asset de este acto se menciona en el acto siguiente?
- [ ] **Alineación Modo-Contenido**: ¿El modo coincide con los tipos de áreas?

### Estructura de cada Acto

```markdown
# Acto N: Título Descriptivo

> **Nivel:** X-Y | **Duración:** X-Y horas | **Tono:** Descriptivo
> **Objetivo:** Qué deben lograr los PJs
> **Fallo:** Qué pasa si no lo logran

## Apertura del Capítulo

**Modo de Juego:** {investigacion|dungeon_lineal|sandbox_urbano|escape|viaje|intriga|confrontacion|downtime}

**Objetivos del Capítulo:**
- Verbo de acción + objetivo específico
- Verbo de acción + objetivo específico
- Verbo de acción + objetivo específico

**Duración Estimada:** X sesiones

**Running Guidance:**
2-4 párrafos (150-400 palabras) explicando estructura, énfasis de tono, y puntos de decisión.

**Asset para el Siguiente Capítulo:**
- **Tipo:** Nombre del asset (propósito)

## Adventure Background

[Contexto para el DM]

## Áreas Numeradas

### Área X: Nombre Descriptivo

>> **Texto para Leer:** *2-4 párrafos, segunda persona presente, 100-600 palabras total. Solo detalles sensoriales.*

**Descripción para el DM:**
3-5 párrafos con DCs específicos:
- **Percepción DC XX:** Qué notan
- **Investigación DC XX:** Qué descubren
- **Arcano/Religión/Naturaleza DC XX:** Conocimiento especializado

**Criaturas:**
- Cantidad + **Nombre Exacto** del bestiary.md

**Treasure:**
- **XP total:** XXX XP
- **Moneda:** XX gp, XX sp
- **Objetos:** **Nombre** (efecto breve)

**Connections:**
- → Área X (dirección, descripción)
- ← Área Y (bidireccional obligatorio)

**Secrets y Trampas:**
- **Detectar:** Percepción DC XX
- **Mecanismo:** Cómo funciona
- **Consecuencia:** Damage/efecto
- **Desactivar:** Herrería/Juego de Manos DC XX

**Desarrollo:**
- **Si [acción]:** [consecuencia]
  - *Recuperación:* [alternativa si falla]

**Decision Points:**
- **IF** los PJs [acción concreta], **THEN** [consecuencia explícita]
  - **Affects:** Área X, Acto N
  - **World State:** [NPCs, facciones, pistas, quests afectados]

**Character Hooks:**
- **[Background/Class]:** [Gancho accionable]

**Cómo Dirigir esta Escena:**
1. **Preparación:** Qué necesita el DM
2. **Ritmo Sugerido:** Timing estimado
3. **Señales de los Jugadores:** Cuándo están enganchados
4. **Cuándo Improvise:** Dónde está OK desviarse
5. **Cuándo ceñirse al Guión:** Elementos NO negociables
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
        id: "act-{n}",
        type: "act",
        content: "Resumen del acto...",
        entity_references: [
          { entity_id: "npc-001", location: "act_{n}" },
          { entity_id: "monster-001", location: "act_{n}" },
          { entity_id: "location-001", location: "act_{n}" }
        ]
      }
    )
    
    IF result.status == "approved":
        validation_passed = true
    ELSE:
        retry_count += 1
        # Corregir issues específicos del feedback
        Fix issues based on result.feedback
    
IF validation_passed:
    save_areas(campaign="{campaign_name}", content=...)
ELSE:
    Report failure: "Validation failed after 3 retries"
    DO NOT save content
```

## WotC Quality Validators

El agente DEBE auto-validar antes de guardar:

### ValidateDevelopments
- ✅ Sección `### Developments` presente
- ✅ Mínimo 3 ramas de decisión
- ✅ Estructura IF-THEN explícita
- ✅ Al menos 1 recovery path

**Ejemplo**:
```markdown
### Developments

**Si los PJs [acción concreta]:**
- **Consecuencia inmediata:** [qué pasa ahora]
- **Consecuencia futura:** [qué pasa en área X o acto N]
- **Recuperación:** [cómo continuar si falla]

**Si los PJs [otra acción]:**
- **Consecuencia inmediata:** [...]
- **Consecuencia futura:** [...]
- **Recuperación:** [...]

**Si los PJs [tercera acción]:**
- **Consecuencia inmediata:** [...]
- **Consecuencia futura:** [...]
- **Recuperación:** [...]
```

### ValidateMultipleSolutions
- ✅ Mínimo 2 paths (stealth/social/combat)
- ✅ DCs NUMÉRICOS
- ✅ Consecuencias para cada path

**Ejemplo**:
```markdown
### Soluciones Alternativas

- **Sigilo:** CD 14 Sigilo para evitar encuentro
- **Social:** CD 13 Persuasión para negociar
- **Combate:** 2 Guardias si atacan
```

### ValidateCharacterHooks
- ✅ Sección `### Character Hooks` presente
- ✅ 2-3 hooks por área
- ✅ Hooks atados a background/class/race/faction

**Ejemplo**:
```markdown
### Character Hooks

- **[PJ Name] ([Background] background):** [Gancho específico conectado a su pasado]
- **[PJ Name] ([Class] class):** [Gancho conectado a su rol/clase]
```

### ValidateBoxedText
- ✅ Formato `>> **Texto para Leer**`
- ✅ 100-600 palabras
- ✅ Segunda persona (ves, escuchas, sientes)
- ✅ Presente (está, hay, son)
- ✅ Sin mecánicas ni spoilers

**Ejemplo**:
```markdown
>> **Texto para Leer:** *El aire húmedo golpea tu rostro. Escuchas gotas cayendo en la distancia. 
Ves formaciones rocosas que parecen figuras retorcidas. La luz de tus antorchas revela pasajes 
que se adentran en la oscuridad.*
```

### ValidateDecisionPoints
- ✅ 3+ decision points por acto
- ✅ Estructura IF-THEN explícita
- ✅ Cross-area propagation documentada
- ✅ World state changes registered

**Ejemplo**:
```markdown
**Decision Points:**
- **IF** los PJs matan al recolector, **THEN** alarma se activa, guardias llegan en 1d4 rondas
  - **Affects:** Área 5 (guardias alertados +2 a Percepción), Acto 2 (Tobias muerto)
  - **World State:** NPCs: Recolector (muerto), Facciones: Guardia (+10), Pistas: clue-003 revelada
- **IF** los PJs sobornan al recolector, **THEN** proporciona información del almacén
  - **Affects:** Área 4 (acceso libre), Facción Contrabandistas (+10 reputación)
  - **World State:** Facciones: Contrabandistas (+10), Pistas: clue-003 (ubicación almacén)
```

### ValidateNPCDescriptions
- ✅ 5+ párrafos de descripción total
- ✅ Appearance física: 3-5 párrafos (rasgos distintivos, vestimenta, expresiones)
- ✅ Personality/voz: 2-3 párrafos (mannerisms, speech patterns, quirks)
- ✅ 3-5 secretos por NPC clave
- ✅ 3-5 líneas de diálogo sample

### ValidateRunningGuidance
- ✅ 150-400 palabras
- ✅ 2-4 párrafos
- ✅ 5 subsecciones: Preparación, Ritmo, Señales, Improvisación, Guión

## Checklist Pre-Guardado (v2.4 WotC)

- [ ] **Prólogo:** Capítulo 1 tiene prólogo de 400-600 palabras (Actos 2+ NO llevan)
- [ ] **Decision Points:** 3+ puntos de decisión con IF-THEN por acto
- [ ] **Cross-area propagation:** Consecuencias documentadas explícitamente
- [ ] **World State:** Cambios registrados (NPCs, facciones, pistas, quests)
- [ ] **Áreas:** 10-15 áreas numeradas (one-shot: 8-12)
- [ ] **Word Count:** 150-200 palabras por área
- [ ] **Boxed Text:** 2-4 párrafos, 100-600 palabras, 2da persona, presente
- [ ] **Character Hooks:** 2+ hooks por área atados a background/class
- [ ] **Developments:** 3+ ramas con recovery paths
- [ ] **Cómo Dirigir:** 5 subsecciones presentes
- [ ] **NPCs:** 5+ párrafos descripción, 3-5 secretos, 3-5 diálogos
- [ ] **Treasure:** XP numérico explícito por área con criaturas
- [ ] **DCs:** Todos NUMÉRICOS (nunca "alto/bajo")
- [ ] **Nombres:** Todos existen en bestiary.md/npcs.md
- [ ] **Connections:** Bidireccionales verificadas
- [ ] **Mode Variety:** No más de 2 actos consecutivos con mismo modo
- [ ] **Asset Chain:** Asset de este acto referenciado en acto siguiente
- [ ] **Running Guidance:** 150-400 palabras, 2-4 párrafos

## Cross-References Format

**OBLIGATORIO usar enlaces markdown, NO texto plano:**

```markdown
❌ MAL: (Ver NPCs: Pipián)
✅ BIEN: [Pipián](npcs/npcs_and_factions.md#pipián)

❌ MAL: (Bestiario: Matón)
✅ BIEN: [Matón](bestiary/bestiary.md#matón)

❌ MAL: → Quest S1: Lúpulo de Verdad
✅ BIEN: → [Quest S1: Lúpulo de Verdad](quests/quests.md#s1)
```

## WotC Quality Validators

El agente DEBE auto-validar antes de guardar:

### ValidateDevelopments
- ✅ Sección `### Developments` presente
- ✅ Mínimo 3 ramas de decisión
- ✅ Estructura IF-THEN explícita
- ✅ Al menos 1 recovery path

### ValidateMultipleSolutions
- ✅ Mínimo 2 paths (stealth/social/combat)
- ✅ DCs NUMÉRICOS
- ✅ Consecuencias para cada path

### ValidateCharacterHooks
- ✅ Sección `### Character Hooks` presente
- ✅ 2-3 hooks por área
- ✅ Hooks atados a background/class/race/faction

### ValidateBoxedText
- ✅ Formato `>> **Texto para Leer**`
- ✅ 100-600 palabras
- ✅ Segunda persona (ves, escuchas, sientes)
- ✅ Presente (está, hay, son)
- ✅ Sin mecánicas ni spoilers

## Error Handling

Si la validación falla:

1. **Analizar feedback específico** del validator
2. **Corregir issues concretos** (ej: "NPC X no existe en canon" → usar NPC Y)
3. **Re-validar** con contenido corregido
4. **Máximo 3 reintentos** — si falla, abortar y reportar al architect

## Output al Architect

```markdown
## Áreas Generadas: {campaign_name}

**Status:** ✅ Complete / ❌ Failed

**Actos:**
- Acto 1: {X} áreas, {Y} palabras, {Z} decision points
- Acto 2: {X} áreas, {Y} palabras, {Z} decision points

**Validación:**
- validate_canon: ✅ Passed
- process_consistency_gate: ✅ Passed

**Cross-References:**
- NPCs referenciados: {count} (todos existen en npcs.md)
- Criaturas referenciadas: {count} (todos existen en bestiary.md)
- Quests referenciadas: {count}
```
