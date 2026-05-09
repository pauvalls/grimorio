---
title: Session Generator / Generador de Sesiones
lang: en/es
---

<div class="lang-selector">
<a href="#english">English</a> | <a href="#espanol">Español</a>
</div>

---

<a name="english"></a>

# Session Generator

## Overview

The Session Generator adapts your campaign content to specific player characters, creating personalized session documents with:

- **Character-Specific Encounters** — Balanced for your party composition
- **Personalized NPC Appearances** — Based on character backstories
- **Tailored Loot** — Appropriate for character classes
- **Custom Scenarios** — Leveraging character abilities and backgrounds

---

## Quick Start

### Generate Session with Specific Characters

```bash
/grimorio generate_session_prep campaign="sunken-city" session_num=1 characters="Raika,Samuel" focus="Ball at the university"
```

### Generate Session with Context

```bash
/grimorio generate_session_prep campaign="sunken-city" session_num=1 characters="Raika,Samuel" focus="Raika doesn't speak Common but knows etiquette (noble from tribal village), Samuel is aristocratic but has stage fright"
```

---

## What Gets Adapted

### 1. Encounters

The Session Generator analyzes your party composition and adjusts encounters accordingly.

**Example Party:**
- Raika (Druid 4) — Nature magic, wild shape
- Samuel (Wizard 4) — Arcane magic, utility

**Generated Encounters:**

```markdown
## Encounter 1: Library Guards

**CR Calculation:**
- Party level: 4 (2 players)
- Adjusted for 2 players: CR 1/2 per encounter

**Adaptations:**
- **For Raika:** Include plants she can command with Wild Shape
- **For Samuel:** Include arcane wards he can dispel

**Stats:**
- 2 Guards (CR 1/2 each)
- 1 Arcane Ward (environmental hazard)
```

### 2. NPC Appearances

NPCs are selected and modified based on character backstories.

**Example:**

```markdown
## NPC Appearance: Dean Thornwood

**Why He Appears:**
- **Samuel Connection:** Samuel's family knows the Dean from aristocratic circles
- **Raika Connection:** The Dean studied tribal druidic traditions

**Dialogue Adaptations:**
- **To Samuel:** "Ah, young noble! Your family's donation made this wing possible."
- **To Raika:** "I've read about your tribe's nature magic. Fascinating stuff."
```

### 3. Character-Specific Hooks

Each character gets 2-3 personalized hooks per session.

```markdown
## Character Hooks

### For Raika
1. **Tribal Symbols:** The university's garden has plants from your homeland. They're wilting unnaturally. Nature DC 13 to identify the cause.

2. **Etiquette Advantage:** You know noble protocols. Advantage on Persuasion when dealing with aristocrats.

### For Samuel
1. **Family Reputation:** Your family name opens doors. The guards let you pass without inspection (for now).

2. **Stage Fright:** You're nervous about performing magic in front of peers. Disadvantage on Performance checks unless alone.
```

### 4. Loot Recommendations

Loot is tailored to character classes.

```markdown
## Loot

**For Raika (Druid):**
- **Druidic Focus** — Carved wooden staff with tribal symbols
- **Potion of Plant Growth** — Can accelerate plant growth

**For Samuel (Wizard):**
- **Spell Scroll (Fireball)** — One-use scroll
- **Arcane Tome** — Contains notes on the egg's magic

**Shared:**
- **Silver Key** — Opens the vault
- **Potion of Healing (2)** — For both
```

---

## Advanced Usage

### Multiple Characters

```bash
/grimorio generate_session_prep campaign="sunken-city" session_num=1 characters="Raika,Samuel,Gromm,Lyra" focus="Full party infiltration"
```

### With Specific Focus

```bash
/grimorio generate_session_prep campaign="sunken-city" session_num=2 characters="Raika,Samuel" focus="Social encounter at the ball — emphasize Samuel's aristocratic background and Raika's tribal nobility"
```

### With Custom Constraints

```bash
/grimorio generate_session_prep campaign="sunken-city" session_num=1 characters="Raika,Samuel" focus="Stealth mission" constraints="No combat, emphasis on social and stealth"
```

---

## Use Cases

### Case 1: One-Shot with Pre-Gens

```bash
# Generate 4 pre-generated characters
/grimorio generate_character campaign="sunken-city" name="Rook" race="human" class="rogue" level=4
/grimorio generate_character campaign="sunken-city" name="Sera" race="tiefling" class="warlock" level=4
/grimorio generate_character campaign="sunken-city" name="Gromm" race="dwarf" class="fighter" level=4
/grimorio generate_character campaign="sunken-city" name="Lyra" race="elf" class="sorcerer" level=4

# Generate session for all 4
/grimorio generate_session_prep campaign="sunken-city" session_num=1 characters="Rook,Sera,Gromm,Lyra" focus="University heist"
```

### Case 2: Ongoing Campaign

```bash
# Session 1
/grimorio generate_session_prep campaign="sunken-city" session_num=1 characters="Raika,Samuel"

# After session, update state
/grimorio update_narrative_state campaign_id="sunken-city" session_num=1 revealed_clues=["clue-001"] key_decisions=["Spared guard"]

# Session 2 adapts to previous decisions
/grimorio generate_session_prep campaign="sunken-city" session_num=2 characters="Raika,Samuel"
```

### Case 3: Drop-in Players

```bash
# Session with 2 players
/grimorio generate_session_prep campaign="sunken-city" session_num=3 characters="Raika,Samuel"

# New player joins
/grimorio generate_character campaign="sunken-city" name="NewCharacter" race="dwarf" class="cleric" level=4

# Session 4 adapts to 3 players
/grimorio generate_session_prep campaign="sunken-city" session_num=4 characters="Raika,Samuel,NewCharacter"
```

---

## Tips for Best Results

### 1. Be Specific with Focus

**Vague:**
```bash
focus="University session"
```

**Specific:**
```bash
focus="Noble ball at university — Raika knows etiquette but doesn't speak Common, Samuel has stage fright"
```

### 2. Include Character Constraints

```bash
focus="Stealth mission"
constraints="Raika can't speak Common (disadvantage on verbal), Samuel has stage fright (disadvantage on Performance)"
```

### 3. Update Narrative State Regularly

This ensures session prep has accurate context:

```bash
/grimorio update_narrative_state campaign_id="sunken-city" session_num=1 revealed_clues=["clue-001"] dead_npcs=["guard-captain"]
```

---

## Troubleshooting

### "Characters not recognized"

Make sure characters exist in the campaign:

```bash
/grimorio list_characters campaign="sunken-city"
```

### "Session prep too generic"

Add more specific focus and constraints:

```bash
/grimorio generate_session_prep campaign="sunken-city" session_num=1 characters="Raika,Samuel" focus="Social encounter at ball" constraints="Emphasize Raika's tribal nobility and Samuel's aristocratic background"
```

### "Encounters too hard/easy"

Adjust party level or specify difficulty:

```bash
/grimorio generate_session_prep campaign="sunken-city" session_num=1 characters="Raika,Samuel" difficulty="easy"
```

---

## Next Steps

- **[Session Tutorial](session-tutorial.md)** — Run your first session
- **[Character Creation](character-creation.md)** — Generate PCs
- **[PDF Compiler](../features/pdf-compiler.md)** — Compile with session prep
- **[DM Guide](../dm-guide.md)** — General DM advice

---

<a name="espanol"></a>

# Generador de Sesiones

## Descripción General

El Generador de Sesiones adapta el contenido de tu campaña a personajes específicos, creando documentos de sesión personalizados con:

- **Encuentros Específicos de Personaje** — Balanceados para tu composición de grupo
- **Apariciones de NPCs Personalizadas** — Basadas en trasfondos de personajes
- **Botín Personalizado** — Apropiado para clases de personajes
- **Escenarios Personalizados** — Aprovechando habilidades y trasfondos

---

## Inicio Rápido

### Generar Sesión con Personajes Específicos

```bash
/grimorio generate_session_prep campaign="ciudad-hundida" session_num=1 characters="Raika,Samuel" focus="Baile en la universidad"
```

### Generar Sesión con Contexto

```bash
/grimorio generate_session_prep campaign="ciudad-hundida" session_num=1 characters="Raika,Samuel" focus="Raika no habla Común pero conoce etiqueta (noble de aldea tribal), Samuel es aristocrático pero tiene miedo escénico"
```

---

## Qué Se Adapta

### 1. Encuentros

El Generador de Sesiones analiza tu composición de grupo y ajusta los encuentros.

**Ejemplo de Grupo:**
- Raika (Druida 4) — Magia de naturaleza, forma salvaje
- Samuel (Mago 4) — Magia arcana, utilidad

**Encuentros Generados:**

```markdown
## Encuentro 1: Guardias de Biblioteca

**Cálculo de CR:**
- Nivel de grupo: 4 (2 jugadores)
- Ajustado para 2 jugadores: CR 1/2 por encuentro

**Adaptaciones:**
- **Para Raika:** Incluye plantas que pueda comandar con Forma Salvaje
- **Para Samuel:** Incluye runas arcanas que pueda disipar

**Estadísticas:**
- 2 Guardias (CR 1/2 cada uno)
- 1 Guarda Arcana (peligro ambiental)
```

### 2. Apariciones de NPCs

Los NPCs se seleccionan y modifican basándose en trasfondos.

**Ejemplo:**

```markdown
## Aparición de NPC: Decano Thornwood

**Por Qué Aparece:**
- **Conexión de Samuel:** La familia de Samuel conoce al Decano de círculos aristocráticos
- **Conexión de Raika:** El Decano estudió tradiciones druídicas tribales

**Adaptaciones de Diálogo:**
- **Para Samuel:** "¡Ah, joven noble! La donación de tu familia hizo posible este ala."
- **Para Raika:** "He leído sobre la magia de naturaleza de tu tribu. Fascinante."
```

### 3. Ganchos Específicos de Personaje

Cada personaje recibe 2-3 ganchos personalizados por sesión.

```markdown
## Ganchos de Personaje

### Para Raika
1. **Símbolos Tribales:** El jardín de la universidad tiene plantas de tu tierra natal. Se están marchitando unnaturally. Naturaleza CD 13 para identificar la causa.

2. **Ventaja de Etiqueta:** Conoces protocolos nobles. Ventaja en Persuasión al tratar con aristócratas.

### Para Samuel
1. **Reputación Familiar:** Tu nombre familiar abre puertas. Los guardias te dejan pasar sin inspección (por ahora).

2. **Miedo Escénico:** Estás nervioso por hacer magia frente a pares. Desventaja en pruebas de Actuación a menos que estés solo.
```

### 4. Recomendaciones de Botín

El botín se adapta a las clases de personajes.

```markdown
## Botín

**Para Raika (Druida):**
- **Enfoque Druida** — Bastón de madera tallada con símbolos tribales
- **Poción de Crecimiento de Plantas** — Puede acelerar crecimiento de plantas

**Para Samuel (Mago):**
- **Pergamino de Hechizo (Bola de Fuego)** — Pergamino de un uso
- **Tomo Arcano** — Contiene notas sobre la magia del huevo

**Compartido:**
- **Llave de Plata** — Abre la bóveda
- **Pociones de Curación (2)** — Para ambos
```

---

## Uso Avanzado

### Múltiples Personajes

```bash
/grimorio generate_session_prep campaign="ciudad-hundida" session_num=1 characters="Raika,Samuel,Gromm,Lyra" focus="Infiltración de grupo completo"
```

### Con Enfoque Específico

```bash
/grimorio generate_session_prep campaign="ciudad-hundida" session_num=2 characters="Raika,Samuel" focus="Encuentro social en el baile — enfatiza el trasfondo aristocrático de Samuel y la nobleza tribal de Raika"
```

### Con Restricciones Personalizadas

```bash
/grimorio generate_session_prep campaign="ciudad-hundida" session_num=1 characters="Raika,Samuel" focus="Misión de sigilo" constraints="Sin combate, énfasis en social y sigilo"
```

---

## Casos de Uso

### Caso 1: One-Shot con Pre-generados

```bash
# Genera 4 personajes pre-generados
/grimorio generate_character campaign="ciudad-hundida" name="Rook" race="humano" class="picaro" level=4
/grimorio generate_character campaign="ciudad-hundida" name="Sera" race="tiefling" class="brujo" level=4
/grimorio generate_character campaign="ciudad-hundida" name="Gromm" race="enano" class="guerrero" level=4
/grimorio generate_character campaign="ciudad-hundida" name="Lyra" race="elfo" class="hechicero" level=4

# Genera sesión para los 4
/grimorio generate_session_prep campaign="ciudad-hundida" session_num=1 characters="Rook,Sera,Gromm,Lyra" focus="Robo en universidad"
```

### Caso 2: Campaña en Curso

```bash
# Sesión 1
/grimorio generate_session_prep campaign="ciudad-hundida" session_num=1 characters="Raika,Samuel"

# Después de sesión, actualiza estado
/grimorio update_narrative_state campaign_id="ciudad-hundida" session_num=1 revealed_clues=["pista-001"] key_decisions=["Perdonaron guardia"]

# Sesión 2 se adapta a decisiones previas
/grimorio generate_session_prep campaign="ciudad-hundida" session_num=2 characters="Raika,Samuel"
```

### Caso 3: Jugadores que Se Unen

```bash
# Sesión con 2 jugadores
/grimorio generate_session_prep campaign="ciudad-hundida" session_num=3 characters="Raika,Samuel"

# Nuevo jugador se une
/grimorio generate_character campaign="ciudad-hundida" name="NuevoPersonaje" race="enano" class="clerigo" level=4

# Sesión 4 se adapta a 3 jugadores
/grimorio generate_session_prep campaign="ciudad-hundida" session_num=4 characters="Raika,Samuel,NuevoPersonaje"
```

---

## Consejos para Mejores Resultados

### 1. Sé Específico con el Enfoque

**Vago:**
```bash
focus="Sesión en universidad"
```

**Específico:**
```bash
focus="Baile noble en universidad — Raika conoce etiqueta pero no habla Común, Samuel tiene miedo escénico"
```

### 2. Incluye Restricciones de Personaje

```bash
focus="Misión de sigilo"
constraints="Raika no puede hablar Común (desventaja en verbal), Samuel tiene miedo escénico (desventaja en Actuación)"
```

### 3. Actualiza Estado Narrativo Regularmente

Esto asegura que la preparación tenga contexto preciso:

```bash
/grimorio update_narrative_state campaign_id="ciudad-hundida" session_num=1 revealed_clues=["pista-001"] dead_npcs=["capitan-guardias"]
```

---

## Solución de Problemas

### "Personajes no reconocidos"

Asegúrate de que los personajes existan en la campaña:

```bash
/grimorio list_characters campaign="ciudad-hundida"
```

### "Preparación de sesión muy genérica"

Añade enfoque y restricciones más específicas:

```bash
/grimorio generate_session_prep campaign="ciudad-hundida" session_num=1 characters="Raika,Samuel" focus="Encuentro social en baile" constraints="Enfatiza nobleza tribal de Raika y trasfondo aristocrático de Samuel"
```

### "Encuentros muy difíciles/fáciles"

Ajusta nivel de grupo o especifica dificultad:

```bash
/grimorio generate_session_prep campaign="ciudad-hundida" session_num=1 characters="Raika,Samuel" difficulty="easy"
```

---

## Próximos Pasos

- **[Tutorial de Sesión](session-tutorial.md)** — Ejecuta tu primera sesión
- **[Creación de Personajes](character-creation.md)** — Genera PJs
- **[Compilador PDF](../features/pdf-compiler.md)** — Compila con preparación
- **[Guía de DM](../dm-guide.md)** — Consejos generales para DM
