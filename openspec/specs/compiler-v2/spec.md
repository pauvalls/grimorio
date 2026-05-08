# Compiler v2 Specification

## Purpose

Define the behavior of the enhanced PDF compiler, which generates professional-quality campaign documents from markdown sources.

## Requirements

### Requirement: Hierarchical Table of Contents

The compiler MUST generate a hierarchical TOC with 3 levels: Act → Area Groups → Individual Areas. The TOC MUST include page number references. The TOC MUST be clickable, linking to each section.

#### Scenario: Campaign with 3 acts

- GIVEN a campaign with 3 acts, each with 10-15 areas
- WHEN the PDF is compiled
- THEN the TOC shows Act 1 → Area Group 1-5 → Individual Areas
- AND each entry links to its page

### Requirement: Clickable Cross-References

All references to other areas MUST be hyperlinks (e.g., "see Area 3" links to Area 3). All references to NPCs in appendices MUST link to the NPC entry. All references to creatures in the bestiary MUST link to the stat block.

#### Scenario: Area with connections

- GIVEN Area 2 references "→ Area 5" and "see **Noska Ur'gray**"
- WHEN the PDF is compiled
- THEN "Area 5" is a clickable link to that area
- AND "Noska Ur'gray" links to the NPC appendix entry

### Requirement: Inline Stat Blocks

The compiler MUST embed stat blocks inline for creatures unique to the campaign (not in MM). For MM creatures, the compiler MUST display a compact reference (name, CR, page number). Stat blocks MUST use the standard two-column WotC format.

#### Scenario: Unique creature

- GIVEN the bestiary contains "Murmuring Specter" (custom creature)
- WHEN the PDF is compiled
- THEN the stat block appears inline the first time the creature is encountered
- AND it uses the standard stat block format

### Requirement: Area Number Highlighting

Area numbers MUST be visually prominent in the PDF (larger font, bold, or colored background). Area names MUST be clearly separated from body text.

#### Scenario: Reading the PDF

- GIVEN a DM is searching for "Area 7"
- WHEN they scan the document
- THEN Area 7 is visually distinct and easy to locate

### Requirement: Two-Column Layout with Breaks

The PDF MUST use a two-column layout for body text. Area headers MUST span both columns. Maps and stat blocks MAY span both columns. Read-aloud text MUST be visually distinct (boxed or shaded).

#### Scenario: Complex area

- GIVEN an area with 200 words, a stat block, and a map reference
- WHEN the PDF is compiled
- THEN body text flows in two columns
- AND the stat block spans both columns for readability

### Requirement: Handout Pages

The compiler MUST append handout pages at the end of the PDF. Handout pages MUST be clearly labeled "Player Handouts". Handout pages MUST be designed for printing and cutting.

#### Scenario: Full campaign PDF

- GIVEN a complete campaign with player maps, clues, and NPC references
- WHEN the PDF is compiled
- THEN handout pages appear after the main content
- AND each handout is on its own page or clearly separated
