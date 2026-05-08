# Bestiary Generation (v2 Delta) Specification

## Purpose

Define the modified behavior of the `grimorio-bestiary` agent for v2.0. This spec adds WotC-standard stat block formatting, creature role classification, encounter group composition, and explicit source references.

## ADDED Requirements

### Requirement: Creature Role Classification

Each creature MUST have a combat role assigned from the standard set: Skirmisher, Tank, Controller, Striker, Supporter, or Artillery. The role MUST inform the creature's stat block design and tactical behavior.

#### Scenario: Tank creature

- GIVEN a creature is designed as a frontline defender
- WHEN the creature is generated
- THEN it has the role "Tank"
- AND its stat block reflects high AC and hit points

#### Scenario: Controller creature

- GIVEN a creature manipulates the battlefield with area effects
- WHEN the creature is generated
- THEN it has the role "Controller"
- AND its abilities include crowd-control or terrain manipulation

### Requirement: Encounter Groups

Each creature entry MUST specify which other creatures it commonly appears with in encounters. Encounter groups MUST be listed as a comma-separated set of creature names with optional quantity hints (e.g., "2x Goblin, 1x Goblin Boss").

#### Scenario: Goblin encounter group

- GIVEN a goblin warrior is generated
- WHEN the bestiary entry is complete
- THEN it includes an "Encounter Groups" field
- AND the field lists compatible creatures such as "Goblin, Goblin Boss, Wolf"

### Requirement: Source Reference

Every creature stat block MUST include a source reference. Standard Monster Manual creatures MUST use the format "(MM p. {page})". Custom or modified creatures MUST be marked "(custom)". Unique module-specific creatures MUST use an inline stat block and be marked "(custom — unique to this module)".

#### Scenario: Standard creature reference

- GIVEN a creature is a standard Monster Manual entry (e.g., Guard)
- WHEN the creature is referenced
- THEN it displays "Use **Guard** (MM p. 347)"
- AND no inline stat block is generated

#### Scenario: Custom unique creature

- GIVEN a creature is unique to the campaign module
- WHEN the creature is generated
- THEN it has a full inline stat block
- AND it is marked "(custom — unique to this module)"

#### Scenario: Modified standard creature

- GIVEN a creature is based on a Monster Manual entry but with significant modifications
- WHEN the creature is generated
- THEN it has a full inline stat block
- AND it includes "(custom — modified from **Zombie** MM p. 316)"

## MODIFIED Requirements

### Requirement: Stat Block Format

(Previously: Standard D&D 5e stat block with header, base stats, attributes, defenses, special abilities, and actions)

All stat blocks MUST follow the WotC-standard two-column layout and terminology. The format MUST include, in order:

1. **Name** and size/type/tagline (e.g., "Medium undead, lawful evil")
2. **Armor Class** with armor type
3. **Hit Points** with hit dice (e.g., "22 (5d8)")
4. **Speed** with all movement types
5. **Ability Scores** table (FUE, DES, CON, INT, SAB, CAR) with modifiers
6. **Saving Throws** (if proficient)
7. **Skills** (if proficient)
8. **Damage Resistances / Immunities / Vulnerabilities** (as applicable)
9. **Condition Immunities** (as applicable)
10. **Senses** and passive Perception
11. **Languages** (or "—")
12. **Challenge** rating and XP value
13. **Features** (2–4 special abilities)
14. **Actions** (attacks with bonus, reach, target, hit, damage)
15. **Tactics** (how the creature fights)
16. **Role** (combat role classification)
17. **Encounter Groups** (compatible creatures)
18. **Source** (MM reference or custom mark)

#### Scenario: Formatted stat block

- GIVEN all required fields are provided for a custom creature
- WHEN the stat block is generated
- THEN it follows the exact 18-section structure above
- AND all numeric values use the WotC presentation style

#### Scenario: Standard creature by reference

- GIVEN a standard Monster Manual creature is needed
- WHEN the creature is referenced in an encounter
- THEN only the name, source, role, encounter groups, and tactics are generated
- AND no full stat block is inlined

### Requirement: Tactics Section

(Previously: Tactics were included as free-form DM guidance at the end of each creature entry)

The Tactics section MUST be a structured, bulleted list with the following subsections: **Opening**, **Priorities**, **Retreat**, and **Synergy**. Each subsection MUST contain at most 2 sentences.

#### Scenario: Structured tactics

- GIVEN a creature has combat behavior defined
- WHEN the tactics section is generated
- THEN it contains exactly four subsections: Opening, Priorities, Retreat, Synergy
- AND each subsection has at most 2 sentences

#### Scenario: Tactics linked to role

- GIVEN a creature has the role "Skirmisher"
- WHEN the tactics are generated
- THEN the Priorities subsection emphasizes mobility and hit-and-run
- AND the Retreat subsection specifies at what HP threshold the creature disengages
