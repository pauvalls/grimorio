# Mixed Exploration/Combat Template

## Overview

A balanced location with both exploration challenges and combat encounters. Versatile template suitable for various campaign situations.

## Template Structure

### Features (3-5 recommended)
- **Entry Area**: Gateway, entrance, transition zone
- **Exploration Zones**: Areas requiring skill checks and investigation
- **Combat Arenas**: Clear spaces for tactical encounters
- **Hazard Zones**: Environmental dangers, traps
- **Treasure Locations**: Hidden or guarded rewards
- **Clue Sites**: Information sources, plot advancement

### Encounter Types
- **Combat**: Balanced mix of creatures and humanoids
- **Exploration**: Puzzles, skill challenges, navigation
- **Social**: Brief NPC interactions, informants
- **Environmental**: Weather, terrain, natural hazards

### NPC Probabilities
- Helpful: 40% (guides, locals, quest givers)
- Neutral: 40% (travelers, merchants, bystanders)
- Hostile: 20% (enemies, territorial creatures)

## Usage

```go
template := areaService.selectTemplate("mixed")
area, err := areaService.GenerateAreaWithContext(
    ctx,
    campaignID,
    chapterID,
    areaNumber,
    "borderlands",  // locationHint
    "mixed",        // settingType
    partyLevel,
    factionContext,
    narrativeState,
)
```

## Example Generated Content

**Boxed Text Theme**: Balance of beauty and danger, variety of terrain, sense of adventure

**Development Branches**:
- IF the party explores thoroughly THEN they discover hidden benefits
- IF the party rushes through THEN they miss opportunities and face greater dangers later
