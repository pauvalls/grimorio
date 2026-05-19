# Developer Guide — Grimorio MCP v2

> How to extend, integrate, and build with Grimorio's CanonService and MCP tooling.

---

## Table of Contents

1. [CanonService Interface Contract](#canonservice-interface-contract)
2. [Cache Invalidation Rules](#cache-invalidation-rules)
3. [Subagent MCP Tool Compatibility](#subagent-mcp-tool-compatibility)
4. [Validation Rules Reference](#validation-rules-reference)
5. [Benchmark Conventions](#benchmark-conventions)
6. [Testing Guidelines](#testing-guidelines)

---

## CanonService Interface Contract

`CanonService` is the core domain service. All subagents and handlers interact with the canon through this service.

### Methods

| Method | Input | Output | Cache? | Invalidates? |
|--------|-------|--------|--------|--------------|
| `InitializeCanon(ctx, brief)` | `CampaignBrief` | `*CanonDocument, error` | No | Yes (clears) |
| `LoadCanon(ctx, campaignID)` | `string` | `*CanonDocument, error` | **Yes** | No |
| `SaveCanon(ctx, doc)` | `*CanonDocument` | `error` | No | **Yes** |
| `RegisterFact(ctx, campaignID, fact)` | `string, CanonFact` | `error` | No | **Yes** |
| `QueryEntity(ctx, campaignID, filter)` | `string, EntityFilter` | `[]CanonEntity, error` | **Yes** (via LoadCanon) | No |
| `UpdateEntityState(ctx, campaignID, entityID, state)` | `string, string, EntityState` | `error` | No | **Yes** |
| `ValidateProposal(ctx, campaignID, proposal)` | `string, ContentProposal` | `*ValidationReport, error` | No | No |
| `GetRelationshipGraph(ctx, campaignID)` | `string` | `*RelationshipGraph, error` | **Yes** (via LoadCanon) | No |

### Error Contract

All methods return errors using Go's standard error wrapping (`fmt.Errorf(..., %w, err)`).

| Error Type | When | Handling |
|------------|------|----------|
| `ValidationError` | Input violates domain rules (empty name, invalid kebab-case) | Return 400 to client |
| `fmt.Errorf("canon not found...")` | Repository miss | Initialize canon or return 404 |
| `fmt.Errorf("entity not found...")` | Entity ID not in canon | Return 404 or suggest creation |
| Wrapped repo errors | I/O failures, disk full, permissions | Log and return 500 |

### Degraded Mode

When `SetDegraded(true)` is called:

- `LoadCanon` returns a minimal empty `CanonDocument`
- `ValidateProposal` returns an auto-approved `ValidationReport`
- `QueryEntity` returns an empty slice
- All write methods continue to work normally

This is useful for graceful degradation when the canon repository is unavailable.

---

## Cache Invalidation Rules

The CanonService uses an **in-memory LRU cache** (default size: 100 entries).

### Cache Key

Key is the `campaignID` (string). There is currently no per-query caching for `QueryEntity` or `GetRelationshipGraph` — these rely on `LoadCanon` caching.

### Automatic Invalidation

The cache is automatically invalidated on these operations:
- `InitializeCanon`
- `SaveCanon`
- `RegisterFact`
- `UpdateEntityState`

### Manual Control

If you write a **wrapper** around `CanonService`, you MUST invalidate the cache after any write operation:

```go
// Wrapper example
type MyCanonWrapper struct {
    svc *services.CanonService
}

func (w *MyCanonWrapper) CustomSave(ctx context.Context, doc *domain.CanonDocument) error {
    err := w.svc.SaveCanon(ctx, doc)
    if err != nil {
        return err
    }
    // Cache was already invalidated by SaveCanon — no extra work needed
    return nil
}
```

### Disabling Cache

Set environment variable before startup:

```bash
export CANON_CACHE_DISABLED=1
```

Or adjust size:

```bash
export CANON_CACHE_SIZE=500  # default is 100
```

---

## Subagent MCP Tool Compatibility

Subagents (AI agents that call Grimorio tools) MUST follow these conventions.

### Tool Naming

Use kebab-case tool names:
- ✅ `generate_adventure_bible`
- ❌ `generateAdventureBible`
- ❌ `GenerateAdventureBible`

### Campaign IDs

Campaign IDs MUST be kebab-case (lowercase, hyphens only):
- ✅ `shadows-of-thornvale`
- ❌ `Shadows of Thornvale`
- ❌ `shadows_of_thornvale`

Validation: `domain.IsValidKebabCase(id)`

### Batch Proposal Format

When submitting content through `process_consistency_gate`, use this JSON structure:

```json
{
  "batch_id": "act-3-draft",
  "campaign_id": "shadows-of-thornvale",
  "attempt": 1,
  "artifacts": [
    {
      "id": "act-3",
      "type": "act",
      "content": "# Act III: The Siege\n\nThe party arrives at Thornvale...",
      "entity_references": [
        {"entity_id": "npc-lord-vex", "location": "act_3, scene_2"}
      ]
    }
  ]
}
```

### Required Fields

- `batch_id` — unique identifier for this submission
- `campaign_id` — must exist in the canon repository
- `attempt` — start at 1, increment on retry
- `artifacts` — at least one artifact required

### Handling Gate Results

| Status | Action |
|--------|--------|
| `approved` | Content is canon-safe; proceed with publishing |
| `rejected` | Review `retry_prompt` and `suggestions`; fix and resubmit with `attempt++` |
| `retrying` | Same as rejected; intermediate state |

---

## Validation Rules Reference

The validation engine runs these checks on every proposal:

### Entity Checks

| Rule | Severity | Condition |
|------|----------|-----------|
| `entity_not_found` | Critical | `EntityReference.EntityID` not in canon |
| `entity_state_mismatch` | Error | `EntityReference.RequiredState` ≠ current state |
| `npc_alive_check` | Critical | NPC referenced but marked dead in narrative state |

### Lore Checks

| Rule | Severity | Condition |
|------|----------|-----------|
| `lore_rule_compliance` | Critical | Content contains keywords from a banned/prohibited/forbidden rule |

### Faction Checks

| Rule | Severity | Condition |
|------|----------|-----------|
| `faction_context` | Warning | Faction reputation delta is extreme (<−50 or >50) |

### Consistency Checks

| Rule | Severity | Condition |
|------|----------|-----------|
| `canon_load` | Critical | Canon document could not be loaded |
| `timeline_continuity` | Error | Content references a timeline event that hasn't happened yet |

### Severity Levels

- **Critical** — blocks approval; must be fixed
- **Error** — blocks approval; should be fixed
- **Warning** — does not block; informational

---

## Benchmark Conventions

All benchmarks live in existing `*_test.go` files (colocated with tests).

### Naming

```go
func Benchmark<Component>_<Operation>(b *testing.B) { ... }
```

Examples:
- `BenchmarkCanonService_LoadCanon`
- `BenchmarkValidationEngine_ValidateAct`
- `BenchmarkConsistencyGate_ProcessBatch`

### Seed Data

Benchmarks MUST use **fixed seed data** (no randomness) to ensure reproducibility across runs.

### Running Benchmarks

```bash
# Run all benchmarks
make bench

# Save baseline
make bench-save

# Compare against baseline
make bench-compare
```

### Memory Profiling

```bash
go test ./internal/services/... -bench=. -benchmem -memprofile=mem.out
```

---

## Testing Guidelines

### Unit Tests

- Table-driven tests preferred
- Mock repositories using `repository.NewMemoryCanonRepository()` and `repository.NewMemoryNarrativeStateRepository()`
- Assert specific values, not just non-nil

### Cache Tests

To prove cache hit vs miss:
1. Load document (warms cache)
2. Corrupt/delete repo entry directly
3. Load again — should succeed from cache
4. Invalidate cache
5. Load again — should fail

### Degraded Mode Tests

1. Set `canonService.SetDegraded(true)`
2. Call read methods — assert empty/minimal results
3. Call `ValidateProposal` — assert auto-approved
4. Call write methods — assert they still work

### Coverage Target

- New code: >80% coverage
- Existing code: maintain or improve
- Run: `make coverage`

---

## See Also

- [DM Guide](dm-guide.md) — Using the coherence system as a DM
- [Canon Schema](../internal/domain/canon.go) — Data structures
- [Gate Schema](../internal/domain/gate.go) — BatchProposal and GateResult
- [Cache Implementation](../internal/cache/lru.go) — LRU cache internals

---

## SDD Workflow (Spec-Driven Development)

Grimorio uses **SDD** (Spec-Driven Development) for substantial changes. The orchestrator agent (`gentle-orchestrator`) coordinates the workflow by delegating phases to specialized sub-agents.

### SDD Phases

```
proposal → specs → tasks → apply → verify → archive
               ↑
             design
```

| Phase | Agent | Purpose |
|-------|-------|---------|
| **Explore** | `sdd-explore` | Investigate codebase, compare approaches, clarify requirements. No files created. |
| **Propose** | `sdd-propose` | Formalize intent, scope, approach, and risks. |
| **Spec** | `sdd-spec` | Write delta specs with requirements, scenarios, and acceptance criteria. |
| **Design** | `sdd-design` | Create technical design with file-by-file plan and architecture decisions. |
| **Tasks** | `sdd-tasks` | Break design into concrete, orderable implementation tasks. |
| **Apply** | `sdd-apply` | Implement code changes task by task. |
| **Verify** | `sdd-verify` | Validate implementation against specs, run tests, report compliance. |
| **Archive** | `sdd-archive` | Persist final state, reconcile spec deltas, close the change. |

### Commands

| Command | Description |
|---------|-------------|
| `/sdd-init` | Initialize SDD context; detects stack, bootstraps Engram persistence. |
| `/sdd-explore <topic>` | Investigate an idea; reads codebase, returns report. |
| `/sdd-apply [change]` | Implement tasks in batches. |
| `/sdd-verify [change]` | Validate implementation against specs. |
| `/sdd-archive [change]` | Close a change and persist final state. |

### Meta-Commands

| Command | Description |
|---------|-------------|
| `/sdd-new <change>` | Start a new change by delegating exploration + proposal. |
| `/sdd-ff <name>` | Fast-forward: proposal → specs → design → tasks in one shot. |
| `/sdd-continue [change]` | Run the next dependency-ready phase. |

### Execution Modes

- **Automatic** (`auto`) — All phases run back-to-back without pausing. Final result only.
- **Interactive** (`interactive`) — After each phase, show result and ask to continue.

### Artifact Store

| Mode | Backend | Use Case |
|------|---------|---------|
| `engram` | Engram (default) | Fast, no files created. |
| `openspec` | `openspec/` directory | File-based, shareable artifacts. |
| `hybrid` | Both | Cross-session recovery + local files. |

### Making a Release

After completing an SDD cycle:

```bash
make release-tag
```

This auto-detects the next version from conventional commits, creates and pushes the tag. CI then creates the GitHub release with binaries and updates the CHANGELOG via git-cliff.
