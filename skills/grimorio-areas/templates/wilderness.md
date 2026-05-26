# Wilderness Encounter Zone Template

## Overview

An outdoor location with natural hazards, wildlife, and exploration opportunities. Best suited for parties traveling through untamed lands.

## Template Structure

### Features (3-5 recommended)
- **Natural Clearings**: Open areas for encounters
- **Hazardous Terrain**: Difficult ground, poisonous plants, unstable ground
- **Water Sources**: Streams, ponds, or waterfalls
- **Wildlife Dens**: Animal lairs or nesting grounds
- **Ancient Landmarks**: Old ruins, standing stones, or sacred groves

### Encounter Types
- **Wildlife**: Bears, wolves, giant spiders, etc.
- **Humanoids**: Bandits, hunters, druids
- **Monsters**: Owlbears, trolls, wyverns
- **Environmental**: Weather events, natural disasters

### NPC Probabilities
- Helpful: 30% (rangers, druids, hermits)
- Neutral: 50% (travelers, hunters)
- Hostile: 20% (bandits, territorial creatures)

## Usage

```go
template := areaService.selectTemplate("wilderness")
area, err := areaService.GenerateAreaWithContext(
    ctx,
    campaignID,
    chapterID,
    areaNumber,
    "dense forest",  // locationHint
    "wilderness",    // settingType
    partyLevel,
    factionContext,
    narrativeState,
)
```

## Example Generated Content

**Boxed Text Theme**: Emphasize natural beauty mixed with danger, sensory details of the wilderness

**Development Branches**:
- IF the party respects nature THEN local druids may aid them
- IF the party destroys the environment THEN wildlife becomes hostile
