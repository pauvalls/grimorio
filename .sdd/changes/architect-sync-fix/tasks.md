# Implementation Tasks: architect-sync-fix

**Change ID:** architect-sync-fix  
**Target File:** `agents/grimorio-architect.md`  
**Goal:** Fix phase numbering inconsistencies and add verification checks

---

## Task Checklist

### 1. Backup original file
- **Action:** Copy `grimorio-architect.md` to `grimorio-architect.md.backup`
- **Path:** `/home/pau/Grimorio/agents/`
- **Command:** `cp agents/grimorio-architect.md agents/grimorio-architect.md.backup`
- **Expected:** Backup file created
- **Verification:** `ls -la agents/grimorio-architect.md.backup`

### 2. Fix Phase 4d heading
- **Action:** Add missing heading `### Phase 4d: Update Narrative State`
- **Location:** Line ~306 (after Phase 4c validation block)
- **Expected:** Phase 4d heading present before narrative state update delegation
- **Verification:** `grep -n "Phase 4d" agents/grimorio-architect.md`

### 3. Renumber Phase 3b-3e
- **Action:** Update phase labels at lines 172, 180, 209, 217
- **Changes:**
  - Line 172: `### Phase 3b: Monitor Batch 1` → keep as is
  - Line 180: `### Phase 3c: Validate Batch 1` → keep as is
  - Line 209: `### Phase 3d: Report Batch 1` → keep as is
  - Line 217: `### Phase 4: Batch 2` → keep as is
- **Expected:** Phase 3b, 3c, 3d, 4 labels correct
- **Verification:** `grep -n "^### Phase 3" agents/grimorio-architect.md`

### 4. Renumber Phase 4a-4e
- **Action:** Update phase labels at lines 217, 262, 270, 318
- **Changes:**
  - Line 217: `### Phase 4: Batch 2` → `### Phase 4a: Batch 2`
  - Line 262: `### Phase 4b: Monitor Batch 2` → keep as is
  - Line 270: `### Phase 4c: Validate Batch 2` → keep as is
  - Line 318: Add `### Phase 4d: Update Narrative State` (see task 2)
  - Line 326: `### Phase 4e: Report Batch 2` → keep as is
- **Expected:** Phase 4a, 4b, 4c, 4d, 4e sequence complete
- **Verification:** `grep -n "^### Phase 4" agents/grimorio-architect.md`

### 5. Renumber Phase 5a-5e
- **Action:** Update phase labels at lines 326, 355, 363, 398, 407
- **Changes:**
  - Line 326: `### Phase 5: Batch 3` → `### Phase 5a: Batch 3`
  - Line 355: `### Phase 5b: Monitor Batch 3` → keep as is
  - Line 363: `### Phase 5c: Validate Batch 3` → keep as is
  - Line 398: `### Phase 5d: Report Batch 3` → keep as is
  - Line 407: `### Phase 5e: Appendices` → keep as is
- **Expected:** Phase 5a, 5b, 5c, 5d, 5e sequence complete
- **Verification:** `grep -n "^### Phase 5" agents/grimorio-architect.md`

### 6. Add Batch 1 verification
- **Action:** Insert file existence checks after line 159 (end of Phase 3a)
- **Content to add:**
  ```markdown
  ### Phase 3a.1: Verify Batch 1 Files
  
  ```bash
  # Verify all Batch 1 files exist before proceeding
  test -f {campaign_path}/npcs/npcs_and_factions.md || echo "❌ NPCs file missing"
  test -f {campaign_path}/bestiary/bestiary.md || echo "❌ Bestiary file missing"
  test -f {campaign_path}/maps/maps.md || echo "❌ Maps file missing"
  ```
  ```
- **Expected:** File existence validation before Batch 1 delegation
- **Verification:** `grep -A 5 "Phase 3a.1" agents/grimorio-architect.md`

### 7. Add Batch 2 verification
- **Action:** Insert file existence checks after line 260 (end of Phase 4a setup)
- **Content to add:**
  ```markdown
  ### Phase 4a.1: Verify Batch 2 Files
  
  ```bash
  # Verify all Batch 2 files exist before proceeding
  test -f {campaign_path}/lore.md || echo "❌ Lore file missing"
  test -f {campaign_path}/setting-guide.md || echo "❌ Setting Guide file missing"
  test -f {campaign_path}/quests/*.md || echo "❌ Quests files missing"
  test -f {campaign_path}/encounters/encounters.md || echo "❌ Encounters file missing"
  test -f {campaign_path}/characters/*.md || echo "❌ Characters files missing"
  ```
  ```
- **Expected:** File existence validation before Batch 2 delegation
- **Verification:** `grep -A 5 "Phase 4a.1" agents/grimorio-architect.md`

### 8. Add Batch 3 verification
- **Action:** Insert file existence checks after line 353 (end of Phase 5a setup)
- **Content to add:**
  ```markdown
  ### Phase 5a.1: Verify Batch 3 Files
  
  ```bash
  # Verify all Batch 3 files exist before proceeding
  test -f {campaign_path}/maps/*.svg || echo "❌ SVG maps missing"
  test -f {campaign_path}/assets/divider-*.svg || echo "❌ Dividers missing"
  for i in $(seq 1 $act_count); do
    test -f {campaign_path}/acts/act-$i.md || echo "❌ Act $i missing"
  done
  ```
  ```
- **Expected:** File existence validation before Batch 3 delegation
- **Verification:** `grep -A 5 "Phase 5a.1" agents/grimorio-architect.md`

### 9. Replace Phase 9
- **Action:** Replace lines 509-531 (Phase 9: Final Consistency Check)
- **Changes:**
  - Add mention of Check 13A-E (from consistency gate)
  - Add retry logic for failed validations
  - Include explicit blocking gate language
- **Expected:** Phase 9 includes Check 13A-E reference and retry logic
- **Verification:** `grep -A 20 "Phase 9:" agents/grimorio-architect.md | grep -E "Check 13|retry"`

### 10. Delete Phase X
- **Action:** Remove lines 533-754 completely (Phase X: WotC Validation)
- **Reason:** Validation logic moved to narrative-custodian agent
- **Expected:** Phase X section removed entirely
- **Verification:** `grep -c "Phase X" agents/grimorio-architect.md` → should return 0

### 11. Renumber Phase 10→11 (Compile PDF)
- **Action:** Update phase label at line ~769
- **Change:** `### Phase 10: Compile PDF` → `### Phase 11: Compile PDF`
- **Expected:** PDF compilation is Phase 11
- **Verification:** `grep -n "Phase 11: Compile PDF" agents/grimorio-architect.md`

### 12. Renumber Phase 10→12 (Final Report)
- **Action:** Update phase label at line ~811
- **Change:** `### Phase 10: Final Report to User` → `### Phase 12: Final Report to User`
- **Expected:** Final report is Phase 12
- **Verification:** `grep -n "Phase 12: Final Report" agents/grimorio-architect.md`

### 13. Update phase diagram
- **Action:** Update header comment showing phase structure (lines 1-45)
- **Change:** Update workflow comment to show Phases 1-12 structure
- **Content:**
  ```markdown
  ## Workflow (STRICT ORDER — sequential phases, each waits for previous)
  
  **Phase Structure:**
  1. Gather Requirements
  2. Create Campaign Structure + Adventure Bible
  3. Batch 1 (3a-3e): NPCs, Bestiary, Maps → Validate → Report
  4. Batch 2 (4a-4e): Lore, Quests, Encounters, Characters → Validate → Update State → Report
  5. Batch 3 (5a-5e): SVG Maps, Acts → Validate → Report → Appendices
  6. Artist Batch Spec
  7. Generate AI Images
  8. Update Markdown References
  9. Final Consistency Check (Check 13A-E)
  10. WotC Validation (removed - moved to custodian)
  11. Compile PDF
  12. Final Report
  ```
- **Expected:** Header shows complete 1-12 phase structure
- **Verification:** `grep -A 15 "Phase Structure:" agents/grimorio-architect.md`

### 14. Verify implementation
- **Action:** Run grep checks for phase numbers
- **Commands:**
  ```bash
  # Check all phase labels are sequential
  grep "^### Phase" agents/grimorio-architect.md
  
  # Verify no duplicate phase numbers
  grep "^### Phase" agents/grimorio-architect.md | sort | uniq -d
  
  # Verify Phase X is removed
  grep "Phase X" agents/grimorio-architect.md
  
  # Verify Phase 11 and 12 exist
  grep -E "Phase 1[12]:" agents/grimorio-architect.md
  ```
- **Expected:** All phases numbered 1-12, no duplicates, no Phase X
- **Verification:** Manual review of grep output

### 15. Test with new campaign
- **Action:** Generate a test campaign to validate flow
- **Command:** Create a minimal test campaign using the updated architect
- **Expected:** Campaign generation completes through all 12 phases
- **Verification:** 
  ```bash
  # Check campaign structure created
  ls -la campaigns/test-campaign/
  
  # Check PDF generated
  ls -la campaigns/test-campaign/campaign.pdf
  
  # Check all expected files exist
  test -f campaigns/test-campaign/lore.md && echo "✅ Lore"
  test -f campaigns/test-campaign/npcs/npcs_and_factions.md && echo "✅ NPCs"
  test -f campaigns/test-campaign/bestiary/bestiary.md && echo "✅ Bestiary"
  test -f campaigns/test-campaign/acts/act-1.md && echo "✅ Act 1"
  ```

---

## Verification Summary

After completing all tasks, run:

```bash
# Full phase structure check
echo "=== Phase Structure ==="
grep "^### Phase" agents/grimorio-architect.md | nl

# Verify no Phase X remains
echo "=== Phase X Check (should be empty) ==="
grep "Phase X" agents/grimorio-architect.md

# Verify Phase 11 and 12 exist
echo "=== Final Phases ==="
grep -E "Phase 1[12]:" agents/grimorio-architect.md

# Verify backup exists
echo "=== Backup File ==="
ls -la agents/grimorio-architect.md.backup
```

**Expected Output:**
- 12 phases numbered sequentially (1-12, with sub-phases a-e)
- No Phase X references
- Phase 11: Compile PDF
- Phase 12: Final Report to User
- Backup file exists

---

## Rollback Plan

If issues are found:

```bash
# Restore from backup
cp agents/grimorio-architect.md.backup agents/grimorio-architect.md

# Verify restoration
git diff agents/grimorio-architect.md  # Should show no changes
```
