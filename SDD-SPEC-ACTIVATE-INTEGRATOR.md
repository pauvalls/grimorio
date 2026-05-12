# SDD Delta Spec: activate-integrator

## Overview

**Change ID**: `activate-integrator`  
**Version**: 1.0  
**Date**: 2026-05-11  
**Author**: SDD Cycle  
**Status**: Spec Complete

## Requirements

### Functional Requirements

#### FR-1: Phase 5f Insertion
- **ID**: FR-1
- **Title**: Phase 5f must be inserted after Phase 5e
- **Description**: Add new Phase 5f (Integration) section to `agents/grimorio-architect.md`
- **Location**: After line 451 (Phase 5e: Appendices), before Phase 6 (Artist)
- **Acceptance Criteria**:
  - Phase 5f header appears after Phase 5e content
  - Phase 6 header remains unchanged but follows Phase 5f
  - No content from Phase 5e or Phase 6 is removed or modified

#### FR-2: Delegation Prompt
- **ID**: FR-2
- **Title**: Integrator delegation prompt must include all 7 responsibilities
- **Description**: The delegation prompt passed to `grimorio-integrator` must explicitly list all integration tasks
- **Acceptance Criteria**:
  - Prompt includes: Cross-Reference Audit
  - Prompt includes: Technical Standardization
  - Prompt includes: Balance Audit
  - Prompt includes: Integration (inline stat blocks, quick refs, treasure summaries)
  - Prompt includes: Handouts generation
  - Prompt includes: Auto-Fix for obvious issues
  - Prompt includes: Final Validation (validate_canon + check_consistency)
  - Prompt specifies reading ALL source files before execution

#### FR-3: Monitoring Loop
- **ID**: FR-3
- **Title**: Workflow must wait for integrator completion
- **Description**: Implement monitoring loop that blocks until integrator delegation completes
- **Acceptance Criteria**:
  - WHILE loop checks `delegation_list` for integrator status
  - Loop exits only when integrator delegation is complete
  - No Phase 6 activities begin until loop exits

#### FR-4: Validation Gate
- **ID**: FR-4
- **Title**: Validation gate with auto-retry logic
- **Description**: Implement validation gate using `grimorio-narrative-custodian` with max 2 retries
- **Acceptance Criteria**:
  - `max_retries = 2`
  - `retry_count` starts at 0
  - Validation delegates to `grimorio-narrative-custodian`
  - If status == "approved": validation_passed = true, proceed to Phase 6
  - If status == "rejected" and retry_count < max_retries: increment retry_count, fix issues, retry
  - If status == "rejected" and retry_count >= max_retries: report failure, DO NOT proceed to Phase 6

#### FR-5: Progress Report
- **ID**: FR-5
- **Title**: Progress report after Phase 5f completion
- **Description**: Output a structured progress report to the user after Phase 5f completes
- **Acceptance Criteria**:
  - Report includes status for each of 7 integrator responsibilities
  - Report lists issues found (if any)
  - Report lists files modified (if any)
  - Report indicates next phase (Phase 6: Artist)
  - Format matches existing phase reports (Phase 3d, 4e, 5d)

#### FR-6: Blocking Behavior
- **ID**: FR-6
- **Title**: Validation failure must block Phase 6
- **Description**: If validation fails after max retries, Phase 6 must NOT execute
- **Acceptance Criteria**:
  - Code path exists that prevents Phase 6 delegation when validation_passed == false
  - User is informed of failure with specific issues
  - Manual intervention is required to proceed

### Non-Functional Requirements

#### NFR-1: Consistency with Existing Patterns
- **ID**: NFR-1
- **Title**: Follow existing phase patterns
- **Description**: Phase 5f implementation must match the structure and style of Phases 3c, 4c, and 5c
- **Acceptance Criteria**:
  - Uses same delegation pattern (`delegate(agent="...", prompt="...")`)
  - Uses same monitoring pattern (`WHILE ... delegation_list`)
  - Uses same validation pattern (max_retries, retry_count, validation_passed)
  - Uses same report format (checkmarks, counts, "Iniciando Phase X")

#### NFR-2: Code Quality
- **ID**: NFR-2
- **Title**: Maintain markdown code quality
- **Description**: Added markdown must be properly formatted and indented
- **Acceptance Criteria**:
  - No trailing whitespace
  - Consistent indentation (4 spaces for code blocks)
  - Proper markdown header hierarchy (### for phase, #### for sub-sections)
  - Code blocks use triple backticks with language hint where applicable

#### NFR-3: Documentation
- **ID**: NFR-3
- **Title**: Self-documenting code
- **Description**: Comments and structure should make the phase self-explanatory
- **Acceptance Criteria**:
  - Phase header clearly states purpose
  - Validation gate includes inline comments explaining retry logic
  - Blocking check includes comment explaining why it's critical

## Scenarios

### Scenario 1: Happy Path - Integration Succeeds on First Try

**Given**: All campaign content has been generated (Phases 1-5e complete)  
**When**: Phase 5f begins execution  
**Then**:
1. Integrator delegation is created with full prompt
2. Monitoring loop waits for completion
3. Validation gate delegates to narrative-custodian
4. Validation returns status "approved"
5. Progress report is output with all checkmarks
6. Phase 6 begins execution

**Example**:
```
## Phase 5f Completada — Integración

✅ Cross-Reference Audit: 47/47 referencias verificadas
✅ Technical Standardization: 12 correcciones aplicadas
✅ Balance Audit: XP por acto dentro de rango esperado
✅ Integration: 8 tablas de referencia agregadas
✅ Handouts: 3 handouts generados
✅ Auto-Fix: 5 issues auto-corregidos
✅ Final Validation: PASSED

Issues encontrados: Ninguno
Archivos modificados: acts/act-1.md, acts/act-2.md, encounters/encounters.md

Iniciando Phase 6: Artist batch-spec...
```

### Scenario 2: Validation Fails, Retry Succeeds

**Given**: Integrator completed but validation finds issues  
**When**: Validation gate executes first attempt  
**Then**:
1. Validation returns status "rejected" with feedback
2. retry_count increments to 1
3. Issues are fixed based on feedback
4. Validation retries (second attempt)
5. Validation returns status "approved"
6. Phase 5f completes successfully

**Example**:
```
Validation attempt 1: REJECTED
Issues: DCs not standardized in Area 3, missing connection in Area 7

Fixing issues...
Re-delegating integrator for auto-fix...

Validation attempt 2: APPROVED
Proceeding to progress report.
```

### Scenario 3: Validation Fails After Max Retries

**Given**: Integrator completed but validation finds critical issues  
**When**: Validation gate exhausts all retries  
**Then**:
1. Validation returns status "rejected" (attempt 1)
2. Issues fixed, retry_count = 1
3. Validation returns status "rejected" (attempt 2)
4. Issues fixed, retry_count = 2
5. Validation returns status "rejected" (attempt 3)
6. retry_count >= max_retries, loop exits
7. Failure is reported to user
8. Phase 6 does NOT begin
9. Manual intervention required

**Example**:
```
## Phase 5f: Integration FAILED

❌ Validation failed after 2 retries.

Issues:
- Area 12 references creature "Goblin Boss" not in bestiary
- Treasure format inconsistent in Area 5
- Missing bidirectional connection between Area 3 and Area 8

**Next Action:**
- Manual review required
- Fix issues in source files
- Re-run Phase 5f manually or restart campaign generation
```

### Scenario 4: Integrator Reports No Changes Needed

**Given**: Campaign content is already well-formed  
**When**: Integrator executes  
**Then**:
1. Integrator reads all source files
2. Cross-reference audit finds all references valid
3. Technical standardization finds all DCs numeric
4. Balance audit finds XP budgets within range
5. No auto-fixes are applied
6. Validation passes
7. Progress report shows "0 correcciones" but all checkmarks

**Example**:
```
## Phase 5f Completada — Integración

✅ Cross-Reference Audit: 52/52 referencias verificadas
✅ Technical Standardization: 0 correcciones aplicadas
✅ Balance Audit: XP por acto dentro de rango esperado
✅ Integration: 0 tablas de referencia agregadas (ya completas)
✅ Handouts: 3 handouts generados
✅ Auto-Fix: 0 issues auto-corregidos
✅ Final Validation: PASSED

Issues encontrados: Ninguno
Archivos modificados: Ninguno (contenido ya cumple estándares)

Iniciando Phase 6: Artist batch-spec...
```

## Edge Cases

### Edge Case 1: Empty Campaign

**Condition**: Campaign has minimal content (one-shot, 1 act, 2 NPCs, 3 encounters)  
**Behavior**: Integrator still runs all 7 phases, but counts are lower  
**Expected**: No errors, validation passes with lower numbers

### Edge Case 2: Massive Campaign

**Condition**: Campaign has 3 acts, 20 NPCs, 30 encounters, 50+ references  
**Behavior**: Integrator may take longer, but completes all phases  
**Expected**: No timeout, validation handles large content sets

### Edge Case 3: Integrator Agent Unavailable

**Condition**: `grimorio-integrator` agent definition is missing or malformed  
**Behavior**: Delegation fails immediately  
**Expected**: Error reported, Phase 5f marked as failed, Phase 6 blocked

### Edge Case 4: Partial File Corruption

**Condition**: One source file (e.g., acts/act-2.md) is corrupted or unreadable  
**Behavior**: Integrator reports read error for that file  
**Expected**: Validation fails, retry may succeed if integrator handles gracefully

## Acceptance Tests

### Test 1: Structural Verification
```bash
# Verify Phase 5f exists in architect.md
grep -c "### Phase 5f: Integration" agents/grimorio-architect.md
# Expected: 1

# Verify Phase 5f is between 5e and 6
grep -n "Phase 5e\|Phase 5f\|Phase 6" agents/grimorio-architect.md
# Expected: 5e line < 5f line < 6 line
```

### Test 2: Delegation Pattern Verification
```bash
# Verify delegation to grimorio-integrator
grep -c 'delegate(agent="grimorio-integrator"' agents/grimorio-architect.md
# Expected: 1

# Verify prompt includes all 7 responsibilities
grep -A 30 'grimorio-integrator' agents/grimorio-architect.md | grep -c "Cross-Reference\|Standardization\|Balance\|Integration\|Handouts\|Auto-Fix\|Validation"
# Expected: >= 7
```

### Test 3: Validation Gate Verification
```bash
# Verify max_retries = 2
grep -A 10 "Phase 5f" agents/grimorio-architect.md | grep -c "max_retries = 2"
# Expected: 1

# Verify retry logic exists
grep -A 20 "Phase 5f" agents/grimorio-architect.md | grep -c "retry_count\|validation_passed"
# Expected: >= 3
```

### Test 4: Blocking Behavior Verification
```bash
# Verify Phase 6 is conditional on validation_passed
grep -A 50 "Phase 5f" agents/grimorio-architect.md | grep -B 5 "Phase 6"
# Expected: Contains "IF validation_passed" or similar conditional
```

---

**Status**: Spec Complete  
**Next**: Design phase - technical approach and architecture
