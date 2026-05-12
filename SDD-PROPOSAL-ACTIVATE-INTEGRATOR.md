# SDD Proposal: activate-integrator

## Change ID
`activate-integrator`

## Intent
Activate the existing `grimorio-integrator` agent within the grimorio-architect workflow by adding Phase 5f (Integration) between Appendices (5e) and Artist (6).

## Problem Statement
The `grimorio-integrator` agent exists (334 lines, fully implemented with 7 phases) but is **NEVER invoked** in the current grimorio-architect workflow. This means campaigns are generated without:
- Cross-reference audits (broken creature/NPC references)
- Technical standardization (relative DCs, non-standard treasure format)
- Balance audits (XP budget verification, difficulty curve)
- Integration artifacts (inline stat blocks, quick reference tables)
- Player handouts (maps, clues, NPC lists)
- Auto-fixes for common issues
- Final validation before image generation

## Scope

### In Scope
- Add Phase 5f (Integration) to `agents/grimorio-architect.md`
- Insert after Phase 5e (Appendices), before Phase 6 (Artist)
- Include delegation prompt, monitoring loop, validation gate with 2 retries, and progress report
- Save artifacts to Engram topic: `sdd/activate-integrator/*`

### Out of Scope
- Modifying `grimorio-integrator.md` (already complete)
- Changing other phases in the workflow
- Modifying MCP tools
- Testing with full campaign generation (covered by separate test plan)

## Approach

1. **Insert Phase 5f** at line ~451 in `agents/grimorio-architect.md` (after Phase 5e)
2. **Delegation pattern**: Follow existing pattern from other phases (delegate → monitor → validate → report)
3. **Validation gate**: Use `grimorio-narrative-custodian` with max 2 retries (consistent with Batches 1-3)
4. **Blocking behavior**: If validation fails after retries, do NOT proceed to Phase 6 (Artist)

## Why This Order Matters

The integrator MUST run:
1. **AFTER all content is generated** (acts, NPCs, bestiary, encounters, appendices) — because it cross-references everything
2. **BEFORE images are generated** — because the artist needs accurate references to generate correct NPC portraits and monster illustrations

**If integrator runs BEFORE appendices**: Missing reference material  
**If integrator runs AFTER artist**: Images may reference entities that get renamed/removed during integration

## Success Criteria

1. ✅ Phase 5f appears in grimorio-architect.md workflow between 5e and 6
2. ✅ Delegation prompt includes all 7 integrator responsibilities
3. ✅ Monitoring loop waits for completion before proceeding
4. ✅ Validation gate with 2 retries implemented
5. ✅ Progress report format matches existing phase reports
6. ✅ Blocking behavior on validation failure (does not proceed to Phase 6)
7. ✅ All SDD artifacts saved to Engram

## Risks

| Risk | Probability | Mitigation |
|------|-------------|------------|
| Integrator modifies files and breaks references | Medium | Only auto-fix OBVIOUS issues; detailed report of every change |
| Integration takes too long | Medium | Accept as necessary step; runs after all content exists |
| Validation gate too strict | Low | 2 retries allowed; manual intervention if still failing |
| Conflicts with Artist phase | Low | Integrator runs FIRST, stabilizes references; Artist reads stable state |

## Timeline

- **Propose**: Complete
- **Spec**: Next phase
- **Design**: Next phase
- **Tasks**: Next phase
- **Apply**: Next phase
- **Verify**: Next phase
- **Archive**: Final phase

## Stakeholders

- **Primary**: grimorio-architect agent (workflow owner)
- **Consumer**: grimorio-integrator agent (will be invoked)
- **Validator**: grimorio-narrative-custodian (validation gate)
- **Downstream**: grimorio-artist (reads stabilized references)

---

**Status**: Ready for Spec phase  
**Next**: Create delta spec with requirements and scenarios
