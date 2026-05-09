---
title: Character Creation / Creación de Personajes
lang: en/es
---

<div class="lang-selector">
<a href="#english">English</a> | <a href="#espanol">Español</a>
</div>

---

<a name="english"></a>

# Character Creation

## Overview

Grimorio provides multiple ways to create characters for your campaign:

1. **Pre-generated Characters** — Ready-to-play NPCs with backstory hooks
2. **Custom Character Generation** — Generate PCs with specific race/class/background
3. **Character Worksheets** — Interactive sheets for player customization

---

## Quick Start

### Generate a Character

```bash
/grimorio generate_character campaign="sunken-city" name="Raika" race="elf" class="druid" level=4 background="noble"
```

### List All Characters

```bash
/grimorio list_characters campaign="sunken-city"
```

### Get Character Details

```bash
/grimorio get_character campaign="sunken-city" name="Raika"
```

---

## Available Options

### Races

| Race | Description |
|------|-------------|
| Human | Versatile, adaptable |
| Elf | Agile, magical affinity |
| Dwarf | Tough, resilient |
| Halfling | Lucky, stealthy |
| Half-Elf | Charismatic, versatile |
| Half-Orc | Strong, fierce |
| Gnome | Intelligent, inventive |
| Tiefling | Infernal heritage, charismatic |
| Dragonborn | Draconic ancestry, powerful |
| Goblin | Small, cunning |
| Firbolg | Giant-kin, nature-connected |
| Gith | Otherworldly, disciplined |

### Classes

| Class | Description |
|-------|-------------|
| Barbarian | Rage-fueled warrior |
| Bard | Magical performer |
| Warlock | Pact-bound spellcaster |
| Cleric | Divine servant |
| Druid | Nature guardian |
| Fighter | Martial master |
| Sorcerer | Innate magic |
| Wizard | Studied magic |
| Monk | Martial artist |
| Paladin | Holy warrior |
| Rogue | Skilled infiltrator |
| Ranger | Wilderness tracker |

### Backgrounds

| Background | Description |
|------------|-------------|
| Acolyte | Religious servant |
| Criminal | Outlaw, thief |
| Noble | Aristocrat, high-born |
| Sage | Scholar, researcher |
| Soldier | Military veteran |
| Guild Artisan | Skilled craftsman |
| Sailor | Sea-faring |
| Hermit | Reclusive, spiritual |
| Charlatan | Deceiver, trickster |
| Folk Hero | Common champion |
| Entertainer | Performer, actor |
| Gladiator | Arena fighter |
| Mercenary | Hired sword |
| Pirate | Sea raider |

---

## Pre-Generated Characters

### What's Included

Pre-generated characters come with:

- **Full Stat Block** — Abilities, skills, proficiencies
- **Backstory Hooks** — 2-3 plot connections to the campaign
- **Personality Traits** — Roleplay guidance
- **Secrets** — Hidden information the DM can reveal
- **Goals** — Character motivations and objectives
- **Spell List** — Prepared spells for spellcasters

---

## Custom Character Generation

### Basic Generation

```bash
/grimorio generate_character campaign="sunken-city" name="Samuel" race="human" class="wizard" level=4 background="noble" alignment="LN"
```

### With Custom Options

```bash
/grimorio generate_character campaign="sunken-city" name="Samuel" race="human" class="wizard" level=4 background="noble" alignment="LN" backstory_depth="detailed" include_spells=true include_secrets=true
```

**Options:**
- `backstory_depth`: `minimal` | `standard` | `detailed`
- `include_spells`: `true` | `false`
- `include_secrets`: `true` | `false`
- `include_goals`: `true` | `false`

---

## Session-Adapted Characters

### Generate Characters for Specific Session

When using the Session Generator, you can specify which characters are attending:

```bash
/grimorio generate_session_prep campaign="sunken-city" session_num=1 characters="Raika,Samuel" focus="Ball at the university"
```

This adapts:
- **Encounters** — Balanced for Raika (druid) + Samuel (wizard)
- **NPC Appearances** — Based on character backstories
- **Character Hooks** — Specific to their backgrounds
- **Loot** — Appropriate for their classes

### Example: Raika & Samuel at the Ball

```bash
/grimorio generate_session_prep campaign="sunken-city" session_num=1 characters="Raika,Samuel" focus="Noble ball at the university — Raika knows etiquette but doesn't speak Common, Samuel is aristocratic but has stage fright"
```

**Generated adaptations:**

- **Raika (Druid, Noble from tribal village):**
  - Advantage on Etiquette checks
  - Disadvantage on Common language interactions
  - Hook: Recognizes tribal symbols in university decorations

- **Samuel (Wizard, Aristocrat):**
  - Advantage on Arcana and Noble knowledge
  - Disadvantage on Performance (stage fright)
  - Hook: Knows the dean personally

---

## Tips for DMs

### 1. Generate Characters Early

Create pre-generated characters before Session Zero:

```bash
/grimorio generate_character campaign="sunken-city" name="CharacterName" race="..." class="..." level=4
```

### 2. Use Character Hooks

Each character comes with 2-3 hooks. Weave them into your sessions:

```markdown
**For Raika:** The university's garden has plants from her homeland. She can identify them with Nature DC 12.
```

### 3. Track Character Secrets

Secrets are hidden from players but available to you:

```bash
/grimorio get_character campaign="sunken-city" name="Raika" include_secrets=true
```

### 4. Update Characters After Sessions

Record XP, level-ups, and new abilities:

```bash
/grimorio get_character campaign="sunken-city" name="Raika"
# Then update manually or regenerate with new level
```

---

## Troubleshooting

### "Character not appearing in PDF"

Make sure the character file is in the `characters/` directory:

```bash
ls ~/campaigns/sunken-city/characters/
```

### "Need more diverse races"

Grimorio supports goblins, firbolg, gith, and more:

```bash
/grimorio generate_character campaign="sunken-city" name="Grok" race="goblin" class="rogue"
```

### "Want custom backgrounds"

Use the `background` field flexibly:

```bash
/grimorio generate_character campaign="sunken-city" name="Character" background="custom" backstory_depth="detailed"
```

---

## Next Steps

- **[Session Tutorial](session-tutorial.md)** — Run your first session
- **[Session Generator](session-generator.md)** — Adapt sessions to your party
- **[PDF Compiler](../features/pdf-compiler.md)** — Compile with character sheets
- **[DM Guide](../dm-guide.md)** — General DM advice

---

<a name="espanol"></a>

# Creación de Personajes

## Descripción General

Grimorio proporciona múltiples formas de crear personajes para tu campaña:

1. **Personajes Pre-generados** — PJs listos para jugar con ganchos de trasfondo
2. **Generación Personalizada** — Genera PJs con raza/clase/trasfondo específicos
3. **Hojas de Trabajo** — Hojas interactivas para personalización de jugadores

---

## Inicio Rápido

### Generar un Personaje

```bash
/grimorio generate_character campaign="ciudad-hundida" name="Raika" race="elfo" class="druida" level=4 background="noble"
```

### Listar Todos los Personajes

```bash
/grimorio list_characters campaign="ciudad-hundida"
```

### Obtener Detalles de Personaje

```bash
/grimorio get_character campaign="ciudad-hundida" name="Raika"
```

---

## Opciones Disponibles

### Razas

| Raza | Descripción |
|------|-------------|
| Humano | Versátil, adaptable |
| Elfo | Ágil, afinidad mágica |
| Enano | Resistente, resiliente |
| Mediano | Afortunado, sigiloso |
| Semielfo | Carismático, versátil |
| Semiorco | Fuerte, feroz |
| Gnomo | Inteligente, inventivo |
| Tiefling | Herencia infernal, carismático |
| Draconido | Ancestro dracónico, poderoso |
| Trasgo (Goblin) | Pequeño, astuto |
| Firbolg | Gigante, conectado con naturaleza |
| Gith | Extraterrenal, disciplinado |

### Clases

| Clase | Descripción |
|-------|-------------|
| Barbaro | Guerrero impulsado por ira |
| Bardo | Ejecutante mágico |
| Brujo | Lanzador atado a pacto |
| Clerigo | Sirviente divino |
| Druida | Guardián de naturaleza |
| Guerrero | Maestro marcial |
| Hechicero | Magia innata |
| Mago | Magia estudiada |
| Monje | Artista marcial |
| Paladín | Guerrero sagrado |
| Pícaro | Infiltrador experto |
| Explorador | Rastreador de naturaleza |

### Trasfondos

| Trasfondo | Descripción |
|------------|-------------|
| Acolito | Sirviente religioso |
| Criminal | Forajido, ladrón |
| Noble | Aristócrata, nacido alto |
| Sabio | Erudito, investigador |
| Soldado | Veterano militar |
| Artesano | Artesano experto |
| Marinero | Navegante marino |
| Ermitaño | Reclusivo, espiritual |
| Charlatán | Engañador, tramposo |
| Héroe del Pueblo | Campeón común |
| Animador | Ejecutante, actor |
| Gladiador | Luchador de arena |
| Mercenario | Espada alquilada |
| Pirata | Asaltante marino |

---

## Personajes Pre-generados

### Qué Está Incluido

Los personajes pre-generados incluyen:

- **Estadísticas Completas** — Habilidades, competencias, pericias
- **Ganchos de Trasfondo** — 2-3 conexiones de trama a la campaña
- **Rasgos de Personalidad** — Guía para roleplay
- **Secretos** — Información oculta que el DM puede revelar
- **Objetivos** — Motivaciones y objetivos del personaje
- **Lista de Hechizos** — Hechizos preparados para lanzadores

---

## Generación Personalizada

### Generación Básica

```bash
/grimorio generate_character campaign="ciudad-hundida" name="Samuel" race="humano" class="mago" level=4 background="noble" alignment="LN"
```

### Con Opciones Personalizadas

```bash
/grimorio generate_character campaign="ciudad-hundida" name="Samuel" race="humano" class="mago" level=4 background="noble" alignment="LN" backstory_depth="detailed" include_spells=true include_secrets=true
```

**Opciones:**
- `backstory_depth`: `minimal` | `standard` | `detailed`
- `include_spells`: `true` | `false`
- `include_secrets`: `true` | `false`
- `include_goals`: `true` | `false`

---

## Personajes Adaptados a Sesión

### Generar Personajes para Sesión Específica

Cuando uses el Generador de Sesiones, puedes especificar qué personajes asisten:

```bash
/grimorio generate_session_prep campaign="ciudad-hundida" session_num=1 characters="Raika,Samuel" focus="Baile en la universidad"
```

Esto adapta:
- **Encuentros** — Balanceados para Raika (druida) + Samuel (mago)
- **Apariciones de NPCs** — Basadas en trasfondos de personajes
- **Ganchos de Personajes** — Específicos a sus trasfondos
- **Botín** — Apropiado para sus clases

### Ejemplo: Raika y Samuel en el Baile

```bash
/grimorio generate_session_prep campaign="ciudad-hundida" session_num=1 characters="Raika,Samuel" focus="Baile noble en la universidad — Raika conoce etiqueta pero no habla Común, Samuel es aristocrático pero tiene miedo escénico"
```

**Adaptaciones generadas:**

- **Raika (Druida, Noble de aldea tribal):**
  - Ventaja en pruebas de Etiqueta
  - Desventaja en interacciones en Común
  - Gancho: Reconoce símbolos tribales en decoraciones de universidad

- **Samuel (Mago, Aristócrata):**
  - Ventaja en Arcanos y conocimiento Noble
  - Desventaja en Actuación (miedo escénico)
  - Gancho: Conoce al decano personalmente

---

## Consejos para DMs

### 1. Genera Personajes Temprano

Crea personajes pre-generados antes de la Sesión Cero:

```bash
/grimorio generate_character campaign="ciudad-hundida" name="NombrePersonaje" race="..." class="..." level=4
```

### 2. Usa Ganchos de Personaje

Cada personaje viene con 2-3 ganchos. Intégralos en tus sesiones:

```markdown
**Para Raika:** El jardín de la universidad tiene plantas de su tierra natal. Puede identificarlas con Naturaleza CD 12.
```

### 3. Rastrea Secretos de Personajes

Los secretos están ocultos a los jugadores pero disponibles para ti:

```bash
/grimorio get_character campaign="ciudad-hundida" name="Raika" include_secrets=true
```

### 4. Actualiza Personajes Después de Sesiones

Registra XP, subidas de nivel, y nuevas habilidades:

```bash
/grimorio get_character campaign="ciudad-hundida" name="Raika"
# Luego actualiza manualmente o regenera con nuevo nivel
```

---

## Solución de Problemas

### "Personaje no aparece en PDF"

Asegúrate de que el archivo esté en el directorio `characters/`:

```bash
ls ~/campaigns/ciudad-hundida/characters/
```

### "Necesito razas más diversas"

Grimorio soporta goblins, firbolg, gith, y más:

```bash
/grimorio generate_character campaign="ciudad-hundida" name="Grok" race="trasgo" class="picaro"
```

### "Quiero trasfondos personalizados"

Usa el campo `background` flexiblemente:

```bash
/grimorio generate_character campaign="ciudad-hundida" name="Personaje" background="custom" backstory_depth="detailed"
```

---

## Próximos Pasos

- **[Tutorial de Sesión](session-tutorial.md)** — Ejecuta tu primera sesión
- **[Generador de Sesiones](session-generator.md)** — Adapta sesiones a tu grupo
- **[Compilador PDF](../features/pdf-compiler.md)** — Compila con hojas de personaje
- **[Guía de DM](../dm-guide.md)** — Consejos generales para DM
