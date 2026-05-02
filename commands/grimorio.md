---
description: Generate a complete D&D 5e one-shot or campaign from an idea
argument-hint: [campaign-idea]
allowed-tools: [
  "mcp__plugin_grimorio_grimorio__create_campaign",
  "mcp__plugin_grimorio_grimorio__get_template",
  "mcp__plugin_grimorio_grimorio__save_act",
  "mcp__plugin_grimorio_grimorio__save_npcs",
  "mcp__plugin_grimorio_grimorio__save_bestiary",
  "mcp__plugin_grimorio_grimorio__save_encounters",
  "mcp__plugin_grimorio_grimorio__save_maps",
  "mcp__plugin_grimorio_grimorio__compile_pdf"
]
---

# Grimorio Campaign Generation Workflow

## Step 1: Gather Requirements

Before generating, ask the user:
- Type: one-shot or full campaign
- Approximate player level (1-3, 4-6, 7-10, 11-15, 16-20)
- Desired tone (heroic, dark, humorous, political intrigue)
- Duration: one-shot, 3-5 sessions, or long campaign
- Campaign name (kebab-case, e.g. "sunken-city")

If no name is provided, suggest one based on the idea.

## Step 2: Create Campaign Structure

Use MCP tool `create_campaign` with:
- name: kebab-case name
- title: readable title
- setting: brief setting description

## Step 3: Generate Lore and Setting

Get template `lore` with `get_template`, then generate complete lore using your D&D 5e knowledge. Save as `lore.md` in the campaign directory (use Write tool directly to the output directory).

Required content:
- 3-5 paragraph general synopsis
- World geography and history
- Central conflict
- Themes and tone
- Narrative inflection points

## Step 4: Generate Acts (minimum 3 for campaigns, 1 for one-shots)

For each act:
1. Get template `act` with `get_template`
2. Generate complete act content
3. Use `save_act` to save it

Each act must include:
- Act summary
- 3 adventure hooks
- Key scenes with read-aloud descriptions
- Clues and information
- Success/failure consequences
- Connection to the next act

## Step 5: Generate NPCs and Factions

Get template `npc`, generate content, save with `save_npcs`.

Include:
- Minimum 5 main NPCs with personality, motivation, and secret
- 2-3 factions with goals and resources
- Representative quotes

## Step 6: Generate Bestiary

Get template `monster`, generate content, save with `save_bestiary`.

Include:
- 3-5 monsters/creatures with formatted stat blocks
- Combat tactics
- Physical descriptions
- Difficulty adjustments by party level

## Step 7: Generate Encounters

Get template `encounter`, generate content, save with `save_encounters`.

Include:
- 3-5 balanced encounters
- Difficulty tables and level adjustments
- Terrain description
- Loot

## Step 8: Generate Maps and Scenes

Get template `map`, generate content, save with `save_maps`.

Include:
- Key scene descriptions
- DM notes
- Area connections

## Step 9: Compile PDF

Use `compile_pdf` to generate the final D&D-styled PDF.

Inform the user:
- Location of all Markdown files
- Location of the final PDF
- Summary of what was generated
