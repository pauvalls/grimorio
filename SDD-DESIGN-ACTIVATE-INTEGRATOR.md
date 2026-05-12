# SDD Design: activate-integrator

## Technical Design

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                   grimorio-architect                         │
│                     (Orchestrator)                           │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  Phase 5e: Appendices → grimorio-appendices                 │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  Phase 5f: Integration → grimorio-integrator    ← NEW       │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ 1. Cross-Reference Audit                            │   │
│  │ 2. Technical Standardization                        │   │
│  │ 3. Balance Audit                                    │   │
│  │ 4. Integration (stat blocks, quick refs)            │   │
│  │ 5. Handouts                                         │   │
│  │ 6. Auto-Fix                                         │   │
│  │ 7. Final Validation                                 │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  Validation Gate → grimorio-narrative-custodian             │
│  (max_retries = 2, blocking on failure)                     │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  Phase 6: Artist → grimorio-artist                          │
└─────────────────────────────────────────────────────────────┘
```

### Component Interactions

#### 1. grimorio-architect (Orchestrator)
- **Role**: Workflow coordinator
- **Responsibility**: Delegate Phase 5f, monitor completion, validate results, report progress
- **State**: Maintains retry_count, validation_passed flags

#### 2. grimorio-integrator (Worker)
- **Role**: Integration specialist
- **Responsibility**: Execute 7-phase integration process
- **Input**: Full prompt with campaign context
- **Output**: Integration report, modified markdown files

#### 3. grimorio-narrative-custodian (Validator)
- **Role**: Quality gate
- **Responsibility**: Validate integration results against canon
- **Input**: Integration report, campaign files
- **Output**: Validation status (approved/rejected) with feedback

### Data Flow

```
Phase 5e Complete
       │
       ▼
┌──────────────────────────────────────┐
│ Delegate to grimorio-integrator      │
│ Prompt includes:                     │
│ - campaign_name                      │
│ - campaign_path                      │
│ - All 7 responsibilities             │
│ - File read list                     │
└──────────────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────┐
│ Monitoring Loop                      │
│ WHILE delegation running:            │
│   delegation_list                    │
└──────────────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────┐
│ Validation Gate                      │
│ WHILE retry_count <= 2:              │
│   delegate to narrative-custodian    │
│   IF approved: break                 │
│   ELSE: fix issues, retry_count++    │
└──────────────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────┐
│ IF validation_passed:                │
│   Report progress                    │
│   Proceed to Phase 6                 │
│ ELSE:                                │
│   Report failure                     │
│   Block Phase 6                      │
└──────────────────────────────────────┘
```

## Implementation Strategy

### Insertion Point Analysis

**File**: `agents/grimorio-architect.md`  
**Current Structure**:
- Line 446-451: Phase 5e: Appendices
- Line 453-469: Phase 6: Artist

**Insertion**: Between lines 451 and 453  
**Lines to Add**: ~80-100 lines

### Code Structure

```markdown
### Phase 5f: Integration

```
delegate(agent="grimorio-integrator", prompt="...")
```

### Phase 5f: Monitor Integration

```
WHILE integrator is running:
  delegation_list
```

**Do NOT proceed until integration completes.**

### Phase 5f: Validation Gate (Auto-Retry)

```
max_retries = 2
retry_count = 0
validation_passed = false

WHILE retry_count <= max_retries AND NOT validation_passed:
    delegate(agent="grimorio-narrative-custodian", prompt="...")
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
            Report failure
            DO NOT proceed to Phase 6

IF validation_passed:
    Proceed to Phase 6
```

**BLOCKING CHECK:** If validation returns **rejected** after max retries, DO NOT proceed.

### Phase 5f: Progress Report

```
## Phase 5f Completada — Integración

✅ Cross-Reference Audit: {X/Y} referencias verificadas
...
```

### Design Decisions

#### Decision 1: Placement After Appendices

**Rationale**: 
- Appendices consolidate all reference material (monsters, NPCs, items, handouts)
- Integrator needs complete reference set for cross-reference audit
- Running before appendices would miss appendix content in audit

**Tradeoff**: 
- Adds one more sequential step (increases total generation time)
- Mitigation: Integrator runs comprehensive validation that prevents downstream errors

#### Decision 2: Validation Gate with Retries

**Rationale**:
- Consistent with Batches 1-3 validation pattern
- Allows auto-fix of common issues without manual intervention
- 2 retries balances thoroughness with time constraints

**Tradeoff**:
- May delay campaign generation if issues found
- Mitigation: Most issues are auto-fixable (DCs, connections, references)

#### Decision 3: Blocking on Validation Failure

**Rationale**:
- Prevents Artist from generating images based on unstable references
- Forces manual review of critical issues
- Maintains campaign quality standards

**Tradeoff**:
- Campaign generation halts on failure
- Mitigation: Clear error messages guide manual fix

#### Decision 4: Delegation vs Direct MCP Calls

**Rationale**:
- Consistent with architect's orchestrator role
- Integrator agent has specialized knowledge (334 lines of logic)
- Separation of concerns: architect orchestrates, integrator integrates

**Tradeoff**:
- Adds delegation overhead
- Mitigation: Pattern already established in workflow

## Error Handling

### Error Categories

1. **Delegation Failure**: Integrator agent unavailable
   - **Detection**: delegation_list shows error status
   - **Recovery**: Report error, halt workflow

2. **Validation Failure (Retryable)**: Fixable issues found
   - **Detection**: validation status = "rejected" with fix suggestions
   - **Recovery**: Auto-fix, retry up to 2 times

3. **Validation Failure (Critical)**: Unfixable issues
   - **Detection**: validation status = "rejected" after max retries
   - **Recovery**: Report failure, require manual intervention

4. **Timeout**: Integrator takes too long
   - **Detection**: delegation_list shows running beyond expected time
   - **Recovery**: Continue waiting (no timeout in current pattern)

## Testing Approach

### Unit Tests (Static Analysis)

1. **Structural Tests**:
   - Verify Phase 5f exists in correct location
   - Verify all 7 responsibilities mentioned in prompt
   - Verify validation gate pattern matches Batches 1-3

2. **Pattern Tests**:
   - Verify delegation syntax matches existing phases
   - Verify monitoring loop syntax matches existing phases
   - Verify report format matches existing phases

### Integration Tests (Runtime)

1. **Happy Path**:
   - Run campaign generation with well-formed content
   - Verify Phase 5f executes and completes
   - Verify Phase 6 begins after Phase 5f

2. **Retry Path**:
   - Inject validation failures (e.g., broken references)
   - Verify retry logic executes
   - Verify success on retry

3. **Failure Path**:
   - Inject unfixable validation failures
   - Verify Phase 6 is blocked
   - Verify error message is clear

## Rollback Plan

If Phase 5f causes issues in production:

1. **Immediate Rollback**:
   ```bash
   git revert <commit-hash>
   ```

2. **Partial Disable**:
   - Comment out Phase 5f delegation in architect.md
   - Keep validation gate structure for future re-enablement

3. **Fallback**:
   - Campaigns generate without integration (previous behavior)
   - Manual integration can be run post-generation

## Success Metrics

1. **Adoption**: Phase 5f executes in 100% of campaign generations
2. **Quality**: Cross-reference errors reduced by >90%
3. **Efficiency**: Auto-fix resolves >80% of validation issues on first retry
4. **User Satisfaction**: Clear progress reports, no silent failures

---

**Status**: Design Complete  
**Next**: Tasks phase - implementation checklist
