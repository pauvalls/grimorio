# NPC Generation (v2 Delta) Specification

## Purpose

Define the modified behavior of the `grimorio-npc` agent for v2.0. This spec adds alignment, combat stats, location, quest involvement, and secrets to NPC generation.

## ADDED Requirements

### Requirement: NPC Alignment

Each NPC MUST have an alignment specified using the 3x3 grid format (e.g., LG, CG, NE, CN). The alignment MUST influence the NPC's behavior and faction relations.

#### Scenario: Noble NPC

- GIVEN an NPC is a noble with secret dealings
- WHEN the NPC is generated
- THEN it has an alignment (e.g., LE)
- AND the alignment is reflected in motivations and secrets

### Requirement: NPC Location

Each NPC MUST have an initial location specified as an area number (e.g., "Area 3" or "Area M4"). The location MUST be a valid area in the campaign.

#### Scenario: Tavern keeper

- GIVEN an NPC is a tavern keeper
- WHEN the NPC is generated
- THEN it has a location like "Area 2: The Rusty Anchor"
- AND the area description references the NPC

### Requirement: NPC Combat Stats

Each NPC that MAY participate in combat MUST have either: (a) a stat block reference ("Use **Guard** MM p. 347"), or (b) an inline stat block with modifications. NPCs that are non-combatants MUST be explicitly marked as such.

#### Scenario: Combat-capable NPC

- GIVEN an NPC is a city guard captain
- WHEN the NPC is generated
- THEN it includes "Use **Guard Captain** (veteran MM p. 350)"
- OR it includes a custom stat block

#### Scenario: Non-combatant NPC

- GIVEN an NPC is an elderly librarian
- WHEN the NPC is generated
- THEN it is marked "Non-combatant"
- AND it has no combat stats

### Requirement: NPC Quest Involvement

Each NPC MUST be associated with at least one quest or plot thread. The quest involvement MUST specify the NPC's role (quest giver, ally, obstacle, target, informant).

#### Scenario: Quest giver

- GIVEN an NPC starts a major quest
- WHEN the NPC is generated
- THEN it includes "Quest: 'Find the Lost Key' — Quest Giver"

### Requirement: NPC Secrets

Each NPC MUST have 1-2 secrets. Secrets MUST be relevant to the campaign plot. Secrets MUST be discoverable by players through investigation or social interaction.

#### Scenario: Secret traitor

- GIVEN an NPC appears to be an ally
- WHEN the NPC is generated
- THEN it has a secret like "Secret: Works for the villain, reports party movements"
- AND the secret is discoverable with DC 18 Insight

## MODIFIED Requirements

### Requirement: NPC Format

(Previously: Simple list with name, role, description, motivation)

Each NPC MUST follow this format:
```
### Name
*Alignment Race Class*

**Location:** Area X
**Role:** Function in story
**Motivation:** What they want
**Secret:** Hidden information
**Quest:** Involvement in quests
**Stats:** Reference or inline
**Quote:** Characteristic line
```

#### Scenario: Formatted NPC

- GIVEN all required fields are provided
- WHEN the NPC is formatted
- THEN it follows the exact structure above
- AND all fields are populated
