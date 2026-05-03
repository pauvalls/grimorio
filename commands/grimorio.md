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

### Phase 3: Launch Orchestrator (SINGLE delegate call)
Launch the **grimorio-orchestrator** subagent with ALL campaign parameters.

You MUST pass these parameters in the prompt:
- `campaign_path` — the full path returned by create_campaign
- `campaign_name` — the kebab-case campaign name
- `setting` — the campaign description/setting
- `level_range` — e.g., "1-3", "4-6"
- `tone` — e.g., "heroic", "dark"
- `duration` — e.g., "one-shot", "3-5 sessions"
- `is_oneshot` — true if one-shot, false if campaign

Example:
```
delegate(
  agent="grimorio-orchestrator",
  prompt="Coordinate campaign generation for 'sunken-city'.\n\ncampaign_path: /home/pau/campaigns/sunken-city\ncampaign_name: sunken-city\nsetting: A sunken city where nobles are aquatic vampires...\nlevel_range: 4-6\ntone: dark\nduration: 3-5 sessions\nis_oneshot: false"
)
```

**CRITICAL:** This is the ONLY `delegate` call you make. The orchestrator handles ALL other subagents internally. Do NOT launch any other subagents from this thread.

### Phase 4: Report
After the orchestrator completes, report to the user:
- Where the PDF was saved
- What content was generated
- Any issues encountered

**DO NOT call `delegation_list` repeatedly. Launch the orchestrator once and wait for it to complete.**