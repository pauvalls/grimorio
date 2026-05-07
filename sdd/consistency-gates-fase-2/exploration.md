## Exploration: consistency-gates-fase-2

### Current State

Phase 1 is complete and solid. The codebase has:

**Services (4 MCP tools active):**
- `CanonService` — InitializeCanon, LoadCanon, SaveCanon, RegisterFact, QueryEntity, ValidateProposal, UpdateEntityState, GetRelationshipGraph
- `NarrativeStateService` — Load, Save, Update (batch), GetSessionPrepContext
- `ValidationEngine` — ValidateAct, ValidateQuest, CheckConsistency
- 4 MCP handlers wired in `internal/mcp/server.go`: `generate_adventure_bible`, `validate_canon`, `update_narrative_state`, `check_consistency`

**Validation rules currently implemented (4 rules):**
1. `entity_not_found` — referenced entity doesn't exist in canon
2. `npc_alive_check` — dead NPC appears in new content
3. `lore_rule_compliance` — content violates a canon rule (naive keyword matching)
4. `mcguffin_continuity` — mcguffin location/possession contradicts narrative state

**Domain model:** Well-structured with `CanonDocument`, `NarrativeState`, `ValidationReport`, `ConsistencyReport`, `ContentProposal`, `EntityReference`, `DMOverride`, etc. All structs have `Validate()` methods.

**Persistence:** Dual repository pattern — `FilesystemCanonRepository` and `MemoryCanonRepository` for both canon and narrative state. JSON files stored at `campaigns/{name}/canon/canon.json` and `narrative_state.json`. Atomic writes via temp+rename.

**Test coverage:** domain 93.1%, services 83.8%, repository 63.9%.

---

### Specific Gaps for Phase 2

| Gap | Severity | Detail |
|-----|----------|--------|
| **ConsistencyGate doesn't exist** | Critical | No checkpoint service between batches. No approve/reject/retry state machine. |
| **Only 4/10 validation rules** | Critical | Need 6 more: `quest_reward_existence`, `level_encounter_balance`, `location_existence`, `timeline_consistency`, `prerequisite_clue_check`, plus one more to reach 10. |
| **No pipeline batch state** | High | No concept of "Batch 1", "Batch 2", or `canon_lock`. Subagents generate without coordination. |
| **No structured retry feedback** | High | `ValidationReport` has `Suggestions` but no mechanism for subagents to consume them as retry prompts. |
| **Auto-save post-gate missing** | Medium | Canon + state persistence exists but no automatic save triggered by gate approval. |
| **check_consistency is partial** | High | Only checks npc_alive, mcguffin_continuity, and lore rules. Missing: quests, acts, encounters, locations, timeline. |
| **lore_rule_compliance is naive** | Medium | Uses simple keyword overlap (`banned` + word match). Needs semantic understanding or at least better parsing. |
| **No file locking** | Medium | `FilesystemCanonRepository` has atomic writes but no `flock`. Race conditions if two agents write simultaneously. |
| **DMOverride not wired** | Medium | `DMOverride` struct exists in `NarrativeState` but `ValidationEngine` doesn't consult it. |
| **No `canon_lock` mechanism** | High | Roadmap mentions locking canon during batch generation. No implementation exists. |

---

### Affected Areas

- `internal/services/validation_engine.go` — add 6 new rules, expand CheckConsistency to full scope
- `internal/services/validation_engine_test.go` — tests for new rules
- `internal/services/consistency_gate.go` **(NEW)** — gate orchestrator service
- `internal/services/consistency_gate_test.go` **(NEW)** — gate tests
- `internal/domain/validation.go` — extend with GateStatus, BatchState types
- `internal/mcp/handlers/canon.go` — add HandleConsistencyGate tool
- `internal/mcp/server.go` — register new MCP tool
- `internal/services/canon_service.go` — add auto-save hook, canon_lock helper
- `internal/repository/interfaces.go` — possibly add `GateRepository` interface
- `internal/repository/filesystem_canon.go` — add file locking (flock)

---

### Approaches

#### 1. ConsistencyGate as a Service (Recommended)

Create a `ConsistencyGateService` that orchestrates batch validation:

```go
type ConsistencyGateService struct {
    canonService     *CanonService
    stateService     *NarrativeStateService
    validationEngine *ValidationEngine
}

func (s *ConsistencyGateService) ProcessBatch(ctx, campaignID, batchID string, proposals []ContentProposal) (*GateResult, error)
```

**Flow:**
1. Lock canon (`canon_lock` flag in narrative state or separate lock file)
2. Validate each proposal via `ValidationEngine`
3. Aggregate into `GateResult` with overall `approved`/`rejected`/`retry`
4. If approved → auto-save canon + state, unlock
5. If rejected → return `ValidationReport` per proposal with structured `Suggestions`
6. If retry → keep lock, return retry prompts

**Pros:**
- Fits existing service-layer architecture
- Easy to test with `MemoryRepo`
- Can be exposed as single MCP tool (`process_consistency_gate`)

**Cons:**
- Requires adding lock/unlock logic to canon persistence
- Gate state needs persistence (what if server crashes mid-batch?)

**Effort:** Medium

---

#### 2. ConsistencyGate as Middleware/Pipeline Decorator

Wrap the existing MCP handlers so every `save_act`, `save_npcs`, etc. automatically triggers validation before persisting markdown.

**Pros:**
- Transparent to subagents — they don't need to change
- Every write is automatically validated

**Cons:**
- Breaks the batch concept — each individual write is validated, not the batch as a unit
- Harder to implement "retry with suggestions" because the write has already happened
- Doesn't solve the coordination problem between parallel agents

**Effort:** High

---

#### 3. ConsistencyGate as External Orchestrator (Out of Scope)

Build a separate orchestrator service/agent that manages the pipeline and calls Grimorio MCP tools.

**Pros:**
- Keeps Grimorio MCP server stateless
- Pipeline logic lives in the agent, not the server

**Cons:**
- Roadmap specifies gate is a "checkpoint programático" inside the MCP server
- Adds deployment complexity

**Effort:** High

---

### Recommendation

**Go with Approach 1: ConsistencyGate as a Service.**

Reasoning:
- It aligns with the existing architecture (service layer + repository pattern)
- It enables explicit batch validation as specified in the roadmap (D2.1)
- It supports the `approve`/`reject`/`retry` tri-state required
- It allows structured feedback (`Suggestions` → retry prompts)
- It can be exposed as a single MCP tool that subagents call

The gate should manage:
- **BatchState** (pending → validating → approved/rejected)
- **Canon lock** (prevents parallel writes during validation)
- **Auto-save** (persist canon + state after approval)
- **Retry context** (return suggestions formatted as natural language prompts)

---

### Risks

1. **File locking race conditions** — Two parallel subagents could write canon simultaneously. Mitigation: add `flock` to `FilesystemCanonRepository.Save()` or use a `.canon.lock` file.

2. **False positives in validation** — The naive `lore_rule_compliance` (keyword matching) will flag legitimate content. Mitigation: improve rule parsing or add a `fast_mode` flag that skips semantic checks.

3. **LLM subagents ignore structured suggestions** — The roadmap explicitly warns about this (RT3). Mitigation: format `Suggestion` as both JSON and natural language; limit retries to 2.

4. **Gate becomes a bottleneck** — Sequential validation of each proposal in a batch could slow generation. Mitigation: validate proposals in parallel within the gate; cache entity lookups.

5. **Scope creep on "10 rules"** — The roadmap says 10 rules but doesn't exhaustively list them. We have 4 now, the user listed 6 new ones. Need to define exactly what the 10th rule is.

6. **No transaction boundary** — If gate approves but auto-save fails, the batch is approved but not persisted. Mitigation: save BEFORE returning approval; rollback on error.

---

### Ready for Proposal

**Yes.**

The orchestrator should tell the user:
- Phase 1 foundation is solid and well-tested
- Phase 2 requires creating a new `ConsistencyGateService` and expanding the `ValidationEngine` from 4 to 10 rules
- The main architectural decision is whether to add file locking (flock) now or defer to Phase 5
- We need clarification on what the 10th validation rule should be (user listed 6 new + 3 existing = 9)
