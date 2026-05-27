# Dungeon Complex Template

## Overview

An indoor location with traps, monsters, treasure, and exploration challenges. Classic dungeon crawl environment with interconnected rooms and passages.

## Template Structure

### Features (3-5 recommended)
- **Main Chambers**: Large rooms with significant encounters
- **Passages/Corridors**: Connecting tunnels, narrow passages
- **Traps**: Mechanical or magical hazards
- **Treasure Vaults**: Locked rooms with valuable loot
- **Secret Areas**: Hidden doors, concealed compartments
- **Guard Posts**: Monster lairs, sentinel positions

### Encounter Types
- **Undead**: Skeletons, zombies, ghosts, liches
- **Monsters**: Oozes, constructs, aberrations
- **Humanoids**: Cultists, prisoners, rival adventurers
- **Traps**: Pit traps, poison darts, magical runes

### NPC Probabilities
- Helpful: 10% (prisoners, trapped spirits)
- Neutral: 20% (lost adventurers, neutral creatures)
- Hostile: 70% (monsters, guards, undead)

## Usage

```go
template := areaService.selectTemplate("dungeon")
area, err := areaService.GenerateAreaWithContext(
    ctx,
    campaignID,
    chapterID,
    areaNumber,
    "ancient crypt",  // locationHint
    "dungeon",        // settingType
    partyLevel,
    factionContext,
    narrativeState,
)
```

## Example Generated Content

**Boxed Text Theme**: Darkness, decay, ancient architecture, sense of dread, echoing sounds

**Development Branches**:
- IF the party clears the dungeon THEN they gain control of the area
- IF the party triggers alarms THEN additional enemies arrive
