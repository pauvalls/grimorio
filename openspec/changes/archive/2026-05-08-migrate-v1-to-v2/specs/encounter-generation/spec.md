# Encounter Generation (v2 Delta) Specification

## Purpose

Define the modified behavior of the `grimorio-encounters` agent for v2.0. This spec converts encounters into reusable templates referenced by name from areas, adds mandatory tactical maps and conditions, and formalizes round-by-round development and alternative resolution.

## ADDED Requirements

### Requirement: Encounter Templates

Each encounter MUST be generated as a standalone template with a unique name. Areas MUST reference encounters by name only (e.g., "Encounter: **Ambush in the Sewers**"), NOT by embedding the full encounter text inline. The encounter template MUST be stored in the encounters file and referenced from the area file.

#### Scenario: Area references encounter by name

- GIVEN an area contains a combat challenge
- WHEN the area is generated
- THEN it includes a reference such as "Encounter: **The Dockside Ambush**"
- AND the full encounter details exist only in `encounters/encounters.md`

#### Scenario: Multiple areas share one encounter

- GIVEN two different areas feature the same enemy patrol
- WHEN the areas are generated
- THEN both areas reference the same encounter template by name
- AND the encounter template is defined once in the encounters file

### Requirement: Tactical Map

Each encounter MUST include a "Tactical Map" section that describes the combat terrain in sufficient detail for a DM to sketch or visualize it. The Tactical Map MUST specify: approximate dimensions, notable terrain features, cover positions, elevation changes, and interactive elements.

#### Scenario: Warehouse encounter tactical map

- GIVEN an encounter takes place in a warehouse
- WHEN the encounter is generated
- THEN it includes a Tactical Map with dimensions (e.g., "30 ft × 40 ft")
- AND it lists cover (crates), elevation (loading dock, 3 ft), and interactive elements (hanging chains, oil barrels)

### Requirement: Encounter Conditions

Each encounter MUST list the environmental conditions that affect combat. Conditions MUST include: illumination level (bright light, dim light, darkness), difficult terrain (yes/no, specify type), and cover availability (none, partial, abundant). Additional conditions such as weather, temperature, or magical effects SHOULD be included when relevant.

#### Scenario: Cave encounter conditions

- GIVEN an encounter occurs in an underground cavern
- WHEN the encounter is generated
- THEN it specifies "Illumination: Dim light (bioluminescent fungi)"
- AND "Difficult Terrain: Yes (uneven stalagmites)"
- AND "Cover: Partial (rock formations)"

### Requirement: Alternative Resolution

Each encounter MUST provide at least one non-combat resolution path. The alternative resolution MUST be concrete and actionable, not generic advice. It MUST specify what skill checks or conditions enable the alternative, and what the outcome is.

#### Scenario: Diplomatic resolution

- GIVEN an encounter involves hostile guards
- WHEN the encounter is generated
- THEN it includes an alternative resolution such as "DC 15 Persuasion: The guards accept a bribe of 20 gp and look the other way"
- AND the outcome is clearly stated

#### Scenario: Stealth resolution

- GIVEN an encounter involves patrolling creatures
- WHEN the encounter is generated
- THEN it includes an alternative resolution such as "DC 14 Stealth (group check): The party sneaks past while the patrol is at the far end of the corridor"
- AND failure triggers the combat version of the encounter

## MODIFIED Requirements

### Requirement: Round-by-Round Development

(Previously: Encounters included a step-by-step development with 3–6 phases, describing enemy actions and conditions of change)

Each combat encounter MUST include a "Development" section structured explicitly by combat rounds (Round 1, Round 2, Round 3, etc.) up to a maximum of 6 rounds. Each round MUST specify: enemy actions, tactical shifts, and trigger conditions for state changes (reinforcements, retreat, environmental events). After Round 6, a "Resolution" subsection MUST describe how the encounter concludes.

#### Scenario: Three-round encounter

- GIVEN a combat encounter with bandits is designed
- WHEN the development section is generated
- THEN it contains "Round 1", "Round 2", and "Round 3"
- AND each round lists specific enemy actions and tactical shifts
- AND Round 3 specifies retreat conditions (e.g., "Bandits flee if 50% are defeated")

#### Scenario: Encounter with reinforcements

- GIVEN an encounter includes enemy reinforcements
- WHEN the development section is generated
- THEN a specific round (e.g., "Round 3") states "Reinforcements arrive: 2 additional bandits from the north door"
- AND the trigger condition is explicit (e.g., "if the alarm was raised in Round 1")

### Requirement: Encounter Header

(Previously: Header included name, difficulty, total XP, and atmospheric setting)

The encounter header MUST include, in this order: **Name**, **Difficulty** (Easy / Medium / Hard / Deadly), **Total XP**, **Adjusted XP** (accounting for monster quantity multiplier), **Encounter Reference** (link to bestiary entries), and **Area** (the area number where this encounter is referenced).

#### Scenario: Complete encounter header

- GIVEN an encounter is generated for Area 7
- WHEN the header is formatted
- THEN it displays all fields: Name, Difficulty, Total XP, Adjusted XP, Encounter Reference, Area
- AND Encounter Reference links to bestiary entries by name

### Requirement: Enemy Listing

(Previously: Enemies were listed in a table with creature, quantity, CR, and XP)

The enemy listing MUST reference creatures from the bestiary by exact name. For standard Monster Manual creatures, the listing MUST use the format "**{Creature Name}** (MM p. {page}) — Qty: {n}". For custom creatures, it MUST use "**{Creature Name}** (custom) — Qty: {n}". The table MUST include a "Source" column.

#### Scenario: Mixed enemy listing

- GIVEN an encounter uses both standard and custom creatures
- WHEN the enemy table is generated
- THEN standard creatures show "(MM p. 347)" in the Source column
- AND custom creatures show "(custom)" in the Source column
- AND all creatures are referenced by exact bestiary name
