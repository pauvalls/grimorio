# Proposal: Migrate Grimorio v1.x to v2.0 — Area-Based Campaign Generation

## Intent

Grimorio v1.x generates "narrative scenes" (3-5 per act, 350 words each) that read like novel summaries. Official WotC adventures (WDH, SKT) use "numbered playable areas" (10-36 per chapter, 150-200 words each) with specific DCs, treasure, and mechanics. DMs using Grimorio must improvise ~70% of content at the table. This migration replaces scene-based generation with area-based generation to match professional module quality.

## Scope

### In Scope
- Replace `grimorio-acts` agent with `grimorio-areas` (numbered locations)
- Create `grimorio-integrator` agent (cross-reference validation)
- Update all templates to WotC-style technical format
- Enhance compiler: hierarchical TOC, clickable cross-references, inline stat blocks
- Add player handouts (maps, clues, NPC tracker)
- Implement area numbering with bidirectional connections

### Out of Scope
- AI image generation improvements (deferred to v2.1)
- New campaign settings or lore systems
- Multi-language support beyond Spanish/English
- Web interface or CLI rewrite

## Capabilities

### New Capabilities
- `area-generation`: Generate 10-15 numbered areas per act with WotC format (Read-Aloud → Features → Mechanics → Treasure → Secrets)
- `campaign-integration`: Cross-reference validation between areas, NPCs, bestiary, encounters; XP budget calculation; consistency checks
- `handout-generation`: Auto-generate player-facing maps, clue lists, and NPC quick-reference sheets
- `compiler-v2`: Hierarchical TOC, clickable area references, inline stat blocks, DM/Player map versions

### Modified Capabilities
- `npc-generation`: Add alignment, location, combat stats, quest involvement, secrets
- `bestiary-generation`: Add tactics, encounter groups, role classification (skirmisher/tank/controller)
- `encounter-generation`: Add round-by-round development, terrain details, XP totals

## Approach

Implement in 5 phases: Foundation (lore/NPCs/bestiary) → Areas (numbered locations) → Integration (validation) → Visuals (maps/dividers) → Compilation (PDF v2). Each phase produces verified output before next begins. The integrator runs between Areas and Compilation, enforcing cross-references and balance.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `agents/grimorio-acts.md` | Removed | Replaced by `grimorio-areas.md` |
| `agents/grimorio-areas.md` | New | Numbered area generation agent |
| `agents/grimorio-integrator.md` | New | Cross-reference and balance validator |
| `agents/grimorio-npc.md` | Modified | Add stats, alignment, location, secrets |
| `agents/grimorio-bestiary.md` | Modified | Add tactics, encounter groups |
| `agents/grimorio-encounters.md` | Modified | Add round-by-round development |
| `internal/compiler/compiler.go` | Modified | Hierarchical TOC, links, stat blocks |
| `internal/compiler/templates/*.tmpl` | Modified | WotC technical format |
| `internal/compiler/templates/dnd-style.css` | Modified | Area numbers, stat block borders |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| LLM still generates "scenes" not "areas" | High | Strict prompt engineering with WotC examples; format validation in integrator; reject/retry loop |
| v1.x campaigns break with new format | Medium | Maintain `grimorio-acts.md` as legacy mode; migration script in `cmd/migrate-v1-to-v2` |
| PDF size increases significantly | Medium | Use MM references by default; inline stats only for unique creatures; compress images |
| Integration becomes bottleneck | Medium | Cache validation results; allow manual override for edge cases |

## Rollback Plan

1. Revert agent files from git history
2. Restore `grimorio-acts.md` as primary agent
3. Use compiler v1 (no hierarchical TOC) via flag `--compiler-version=1`
4. Legacy campaigns continue working with old pipeline

## Dependencies

- wkhtmltopdf (already required)
- Existing markdown campaigns for testing

## Success Criteria

- [ ] Each act has 10-15 numbered areas (not 3-5 scenes)
- [ ] 90%+ of areas have specific DCs (numeric, not "high/low")
- [ ] 70%+ of areas with creatures have treasure with XP values
- [ ] All creature names in areas exist in bestiary.md
- [ ] PDF has hierarchical TOC with clickable cross-references
- [ ] Player handouts auto-generated (map, clues, NPCs)
- [ ] `make test` passes with no regressions
- [ ] Generated campaign matches WDH quality within 20% on all metrics
