# Campaign Integration Specification

## Purpose

Define the behavior of the `grimorio-integrator` agent, which validates cross-references, checks balance, and ensures consistency across all campaign components.

## Requirements

### Requirement: Cross-Reference Validation

The system MUST verify that every creature name referenced in areas exists in the bestiary file. The system MUST verify that every NPC name referenced in areas exists in the NPCs file. The system MUST verify that every area number referenced in connections exists.

#### Scenario: Missing creature reference

- GIVEN an area references "Shadow Wraith"
- AND "Shadow Wraith" does not exist in bestiary.md
- WHEN the integrator runs
- THEN it reports the missing reference
- AND suggests adding the creature to bestiary.md or correcting the name

#### Scenario: Broken area connection

- GIVEN Area 3 references "→ Area 7"
- AND Area 7 does not exist in the act
- WHEN the integrator runs
- THEN it reports the broken connection
- AND suggests creating Area 7 or correcting the reference

### Requirement: XP Budget Calculation

The system MUST calculate total XP per act by summing all creature XP and encounter XP. The system MUST calculate XP per player character by dividing total XP by assumed party size (4-5). The system MUST compare XP per PC against level-appropriate thresholds.

#### Scenario: Level 1 act balance

- GIVEN an act designed for level 1 characters
- WHEN XP is calculated
- THEN total XP per PC MUST be between 300-400 XP
- AND the integrator reports "balanced" or "adjust needed"

### Requirement: Treasure Consistency

The system MUST verify that every area containing creatures or NPCs has treasure documented. The system MUST verify that treasure format follows the standard: "**Treasure.** [Location]: [Amount] gp, [Items with value]".

#### Scenario: Missing treasure

- GIVEN Area 4 contains 2 goblins
- AND Area 4 has no treasure section
- WHEN the integrator runs
- THEN it reports missing treasure
- AND suggests adding appropriate treasure for CR 1/4 creatures

### Requirement: Bidirectional Connection Check

The system MUST verify that all area connections are bidirectional. If Area A references Area B, then Area B MUST reference Area A.

#### Scenario: One-way connection

- GIVEN Area 2 has "→ Area 5"
- AND Area 5 does not reference Area 2
- WHEN the integrator runs
- THEN it reports the one-way connection
- AND suggests adding the reverse connection

### Requirement: Completeness Check

The system MUST verify that each area has at least one interactive element (creature, NPC, trap, puzzle, or treasure). The system MUST verify that all DC values are numeric. The system MUST verify that no NPC appears in two locations simultaneously.

#### Scenario: Empty area

- GIVEN Area 8 has no creatures, NPCs, traps, or treasure
- WHEN the integrator runs
- THEN it reports the empty area
- AND suggests adding an environmental hazard or decorative clue
