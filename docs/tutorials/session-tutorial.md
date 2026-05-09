---
title: Session Tutorial / Tutorial de Sesión
lang: en/es
---

<div class="lang-selector">
<a href="#english">English</a> | <a href="#espanol">Español</a>
</div>

---

<a name="english"></a>

# Session Tutorial

## Running Your First Game

This tutorial walks you through running a complete session with Grimorio, from prep to post-session tracking.

## Prerequisites

- A generated campaign (see [Getting Started](getting-started.md))
- Players ready to play
- Your AI assistant with Grimorio MCP server

---

## Phase 1: Pre-Session Prep

### Generate a DM Prep Sheet

Before the session, generate a prep sheet tailored to your planned content:

```
/grimorio generate_session_prep campaign="sunken-city" session_num=1
```

Or with context:

```
/grimorio generate_session_prep campaign="sunken-city" session_num=1 focus="Players arrive at the university for the heist"
```

**The prep sheet includes:**
- **"Previously On"** — Recap of last session (if applicable)
- **Active Quests** — Current objectives and their status
- **Relevant NPCs** — Who the players might encounter
- **Likely Scenarios** — 3-4 predicted situations based on player goals
- **Encounter Recommendations** — Balanced fights for your party
- **Loot Suggestions** — Magic items appropriate for party level
- **DM Reminders** — Important plot points, faction tensions, secrets

### Review the Prep Sheet

Example output:

```markdown
# Session Prep — Sunken City (Session 1)

## Previously On
*First session — no previous content*

## Active Quests
- **The Perfect Heist** (Main) — Steal the arcane egg from the university
  - Giver: "The Key"
  - Status: Active

## Relevant NPCs
- **Prof. Elara Vex** — Researcher who knows about the egg
  - Location: Laboratory (Area 5)
  - Secret: Her family awakened the Dark Witness 200 years ago
  - Motivation: Contain the egg at all costs

## Likely Scenarios
1. **Library Infiltration** — Stealth through the main library
   - DC 14 Dexterity (Stealth) to avoid creaky floorboards
   - Orin won't hear, but the head librarian will

2. **Laboratory Confrontation** — Prof. Vex discovers the party
   - If persuaded: She helps (Insight DC 13 to see she's scared)
   - If hostile: She activates the Golem (CR 5)

3. **The Vault** — Final confrontation with the egg
   - The egg is about to hatch (3 rounds remaining)
   - Three choices: deliver, destroy, or contain

## Encounter Recommendations
| Encounter | Party Level | Type | CR |
|-----------|-------------|------|-----|
| Guard Patrol (2 guards) | 4 | Combat | 1/2 + 1/2 |
| Golem of Observatory | 4 | Boss | 5 |
| Arcane Ward Hazard | 4 | Environmental | — |

## Loot Suggestions
- **Silver Dagger +1** — Hidden in Prof. Vex's desk
- **Cathedral Key** — Worn by the guard captain
- **Potion of Invisibility** — In the vault's containment unit

## DM Reminders
⚠️ **Faction Tension:** The Umbral Guild is watching. If the party takes too long, send a rival thief.

💡 **Secret:** The egg is one of seven seals. Don't reveal this yet — drop hints through visions (DC 15 Wisdom save or gain 1 Shock Point).
```

### Prepare Player Materials

Generate handouts for your players:

```
/grimorio generate_handouts campaign="sunken-city" handout_type="quest" content_refs=["quest-001"] version="player"
```

This creates player-facing quest descriptions without spoilers.

---

## Phase 2: Running the Session

### Use the Generated Acts

Your campaign's `areas/` directory contains numbered areas in WotC format:

```markdown
### Area 1: Entrance Hall

>> The oak doors of the university rise before you, carved with runes that glow faintly in the darkness. The air smells of old parchment and ozone.

**Character Hooks:**
- **Rook:** You notice the third stair is worn — it will creak. DC 14 Dexterity to step silently.
- **Lyra:** The shadows behind the portraits don't behave as they should. DC 15 Arcana: they are Watchful Eyes.

**Mechanics:**
- Stealth check (DC 14) to avoid alerting guards
- Arcana check (DC 15) to identify the wards

**Treasure:** Guard captain's silver key (opens Area 4)
```

### Running Tips

1. **Read Aloud Text** — The `>>` sections are meant to be read verbatim to players
2. **Character Hooks** — Call out specific hooks for each PC when relevant
3. **Developments** — Each area has 3+ narrative branches with recovery paths
4. **Running the Scene** — Check the DM guidance for pacing and improvisation tips

### Adapting on the Fly

If players go off-script, use the Session Generator:

```
/grimorio generate_session_prep campaign="sunken-city" session_num=1 focus="Players decided to talk to the dean instead of infiltrating"
```

This regenerates scenarios based on the new direction.

---

## Phase 3: Post-Session Tracking

### Update Narrative State

After the session, record what happened:

```
/grimorio update_narrative_state campaign_id="sunken-city" session_num=1 revealed_clues=["clue-egg-is-alive"] dead_npcs=[] completed_quests=[] key_decisions=["Players spared the guard captain"]
```

**Parameters:**
- `revealed_clues` — Clue IDs the players discovered
- `dead_npcs` — NPCs who died (prevents them appearing alive later!)
- `completed_quests` — Quest IDs finished
- `key_decisions` — Major choices that affect the story
- `xp_awarded` — Total XP given to the party
- `loot_acquired` — Items players took
- `session_summary` — Brief summary for "Previously On" next time

### Evaluate Consequences

Check what ripple effects the players' actions created:

```
/grimorio evaluate_consequences campaign_id="sunken-city"
```

This returns consequences based on the `consequences.json` rules.

### Update Faction Reputation

If the party's actions affected factions:

```
/grimorio update_faction_reputation campaign_id="sunken-city" faction_id="university_guards" party_id="party-1" delta=10 reason="Spared the guard captain"
```

This automatically propagates to allied factions.

---

## Phase 4: Preparing Session 2

### Generate Session 2 Prep

```
/grimorio generate_session_prep campaign="sunken-city" session_num=2
```

The prep sheet now includes:
- **"Previously On"** — Auto-generated from session 1 summary
- **Updated NPC status** — Guards are now friendly (if you spared them)
- **New consequences** — The captain might help you later
- **Adjusted encounters** — Guards won't attack on sight

### Review Character Progression

Check if any characters leveled up:

```
/grimorio get_character campaign="sunken-city" name="Rook"
```

Update their sheet if needed.

---

## Troubleshooting

### "NPC appeared even though they died!"

The validator should prevent this. Run a consistency check:

```
/grimorio check_consistency campaign_id="sunken-city"
```

If it finds issues, it will suggest fixes.

### "Players went completely off-track!"

Use the Session Generator to adapt:

```
/grimorio generate_session_prep campaign="sunken-city" session_num=1 focus="Players are negotiating with the dean instead of stealing"
```

This creates new scenarios based on the current situation.

### "I need a custom encounter!"

Generate a contextual random table:

```
/grimorio generate_random_tables campaign_id="sunken-city" table_type="encounter" location_hint="library" party_size=4 level_range="4-6"
```

---

## Next Steps

- **[Session Generator](session-generator.md)** — Deep dive into adapting sessions
- **[Character Creation](character-creation.md)** — Generate and customize PCs
- **[PDF Compiler](../features/pdf-compiler.md)** — Compile your campaign with customizations
- **[DM Guide](../dm-guide.md)** — General advice for running games

---

<a name="espanol"></a>

# Tutorial de Sesión

## Ejecutando Tu Primer Juego

Este tutorial te guía para ejecutar una sesión completa con Grimorio, desde la preparación hasta el seguimiento post-sesión.

---

## Prerrequisitos

- Una campaña generada (ver [Empezando](getting-started.md))
- Jugadores listos para jugar
- Tu asistente IA con el servidor MCP de Grimorio

---

## Fase 1: Preparación Pre-Sesión

### Generar Hoja de Preparación para DM

Antes de la sesión, genera una hoja de preparación adaptada a tu contenido planeado:

```
/grimorio generate_session_prep campaign="ciudad-hundida" session_num=1
```

O con contexto:

```
/grimorio generate_session_prep campaign="ciudad-hundida" session_num=1 focus="Los jugadores llegan a la universidad para el robo"
```

**La hoja de preparación incluye:**
- **"Anteriormente en"** — Resumen de la sesión anterior (si aplica)
- **Misiones Activas** — Objetivos actuales y su estado
- **NPCs Relevantes** — A quiénes podrían encontrar los jugadores
- **Escenarios Probables** — 3-4 situaciones predichas basadas en los objetivos
- **Recomendaciones de Encuentros** — Peleas balanceadas para tu grupo
- **Sugerencias de Botín** — Objetos mágicos apropiados para el nivel del grupo
- **Recordatorios para DM** — Puntos importantes de la trama, tensiones de facciones, secretos

### Revisar la Hoja de Preparación

Ejemplo de salida:

```markdown
# Preparación de Sesión — Ciudad Hundida (Sesión 1)

## Anteriormente En
*Primera sesión — sin contenido previo*

## Misiones Activas
- **El Robo Perfecto** (Principal) — Robar el huevo arcano de la universidad
  - Dador: "La Llave"
  - Estado: Activa

## NPCs Relevantes
- **Prof. Elara Vex** — Investigadora que sabe sobre el huevo
  - Ubicación: Laboratorio (Área 5)
  - Secreto: Su familia despertó al Testigo Oscuro hace 200 años
  - Motivación: Contener el huevo a toda costa

## Escenarios Probables
1. **Infiltración en la Biblioteca** — Sigilo a través de la biblioteca principal
   - CD 14 Destreza (Sigilo) para evitar tablones que crujen
   - Orin no escuchará, pero la bibliotecaria jefe sí

2. **Confrontación en el Laboratorio** — La Prof. Vex descubre al grupo
   - Si la persuaden: Ella ayuda (Intuición CD 13 para ver que está asustada)
   - Si es hostil: Ella activa el Gólem (CR 5)

3. **La Bóveda** — Confrontación final con el huevo
   - El huevo está a punto de eclosionar (3 rondas restantes)
   - Tres opciones: entregar, destruir, o contener

## Recomendaciones de Encuentros
| Encuentro | Nivel del Grupo | Tipo | CR |
|-----------|-------------|------|-----|
| Patrulla de Guardias (2 guardias) | 4 | Combate | 1/2 + 1/2 |
| Gólem del Observatorio | 4 | Jefe | 5 |
| Peligro de Guarda Arcana | 4 | Ambiental | — |

## Sugerencias de Botín
- **Daga de Plata +1** — Oculta en el escritorio de la Prof. Vex
- **Llave de la Catedral** — Llevada por el capitán de guardias
- **Poción de Invisibilidad** — En la unidad de contención de la bóveda

## Recordatorios para DM
⚠️ **Tensión de Facción:** El Gremio Umbrío está observando. Si el grupo tarda demasiado, envía un ladrón rival.

💡 **Secreto:** El huevo es uno de siete sellos. No lo reveles aún — da pistas a través de visiones (CD 15 Sabiduría o gana 1 Punto de Shock).
```

### Preparar Materiales para Jugadores

Genera handouts para tus jugadores:

```
/grimorio generate_handouts campaign="ciudad-hundida" handout_type="quest" content_refs=["quest-001"] version="player"
```

Esto crea descripciones de misiones para jugadores sin spoilers.

---

## Fase 2: Ejecutando la Sesión

### Usar los Actos Generados

El directorio `areas/` de tu campaña contiene áreas numeradas en formato WotC:

```markdown
### Área 1: Hall de Entrada

>> Las puertas de roble de la universidad se alzan ante vosotros, talladas con runas que brillan débilmente en la oscuridad. El aire huele a pergamino viejo y ozono.

**Ganchos de Personaje:**
- **Rook:** Notás que el tercer escalón está desgastado — va a crujir. CD 14 Destreza para pisarlo en silencio.
- **Lyra:** Las sombras detrás de los retratos no se comportan como deberían. CD 15 Arcanos: son Ojos de Vigilante.

**Mecánicas:**
- Prueba de Sigilo (CD 14) para evitar alertar guardias
- Prueba de Arcanos (CD 15) para identificar las runas

**Tesoro:** Llave de plata del capitán (abre Área 4)
```

### Consejos para Dirigir

1. **Texto para Leer en Voz Alta** — Las secciones `>>` están hechas para leerse textualmente a los jugadores
2. **Ganchos de Personaje** — Menciona ganchos específicos para cada PJ cuando sea relevante
3. **Desarrollos** — Cada área tiene 3+ ramas narrativas con caminos de recuperación
4. **Dirigiendo la Escena** — Revisa la guía de DM para consejos de ritmo e improvisación

### Adaptando Sobre la Marcha

Si los jugadores se salen del guión, usa el Generador de Sesiones:

```
/grimorio generate_session_prep campaign="ciudad-hundida" session_num=1 focus="Los jugadores decidieron hablar con el decano en lugar de infiltrarse"
```

Esto regenera escenarios basados en la nueva dirección.

---

## Fase 3: Seguimiento Post-Sesión

### Actualizar Estado Narrativo

Después de la sesión, registra lo que pasó:

```
/grimorio update_narrative_state campaign_id="ciudad-hundida" session_num=1 revealed_clues=["pista-huevo-esta-vivo"] dead_npcs=[] completed_quests=[] key_decisions=["Los jugadores perdonaron al capitán de guardias"]
```

**Parámetros:**
- `revealed_clues` — IDs de pistas que los jugadores descubrieron
- `dead_npcs` — NPCs que murieron (¡evita que aparezcan vivos después!)
- `completed_quests` — IDs de misiones completadas
- `key_decisions` — Decisiones importantes que afectan la historia
- `xp_awarded` — XP total dada al grupo
- `loot_acquired` — Objetos que los jugadores tomaron
- `session_summary` — Resumen breve para "Anteriormente en" la próxima vez

### Evaluar Consecuencias

Revisa qué efectos secundarios crearon las acciones de los jugadores:

```
/grimorio evaluate_consequences campaign_id="ciudad-hundida"
```

Esto devuelve consecuencias basadas en las reglas de `consequences.json`.

### Actualizar Reputación de Facción

Si las acciones del grupo afectaron facciones:

```
/grimorio update_faction_reputation campaign_id="ciudad-hundida" faction_id="guardias-universidad" party_id="party-1" delta=10 reason="Perdonaron al capitán de guardias"
```

Esto se propaga automáticamente a facciones aliadas.

---

## Fase 4: Preparando Sesión 2

### Generar Preparación Sesión 2

```
/grimorio generate_session_prep campaign="ciudad-hundida" session_num=2
```

La hoja de preparación ahora incluye:
- **"Anteriormente en"** — Autogenerado del resumen de sesión 1
- **Estado actualizado de NPCs** — Los guardias ahora son amistosos (si los perdonaste)
- **Nuevas consecuencias** — El capitán podría ayudarte después
- **Encuentros ajustados** — Los guardias no atacarán a primera vista

### Revisar Progresión de Personajes

Revisa si algún personaje subió de nivel:

```
/grimorio get_character campaign="ciudad-hundida" name="Rook"
```

Actualiza su hoja si es necesario.

---

## Solución de Problemas

### "¡El NPC apareció aunque murió!"

El validador debería prevenir esto. Ejecuta un chequeo de consistencia:

```
/grimorio check_consistency campaign_id="ciudad-hundida"
```

Si encuentra problemas, sugerirá arreglos.

### "¡Los jugadores se salieron completamente del camino!"

Usa el Generador de Sesiones para adaptar:

```
/grimorio generate_session_prep campaign="ciudad-hundida" session_num=1 focus="Los jugadores están negociando con el decano en lugar de robar"
```

Esto crea nuevos escenarios basados en la situación actual.

### "¡Necesito un encuentro personalizado!"

Genera una tabla aleatoria contextual:

```
/grimorio generate_random_tables campaign_id="ciudad-hundida" table_type="encounter" location_hint="biblioteca" party_size=4 level_range="4-6"
```

---

## Próximos Pasos

- **[Generador de Sesiones](session-generator.md)** — Profundiza en adaptar sesiones
- **[Creación de Personajes](character-creation.md)** — Genera y personaliza PJs
- **[Compilador PDF](../features/pdf-compiler.md)** — Compila tu campaña con personalizaciones
- **[Guía de DM](../dm-guide.md)** — Consejos generales para dirigir juegos
