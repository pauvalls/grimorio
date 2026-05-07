---
name: D&D 5e Campaign Design & SRD
version: "2.0.0"
description: D&D 5e System Reference Document + official adventure design patterns, narrative coherence, encounter balance, and campaign structure for high-quality adventure generation.
---

# D&D 5e Campaign Design Context

## 1. Official Adventure Structure Patterns

Based on analysis of Wizards of the Coast official adventures (Waterdeep: Dragon Heist, Ghosts of Saltmarsh, Out of the Abyss, Curse of Strahd, Rise of Tiamat, Icewind Dale).

### Pattern 1: Adventure Background (DM-Only)

Every official campaign starts with a DM-only setup section BEFORE any player-facing content:

- **Timeline of Previous Events**: Chronological sequence explaining how the current situation came to be (last weeks/months/years)
- **Storyline Summary**: 2-4 paragraphs for the DM showing the complete narrative arc from start to finish
- **Geopolitical Context**: Active factions, powers at play, how the world "breathes" independently of the PCs
- **Hooks to Connect PCs**: Multiple options based on class, background, or alignment — never "you are the chosen ones"

### Pattern 2: Chapters as Narrative Units with Distinct Game Modes

Each chapter adopts a **dominant game mode** that gives it mechanical identity:

| Chapter Example | Game Mode | PC Objective |
|---|---|---|
| Investigation + Linear Dungeon | Rescue NPC, obtain reward |
| Urban Sandbox + Downtime | Restore location, manage factions |
| Investigation + Confrontation | Solve mystery, track McGuffin |
| Chase + Villain Choice | Recover McGuffin before villain |
| Hub + Regional Exploration | Establish base, investigate rumors |
| Dungeon Exploration + Stealth/Combat | Dismantle enemy operation |
| Diplomacy + Alliances | Negotiate with faction |
| Escape + Survival | Flee from captors |
| Travel + Survival | Navigate dangerous territory |
| Intrigue + Foreign Culture | Resolve local conflicts connected to main plot |

**Critical Rule**: Chapter transitions are NEVER arbitrary. Each chapter delivers a **narrative asset** (information, ally, base, key item) that is prerequisite for the next.

### Pattern 3: Locations Broken into Numbered Areas

Within any chapter involving physical exploration, space is atomized into **sequentially numbered areas**. Each area follows a standardized format:

```
## Area X: Descriptive Name

**Read-aloud text:** (2-4 sentences in second person, present tense, sensory description)

Expanded DM description: details PCs can discover with skill checks.

- **Creatures:** Monster list, quantity, stats referenced to MM/embedded stats
- **Treasure:** XP, currency, magic items with standard loot table format
- **Connections:** Which areas connect (referenced by numbers)
- **Secrets/Traps:** DCs to detect, mechanisms, consequences
- **Development:** What happens after if PCs do X, Y, or Z
```

This format is INVARIANT across adventures. Even in city adventures, when PCs explore a villa or theater, the space becomes a "numbered dungeon."

### Pattern 4: NPCs as Companion/Faction/Antagonist Systems

Official campaigns don't treat NPCs as mere quest vectors. They build **NPC systems**:

- **Companions** (OotA): 10+ pre-generated NPCs with personalities, hidden agendas, plot connections
- **Faction Quest Chains** (WDH): 8 factions with tiered quests by level. Each faction has a contact, motivation, and 1 quest per tier
- **Political Factions** (GoS): Traditionalists vs Loyalists with secret third faction. Agendas affect side quests and epilogue

High-impact NPCs (villains, key allies) receive structured profiles: name, appearance, personality, motivation, secret, stats, roleplay notes.

### Pattern 5: Random Tables as Content Generators

Random tables aren't filler — they're **procedural content generators** the DM uses to improvise:

| Table Type | Example | Function |
|---|---|---|
| Travel Encounters | OotA: Random encounters by Underdark region | Sustain constant danger fantasy |
| Rumors | GoS: Tavern rumor table | Give optional clues, red herrings |
| Mood/Ambiance | GoS: Town mood table | Vary hub atmosphere |
| Jobs/Quests | WDH: Faction missions | Provide between-chapter content |
| Weather/Environment | OotA: Underdark travel conditions | Add mechanical variability |
| Treasure | All adventures: Treasure tables by CR | Standardize rewards |

### Pattern 6: Appendices as Asset Repositories

20-30% of an official campaign lives in appendices:

- **New Magic Items**: Standard format with history, appearance, properties
- **New Monsters**: Stat blocks with lore
- **Maps**: Player-facing (no secrets) and DM-facing (with secrets)
- **Handouts**: Letters, codes, in-world documents for physical delivery
- **NPC Profiles**: Quick reference for all NPCs

---

## 2. Narrative Coherence System (v2.0)

### 2.1 The Anti-Pattern: "Hope-Driven Consistency"

Never rely on LLM competence alone for campaign coherence. The current architecture operates under "hope-driven consistency" — expecting LLMs to generate content that "fits" by pure model competence. This fails at scale:

- **One-shot (1 act)**: ~5% failure rate
- **Mini-campaign (2-3 acts)**: ~25% failure rate  
- **Standard campaign (5 acts)**: ~55% failure rate
- **Long campaign (8+ acts)**: ~80% failure rate

### 2.2 Required Coherence Mechanisms

**Adventure Bible / Canon Document**: A single source of truth declaring immutable facts:
- "The curse comes from the dead god Morbus"
- "Queen of Shadows is actually the King's twin sister; nobody knows except her and the butler"
- "Arcane magic is banned in the city"

**Narrative State Tracker**: Mutable record of what PCs have experienced:
- Revealed clues (with critical/optional flag)
- Active/completed/failed quests
- Dead NPCs (with session, cause, killer)
- Key items (who holds them)
- Faction reputation scores
- World events
- Session log

**Cross-Reference Validation**: Before saving ANY content, validate:
1. All referenced entities exist in canon
2. Dead NPCs don't appear alive
3. No world rule violations (e.g., public magic in banned city)
4. Prerequisite clues have been revealed (or alternative path provided)
5. Motivations are consistent with canon
6. Encounters match party level (CR vs. level check)
7. Loot is balanced for level

**Consequence Tracking**: World reactivity. If PCs burn the village in Act 1, Act 2 must reflect that. Consequences must persist and propagate.

### 2.3 Validation Report Format

```json
{
  "status": "rejected",
  "checks": [
    {
      "rule": "npc_death_state",
      "passed": false,
      "severity": "critical",
      "message": "NPC El Informador is dead (session 2) but appears alive in Act 3, Scene 2",
      "location": "act_3, scene_2, line_45",
      "fix_suggestion": "Replace with new messenger NPC (e.g., 'Gorin, the beggar') or use non-NPC method (letter, vision)"
    }
  ],
  "retry_prompt": "Fix the following consistency issues: 1) Replace 'El Informador' in Act 3 with a new NPC or method..."
}
```

---

## 3. Encounter Design

### 3.1 XP Thresholds by Level

| Level | Easy | Medium | Hard | Deadly |
|-------|------|--------|------|--------|
| 1 | 25 | 50 | 75 | 100 |
| 2 | 50 | 100 | 150 | 200 |
| 3 | 75 | 150 | 225 | 400 |
| 4 | 125 | 250 | 375 | 500 |
| 5 | 250 | 500 | 750 | 1100 |
| 6 | 300 | 600 | 900 | 1400 |
| 7 | 350 | 750 | 1100 | 1700 |
| 8 | 450 | 900 | 1400 | 2100 |
| 9 | 550 | 1100 | 1600 | 2400 |
| 10 | 600 | 1200 | 1900 | 2800 |

### 3.2 Multi-Monster Multiplier

| Monster Count | Multiplier |
|---------------|------------|
| 1 | x1 |
| 2 | x1.5 |
| 3-6 | x2 |
| 7-10 | x2.5 |
| 11-14 | x3 |
| 15+ | x4 |

### 3.3 Encounter Structure (Official Format)

Every encounter must include:

1. **Objective**: What must the PCs achieve (not always "kill everything")
2. **Terrain**: Dimensions, cover, difficult terrain, lighting, interactive elements
3. **Enemies**: Creature table with quantity, CR, XP, and behavior notes
4. **Step-by-step Development**:
   - Round 1: Initial enemy positions and actions
   - Rounds 2-3: Escalation, reinforcements, tactical shifts
   - Round 4+: Climax events, condition changes
5. **Alternative Resolutions**: At least 2 non-combat solutions (diplomacy, stealth, ingenuity) with suggested DCs
6. **Loot**: XP per PC, currency, magic items, information/clues
7. **Scaling**: Adjustments for large (6+) and small (2-3) parties
8. **Narrative Consequences**: How victory/defeat/avoidance affects the world

### 3.4 Tactical Design Rules

- **Cover**: Every combat area needs cover elements
- **Verticality**: Multiple elevation levels when possible
- **Interactive Elements**: Objects PCs can use (barrels, chandeliers, altars)
- **Enemy Tactics**: Each creature needs priorities (attack weakest? protect leader? flee at 50% HP?)
- **Retreat Conditions**: When do enemies flee or surrender?

---

## 4. Stat Block Format (Official)

```
## Monster Name
*Size type, alignment*

**Armor Class:** XX (armor type)
**Hit Points:** XX (XdX + X)
**Speed:** XX ft.

| STR | DEX | CON | INT | WIS | CHA |
|:---:|:---:|:---:|:---:|:---:|:---:|
| XX (+X) | XX (+X) | XX (+X) | XX (+X) | XX (+X) | XX (+X) |

**Saving Throws:** Str +X, Dex +X
**Skills:** Athletics +X, Perception +X
**Damage Resistances:** cold, fire
**Damage Immunities:** necrotic
**Condition Immunities:** charmed, frightened
**Senses:** darkvision 60 ft., passive Perception 13
**Languages:** Common, Infernal
**Challenge:** X (XXX XP)

### Traits

**Trait Name.** Description of passive or active trait.

### Actions

**Weapon Attack.** *Weapon Attack:* +X to hit, reach 5 ft., one target. *Hit:* X (XdX + X) slashing damage.

### Legendary Actions (if applicable)

The creature can take X legendary actions, choosing from the options below. Only one legendary action can be used at a time and only at the end of another creature's turn.

- **Legendary Action 1.** Description
- **Legendary Action 2.** Description

### Reactions (if applicable)

**Reaction Name.** Description of trigger and effect.
```

### 4.1 Monster Design Rules

- **CR Balanced for Level**: Level 1 boss should be CR 1-2 with exploitable weaknesses. Minions CR 1/8 to 1/4.
- **Clear Weaknesses**: Every creature needs at least one weakness PCs can discover and exploit
- **Varied Actions**: Each creature needs at least 2 combat options (not just "attack")
- **Tactics Matter**: Tell the DM HOW to use this creature in combat
- **Lore Integration**: Each entry should have description that helps the DM roleplay the creature

---

## 5. NPC Design

### 5.1 High-Impact NPC Profile

Every major NPC must have:

1. **Name and Role**: Memorable name consistent with tone
2. **Race/Class**: With level if applicable
3. **Alignment**: LG, NG, CG, LN, N, CN, LE, NE, CE
4. **Appearance**: Distinctive visual description
5. **Personality**: 2-3 sentences defining speech, tics, mannerisms
6. **Motivation**: What they want. Everyone wants something.
7. **Secret**: Something hidden. Must be RELEVANT — if discovered, it changes something.
8. **Connections**: Who they relate to (other NPCs, factions, locations)
9. **Information They Hold**: What they know that PCs can obtain
10. **Roleplay Notes**: Voice, gestures, typical phrases
11. **Typical Quote**: One dialogue line capturing their essence

### 5.2 NPC Balance

- **Clear Allies** (1-2): Genuinely want to help
- **Useful Neutrals** (2-3): Help if it benefits them
- **Hidden Hostiles** (1): Appear as allies but work for the villain
- **The Villain**: Comprehensible motivation, not evil for evil's sake

### 5.3 Faction System

Each faction must have:
- **Type**: Guild, cult, noble court, etc.
- **Objective**: What the group wants to achieve
- **Relationship with PCs**: Friendly, neutral, hostile, deceptive
- **Leader**: Name and brief description (reference an NPC)
- **Resources**: Weapons, information, shelter, money, influence
- **Group Motivation**: Why this faction exists. What they need.
- **Quests by Tier**: If applicable, quests offered by PC level
- **Reaction Matrix**: What happens if PCs help/oppose/ignore them

---

## 6. Campaign Design Principles

### 6.1 The Three-Pillar Rule
Every session should touch at least 2 of the 3 D&D pillars:
- **Combat**
- **Exploration**  
- **Social Interaction**

### 6.2 Session Structure
- **Opening Hook**: Get players engaged in first 10 minutes
- **Rising Action**: Build tension through scenes
- **Climax**: Major combat, revelation, or decision
- **Falling Action**: Consequences, loot, clues
- **Closing Hook**: Tease next session

### 6.3 Pacing Rules
- Alternate tension and relief; never maintain maximum intensity
- Every act should have combat, exploration/investigation, AND social interaction
- Include "breather" scenes between intense encounters

### 6.4 Player Agency
- Always provide multiple valid approaches to problems
- Failed rolls should complicate, not block, progress ("fail forward")
- Let player choices have meaningful consequences

### 6.5 Foreshadowing
Plant seeds for future plot points 2-3 sessions ahead. Every major revelation should have been hinted at earlier.

### 6.6 Loot Guidelines
- Use Treasure Tables in the DMG
- Magic items should feel earned, not random
- Economic balance: don't break the gold curve (refer to DMG tables)

---

## 7. Ability Scores and Modifiers

| Score | Modifier |
|-------|----------|
| 1 | -5 |
| 2-3 | -4 |
| 4-5 | -3 |
| 6-7 | -2 |
| 8-9 | -1 |
| 10-11 | +0 |
| 12-13 | +1 |
| 14-15 | +2 |
| 16-17 | +3 |
| 18-19 | +4 |
| 20-21 | +5 |
| 22-23 | +6 |
| 24-25 | +7 |
| 26-27 | +8 |
| 28-29 | +9 |
| 30 | +10 |

---

## 8. Size Categories

| Size | Space | Examples |
|------|-------|----------|
| Tiny | 2.5x2.5 ft | Rat, sprite |
| Small | 5x5 ft | Goblin, dwarf |
| Medium | 5x5 ft | Human, elf |
| Large | 10x10 ft | Bear, ogre |
| Huge | 15x15 ft | Young dragon |
| Gargantuan | 20x20+ ft | Ancient dragon |

---

## 9. CR to Proficiency Bonus

| CR | Bonus |
|----|-------|
| 0-4 | +2 |
| 5-8 | +3 |
| 9-12 | +4 |
| 13-16 | +5 |
| 17-20 | +6 |
| 21-24 | +7 |
| 25-28 | +8 |
| 29-30 | +9 |

---

## 10. Damage Types

- **Acid**: Corrosion, dissolution
- **Bludgeoning**: Strikes, falls
- **Cold**: Ice, freezing
- **Fire**: Burns, heat
- **Force**: Pure magical energy
- **Necrotic**: Decay, death
- **Piercing**: Pointed weapons
- **Psychic**: Mental damage
- **Radiant**: Divine light
- **Slashing**: Blades
- **Lightning**: Electricity
- **Thunder**: Sound, vibration

---

## 11. Development Rule for Grimorio

**Every new MCP tool MUST update:**

1. **Relevant agent(s)** that use it (in `/agents/`)
2. **Architecture diagrams** in README.md (EN + ES)
3. **MCP tools table** in README.md
4. **install.sh output** (if user-facing)
5. **This skill** if it introduces new D&D 5e mechanics or design patterns

**Template updates are MANDATORY when:**
- New content types are added (new markdown sections)
- Official WotC format changes (e.g., new stat block layout)
- Coherence requirements demand new fields (e.g., "Prerequisites" in clues)
- New game modes require new structures (e.g., "Downtime Activities")

---

## 12. Key Lessons from Official Adventures

| Adventure | What It Does Well | What Grimorio Should Copy |
|---|---|---|
| **Curse of Strahd** | Coherent sandbox with omnipresent reactive villain | "Villain awareness" system tracking PC actions and generating responses |
| **Out of the Abyss** | Madness escalation; random event tables affecting mood | Contextualized random tables matching campaign tone |
| **Rise of Tiamat** | Faction scorecard with numerical loyalty tracking | `FactionReputation` domain model with change events |
| **Waterdeep: Dragon Heist** | 4 villain variants, each changing the plot | "Seed" system that permutes campaign without breaking coherence |
| **Icewind Dale: Rime of the Frostmaiden** | Chapter 0 with personalized PC hooks | `CharacterHook` generation connected to backgrounds |
| **Candlekeep Mysteries** | Modular one-shots with self-contained consistency | One-shot template with narrative closure validation |
