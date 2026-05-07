# DM Guide — Grimorio MCP v2

> How to run a canon-consistent D&D 5e campaign using Grimorio's narrative coherence system.

---

## Table of Contents

1. [Canon Initialization](#canon-initialization)
2. [Consistency Gates](#consistency-gates)
3. [DM Overrides](#dm-overrides)
4. [Session Prep Workflow](#session-prep-workflow)
5. [Faction Reputation Tracking](#faction-reputation-tracking)
6. [Legacy Mode Safety](#legacy-mode-safety)

---

## Canon Initialization

Every campaign starts with a **Canon Document**. This is the single source of truth for:

- **Facts** — immutable lore (e.g., "Arcane magic is banned in Thornvale")
- **Entities** — NPCs, items, locations, factions
- **Rules** — hard constraints that new content must respect
- **Timeline** — world events that shape the narrative
- **Relationships** — who knows whom, alliances, rivalries

### Via MCP Tool

Use `generate_adventure_bible` with a campaign brief:

```json
{
  "name": "shadows-of-thornvale",
  "level_range": "1-5",
  "tone": "dark",
  "setting_type": "gothic",
  "villain_type": "lich",
  "mcguffin_type": "artifact"
}
```

The tool creates:
- `CanonDocument` (schema v2)
- Empty `NarrativeState` (session tracker)

### After Initialization

Add entities and rules incrementally:
- `validate_canon` — check a content proposal before adding it
- `update_narrative_state` — record session outcomes (deaths, discoveries)
- `update_faction_reputation` — adjust standing after PC actions

---

## Consistency Gates

The **Consistency Gate** is the quality checkpoint for all campaign content.

### How It Works

1. You submit a **batch** of artifacts (acts, quests, NPCs, encounters)
2. The gate validates each artifact against the canon
3. Results:
   - **Approved** — all checks passed, canon updated automatically
   - **Rejected** — critical or error-level violations found; retry prompt given
   - **Retrying** — you resubmit with fixes

### Gate Checks

| Check | Severity | What It Catches |
|-------|----------|-----------------|
| `entity_not_found` | Critical | Referencing an NPC/item that doesn't exist |
| `entity_state_mismatch` | Error | An NPC is marked dead but appears alive |
| `npc_alive_check` | Critical | Dead NPC used in new content |
| `lore_rule_compliance` | Critical | Violates a canon rule (e.g., banned magic) |
| `faction_context` | Warning | Faction reputation not considered |

### Fast Mode

Enable `fast_mode=true` to skip non-critical checks when iterating rapidly. Use full validation before finalizing content.

---

## DM Overrides

Sometimes the story needs to break canon **intentionally**. The DM always has the final word.

### Workflow

1. **Propose** the override as a content batch
2. **Review** the gate rejection report
3. **Decide**:
   - Accept the rejection and adjust the narrative
   - Force the override and update the canon to match

### Forced Override

If you want the gate to approve despite violations:
1. Update the canon directly (add the entity, change the state, relax the rule)
2. Re-run the batch
3. The gate will pass because the canon now supports your narrative

> ⚠️ **Warning**: Forced overrides create ripple effects. Use `evaluate_consequences` after a forced override to see what breaks downstream.

---

## Session Prep Workflow

Before each session, generate a **prep sheet**:

```json
{
  "campaign_id": "shadows-of-thornvale",
  "session_num": 3
}
```

The `generate_session_prep` tool synthesizes:
- **Active quests** and their current status
- **Living world state** — faction standings, consequences pending
- **PC hooks** — personalized reminders for each player
- **Consistency warnings** — known violations that need DM attention

### Post-Session

1. Run `update_narrative_state` with session number and outcomes
2. Run `evaluate_consequences` to see triggered rules
3. Run `update_faction_reputation` for any diplomatic shifts
4. Generate prep for next session

---

## Faction Reputation Tracking

Factions react to PC actions through a **reputation graph**.

### Updating Reputation

```json
{
  "campaign_id": "shadows-of-thornvale",
  "faction_id": "thieves-guild",
  "party_id": "party-alpha",
  "delta": -15,
  "reason": "Refused to return stolen artifact"
}
```

### Propagation

Reputation changes **propagate** to allies and enemies:
- Ally factions: +50% of delta (rounded)
- Enemy factions: −50% of delta (inverted)

Example:
- Thieves Guild: −15 (direct)
- Shadow Council (ally): −8 (propagated)
- City Guard (enemy): +8 (inverted)

### Consequences

Use `evaluate_consequences` to see what triggers when reputation crosses thresholds:
- < −50: Hostile — may send assassins or refuse services
- −50 to 0: Unfriendly — higher prices, limited access
- 0 to 50: Neutral — standard interactions
- 50 to 100: Friendly — discounts, information sharing
- > 100: Allied — military support, exclusive quests

---

## Legacy Mode Safety

`CANON_LEGACY_MODE=1` disables all canon consistency gates. Use this **only** when:

- The canon repository is corrupted or unreadable
- You need to run Grimorio without narrative coherence (e.g., for legacy v1 campaigns)
- Emergency recovery where the gate is blocking legitimate content

### Implications

| Feature | Normal Mode | Legacy Mode |
|---------|-------------|-------------|
| `LoadCanon` | Loads from repo | Returns empty doc |
| `ValidateProposal` | Full validation | Auto-approved |
| `QueryEntity` | Searches canon | Returns empty |
| Consistency Gate | Validates batches | Always approves |

### How to Enable

```bash
export CANON_LEGACY_MODE=1
grimorio
```

Or in Docker:
```bash
docker run -e CANON_LEGACY_MODE=1 grimorio:mcp-v2.0.0
```

> ⚠️ **Warning**: Legacy mode bypasses ALL narrative safety. Content generated in legacy mode may contradict established lore. Always return to normal mode as soon as possible.

---

## See Also

- [Developer Guide](developer-guide.md) — API contracts, subagent authoring
- [Canon Schema](../internal/domain/canon.go) — Full CanonDocument structure
- [Gate Schema](../internal/domain/gate.go) — BatchProposal and GateResult types
