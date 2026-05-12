# SDD Exploration: activate-integrator

## Summary

**Change ID**: activate-integrator  
**Goal**: Add Phase 5f (Integration) to invoke grimorio-integrator after Appendices (5e) and before Artist (6)  
**Problem**: grimorio-integrator exists (334 lines, 7 phases) but is NEVER invoked in workflow  
**Files affected**: `agents/grimorio-architect.md`

---

## Current Flow

```
Phase 5e: Appendices → grimorio-appendices
Phase 6:  Artist → grimorio-artist (batch-spec.json)
```

## New Flow

```
Phase 5e: Appendices → grimorio-appendices
Phase 5f: Integration → grimorio-integrator  ← NEW
Phase 6:  Artist → grimorio-artist (batch-spec.json)
```

---

## Why This Order Matters

The integrator MUST run:
1. **AFTER all content is generated** (acts, NPCs, bestiary, encounters, appendices) — because it cross-references everything
2. **BEFORE images are generated** — because the artist needs accurate references to generate correct NPC portraits and monster illustrations

If integrator runs BEFORE appendices: Missing reference material  
If integrator runs AFTER artist: Images may reference entities that get renamed/removed during integration

---

## Implementation Details

### Insertion Point

**File**: `agents/grimorio-architect.md`  
**Location**: After Phase 5e (line ~451), before Phase 6 (line ~453)

### Phase 5f Code to Insert

```markdown
### Phase 5f: Integration

```
delegate(agent="grimorio-integrator", prompt="Integrate and polish campaign '{campaign_name}' at {campaign_path}.

ALL content has been generated (acts, NPCs, bestiary, encounters, appendices). Your job is to:

1. **Cross-Reference Audit**: Verify ALL creature/NPC/encounter references in acts exist in source files
2. **Technical Standardization**: Convert relative DCs to numbers, ensure bidirectional connections, standardize treasure format
3. **Balance Audit**: Calculate XP budget per act, verify difficulty curve, adjust encounters if needed
4. **Integration**: Add inline stat blocks for unique creatures, NPC quick reference tables, treasure summaries
5. **Handouts**: Generate player-facing materials (maps, clues, known NPCs)
6. **Auto-Fix**: Fix common issues (DCs, connections, missing references) — only obvious fixes
7. **Final Validation**: Run validate_canon + check_consistency + process_consistency_gate

CRITICAL: Read ALL files first:
- canon.json, lore.md, introduction.md
- acts/*.md (ALL chapters)
- npcs/npcs_and_factions.md
- bestiary/bestiary.md
- encounters/encounters.md
- maps/maps.md
- quests/*.md
- appendices.md

Report detailed integration report with all fixes made.
")
```

### Phase 5f Monitoring

```
WHILE integrator is running:
  delegation_list
```

**Do NOT proceed until integration completes.**

### Phase 5f Validation Gate

```
max_retries = 2
retry_count = 0
validation_passed = false

WHILE retry_count <= max_retries AND NOT validation_passed:
    delegate(agent="grimorio-narrative-custodian", prompt="Validate integration results for campaign '{campaign_name}' at {campaign_path}.

    Read integration report and verify:
    - All cross-references resolved
    - All auto-fixes applied correctly
    - Balance audit passed
    - Final validation passed

    Return validation report with status.")
    
    Wait for delegation to complete
    result = delegation_read(id)
    
    IF result.status == "approved":
        validation_passed = true
    ELSE:
        retry_count += 1
        IF retry_count <= max_retries:
            Fix issues based on result.feedback
            Continue loop
        ELSE:
            Report failure: "Integration validation failed after {max_retries} retries. Issues: {result.feedback}"
            DO NOT proceed to Phase 6

IF validation_passed:
    Proceed to Phase 6
```

**BLOCKING CHECK:** If validation returns **rejected** after max retries, DO NOT proceed. Report failure and wait for manual intervention.

### Phase 5f Progress Report

```
## Phase 5f Completada — Integración

✅ Cross-Reference Audit: {X/Y} referencias verificadas
✅ Technical Standardization: {X} correcciones aplicadas
✅ Balance Audit: XP por acto dentro de rango esperado
✅ Integration: {X} tablas de referencia agregadas
✅ Handouts: {X} handouts generados
✅ Auto-Fix: {X} issues auto-corregidos
✅ Final Validation: PASSED

Issues encontrados: {lista}
Archivos modificados: {lista}

Iniciando Phase 6: Artist batch-spec...
```

---

## Integrator Responsibilities (from grimorio-integrator.md)

1. **Cross-Reference Audit**: Verify creatures, NPCs, encounters, objects referenced in acts exist in source files
2. **Technical Standardization**: 
   - Treasure format (XP + currency + items)
   - Creature format (exact names with bestiary references)
   - Bidirectional connections
   - Numeric DCs (no "high/low")
3. **Balance Audit**:
   - XP budget per act (Act 1: 300-400 XP/PJ, Act 2: 600-900, Act 3: 1200-1800)
   - Encounter difficulty labeling
   - Difficulty curve (easy → medium → hard → boss)
4. **Integration**:
   - Inline stat blocks for unique creatures
   - NPC quick reference tables
   - Treasure summaries
   - Connection maps
5. **Handouts**: Player maps, clue lists, known NPCs
6. **Auto-Fix**: Fix obvious issues (DCs, connections, missing references)
7. **Final Validation**: validate_canon + check_consistency + process_consistency_gate

---

## Risks and Mitigations

| Risk | Probability | Mitigation |
|------|-------------|------------|
| Integrator modifies files and breaks references | Medium | Only auto-fix OBVIOUS issues; if doubt → warning. Detailed report of every change |
| Integration takes too long | Medium | Run in parallel with Appendices validation (not possible — needs Appendices output). Accept as necessary step |
| Integrator conflicts with Artist | Low | Integrator runs FIRST, stabilizes all references. Artist reads stable references |
| Validation gate too strict | Low | 2 retries allowed. Manual intervention if still failing |

---

## Dependencies

- ✅ grimorio-integrator agent exists (334 lines, fully implemented)
- ✅ grimorio-appendices runs before (Phase 5e)
- ✅ grimorio-artist runs after (Phase 6)
- ✅ MCP tools available: validate_canon, check_consistency, process_consistency_gate, save_areas, save_npcs, save_encounters

---

## Testing Strategy

After implementation:
1. Run a test campaign generation
2. Verify Phase 5f executes after Phase 5e
3. Check integration report shows cross-reference results
4. Verify Phase 6 (Artist) reads corrected references
5. Confirm no broken references in final PDF

---

## Files to Modify

- `agents/grimorio-architect.md` — Insert Phase 5f after Phase 5e (~line 451)

Total lines to add: ~80-100 lines (delegation + monitoring + validation + report)
