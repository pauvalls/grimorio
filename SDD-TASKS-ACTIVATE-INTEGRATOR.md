# SDD Tasks: activate-integrator

## Implementation Checklist

### Task 1: Read Target File
- [ ] Read `agents/grimorio-architect.md` to confirm current structure
- [ ] Identify exact insertion point (after Phase 5e, before Phase 6)
- [ ] Note line numbers for Phase 5e end and Phase 6 start

### Task 2: Draft Phase 5f Content
- [ ] Write Phase 5f header: `### Phase 5f: Integration`
- [ ] Write delegation block with full prompt including:
  - [ ] Cross-Reference Audit responsibility
  - [ ] Technical Standardization responsibility
  - [ ] Balance Audit responsibility
  - [ ] Integration responsibility (stat blocks, quick refs, treasure summaries)
  - [ ] Handouts generation responsibility
  - [ ] Auto-Fix responsibility
  - [ ] Final Validation responsibility (validate_canon + check_consistency)
  - [ ] File read list (canon.json, lore.md, acts/*.md, npcs/, bestiary/, etc.)
- [ ] Write monitoring loop section
- [ ] Write validation gate section with:
  - [ ] max_retries = 2
  - [ ] retry_count initialization
  - [ ] validation_passed flag
  - [ ] WHILE loop structure
  - [ ] Delegation to narrative-custodian
  - [ ] IF/ELSE logic for approved/rejected
  - [ ] Fix issues comment
  - [ ] Blocking check comment
- [ ] Write progress report template with all 7 checkmarks

### Task 3: Insert Phase 5f
- [ ] Use Edit tool to insert Phase 5f after Phase 5e
- [ ] Verify insertion point is correct (after line 451, before Phase 6)
- [ ] Ensure proper spacing between phases
- [ ] Verify markdown formatting is correct

### Task 4: Verify Structural Integrity
- [ ] Run grep to confirm Phase 5f exists: `grep -c "Phase 5f: Integration" agents/grimorio-architect.md`
- [ ] Verify phase order: 5e < 5f < 6
- [ ] Check for any duplicate headers or content
- [ ] Verify no content was accidentally removed from Phase 5e or Phase 6

### Task 5: Verify Pattern Consistency
- [ ] Compare Phase 5f structure to Phase 3c (Batch 1 validation)
- [ ] Compare Phase 5f structure to Phase 4c (Batch 2 validation)
- [ ] Compare Phase 5f structure to Phase 5c (Batch 3 validation)
- [ ] Ensure same variable names (max_retries, retry_count, validation_passed)
- [ ] Ensure same delegation pattern
- [ ] Ensure same monitoring pattern
- [ ] Ensure same report format

### Task 6: Save SDD Artifacts to Engram
- [ ] Save proposal artifact: `sdd/activate-integrator/proposal`
- [ ] Save spec artifact: `sdd/activate-integrator/spec`
- [ ] Save design artifact: `sdd/activate-integrator/design`
- [ ] Save tasks artifact: `sdd/activate-integrator/tasks`
- [ ] Save apply artifact: `sdd/activate-integrator/apply` (after implementation)
- [ ] Save verify artifact: `sdd/activate-integrator/verify` (after verification)
- [ ] Save archive artifact: `sdd/activate-integrator/archive` (after completion)

### Task 7: Run Acceptance Tests
- [ ] Test 1: Structural Verification
  ```bash
  grep -c "### Phase 5f: Integration" agents/grimorio-architect.md
  # Expected: 1
  ```
- [ ] Test 2: Delegation Pattern Verification
  ```bash
  grep -c 'delegate(agent="grimorio-integrator"' agents/grimorio-architect.md
  # Expected: 1
  ```
- [ ] Test 3: Validation Gate Verification
  ```bash
  grep -c "max_retries = 2" agents/grimorio-architect.md
  # Expected: 1 (in Phase 5f context)
  ```
- [ ] Test 4: Verify Phase 6 is conditional
  ```bash
  grep -A 50 "Phase 5f" agents/grimorio-architect.md | grep -B 5 "Phase 6"
  # Expected: Contains conditional logic
  ```

### Task 8: Final Review
- [ ] Review entire Phase 5f section for clarity
- [ ] Verify all acceptance criteria from spec are met
- [ ] Check for typos or formatting issues
- [ ] Confirm blocking behavior is clearly documented
- [ ] Verify progress report template is complete

---

## Task Dependencies

```
Task 1 (Read) → Task 2 (Draft) → Task 3 (Insert) → Task 4 (Verify Structure)
                                           ↓
                                    Task 5 (Verify Pattern)
                                           ↓
                                    Task 6 (Save Artifacts)
                                           ↓
                                    Task 7 (Acceptance Tests)
                                           ↓
                                    Task 8 (Final Review)
```

## Estimated Effort

- **Task 1-2**: 10 minutes (reading and drafting)
- **Task 3**: 5 minutes (insertion)
- **Task 4-5**: 10 minutes (verification)
- **Task 6**: 10 minutes (Engram saves)
- **Task 7**: 5 minutes (acceptance tests)
- **Task 8**: 5 minutes (final review)

**Total**: ~45 minutes

## Definition of Done

- [ ] Phase 5f inserted in correct location
- [ ] All 7 integrator responsibilities included in prompt
- [ ] Validation gate with 2 retries implemented
- [ ] Blocking behavior on failure implemented
- [ ] Progress report template complete
- [ ] Pattern consistent with existing phases
- [ ] All acceptance tests pass
- [ ] All SDD artifacts saved to Engram

---

**Status**: Ready for Apply phase  
**Next**: Execute tasks and implement changes
