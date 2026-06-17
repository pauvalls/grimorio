---
name: grimorio-chapters
version: "2.0.0"
description: Generate self-contained WotC-fidelity chapters with DM sidebars, roleplay cues, random encounter tables, per-encounter XP, multi-DC traps, faction trackers, and "What's Next?" handoffs. Absorbs the legacy grimorio-areas rule set (Reglas 1-14).
---

# grimorio-chapters — Chapter Designer (WotC Fidelity v2)

> **Cambio mayor v2**: este skill absorbe las 14 reglas legadas de `grimorio-areas/SKILL.md` y agrega 7 bloques WotC obligatorios (DM Sidebar, Roleplay, Random Encounter Table, Per-Encounter XP, Multi-DC Trap, Faction Tracker, What's Next? handoff). Las campañas existentes con `areas/` siguen funcionando vía backwards compatibility.

## Template Requerido

**ANTES de generar contenido, LEER el template:**

```
get_template(type="chapter")
```

El template define el formato WotC obligatorio para capítulos auto-contenidos. El template soporta campos en ESPAÑOL (`Título`, `ModoDeJuego`, `Duración`) y EN INGLÉS (`Title`, `GameMode`, `Duration`).

## Herramientas Disponibles

**MCP Tools (USAR para guardar contenido):**
- `save_chapter_part` — Guardar una parte del capítulo (generación secuencial)
- `finalize_chapter` — Ensamblar partes, validar y guardar capítulo final
- `save_chapter` — Guardar capítulo completo (legacy, aún soportado)
- `validate_canon` — Validar contra canon.json
- `get_template` — Obtener template WotC
- `check_consistency` — Chequeo de consistencia
- `process_consistency_gate` — Validación batch con auto-retry

**NO usar Write para contenido creativo** — El frontmatter del agente ya no incluye Write para forzar el uso de MCP save tools.

## Sequential Chapter Workflow (Recomendado)

Generar capítulos parte por parte para mejor control de calidad y coherencia narrativa:

| Parte | Nombre | Presupuesto | Contenido |
|-------|--------|-------------|-----------|
| 1 | opener | 500-800 | Header, game mode, objectives, adventure background |
| 1.5 | general-features | 100-200 | Environmental properties (ceilings, doors, light, sound) con ***Name.*** inline |
| 2 | npcs | 800-1500 | 2-5 inline NPC cards con roleplay cues |
| 3 | encounters | 800-1500 | 2-4 encounter cards con XP, tactics, alternative resolution |
| 4 | areas-1 | 1500-3000 | Áreas 1-7 (o 1-5 para capítulos pequeños) |
| 5 | areas-2 | 1500-3000 | Áreas 8-15 (o 6-10) |
| 6 | closing | 400-800 | Consequences, transition, faction tracker, What's Next |

### Pasos

```
1. LEER contexto (canon, lore, capítulo anterior, narrative_state)
2. LEER template: get_template(type="chapter")
3. GENERAR cada parte secuencialmente:
   - save_chapter_part(campaign, chapter_number, part_name, content)
   - Usar parts_received y accumulated_words del response para trackear progreso
   - Mantener continuidad narrativa (NPCs, áreas, encuentros de partes anteriores)
4. finalize_chapter(campaign, chapter_number, title) → ensambla, valida, guarda
```

### Prologue Chapter (chapter_00)

Para el prólogo, usar `chapter_number: 0` e incluir `is_prologue: true` en el frontmatter.
Las áreas del prólogo son encuentros sociales (no requieren stat blocks de combate).
Incluir 3-5 áreas sociales, introducción de NPCs, y presentación de character hooks.

## Bilingual Support (ES/EN)

Todos los validadores aceptan marcadores en español E inglés. NO mezclar idiomas en el mismo capítulo.

| Patrón | Español | English |
|--------|---------|---------|
| Boxed text | Texto para Leer | Read-Aloud Text |
| If-then | Si [condición] | If the PCs [condition] |
| Consequence | Consecuencia | Consequence |
| Recovery | Recuperación | Recovery |
| Location | Ubicación | Location |
| Combat Stats | Estadísticas de Combate | Combat Stats |

## WotC Word Count Standards

- Área: 150-600 palabras
- Boxed text: 50-400 palabras
- Total capítulo: 3000-16000 palabras
- Áreas por capítulo: 7-15
- Áreas con letras (A1-A7, E1-E7) soportadas para capítulos urbanos

## Workflow Obligatorio (Legacy — aún soportado)

```
1. LEER contexto:
   - canon.json (hechos canónicos, entidades)
   - lore.md (tono, conflicto, setting)
   - chapters/chapter_{N-1}.md (si existe, para continuidad)
   - narrative_state.json (estado actual)

2. LEER template:
   - get_template(type="chapter")

3. GENERAR capítulo siguiendo el template:
   - Apertura narrativa (200-400 palabras)
   - NPCs inline (2-5, perfiles condensados 150-300w)
   - Encuentros (2-4, con dificultad y monstruos por nombre)
   - 10-15 áreas por capítulo (one-shot: 8-12)
   - DM Sidebars (≥ 1 por área de riesgo)
   - Roleplay cues (≥ 1 por NPC)
   - Random Encounter Tables (≥ 1 por capítulo)
   - Per-encounter XP awards
   - Multi-DC trap blocks (≥ 1 si hay trampas)
   - Faction tracker (≥ 1 entrada si hay facciones)
   - What's Next? handoff (1 al cierre)
   - Consecuencias y transición

4. VALIDAR antes de guardar:
   - validate_canon() con entity_references
   - Mínimo 10 áreas, máximo 15
   - Todos los monstruos referenciados por nombre (NO stat blocks inline)
   - ≥ 1 DM Sidebar, ≥ 1 Roleplay cue, ≥ 1 Random Encounter Table
   - Per-encounter XP presente
   - Faction tracker presente si la campaña tiene facciones
   - What's Next? handoff al final

5. GUARDAR solo si validación pasa:
   - save_chapter(campaign, chapter_number, title, content)

6. REPORTAR al architect
```

---

## Formato WotC Obligatorio — Reglas Heredadas (1-14)

> Las Reglas 1-14 absorben el set legacy de `grimorio-areas`. Aplican siempre.

### Regla 1: Capítulo Auto-Contenido

Cada capítulo DEBE ser un único archivo markdown que el DM pueda ejecutar sin abrir otros archivos.

**Estructura obligatoria:**

```markdown
# Capítulo N: Título

> **Modo de Juego:** {uno de los 8 canónicos}
> **Duración:** N sesión(es) | **Nivel:** X-Y | **Tono:** descriptivo
> **Objetivos del Capítulo:**
> - Verbo de acción + objetivo específico
> - Verbo de acción + objetivo específico
> **Asset para el Siguiente Capítulo:** tipo + nombre

## Apertura Narrativa
## Adventure Background
## NPCs en este Capítulo
## Encounters
## Areas (10-15)
## Consecuencias y Transición
## What's Next? (handoff al siguiente capítulo)
```

### Regla 2: Chapter Opener — Modo de Juego

Cada acto DEBE declarar un modo de juego canónico. Tabla de referencia:

| Posición | Modo Recomendado | Justificación |
|----------|------------------|---------------|
| Acto 1 | `investigacion` o `dungeon_lineal` | Enganche rápido, primer éxito |
| Acto 2 | `sandbox_urbano` o `downtime` | Expansión, base building |
| Acto 3-N | variado | No repetir modo más de 2 veces consecutivas |
| Acto Final | `confrontacion` | Clímax combativo o dramático |

**Modos válidos** (lowercase con underscores, usar exactamente así):
- `investigacion`
- `sandbox_urbano`
- `dungeon_lineal`
- `escape`
- `viaje`
- `intriga`
- `confrontacion`
- `downtime`

**Modo híbrido**: Si el capítulo combina dos modos (ej: investigación + dungeon), declarar como `investigacion + dungeon_lineal`. El primer modo es el dominante.

### Regla 3: Objetivos del Capítulo

Generar 2-3 objetivos que:
- Comienzan con verbo de acción (Rescatar, Obtener, Descubrir, Derrotar, Explorar)
- Son específicos (se sabe cuándo se completaron)
- Al menos uno conecta con el arco de campaña (no solo local)

```markdown
**Objetivos del Capítulo:**
- Rescatar a Floon Blagmar de los capos del crimen
- Obtener la recompensa de Renaer Neverember
- Descubrir la primera pista sobre la Piedra de Golorr
```

### Regla 4: Duración Estimada

Formato obligatorio:
- `1 sesión` (exactamente así, singular)
- `2-3 sesiones` (rango numérico)
- `4-6 sesiones`

**NO usar**: "varias sesiones", "largo", "2 horas"

### Regla 5: Adventure Background (Lore Local)

2-3 párrafos (150-300 palabras) sobre el contexto narrativo local del acto:
- ¿Por qué los PJs están aquí AHORA?
- ¿Quién o qué controla el lugar?
- ¿Qué se siente al estar aquí (tono)?

**Regla**: el background conecta con `lore.md` y NO introduce contradicciones canónicas.

### Regla 6: NPC Cards (Inline)

Cada NPC recibe una tarjeta inline con: nombre, rol/alineamiento, descripción condensada, motivación, secreto, hook para PJs.

**Cantidad**: 2-5 NPCs por capítulo (one-shot: 1-3).

**NO incluir stat blocks completos inline** — referenciar `appendices.md#appendix-b`.

```markdown
### Capitán Veleron Garra de Sal
*Legal Neutral, Humano, Guerrero Nivel 5*

Capitán del "Diente de Tiburón" desde hace 12 años. Cicatriz en la ceja
izquierda. Siempre lleva un trinkete de su madre fallecida.

**Motivación:** Proteger a su tripulación, pagar la deuda con el Gremio.
**Secreto:** Trabajó para la Perla hace 10 años; dejó el servicio.
**Hook:** Sabe la ubicación de un naufragio con tesoros.
```

### Regla 7: Encounter Cards (Inline)

Cada encuentro incluye: nombre, dificultad, monstruos por nombre, descripción, recompensas, tácticas.

**Cantidad**: 2-4 encuentros por capítulo (one-shot: 1-2).

**Incluir Per-Encounter XP** (ver Regla 16, sección WotC Nueva 2).

```markdown
### Encuentro 1: Emboscada en la Cala Oscura
*Dificultad: Medium*

3x Bandido y 1x Líder Bandido emergen de entre las rocas al paso de los PJs.

**Tácticas:** El líder usa Cargar en el primer turno; los bandidos flanquean.
**Recompensas:** 100 XP total, 25 gp, un mapa de la costa norte.
```

### Regla 8: Area Structure (WotC)

Cada área DEBE incluir:
- **Boxed text** (100-600 palabras, read-aloud)
- **DCs** para habilidades clave (Perception, Investigation, opcionalmente Arcana/Religion/Nature)
- **Criaturas** referenciadas por nombre (no stat blocks)
- **Treasure** con valor en gp
- **Desarrollo** (qué pasa después)
- **Ganchos** (≥ 2 por área)
- **DM Sidebar** (ver Regla 15, WotC Nueva 1) si el área tiene secretos

```markdown
### Área 3: La Cueva de los Títeres Rotos

> Read-Aloud: El aire huele a salitre y madera podría...
```

### Regla 9: Treasure Generation

Cada área con tesoro DEBE listar:
- Cantidad exacta en gp / sp / cp (no rangos)
- Items con nombre y descripción
- Si el item es mágico, referenciar `appendices.md#appendix-a`

**Valores de referencia por nivel del grupo:**

| Nivel Party | gp por área estándar | gp por área elite |
|-------------|----------------------|--------------------|
| 1-4 | 50-150 | 150-300 |
| 5-10 | 200-500 | 500-1000 |
| 11-16 | 800-1500 | 1500-3000 |
| 17+ | 2000-4000 | 4000+ |

### Regla 10: Developments (Consecuencias por Decisión)

Si los PJs hacen X, qué pasa. Cada decisión clave del área DEBE tener una entrada en Developments:

```markdown
**Si los PJs liberan a los prisioneros:**
- **Immediate consequence:** Los prisioneros dan información sobre la ruta de escape
- **Future consequence:** Los prisioneros llegan a Puerto Salino y difunden la historia
- **Recovery:** Si los PJs necesitan aliados, los prisioneros están en deuda
```

### Regla 11: Connections (Entre Áreas)

Cada área DEBE declarar su conexión con al menos 1 área adyacente:

```markdown
- → [Área 4: El Sótano](chapter_2.md#area-4) — Escalera descendente oculta tras las cajas
- ← [Área 2: El Muelle](chapter_2.md#area-2) — Volver por la puerta principal
```

### Regla 12: Foreshadowing (Anticipación)

Cada capítulo DEBE plantar al menos 1 semilla narrativa que pague en capítulos futuros. Las semillas pueden ser:
- Un objeto que parece decorativo pero es importante
- Un NPC menor que reaparece
- Una mención offhand que se revela después
- Un detalle del entorno que los PJs notan (o ignoran)

```markdown
**Foreshadowing (Capítulo 3):** El tatuaje del tabernero es el mismo símbolo
que se ve en las velas de los barcos de la Perla. Los PJs pueden notarlo con
Perception DC 13.
```

### Regla 13: Handoff (Transición)

Cada capítulo cierra con un handoff explícito al siguiente:
- Qué tienen los PJs ahora (asset concreto)
- Qué deben hacer después (gancho)
- Una pregunta abierta que los PJs probablemente se hagan

### Regla 14: Anti-Patterns (Prohibiciones Explícitas)

- **NO** generar stat blocks de monstruos inline — siempre por nombre
- **NO** usar el mismo NPC en 3+ capítulos sin evolución
- **NO** resolver encuentros sin opción no-combativa
- **NO** usar "lore dump" en aventuras — el lore se gana, no se entrega
- **NO** dejar áreas sin treasure o sin desarrollo
- **NO** olvidar el handoff al siguiente capítulo

---

## Formato WotC Obligatorio — Reglas Nuevas (15-21)

> Estas 7 reglas son NUEVAS en v2. Cubren los bloques estructurales que WotC usa en sus aventuras hardcover y que las campañas generadas con v1 no tenían.

### Regla 15: DM Sidebar (WotC Nueva 1)

Cada área con secretos, trampas, o información crucial DEBE incluir un `##### DM Sidebar:` blockquote. El sidebar es el bloque que el DM lee en privado y que el jugador NO ve. Contiene:

- Secrets the DM knows but the PCs do not
- Hidden NPC motivations
- Traps triggered by specific conditions
- Difficulty variations by party level
- Connections to campaign lore

**Estructura obligatoria:**

```markdown
> ##### DM Sidebar: Secretos del Área N
>
> - **Secreto #1:** La puerta está sellada con un hechizo de alarma
> - **Si los PJs intentan abrirla:** 3x Skeletons aparecen desde las sombras
> - **Motivación NPC:** Garra de Sal sabe quién es el traidor, pero calla por miedo
> - **Variación Hard Mode:** Añadir 1x Wight en el segundo piso
> - **Conexión a Campaña:** Esta es la primera pista de la Perla
```

**Cantidad**: ≥ 1 DM Sidebar por capítulo. Áreas de riesgo (trampas, NPCs hostiles) requieren sidebar obligatorio.

### Regla 16: Roleplaying Cues (WotC Nueva 2)

Cada NPC importante DEBE tener al menos una `#### Roleplaying [NPC]` cue que indique al DM:
- Tono de voz
- Catchphrase o tic verbal
- Postura / expresión física
- Cómo reacciona ante PJs que mienten, intimidan, o son amables

```markdown
#### Roleplaying Veleron Garra de Sal

- **Tono:** Voz ronca, frases cortas, nunca levanta la voz
- **Catchphrase:** "El mar no perdona, y yo tampoco"
- **Postura:** Manos en jarras, peso en una pierna
- **Si lo intimidan:** Se ríe y los reta a un combate justo
- **Si son amables:** Ofrece ron, baja la guardia lentamente
- **Si mienten:** Lo nota por los ojos; no confronta, recuerda
```

**Cantidad**: ≥ 1 roleplay cue por NPC principal.

### Regla 17: Random Encounter Table (WotC Nueva 3)

Cada capítulo DEBE incluir al menos una `> Random Encounter Table:` (en bloque) seguida de una tabla markdown con ≥ 3 filas. Las tablas cubren 3 categorías:

1. **Sea encounters** (si el capítulo es marítimo)
2. **Dungeon encounters** (si hay exploración subterránea)
3. **Time-of-day variants** (día / noche / crepúsculo)

**Estructura obligatoria:**

```markdown
> **Random Encounter Table:** Tier 1-3 (Calm Sea)
>
> | d6 | Encounter |
> |----|-----------|
> | 1 | 2x Sahuagin merodean la zona |
> | 2 | Restos de un barco hundido flotan a la deriva |
> | 3 | Una manada de delfines guía a los PJs a una cala escondida |
> | 4 | Niebla espesa reduce visibilidad a 30 ft durante 1 hora |
> | 5 | Un náufrago aferrado a un barril (lleva un mapa) |
> | 6 | 1x Merrow sale del agua y ataca sin aviso |
```

**Cantidad**: ≥ 1 tabla por capítulo.

### Regla 18: Per-Encounter XP Award (WotC Nueva 4)

Cada encuentro DEBE declarar explícitamente el XP por PC al final, con la sintaxis `**XP:** N PP`. Esto es CRÍTICO para que el DM asigne XP sin tener que calcular.

**Estructura obligatoria (dentro del bloque del encuentro):**

```markdown
### Encuentro 2: Asalto al Muelle

*Dificultad: Hard*

**XP:** 450 XP por PC (4 PJs nivel 3)

**Setup:** Los PJs llegan al muelle justo cuando la marea cambia.
```

**Cantidad**: 1 declaración de XP por encuentro. **Nunca** omitir.

### Regla 19: Multi-DC Trap Block (WotC Nueva 5)

Cada trampa en un capítulo DEBE incluir el bloque multi-DC completo con 4 sub-bloques: `Detect:`, `Save:`, `Disarm:`, `Escape:`. Esto cubre el ciclo completo de interacción con la trampa (descubrirla, evitarla, desactivarla, escapar si se activa).

**Estructura obligatoria:**

```markdown
> ##### Trampa: Dardo Envenenado en el Sarcófago
>
> - **Trigger:** Abrir el sarcófago sin desactivar primero
> - **Detect:** Perception DC 15 (nota los agujeros en la piedra)
> - **Save:** Dexterity DC 13 (esquiva los dardos)
> - **Disarm:** Thieves' Tools DC 15 (quita el mecanismo)
> - **Damage:** 2d4 piercing + 2d6 poison en fallo de save
> - **Escape:** Si los PJs están en la zona, pueden forzarse una salida
>   con Athletics DC 14 (salto por la trampilla)
```

**Cantidad**: ≥ 1 trap block multi-DC por capítulo (si hay trampas en el capítulo).

### Regla 20: Faction Tracker (WotC Nueva 6)

Cada capítulo con ≥ 2 facciones relevantes DEBE incluir una tabla de reputación de facciones. La tabla se publica al final del capítulo Y al inicio de los apéndices.

**Estructura obligatoria:**

```markdown
> **Faction Tracker (Estado al Final del Capítulo)**
>
> | Faction | Starting | Current | Delta | Trend |
> |---------|----------|---------|-------|-------|
> | Gremio de Capitanes | +10 | +12 | +2 | ↑ |
> | La Perla (Crimson Fleet) | -5 | -8 | -3 | ↓ |
> | Autoridad Portuaria | 0 | -2 | -2 | ↓ |
> | Población Local | 0 | +5 | +5 | ↑ |
```

**Cantidad**: ≥ 1 fila de facción por facción activa en el capítulo.

### Regla 21: "What's Next?" Handoff (WotC Nueva 7)

2-3 párrafos de prosa narrativa libre (100-400 palabras) que cierren el capítulo
con guidance condicional. NO usar campos estructurados. Incluir de forma natural:
- Qué tienen los PJs ahora (asset concreto)
- Una pregunta abierta sin resolver
- Lead al próximo capítulo (nombre/lugar/NPC)
- Cambio de tono emocional

**Ejemplo:**

```markdown
## What's Next?

Los PJs llevan el mapa del naufragio y saben DÓNDE buscar la Perla.
La tripulación del Diente de Tiburón los mira con recelo — algunos
saben que Garra de Sal ocultó algo durante años.

Pero ¿quién dentro de la tripulación está dispuesto a traicionar
al capitán? La pregunta flota en el aire salado del puerto.

En Puerto Salino, 'La Dama de las Olas' espera como contacto
en la taberna del Muelle Roto. El tono cambia de investigación
a confrontación directa — las espadas hablarán más que las palabras.
```

**Cantidad**: 1 al cierre de cada capítulo.

### Regla 22: General Features (WotC Ambiental)

Una sección `## General Features` por ubicación compleja, antes de las áreas.
Usar `***Name.***` bold-italic inline para sub-features ambientales: ceilings,
doors, light, sound, air, walls. Evita repetir info ambiental en cada área.

Esta sección se guarda como parte `general-features` (entre opener y npcs) en
el workflow secuencial.

**Estructura obligatoria:**

```markdown
## General Features

***Ceilings.*** 30 feet high, vaulted stone with moss.
***Doors.*** Heavy oak reinforced with iron bands.
***Light.*** Dim torchlight every 30 feet.
***Sound.*** Dripping water echoes from the east.
***Air.*** Stale and cold, smells of copper.
***Walls.*** Rough-hewn granite, slick with moisture.
```

**Referencia WotC:** LMoP Cragmaw Hideout — la entrada describe las
características generales de la cueva antes de listar las áreas individuales.

**Cantidad**: 1 por ubicación compleja (dungeon, edificio grande, complejo).
Opcional para ubicaciones simples.

---

## Checklist de Validación v2 (WotC Fidelity)

Antes de guardar, verificá que el capítulo tenga:

- [ ] **Modo de Juego:** Uno de los 8 canónicos
- [ ] **Variedad de Modos:** Cumple el límite de 2 actos consecutivos
- [ ] **Objetivos:** 2-3 con verbos de acción
- [ ] **Duración:** Formato "1 sesión" o "X-Y sesiones"
- [ ] **Adventure Background:** 150-300 palabras, conecta con lore.md
- [ ] **NPCs Inline:** 2-5 NPCs con roleplay cue (Regla 16)
- [ ] **Encuentros:** 2-4 con XP por PC explícito (Regla 18)
- [ ] **Áreas:** 10-15 (one-shot: 8-12), cada una con DM Sidebar si tiene secretos (Regla 15)
- [ ] **DM Sidebars:** ≥ 1 por capítulo (Regla 15)
- [ ] **Roleplay Cues:** ≥ 1 por NPC principal (Regla 16)
- [ ] **Random Encounter Table:** ≥ 1 (Regla 17)
- [ ] **Per-Encounter XP:** 1 por encuentro (Regla 18)
- [ ] **Multi-DC Traps:** ≥ 1 si hay trampas (Regla 19)
- [ ] **Faction Tracker:** ≥ 1 fila si hay facciones (Regla 20)
- [ ] **General Features:** 1 por ubicación compleja, antes de áreas (Regla 22)
- [ ] **What's Next? handoff:** 1 al cierre, prosa narrativa libre 100-400 palabras (Regla 21)
- [ ] **Consecuencias y Transición:** Tres ramas (éxito, fallo, decisiones clave)
- [ ] **Cross-References:** Todos los monstruos referenciados existen en bestiary
- [ ] **Asset Handoff:** Concreto (objeto / información / aliado / base)

## Cross-References Format

Los cross-references entre capítulos y apéndices usan el formato:
- `[Appendix A: Magic Items](appendices.md#appendix-a-magic-items)`
- `[Bestiary: Goblin](bestiary.md#goblin)`
- `[Area 5: La Cueva](chapter_3.md#area-5)`

Los anchor IDs DEBEN ser estables y kebab-case para que el compilador los resuelva correctamente.

## Output al Architect

Al terminar, reportá:

```
Capítulo N: {título} — guardado en campaigns/{name}/chapters/chapter_N.md
  - {X} áreas, {Y} NPCs, {Z} encuentros
  - {W} DM sidebars, {V} roleplay cues
  - XP total estimado: {N} por PC
  - Validación: ✓ canon ✓ consistencia
```

## Notas de Deprecación

Este skill REEMPLAZA a `grimorio-areas` para nuevas campañas. Las campañas existentes con `areas/` siguen funcionando (backwards compatibility en el compilador — ver WU2 de v5.0.2 que agrega warning cuando ambos directorios coexisten).
