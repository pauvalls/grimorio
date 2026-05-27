# Urban District Template

## Overview

A city or town location with social encounters, shops, guards, and urban intrigue. Best for political maneuvering, information gathering, and social challenges.

## Template Structure

### Features (3-5 recommended)
- **Marketplace**: Shops, vendors, trading posts
- **Guard Posts**: City watch stations, checkpoints
- **Taverns/Inns**: Social hubs, rumor mills
- **Noble Estates**: Mansions, palaces, government buildings
- **Secret Locations**: Thieves' dens, hidden meeting spots

### Encounter Types
- **Social**: Merchants, nobles, beggars, informants
- **Guard Encounters**: City watch, patrols, inspections
- **Criminal**: Thieves, smugglers, assassins
- **Environmental**: Crowds, festivals, curfews

### NPC Probabilities
- Helpful: 50% (allies, informants, merchants)
- Neutral: 40% (citizens, guards, officials)
- Hostile: 10% (criminals, rival factions)

## Usage

```go
template := areaService.selectTemplate("urban")
area, err := areaService.GenerateAreaWithContext(
    ctx,
    campaignID,
    chapterID,
    areaNumber,
    "busy marketplace",  // locationHint
    "urban",            // settingType
    partyLevel,
    factionContext,
    narrativeState,
)
```

## Example Generated Content

**Boxed Text Theme**: Bustling activity, diverse crowds, architecture, sounds and smells of the city

**Development Branches**:
- IF the party gains favor with city officials THEN they receive official support
- IF the party breaks the law THEN they become wanted criminals
