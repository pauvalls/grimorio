# Tasks: Migrate Grimorio v1.x to v2.0 — Area-Based Campaign Generation

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 1,500–2,500 |
| 400-line budget risk | **High** |
| Chained PRs recommended | **Yes** |
| Suggested split | PR 1 (F1 Foundation) → PR 2 (F2 Areas + F3 Integration) → PR 3 (F4 Visuals + F5 Compilation) |
| Delivery strategy | exception-ok |
| Chain strategy | pending |

Decision needed before apply: No
Chained PRs recommended: Yes
Chain strategy: pending
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | F1: Update agents (npc, bestiary, encounters) + retrocompat flag | PR 1 | Base: main; includes unit tests |
| 2 | F2+F3: Areas agent, area validator, integrator v2 | PR 2 | Base: PR 1 branch; validation gates |
| 3 | F4+F5: Handout service, compiler v2, templates, CSS | PR 3 | Base: PR 2 branch; end-to-end test |

---

## Phase 1: Foundation

- [x] 1.1 Modify `agents/grimorio-npc.md` — add alignment, location, combat stats, quest involvement, secrets
- [x] 1.2 Modify `agents/grimorio-bestiary.md` — add role classification, encounter groups, source reference, structured tactics
- [x] 1.3 Modify `agents/grimorio-encounters.md` — add round-by-round, tactical map, conditions, alternative resolution, encounter templates
- [x] 1.4 Modify `cmd/grimorio/main.go` — add `--compiler-version={1|2}` flag for retrocompatibility
- [x] 1.5 RED: Write failing tests for updated NPC/bestiary/encounter output format
- [x] 1.6 GREEN: Run tests; fix agent output until tests pass
- [x] 1.7 REFACTOR: Extract shared format helpers in agent prompts if duplicated

## Phase 2: Areas

- [x] 2.1 Modify `agents/grimorio-areas.md` — enforce 150-200 words/area, numeric DCs, bidirectional connections, WotC format
- [x] 2.2 Delete `agents/grimorio-acts.md` (or archive as `agents/grimorio-acts-legacy.md`)
- [x] 2.3 Create `internal/validators/area.go` — count areas, word count, numeric DCs, bidirectional connections
- [x] 2.4 RED: Write `internal/validators/area_test.go` with table-driven cases for all validation rules
- [x] 2.5 GREEN: Implement validators until all tests pass
- [x] 2.6 REFACTOR: Optimize regex compilation as package-level vars

## Phase 3: Integration

- [x] 3.1 Modify `agents/grimorio-integrator.md` — add programmatic validation phase description and auto-fix capabilities
- [x] 3.2 Create integration test `TestIntegrationValidation` — broken references, missing creatures, one-way connections
- [x] 3.3 RED: Write failing integration test for XP budget calculation per act
- [x] 3.4 GREEN: Implement XP budget check in integrator logic; tests pass
- [x] 3.5 Add treasure consistency check — every area with creatures has treasure with XP
- [x] 3.6 Verify `check_consistency` and `process_consistency_gate` are called by integrator

## Phase 4: Visuals

- [x] 4.1 Create `internal/services/handout.go` — `HandoutGenerator` with player map redaction, clue list, NPC quick-reference
- [x] 4.2 RED: Write `internal/services/handout_test.go` — mock campaign dirs, assert redaction and filtering
- [x] 4.3 GREEN: Implement handout generation until tests pass
- [x] 4.4 Create `internal/compiler/handouts.go` — wire `HandoutRenderer` interface into compiler pipeline
- [x] 4.5 Update map/divider generation if needed for player-facing versions

## Phase 5: Compilation

- [x] 5.1 Modify `internal/compiler/compiler.go` — hierarchical TOC (3 levels), clickable cross-references, inline stat blocks, 2-col layout, handout pages
- [x] 5.2 Modify `internal/compiler/templates/dnd-style.css` — area number highlighting, stat block borders, read-aloud boxed style, handout page styles
- [x] 5.3 Modify `internal/compiler/templates/act.md.tmpl` — WotC format: Read-Aloud → Features → Mechanics → Treasure → Connections → Secrets
- [x] 5.4 Modify `internal/compiler/templates/encounter.md.tmpl` — round-by-round, adjusted XP, tactical map, conditions
- [x] 5.5 Modify `internal/compiler/templates/npc.md.tmpl` — alignment, location, stats, quest, secret fields
- [x] 5.6 Modify `internal/compiler/templates/monster.md.tmpl` — role, encounter groups, source, structured tactics
- [x] 5.7 RED: Write `compiler_test.go` with markdown fixtures — assert TOC structure, cross-reference links, stat block embedding
- [x] 5.8 GREEN: Run compiler tests; fix rendering until pass
- [x] 5.9 Modify `cmd/migrate-v1-to-v2/main.go` — extend to convert scenes → numbered areas with best-effort
- [x] 5.10 Run `make test` with race and coverage; zero regressions
- [ ] 5.11 Validate generated PDF matches WDH quality metrics within 20%
