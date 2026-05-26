# Social Encounter Location Template

## Overview

A court, tavern, meeting hall, or temple focused on NPC interactions, diplomacy, and social challenges. Minimal combat, maximum roleplay opportunities.

## Template Structure

### Features (3-5 recommended)
- **Main Hall**: Central gathering space
- **Private Chambers**: Meeting rooms, VIP areas
- **Service Areas**: Kitchens, storage, servant quarters
- **Stage/Platform**: Performance area, throne, podium
- **Secret Listening Posts**: Hidden areas for eavesdropping

### Encounter Types
- **Negotiations**: Trade deals, alliances, treaties
- **Information Gathering**: Rumors, secrets, clues
- **Social Challenges**: Persuasion, deception, intimidation
- **Entertainment**: Performances, games, competitions

### NPC Probabilities
- Helpful: 60% (allies, contacts, patrons)
- Neutral: 30% (bystanders, officials, servants)
- Hostile: 10% (rivals, spies, saboteurs)

## Usage

```go
template := areaService.selectTemplate("social")
area, err := areaService.GenerateAreaWithContext(
    ctx,
    campaignID,
    chapterID,
    areaNumber,
    "royal court",  // locationHint
    "social",       // settingType
    partyLevel,
    factionContext,
    narrativeState,
)
```

## Example Generated Content

**Boxed Text Theme**: Opulence or simplicity depending on venue, social dynamics, atmosphere of anticipation

**Development Branches**:
- IF the party succeeds in diplomacy THEN they gain powerful allies
- IF the party causes a scandal THEN they are banished or arrested
