# DM Agent Guide — grimorio-dm

> How to run live D&D 5e sessions using the Grimorio AI Dungeon Master.

---

## Table of Contents

1. [Before the Session](#before-the-session)
2. [Session Start Protocol](#session-start-protocol)
3. [The Narrative Loop](#the-narrative-loop)
4. [Scene Transitions](#scene-transitions)
5. [Combat Protocol](#combat-protocol)
6. [End-of-Session Protocol](#end-of-session-protocol)
7. [Dice & Game Modes](#dice--game-modes)
8. [Information Hiding Reference](#information-hiding-reference)
9. [Troubleshooting](#troubleshooting)

---

## Before the Session

### 1. Generate DM Prep

Call `generate_session_prep` with the campaign ID:

```json
{
  "campaign_id": "shadows-of-thornvale",
  "session_num": 3
}
```

This returns:
- `previously_on` — recap of last session
- `active_quests` — what's currently open
- `relevant_npcs` — who might appear
- `reminders` — canon inconsistencies to watch for

### 2. Call `dm_session_context`

This is the PRIMARY tool for session initialization. It returns the full campaign state as a structured JSON payload:

```json
{
  "campaign_id": "shadows-of-thornvale",
  "session_num": 3,
  "include_prologue": true
}
```

The payload includes:
- **canon** — facts, entities, rules, relationships, timeline
- **narrative_state** — current session, revealed clues, dead NPCs, quest states
- **characters** — PC stats, HP, AC, inventory
- **areas** — numbered areas with summaries, read-aloud text, encounters
- **npcs** — motivations, secrets, dialogue voices, stats
- **bestiary** — monsters with descriptive damage cues
- **factions** — reputation scores and attitudes
- **quests** — active, completed, failed
- **prologue** — 4-part narrative opening (if enabled)

---

## Session Start Protocol

### Session 1 — Prologue Opening

If `session_num == 1` and `prologue` exists in the payload:

1. Read **Prologue Part 1** aloud in a boxed text format:
   ```
   ┌─────────────────────────────────────────┐
   │  El viento aúlla entre las torres...    │
   └─────────────────────────────────────────┘
   ```
2. Ask players to introduce their characters **in-character**.
3. Describe **Area 1** using its `player_read_aloud` text.
4. Ask "¿Qué hacen?"

### Session 2+ — Previously On

1. Read the `session_prep.previously_on` summary.
2. Ask "¿Qué están haciendo ahora?"

### Mode Selection (Every Session)

After loading context, ask:

> **¿Modo de dados: automático, manual, o mixto?**

- **Automático**: DM rolls everything. Fast, good for online play.
- **Manual**: Players roll everything. Good for physical tables.
- **Mixto** *(default)*: Players roll for PCs, DM rolls for NPCs.

Then ask:

> **¿Modo de juego: narrativo o táctico?**

- **Narrativo** *(default)*: 1-2 combats max. Social resolution first.
- **Táctico**: 3-5 combats. Full round-by-round combat.

**Store both selections** and respect them for the entire session.

---

## The Narrative Loop

The core loop of a session:

```
Describe Scene → Player Action → Resolve → Describe Outcome → Transition
```

### Describing Scenes

- Use sensory details: sights, sounds, smells, temperature.
- For significant locations, use boxed read-aloud text.
- Mention relevant NPCs by name if they are present or their influence is felt.

### Resolving Actions

1. **Player declares intent**: "Quiero persuadir al guardia."
2. **Determine DC**: Based on difficulty (Easy 10, Medium 15, Hard 20, Very Hard 25).
3. **Roll** (according to dice mode):
   - MANUAL: "Tirá Persuasión (Carisma)."
   - AUTO/MIXED: Roll secretly, describe result.
4. **Describe outcome narratively**:
   - Success: "El guardia baja la lanza y te escucha con atención."
   - Failure: "Te mira con sospecha y aprieta la empuñadura de su espada."

### Social Resolution in NARRATIVE Mode

When a combat encounter begins:

1. **Offer social resolution first**: "Los goblins parecen nerviosos, no agresivos. ¿Querés intentar intimidarlos para que se rindan?"
2. **If players choose combat**, resolve quickly:
   - Single group initiative roll.
   - Describe the fight in 2-3 narrative beats.
   - Use descriptive damage cues from the bestiary.
3. **If players choose social**, run a skill challenge or single contested roll.

---

## Scene Transitions

### Ending a Scene

Before transitioning, ensure:
- The dramatic question of the scene is answered.
- Players have had a chance to act.
- Any immediate consequences are narrated.

### Transition Techniques

- **Hard cut**: "Tres días después..."
- **Soft fade**: "Mientras descansan en la posada..."
- **Cliffhanger**: "Justo cuando creen que están a salvo, escuchan un rugido desde el pozo."

### Pacing by Game Mode

| Mode | Scene Length | Combat Length |
|------|--------------|---------------|
| NARRATIVE | 10-15 min | 5-10 min |
| TÁCTICO | 15-20 min | 20-30 min |

---

## Combat Protocol

### Initiative

1. Call for initiative (or roll secretly in AUTO mode).
2. Announce the order narratively: "Vos reaccionás primero, luego el ogro, y Aric al final."

### Player Turn

1. Ask: "¿Qué hacés?"
2. Resolve the action.
3. Describe the outcome with sensory detail.

### Enemy Turn

1. Determine the enemy's tactical goal (from bestiary `tactics` if available).
2. Roll secretly.
3. Describe the attack and result narratively.
   - **Hit**: "La clava del ogro desciende con fuerza brutal hacia tu posición. Sentís el impacto en el pecho."
   - **Miss**: "El ogro balancea su arma, pero tropezás con una piedra y esquivás por poco."

### Damage Descriptions

Use the `DescriptiveCues` from the bestiary. If missing, use the fallback scale:

| HP % | Description |
|------|-------------|
| 75-100% | Luce fresco, alerta, sin un rasguño. |
| 50-74% | Muestra signos de daño. Respiración agitada. |
| 25-49% | Claramente herido. Se tambalea. |
| 1-24% | Apenas se mantiene en pie. Ojos vidriosos. |
| 0% | Cae al suelo, inmóvil. |

---

## End-of-Session Protocol

### 1. Narrate Closure

Summarize the session:
- What the party accomplished.
- Any unresolved threads.
- End with a cliffhanger if appropriate.

### 2. Award XP (Milestone System)

Do NOT track individual XP. Level up at story beats:

| Level | Typical Milestone |
|-------|-------------------|
| 1→2 | Complete first quest |
| 2→3 | Major discovery or alliance |
| 3→4 | Defeat first significant enemy |
| 4→5 | Complete act 1 |

### 3. Update Narrative State

Call `update_narrative_state`:

```json
{
  "campaign_id": "shadows-of-thornvale",
  "session_num": 3,
  "revealed_clues": ["clue-001", "clue-002"],
  "completed_quests": ["q1"],
  "dead_npcs": ["npc-guard-3"],
  "key_decisions": [
    {
      "id": "dec-001",
      "description": "Spared the goblin chief",
      "choice_made": "mercy"
    }
  ]
}
```

### 4. Evaluate Consequences

Call `evaluate_consequences`:

```json
{
  "campaign_id": "shadows-of-thornvale"
}
```

Check if any consequence rules triggered based on the updated narrative state.

### 5. Update Faction Reputation

For any reputation changes:

```json
{
  "campaign_id": "shadows-of-thornvale",
  "faction_id": "thieves-guild",
  "party_id": "party-1",
  "delta": -10,
  "reason": "Refused to pay protection money"
}
```

### 6. Export Handouts (Optional)

If players acquired maps, letters, or items:

```json
{
  "campaign_id": "shadows-of-thornvale",
  "handout_id": "handout-001",
  "format": "text"
}
```

---

## Dice & Game Modes

### Changing Modes Mid-Session

Modes are set at session start and should NOT change. If players ask to switch:
- **NARRATIVE → TÁCTICO**: "Ok, pero a partir de ahora vamos a resolver los combates ronda por ronda."
- **TÁCTICO → NARRATIVO**: "Podemos acelerar el combate actual, pero el próximo seguirá táctico si quieren."
- **Any dice mode change**: Allow it, but note it in the session log.

### Mixed Mode Best Practices

- Players roll for their PCs.
- DM rolls secretly for NPCs.
- For group checks (Perception, Stealth), DM rolls once for the group.
- For saving throws, players roll their own.

---

## Information Hiding Reference

### What Players Know

| They Know | They DON'T Know |
|-----------|-----------------|
| Their own HP, AC, stats | Enemy HP or AC |
| Their own roll results | Enemy roll results |
| Whether they hit or miss | Enemy attack bonuses or DCs |
| Damage they deal (descriptive) | Exact damage numbers dealt by enemies |

### Describing Rolls

Instead of "Sacaste 15 en Persuasión", say:
- "Tu argumento es convincente. El guardia asiente lentamente."

Instead of "El ogro sacó 18 en ataque", say:
- "El ogro levanta su clava con furia y la descarga sobre vos."

### Describing Damage

Instead of "Recibís 8 de daño", say:
- "El golpe te deja sin aliento. Sentís un dolor agudo en el costado."

Instead of "El goblin tiene 3 HP", say:
- "El goblin se tambalea, sangrando por la boca, pero aún te apunta con su daga."

---

## Troubleshooting

### "The agent revealed enemy HP"

**Cause**: Agent forgot information-hiding rules.
**Fix**: Remind the agent: "Recordá: nunca reveles HP, AC o tiradas de los enemigos. Usá DescriptiveCues."

### "NPC voice changed mid-conversation"

**Cause**: Agent lost track of `dialogue_voice`.
**Fix**: Reference the NPC's `dialogue_voice` property in the context payload at the start of each interaction.

### "Combat is dragging in NARRATIVE mode"

**Cause**: Agent is running full round-by-round combat.
**Fix**: Remind the agent of the game mode. In NARRATIVE mode, resolve combat in 1-2 narrative beats or offer social resolution.

### "Dead NPC appeared alive"

**Cause**: Agent didn't check `narrative_state.dead_npcs`.
**Fix**: Before introducing any NPC, the agent should cross-reference with the dead NPCs list.

### "Payload is too large"

**Cause**: Campaign has too many areas/NPCs/monsters.
**Fix**: The payload is capped at 100KB. If it exceeds this, the agent will receive a warning. For very large campaigns, the DM may need to manually trim the context or run multiple smaller sessions.

---

## Quick Reference Card

| Tool | When to Call |
|------|--------------|
| `dm_session_context` | Start of every session |
| `generate_session_prep` | Before session prep |
| `update_narrative_state` | End of session |
| `evaluate_consequences` | End of session |
| `update_faction_reputation` | When reputation changes |
| `grimorio_export_handout` | When players acquire items/maps |
