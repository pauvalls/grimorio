# Area Generation Specification

## Purpose

Define the behavior of the `grimorio-areas` agent, which generates numbered playable areas for D&D campaigns following WotC adventure format.

## Requirements

### Requirement: Area Count and Size

The system MUST generate 10-15 numbered areas per act. Each area MUST contain 150-200 words. The total word count per act MUST NOT exceed 3,000 words.

#### Scenario: Standard act generation

- GIVEN a campaign with standard complexity
- WHEN the areas agent generates an act
- THEN it produces 10-15 areas
- AND each area has between 150-200 words

#### Scenario: One-shot generation

- GIVEN a one-shot campaign request
- WHEN the areas agent generates the single act
- THEN it produces 8-12 areas
- AND each area has between 150-200 words

### Requirement: Area Numbering and Naming

Each area MUST have a unique number within the act (e.g., "A1", "A2" or "1", "2"). Each area MUST have a descriptive name following the format: "[Number]. [Descriptive Name]".

#### Scenario: Area numbering consistency

- GIVEN an act with multiple areas
- WHEN areas are generated
- THEN each area has a sequential number
- AND no two areas share the same number

### Requirement: Mandatory Area Sections

Each area MUST include the following sections in order:
1. Read-Aloud text (2-4 sentences, second person, present tense) — OPTIONAL for non-essential areas
2. Features list (3-5 bullet points describing physical characteristics)
3. Mechanics (creatures, NPCs, or interactive elements with specific DCs)
4. Treasure (if creatures/NPCs present: XP amount, currency, items with values)
5. Connections (bidirectional references to other areas using numbers)
6. Secrets/Traps (if applicable: detection DC, mechanism, consequence)

#### Scenario: Combat area

- GIVEN an area containing hostile creatures
- WHEN the area is generated
- THEN it includes creature stats reference, treasure with XP, and at least one tactical element

#### Scenario: Social area

- GIVEN an area containing NPCs without combat
- WHEN the area is generated
- THEN it includes NPC motivations, dialogue hooks, and potential information gained

### Requirement: Specific DC Values

All Difficulty Class (DC) values MUST be numeric. The system MUST NOT use relative terms like "high", "low", or "difficult". Standard DC ranges: Easy (10), Moderate (12-14), Hard (15-18), Very Hard (20+).

#### Scenario: Trap detection

- GIVEN an area with a hidden trap
- WHEN the area is generated
- THEN the detection DC is a specific number (e.g., "DC 14 Wisdom (Perception)")
- AND the disarm/disable DC is a specific number

### Requirement: Cross-References

Each area MUST reference other areas using numbered connections (e.g., "→ Area 3"). Each area MUST reference creatures using exact names from the bestiary. Each area MUST reference NPCs using exact names from the NPCs file.

#### Scenario: Area connections

- GIVEN Area 2 connects to Area 5
- WHEN Area 2 is generated
- THEN it includes "→ Area 5" in its Connections section
- AND Area 5 includes "← Area 2" in its Connections section
