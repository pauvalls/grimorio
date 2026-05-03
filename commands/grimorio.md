---
description: Generate a complete D&D 5e campaign or one-shot from an idea
agent: grimorio-architect
subtask: false
---

Generate a D&D 5e campaign or one-shot from the user's idea.

## IMPORTANT: Use the grimorio-architect agent. It handles everything end-to-end.

## Workflow (followed by grimorio-architect)

### Phase 1: Gather Requirements
Ask the user these questions (one at a time, interactively):
1. What's the campaign name? (kebab-case, e.g. "sunken-city")
2. One-shot or full campaign?
3. Player level? (1-3, 4-6, 7-10, 11-15, 16-20)
4. Desired tone? (heroic, dark, humorous, political intrigue)
5. Duration? (one-shot, 3-5 sessions, long campaign)

### Phase 2: Create Campaign Structure
Use the grimorio MCP tool `create_campaign` to create the structure.

### Phase 3-10: End-to-End Orchestration (sequential batches)
The architect follows strict batch ordering — each batch waits for the previous:

- **Batch 1** (parallel): NPCs, bestiary, maps
- **Batch 2** (parallel): lore, quests, encounters, characters
- **Batch 3** (parallel): SVG maps, acts (needs ALL prior content)
- **Sequential**: artist batch-spec (cover + NPCs + scenes + monsters) → generate images (1x1, retry missing) → update ALL references → PDF

The architect reports progress to the user after each phase.

### Final: Report
After completion, report to the user:
- Where the PDF was saved
- What content was generated
- Any issues encountered

**DO NOT launch subagents from the command thread — the architect manages all delegation internally.**
