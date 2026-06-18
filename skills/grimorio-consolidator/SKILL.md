---
name: grimorio-consolidator
version: "1.0.0"
description: Cross-file campaign consistency engine. Detects and fixes entity name collisions, lore contradictions, stat-block drift, duplicate events, duplicate files, stale generated artifacts, and broken map references. Runs after grimorio-architect and before grimorio-integrator.
---

# grimorio-consolidator — Consistency Engine

> **Role:** Autonomous quality gate between content generation (architect) and
> final assembly (integrator). Detects coherence drift across every markdown
> file in a campaign and applies safe fixes without human input. Only genuine
> ambiguities are escalated to the user/agent.

## When to Run

- **Mandatory** after every chapter / macro-phase that wrote markdown to disk
  (lifecycle: architect → consolidator → integrator).
- **Optional but recommended** before `compile_pdf` to catch issues the
  per-chapter `narrative-custodian` gate cannot see (it checks single files,
  not cross-file coherence).

## Tools (MCP)

| Tool | Purpose | Mutates? |
|------|---------|----------|
| `detect_inconsistencies` | Read-only drift detection | No |
| `consolidate_campaign` | Detect + apply safe fixes | Yes |
| `resolve_ambiguity` | Apply user/agent decision to a specific ambiguity | Yes |
| `regenerate_index` | Rebuild `INDEX.md` with breadcrumbs | Yes |
| `verify_campaign_freshness` | Compare `campaign.md`/`INDEX.md` vs sources | No |

The orchestrator pattern is:

```
detect_inconsistencies
  → if remaining_issues == 0 and needs_human == 0: ship
  → elif auto_fix: consolidate_campaign(auto_fix=true)
       → if needs_human is non-empty: ask user, then resolve_ambiguity
  → regenerate_index           # always safe; idempotent
  → verify_campaign_freshness  # report only; do NOT regenerate here
```

## The Seven Checks

Each check produces a `CheckResult` (validation engine) or `HealthFinding`
(health check) with severity `critical` / `error` / `warning`.

| Rule | Severity | What it catches |
|------|----------|-----------------|
| `consolidation_lore_coherence` | critical | Treaty dates that contradict, events placed in multiple files, primordial entities referenced inconsistently |
| `consolidation_stat_block_consistency` | error | Same boss with conflicting CR across bestiary/acts; bestiary wins |
| `consolidation_event_canonical_location` | error | Key events ("Murder of X", "Fall of Y", …) described in multiple acts |
| `consolidation_entity_uniqueness` | warning | Similar entity names (Levenshtein ≥ 0.85 or token-Jaccard ≥ 0.8) |
| `consolidation_map_assets_exist` | error | Map/SVG references that don't resolve to a file |
| `consolidation_no_duplicate_files` | warning | Two files with byte-identical content (kept by the auto-fix, the second is removed) |
| `consolidation_generated_file_freshness` | warning | `campaign.md` / `INDEX.md` older than the newest source |

## Ambiguity Resolution Criteria

The engine emits an `AmbiguityQuestion` (not a Fix) when:

- **Entity name collision** similarity is in `[0.6, 0.85)` — too similar to
  ignore, not similar enough to auto-merge. Options: each name + `keep_both`.
- **Event placement** when a key event shows up in multiple files. Options:
  the list of files (which one is the canonical home).

When asking the user, prefer options in this order:
1. The most likely canonical form (longest name for entities, first-seen
   for events).
2. The explicit alternative form.
3. `keep_both` / first file (i.e. do nothing).

## Rewrite Rules (Safe Auto-Fixes)

These are the ONLY mutations `consolidate_campaign(auto_fix=true)` applies:

1. **Entity renames** — when similarity ≥ 0.85, replace the variant with the
   canonical name in every markdown file (literal string replace).
2. **Duplicate file removal** — when two files have identical content, keep
   the first path (alphabetical), delete the rest.
3. **`INDEX.md` regeneration** — when `regenerate_index` is called, rewrite
   `INDEX.md` from scratch using the current source list.

Anything that would alter canon (treaty dates, boss CR, lore facts) is **NEVER**
auto-fixed; it must be resolved via `resolve_ambiguity` after a user/agent
decision.

## Backup Behavior

Before applying any fix, the engine copies affected files to
`.consolidation/backups/<timestamp>/<rel-path>`. Rollback is always possible
even if git is dirty.

## Pipeline Placement

The architect agent runs `consolidate_campaign` after macro-phases 2 and 3
and before `grimorio-integrator`. The integrator re-runs
`verify_campaign_freshness` to confirm `campaign.md` reflects the latest
sources.

## Failure Modes

| Symptom | Likely cause | Action |
|---------|--------------|--------|
| `consolidation_lore_coherence` always critical | Two agents wrote the same treaty with different years | `resolve_ambiguity` per question, or update one source |
| `consolidation_map_assets_exist` always error | Map generation never ran | Run `generate_map` for the missing slug, or remove the reference |
| `consolidation_no_duplicate_files` always warning | Architect template emits the same boilerplate | Move shared content to a single file and have the others link to it |
| Auto-fixes silently no-op | Files were not committed between phases | Check git status; consolidate only writes if there is a real change |

## Reference

- **Engine code**: `internal/services/consolidation/`
- **Domain types**: `internal/domain/consolidation.go`
- **Adapter for Validation/Health**: `internal/services/consolidation_adapter.go`
- **MCP tools**: `internal/mcp/handlers/consolidation.go`
