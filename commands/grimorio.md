---
description: Generate a complete D&D 5e campaign or one-shot from an idea
agent: grimorio-architect
subtask: false
---

Generate a D&D 5e campaign or one-shot from the user's idea.

## IMPORTANT: Use `delegate` tool to launch ALL subagents. NEVER do the work yourself.

## Workflow

### Phase 1: Gather Requirements
Ask the user these questions (one at a time, interactively):
1. What's the campaign name? (kebab-case, e.g. "sunken-city")
2. One-shot or full campaign?
3. Player level? (1-3, 4-6, 7-10, 11-15, 16-20)
4. Desired tone? (heroic, dark, humorous, political intrigue)
5. Duration? (one-shot, 3-5 sessions, long campaign)

### Phase 2: Create Campaign Structure
Use the grimorio MCP tool `create_campaign` to create the structure.

### Phase 3: Launch ALL Subagents in PARALLEL (using `delegate`)
You MUST use the `delegate` tool to launch these subagents. Do NOT generate content yourself.

**FIRST — Launch grimorio-cartographer (images and maps):**
- Use `delegate` with agent="grimorio-cartographer"
- The cartographer WILL generate:
  - Cover art: `generate_image` (type: cover, filename: cover-art)
  - Battle maps for EVERY scene: `generate_map` (style: dungeon/landscape/city, use labels parameter)
  - NPC portraits: `generate_image` (type: portrait, filename: npc-{name})
  - Scene illustrations: `generate_image` (type: scene, filename: scene-{act}-{scene}-{name})
- The cartographer WILL update markdown files with image references
- **THIS IS MANDATORY — never skip the cartographer**

**IN PARALLEL — Launch text content subagents:**
- `delegate` lore generation: world backstory, setting, conflict
- `delegate` NPCs generation: 5+ NPCs with factions
- `delegate` bestiary generation: 3-5 monsters with stat blocks
- `delegate` encounters generation: 3-5 encounters
- `delegate` maps generation: scene descriptions with zone breakdowns

### Phase 4: Generate Acts (LAST — after ALL other content exists)
- `delegate` acts generation: generate act 1, act 2, act 3
  - Acts MUST reference existing content by name:
    - NPCs from npcs_and_factions.md
    - Monsters from bestiary.md
    - Encounters from encounters.md
    - Maps: `![Mapa](assets/actX-sceneY-name.svg)`
    - Illustrations if generated: `![Escena](assets/scene-actX-sceneY-name.png)`
    - "Zonas del mapa" linking zones to story elements

### Phase 5: Compile PDF
Use grimorio MCP tool `compile_pdf` to generate the final PDF.

### Phase 6: Report
Tell the user where the PDF and markdown files were saved. Report which images were generated.