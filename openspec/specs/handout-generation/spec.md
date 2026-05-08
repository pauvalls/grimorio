# Handout Generation Specification

## Purpose

Define the behavior of the handout generation system, which creates player-facing documents from campaign content.

## Requirements

### Requirement: Player Map Generation

The system MUST generate a player version of each dungeon map. Player maps MUST exclude secret doors, trap locations, hidden treasure, and numbered area labels. Player maps MUST include visible doors, stairs, and major landmarks.

#### Scenario: Dungeon with secrets

- GIVEN a dungeon map with 3 secret doors and 2 traps
- WHEN the player map is generated
- THEN secret doors are not shown
- AND traps are not shown
- AND visible doors and stairs are shown

### Requirement: Clue Handout

The system MUST generate a handout listing all clues the players have discovered. The handout MUST use second-person present tense ("You know that..."). The handout MUST be updated after each session based on narrative state.

#### Scenario: Mid-campaign clues

- GIVEN players have discovered 3 clues across 2 sessions
- WHEN the clue handout is generated
- THEN it lists each clue in player-facing language
- AND it excludes information the players have not discovered

### Requirement: NPC Quick Reference

The system MUST generate a one-page reference sheet for all NPCs the players have met. Each entry MUST include: name, race/class (if known), location, and relationship to party. The sheet MUST be formatted for quick lookup during play.

#### Scenario: Met 5 NPCs

- GIVEN the party has interacted with 5 NPCs
- WHEN the NPC reference is generated
- THEN it includes all 5 NPCs with known information
- AND it excludes NPCs the party has not met

### Requirement: Session Recap

The system MUST generate a session recap handout summarizing key events, decisions, and discoveries from the previous session. The recap MUST be 1-2 paragraphs maximum.

#### Scenario: After session 3

- GIVEN session 3 involved 4 areas, 2 combats, and 1 key decision
- WHEN the recap handout is generated
- THEN it summarizes the key events in 1-2 paragraphs
- AND it highlights the decision point for player reference
