# Generate a D&D 5e campaign or one-shot from the user's idea.

## EXECUTION MODE: Main Thread Orchestration

**You are the orchestrator.** Execute this workflow directly in the main thread. Use MCP tools and delegate to sub-agents as specified below.

### Phase 1: Gather Requirements
Ask the user these questions (one at a time, interactively):
1. What's the campaign name? (kebab-case, e.g. "sunken-city")
2. One-shot or full campaign?
3. **Campaign idea / brief description?** (What story do you want to tell? 2-3 sentences describing the main plot)
4. Player level? (1-3, 4-6, 7-10, 11-15, 16-20)
5. Desired tone? (heroic, dark, humorous, political intrigue)
6. Duration? (one-shot, 3-5 sessions, long campaign)

### Phase 2: Create Campaign Structure
Use the grimorio MCP tool `create_campaign` to create the structure.

### Phase 3-13: End-to-End Orchestration (sequential batches)
Follow strict batch ordering — each batch waits for the previous:

- **Batch 1** (parallel delegate): NPCs, bestiary, maps → Consistency Gate
- **Batch 2** (parallel delegate): lore, quests, encounters, characters → Consistency Gate → Update Narrative State
- **Batch 3** (parallel delegate): SVG maps, areas → Consistency Gate
- **Phase 6**: Artist batch-spec (cover + NPCs + scenes + monsters)
- **Phase 7**: Generate AI images (1x1 sequential, retry missing)
- **Phase 8**: Update ALL markdown references
- **Phase 9**: Living World tools (factions, random tables, handouts, consequences) → Consistency Gate
- **Phase 10**: DM Experience tools (session prep, flowchart)
- **Phase 11**: Final consistency check
- **Phase 12**: Compile PDF (embeds all images + flowchart)
- **Phase 13**: Final report

Report progress to the user after each phase.

### Final: Report
After completion, report to the user:
- Where the PDF was saved
- What content was generated
- Any issues encountered

**Use delegate for content generation sub-agents. Execute orchestration logic in main thread.**
