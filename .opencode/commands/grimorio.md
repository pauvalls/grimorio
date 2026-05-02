---
description: Generate a complete D&D 5e campaign or one-shot from an idea
agent: grimorio-architect
subtask: false
---

Generate a D&D 5e campaign or one-shot from the user's idea.

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

### Phase 3: Generate Content via Subagents
Launch subagents in PARALLEL using `delegate` for each of these tasks:
- Delegate lore generation: generate world backstory, setting, conflict
- Delegate acts generation: generate act 1, act 2, act 3
- Delegate NPCs generation: generate 5+ NPCs with factions
- Delegate bestiary generation: generate 3-5 monsters with stat blocks
- Delegate encounters generation: generate 3-5 encounters
- Delegate maps generation: generate scene descriptions

Each subagent must use the grimorio MCP tools to save their output.

### Phase 4: Compile PDF
After ALL subagents complete, use grimorio MCP tool `compile_pdf` to generate the final PDF.

### Phase 5: Report
Tell the user where the PDF and markdown files were saved.
