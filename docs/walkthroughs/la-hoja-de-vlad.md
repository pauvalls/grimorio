---
title: Walkthrough — La Hoja de Vlad
subtitle: End-to-end example of generating a Grimorio campaign
lang: en/es
---

<div class="lang-selector">
<a href="#english">English</a> | <a href="#espanol">Español</a>
</div>

---

<a name="english"></a>

# Walkthrough: La Hoja de Vlad

> A real-world end-to-end example of generating a Grimorio campaign —
> from a one-sentence idea to a print-ready WotC-style PDF.

`examples/la-hoja-de-vlad/` is a fully-fleshed reference campaign
embedded in this repository. It is the single best way to learn what
Grimorio produces and how the pieces fit together. This walkthrough
narrates the generation journey, points at the most illustrative
excerpts, and distills the lessons learned.

## What "La Hoja de Vlad" Is

A gothic-political 3-act campaign for characters level 1–5. The
premise, in one sentence: _"A blood-moon ritual in Vladgrad —
nobles are vampires, and the cult is hunting the party."_

The campaign ships in this repo with:

- 1 Adventure Bible (`canon.json`)
- 1 setting guide and 1 introduction
- 12 NPCs across 4 factions
- 1 bestiary with custom stat blocks
- 1 prologue (4 parts)
- 3 chapters, each with 5–8 numbered areas
- Handouts, session preps, and a flowchart
- 1 compiled PDF (`campaign.pdf`, ~1 MB)

That is 28 files. This walkthrough explains where each one came
from and why it looks the way it does.

## The Generation Journey

### 1. Brief → Adventure Bible

You start in OpenCode with the `grimorio-architect` agent and a
one-line brief. The architect asks six questions (name, type, idea,
level, tone, duration), then generates the canon in three batches.

The canon is the source of truth — every later artifact references
back to it. For *La Hoja de Vlad* the bible is
[`examples/la-hoja-de-vlad/canon.json`](../examples/la-hoja-de-vlad/canon.json):

```json
{
  "schema_version": "v2",
  "campaign_id": "la-hoja-de-vlad",
  "facts": [
    {"id": "fact-vladgrad-seal", "category": "lore",
     "statement": "Vladgrad was founded on the sealed tomb of Vlad the Impaler"}
  ],
  "entities": [
    {"id": "npc-archbishop-sergei", "name": "Arzobispo Sergei",
     "type": "npc", "canon_state": "alive",
     "motivation": "complete the blood-moon ritual"}
  ],
  "timeline": [
    {"id": "evt-blood-moon", "timestamp": "1483-10-31T23:00:00Z",
     "summary": "First blood moon in 200 years"}
  ]
}
```

The architect invokes `grimorio_save_lore`, `grimorio_save_npcs`,
and `grimorio_save_bestiary` — three MCP tools in a single batch.

### 2. Lore & Setting

`lore.md` is the player-facing world book. The setting guide
(`setting-guide.md`) is the DM-only deep dive: secret histories,
hidden motives, what the NPCs will not say out loud.

The two documents are written from different points of view. Lore is
in-world and concrete; setting guide is meta and includes spoilers.
The architect generates both and the [Campaign Consistency
gate](../campaign-consistency.md) cross-checks that no fact in `lore.md`
contradicts a fact in `canon.json`.

### 3. NPCs & Factions

NPCs come in two flavors:

1. **Canon entities** in `canon.json` (the source of truth).
2. **Stat blocks** in `npcs/npcs_and_factions.md` (the WotC-styled
   format the DM prints).

For *La Hoja de Vlad* the architect generated 12 NPCs. The Arzobispo
Sergei is the main villain; his stat block has the `mcguffin` role
because he is the one who knows where the Hoja is hidden.

### 4. Bestiary

Custom stat blocks live in `bestiary/bestiary.md`. The campaign
includes a `vampiro_noble` template (CR 5) and a `sombra_cultista`
minion (CR 1/2) — both are reused across the three chapters.

The grimorio-bestiary skill enforces the WotC 5e stat block format:
AC, HP, speed, six ability scores, saves, skills, damage
resistances/immunities, senses, languages, challenge, then
traits and actions.

### 5. Chapters & Areas

This is where the bulk of the work happens. The architect invokes
`grimorio_save_chapter` for each act, which writes a single markdown
file with **10–15 numbered areas** following the WotC format.

Excerpt from `chapter_01_sombras_en_la_corte.md` (Área 1, El
Vestíbulo de los Ancestros):

```markdown
> **Read-Aloud:** *Las puertas de roble se abren ante vosotros,
> revelando un vestíbulo iluminado por candelabros de plata...*
<!-- WOTC: boxed_text — 89 words; second person present tense;
     sensory details only. -->

**Ganchos de Personaje:**
- **Katarina (noble):** Reconocés el escudo Voronova en un tapiz.
- **Ivan (religioso):** Los símbolos son marcas contra no-muertos.
<!-- WOTC: character_hook — 2 hooks, each tied to a discoverable
     detail in the room. -->

**Desarrollos:**
1. **SI examinan los retratos:** descubren ancestro del culto.
   *Recuperación:* el Conde les muestra el retrato más tarde.
2. **SI encuentran el pergamino:** lista de víctimas del culto.
3. **SI ignoran el vestíbulo:** pierden ventaja en el banquete.
<!-- WOTC: development_branch — 3 IF-THEN with Consecuencia +
     Recuperación. -->
```

Notice the `<!-- WOTC: ... -->` annotations. They are author-side
notes for future re-readers; the compiler strips them before
rendering the PDF so they never appear in print.

### 6. Maps & Handouts

The cartographer skill generates two kinds of visual assets:

- **Battle maps** (`assets/map_*.svg`) — procedurally drawn grids
  with room labels, secret doors, and light sources.
- **Decorative dividers** (`assets/divider_*.svg`) — chapter
  separators in the WotC style.

Handouts are player-facing documents: invitations, letters, sketches.
The `grimorio_generate_handouts` tool produces them in
bilingual ES/EN so the same file works for English- and
Spanish-speaking parties.

### 7. Validation & Consistency

Before compiling the PDF, the architect runs
[`grimorio validate`](../getting-started.md#validate-your-campaign):

```bash
grimorio validate --scope=all la-hoja-de-vlad
```

The CLI runs 17 consistency rules and prints a report:

```
Campaign Validation Report
==========================
Campaign: la-hoja-de-vlad
Health: fair

❌ [warning] wotc_developments — [chapter_03] Need 3-5 IF-THEN branches
✅ [info] faction_reputation_gate — Faction Nobleza: allied (75), consistent
…

Summary: 0 errors, 3 warnings, 0 criticals (of 17 checks)
```

The architect fixes the warnings, re-runs, and proceeds only when
the report is clean (or accepts the warnings explicitly).

### 8. Compile to PDF

Finally:

```bash
grimorio compile_pdf la-hoja-de-vlad
```

The compiler turns each markdown file into HTML, applies the
WotC CSS theme, then shells out to Chromium to print to PDF.
The output is `~/campaigns/la-hoja-de-vlad/campaign.pdf`.

## Annotated Excerpts

Here are the five excerpts that best illustrate how Grimorio
thinks.

### Excerpt 1: A read-aloud blockquote (chapter 1, Área 1)

```markdown
> **Read-Aloud:** *Las puertas de roble se abren ante vosotros,
> revelando un vestíbulo iluminado por candelabros de plata...*
```

The `**Read-Aloud:**` label is removed during compile (the CSS
provides the visual indicator via a `::before` pseudo-element) and
the rest becomes a `<blockquote class="read-aloud">` in the PDF.

### Excerpt 2: A boxed-text-style scene

```markdown
**Descripción para el DM:**
El vestíbulo es una trampa psicológica. Los retratos están
encantados con magia de adivinación menor—cualquiera que los
observe por más de 1 minuto siente una presencia observándolo.

- **Percepción DC 12:** Notás sangre fresca, no óxido.
- **Investigación DC 14:** El estandarte oculta un compartimento.
- **Arcano DC 15:** Los símbolos son runas anti-posesión.
```

This is DM-only material. It is **not** in a read-aloud blockquote
because the players should not see DCs.

### Excerpt 3: A development table

```markdown
| Decisión | Condición | Consecuencia | Propagación |
|----------|-----------|--------------|-------------|
| **SI** | examinan retratos | **ENTONCES** ventaja social | Área 2, Facción Nobleza |
| **SI** | encuentran pergamino | **ENTONCES** 3 nombres | Área 5, Acto 2 |
| **SI** | ignoran todo | **ENTONCES** mayordomo hostil | Área 2 (servidores hostiles) |
```

This table format is what the grimorio-areas skill
recommends — explicit IF/THEN with cascade effects.

### Excerpt 4: A faction reputation block

```markdown
**Cambios de Estado del Mundo:**
- **NPCs:** Mayordomo (neutral → hostil si ignoran)
- **Facciones:** Sin cambios aún
- **Pistas:** clue-001 (pergamino con nombres), clue-002 (símbolos)
```

This is what the `dm_session_context` MCP tool surfaces to the
AI Dungeon Master so it can react to player decisions.

### Excerpt 5: A stat block

```markdown
### Sombra Cultista
*CE mediano no-muerto, cualquier alineamiento malvado*

- **Clase de Armadura** 13
- **Puntos de Golpe** 22 (5d8)
- **Velocidad** 9 m
- **DES** 14, **CAR** 10, **CON** 12
- **Inmunidades** a daño de veneno; resistencia a necrótico
- **Sentidos** visión en la oscuridad 18 m, Percepción pasiva 10
- **Idiomas** Común
- **Desafío** 1 (200 PX)
```

This is the WotC 5e format the grimorio-bestiary skill enforces.
The compiler renders it as a `<div class="stat-block">` in the PDF.

## Lessons Learned

1. **The canon is the spine.** Every artifact references it. When
   the architect changes a fact in `canon.json`, the lore, NPCs,
   and bestiary must update too. The validation gate catches
   drift, but it's cheaper to keep canon first.

2. **Bilingual ES/EN is a default, not an afterthought.** The
   setting guide, the introduction, and the handouts ship in
   both languages. Players pick a side; DMs read both.

3. **`<!-- WOTC: ... -->` annotations are gold.** They are how the
   author explains to future-self (or a co-author) why a section
   exists. The compiler strips them, so they never appear in the
   PDF but stay in the source.

4. **Validation is not optional.** Running `grimorio validate`
   before `compile_pdf` is the difference between a polished PDF
   and one with `{{PLACEHOLDER}}` strings in it.

5. **Start small.** The first campaign took 28 files; subsequent
   ones take 12–15 because the canon reuses NPCs and stat blocks
   across chapters. The grimorio-integrator skill batches the
   regeneration.

## Where to Go From Here

- [Getting Started](../getting-started.md) — install + first campaign.
- [MCP Tools](../features/mcp-tools.md) — full tool reference.
- [Architecture](../features/architecture.md) — how the engine fits together.
- [Campaign Consistency](../campaign-consistency.md) — the gate pipeline.
- The example campaign itself:
  [`examples/la-hoja-de-vlad/`](../examples/la-hoja-de-vlad/).

---

<a name="espanol"></a>

# Walkthrough: La Hoja de Vlad

> Un ejemplo end-to-end de cómo generar una campaña de Grimorio —
> desde una idea de una línea hasta un PDF estilo WotC listo para
> imprimir.

`examples/la-hoja-de-vlad/` es una campaña de referencia totalmente
desarrollada, embebida en este repositorio. Es la mejor manera de
aprender qué produce Grimorio y cómo encajan las piezas. Este
walkthrough narra el viaje de generación, apunta a los excerpts
más ilustrativos y resume las lecciones aprendidas.

## Qué es "La Hoja de Vlad"

Una campaña gótico-política de 3 actos para personajes nivel 1-5.
La premisa, en una frase: _"Un ritual de luna de sangre en
Vladgrad — los nobles son vampiros y el culto persigue a la
party."_

La campaña se entrega en este repo con:

- 1 Adventure Bible (`canon.json`)
- 1 setting guide y 1 introduction
- 12 NPCs en 4 facciones
- 1 bestiario con stat blocks custom
- 1 prólogo (4 partes)
- 3 capítulos, cada uno con 5-8 áreas numeradas
- Handouts, session preps y un flowchart
- 1 PDF compilado (`campaign.pdf`, ~1 MB)

Son 28 archivos. Este walkthrough explica de dónde vino cada uno
y por qué se ve como se ve.

## El Viaje de Generación

### 1. Brief → Adventure Bible

Arrancás en OpenCode con el agente `grimorio-architect` y un brief
de una línea. El architect te hace seis preguntas (nombre, tipo,
idea, nivel, tono, duración) y luego genera el canon en tres
batches.

El canon es la fuente de verdad — todo artifact posterior
referencia a él. Para *La Hoja de Vlad* la bible es
[`examples/la-hoja-de-vlad/canon.json`](../examples/la-hoja-de-vlad/canon.json):

```json
{
  "schema_version": "v2",
  "campaign_id": "la-hoja-de-vlad",
  "facts": [
    {"id": "fact-vladgrad-seal", "category": "lore",
     "statement": "Vladgrad fue fundada sobre la tumba sellada de Vlad el Empalador"}
  ],
  "entities": [
    {"id": "npc-archbishop-sergei", "name": "Arzobispo Sergei",
     "type": "npc", "canon_state": "alive",
     "motivation": "completar el ritual de luna de sangre"}
  ],
  "timeline": [
    {"id": "evt-blood-moon", "timestamp": "1483-10-31T23:00:00Z",
     "summary": "Primera luna de sangre en 200 años"}
  ]
}
```

El architect invoca `grimorio_save_lore`, `grimorio_save_npcs`
y `grimorio_save_bestiary` — tres herramientas MCP en un solo
batch.

### 2. Lore y Setting

`lore.md` es el world book para los jugadores. La setting guide
(`setting-guide.md`) es el deep-dive solo para el DM: historias
secretas, motivos ocultos, lo que los NPCs no dirán en voz alta.

Los dos documentos están escritos desde puntos de vista
diferentes. El lore es in-world y concreto; la setting guide es
meta e incluye spoilers. El architect genera ambos y el
[Campaign Consistency gate](../campaign-consistency.md)
cross-checkea que ningún hecho de `lore.md` contradiga un hecho
de `canon.json`.

### 3. NPCs y Facciones

Los NPCs vienen en dos sabores:

1. **Canon entities** en `canon.json` (la fuente de verdad).
2. **Stat blocks** en `npcs/npcs_and_factions.md` (el formato
   estilo WotC que el DM imprime).

Para *La Hoja de Vlad* el architect generó 12 NPCs. El Arzobispo
Sergei es el villano principal; su stat block tiene el rol
`mcguffin` porque es quien sabe dónde se oculta la Hoja.

### 4. Bestiario

Stat blocks custom viven en `bestiary/bestiary.md`. La campaña
incluye un template `vampiro_noble` (CR 5) y un minion
`sombra_cultista` (CR 1/2) — ambos se reutilizan a lo largo de
los tres capítulos.

La skill grimorio-bestiary enforce el formato WotC 5e: AC, HP,
velocidad, seis ability scores, saves, skills, damage
resistances/immunities, senses, languages, challenge, y luego
traits y actions.

### 5. Capítulos y Áreas

Acá es donde va la mayor parte del trabajo. El architect invoca
`grimorio_save_chapter` para cada acto, que escribe un único
archivo markdown con **10–15 áreas numeradas** siguiendo el
formato WotC.

Excerpt de `chapter_01_sombras_en_la_corte.md` (Área 1, El
Vestíbulo de los Ancestros):

```markdown
> **Para Leer en Voz Alta:** *Las puertas de roble se abren ante
> vosotros, revelando un vestíbulo iluminado por candelabros de
> plata...*
<!-- WOTC: boxed_text — 89 palabras; segunda persona presente;
     detalles sensoriales únicamente. -->

**Ganchos de Personaje:**
- **Katarina (noble):** Reconocés el escudo Voronova en un tapiz.
- **Ivan (religioso):** Los símbolos son marcas contra no-muertos.
<!-- WOTC: character_hook — 2 ganchos, cada uno atado a un detalle
     descubrible en la sala. -->

**Desarrollos:**
1. **SI examinan los retratos:** descubren ancestro del culto.
   *Recuperación:* el Conde les muestra el retrato más tarde.
2. **SI encuentran el pergamino:** lista de víctimas del culto.
3. **SI ignoran el vestíbulo:** pierden ventaja en el banquete.
<!-- WOTC: development_branch — 3 IF-THEN con Consecuencia +
     Recuperación. -->
```

Notá las anotaciones `<!-- WOTC: ... -->`. Son notas del autor
para futuros re-readers; el compiler las strippea antes de
renderizar el PDF así que nunca aparecen en la versión
impresa.

### 6. Mapas y Handouts

La skill cartographer genera dos tipos de assets visuales:

- **Mapas de batalla** (`assets/map_*.svg`) — grillas dibujadas
  proceduralmente con room labels, secret doors, y light sources.
- **Divisores decorativos** (`assets/divider_*.svg`) —
  separadores de capítulo al estilo WotC.

Los handouts son documentos para los jugadores: invitaciones,
cartas, sketches. La herramienta `grimorio_generate_handouts` los
produce en ES/EN bilingüe para que el mismo archivo sirva a
parties angloparlantes e hispanoparlantes.

### 7. Validación y Consistencia

Antes de compilar el PDF, el architect corre
[`grimorio validate`](../getting-started.md#validate-your-campaign):

```bash
grimorio validate --scope=all la-hoja-de-vlad
```

El CLI corre 17 reglas de consistencia e imprime un reporte:

```
Campaign Validation Report
==========================
Campaign: la-hoja-de-vlad
Health: fair

❌ [warning] wotc_developments — [chapter_03] Need 3-5 IF-THEN branches
✅ [info] faction_reputation_gate — Facción Nobleza: allied (75), consistent
…

Summary: 0 errors, 3 warnings, 0 criticals (de 17 checks)
```

El architect arregla los warnings, vuelve a correr, y solo
procede cuando el reporte está limpio (o acepta los warnings
explícitamente).

### 8. Compilar a PDF

Finalmente:

```bash
grimorio compile_pdf la-hoja-de-vlad
```

El compiler convierte cada archivo markdown a HTML, aplica el
tema CSS WotC, y delega a Chromium para imprimir a PDF. El output
es `~/campaigns/la-hoja-de-vlad/campaign.pdf`.

## Excerpts Anotados

Acá están los cinco excerpts que mejor ilustran cómo piensa
Grimorio.

### Excerpt 1: Un blockquote de read-aloud (capítulo 1, Área 1)

```markdown
> **Para Leer en Voz Alta:** *Las puertas de roble se abren ante
> vosotros, revelando un vestíbulo iluminado por candelabros de
> plata...*
```

El label `**Para Leer en Voz Alta:**` se remueve durante el
compile (el CSS provee el indicador visual via pseudo-elemento
`::before`) y el resto se convierte en un `<blockquote
class="read-aloud">` en el PDF.

### Excerpt 2: Una escena de estilo boxed-text

```markdown
**Descripción para el DM:**
El vestíbulo es una trampa psicológica. Los retratos están
encantados con magia de adivinación menor—cualquiera que los
observe por más de 1 minuto siente una presencia observándolo.

- **Percepción DC 12:** Notás sangre fresca, no óxido.
- **Investigación DC 14:** El estandarte oculta un compartimento.
- **Arcano DC 15:** Los símbolos son runas anti-posesión.
```

Este es material solo para el DM. **No** está en blockquote de
read-aloud porque los jugadores no deberían ver los DCs.

### Excerpt 3: Una tabla de developments

```markdown
| Decisión | Condición | Consecuencia | Propagación |
|----------|-----------|--------------|-------------|
| **SI** | examinan retratos | **ENTONCES** ventaja social | Área 2, Facción Nobleza |
| **SI** | encuentran pergamino | **ENTONCES** 3 nombres | Área 5, Acto 2 |
| **SI** | ignoran todo | **ENTONCES** mayordomo hostil | Área 2 (servidores hostiles) |
```

Este formato de tabla es lo que la skill grimorio-areas
recomienda — IF/THEN explícitos con efectos en cascada.

### Excerpt 4: Un bloque de faction reputation

```markdown
**Cambios de Estado del Mundo:**
- **NPCs:** Mayordomo (neutral → hostil si ignoran)
- **Facciones:** Sin cambios aún
- **Pistas:** clue-001 (pergamino con nombres), clue-002 (símbolos)
```

Esto es lo que la herramienta MCP `dm_session_context` le
surfacea al AI Dungeon Master para que pueda reaccionar a las
decisiones de los jugadores.

### Excerpt 5: Un stat block

```markdown
### Sombra Cultista
*CE mediano no-muerto, cualquier alineamiento malvado*

- **Clase de Armadura** 13
- **Puntos de Golpe** 22 (5d8)
- **Velocidad** 9 m
- **DES** 14, **CAR** 10, **CON** 12
- **Inmunidades** a daño de veneno; resistencia a necrótico
- **Sentidos** visión en la oscuridad 18 m, Percepción pasiva 10
- **Idiomas** Común
- **Desafío** 1 (200 PX)
```

Este es el formato WotC 5e que la skill grimorio-bestiary
enforce. El compiler lo renderiza como un
`<div class="stat-block">` en el PDF.

## Lecciones Aprendidas

1. **El canon es la espina dorsal.** Todo artifact lo referencia.
   Cuando el architect cambia un hecho en `canon.json`, el lore,
   los NPCs y el bestiario deben actualizarse también. El
   validation gate detecta drift, pero es más barato mantener el
   canon primero.

2. **Bilingüe ES/EN es default, no afterthought.** La setting
   guide, la introduction y los handouts se entregan en ambos
   idiomas. Los jugadores eligen un lado; los DMs leen ambos.

3. **Las anotaciones `<!-- WOTC: ... -->` son oro.** Son cómo el
   autor le explica al futuro-self (o a un co-autor) por qué
   existe una sección. El compiler las strippea, así que nunca
   aparecen en el PDF pero quedan en la fuente.

4. **Validar no es opcional.** Correr `grimorio validate` antes
   de `compile_pdf` es la diferencia entre un PDF pulido y uno
   con strings `{{PLACEHOLDER}}` adentro.

5. **Empezá chico.** La primera campaña llevó 28 archivos; las
   subsiguientes llevan 12-15 porque el canon reutiliza NPCs y
   stat blocks entre capítulos. La skill grimorio-integrator
   batchea la regeneración.

## Dónde Ir Desde Acá

- [Empezando](../getting-started.md) — instalación + primera campaña.
- [Herramientas MCP](../features/mcp-tools.md) — referencia completa.
- [Arquitectura](../features/architecture.md) — cómo encaja el engine.
- [Consistencia de Campaña](../campaign-consistency.md) — el pipeline del gate.
- La campaña de ejemplo en sí:
  [`examples/la-hoja-de-vlad/`](../examples/la-hoja-de-vlad/).
