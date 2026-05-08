# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [2.0.0] - 2026-05-08

### Added
- **Area-Based Generation (v2.0 Migration)** — Complete rewrite from scene-based to WotC-style area-based modules
  - `grimorio-areas` — New agent replacing `grimorio-acts`, generates 10-15 numbered areas per act (150-200 words each)
  - `grimorio-integrator` — New mandatory validation agent: cross-reference checks, XP balance audit, consistency validation, auto-fixes
  - Area format: Features → Mechanics (bold) → Treasure → Secrets → Connections, with specific DCs and XP values
  - 90%+ areas have mechanics (vs 50% in v1.x), 70% of combat areas have treasure (vs 30%)
  - Cross-references between areas ("see Area C4"), NPC locations, and bestiary entries
  - `internal/validators/` — New validation package: area format, integration checks, cross-references
  - `internal/services/handout.go` — HandoutGenerator for player maps, clue lists, NPC rosters, session recaps
  - Compiler v2: hierarchical TOC with links, clickable cross-references, inline stat blocks for unique creatures, area number highlighting
  - `--compiler-version={1|2}` flag for backwards compatibility
  - `grimorio-acts-legacy.md` preserved for old campaigns
  - `cmd/migrate-v1-to-v2` — Converts scene-based acts to area-based format (best-effort)
- **Narrative Coherence Subsystem (Fases 1-2)** — Complete canon and validation system
  - `generate_adventure_bible` — Creates canon.json with immutable facts, entities, timeline, rules
  - `validate_canon` — Validates content proposals against canon (10 rules)
  - `update_narrative_state` — Tracks session state (clues, quests, deaths, decisions)
  - `check_consistency` — Full campaign consistency validation
  - `process_consistency_gate` — Batch validation gate with approve/reject/retry
- **Living World Subsystem (Fase 3)** — Dynamic factions, consequences, random tables, handouts
  - `update_faction_reputation` — Modify faction reputation with BFS propagation
  - `generate_random_tables` — Contextualized encounter, rumor, weather, treasure tables
  - `generate_handouts` — Dual-version handouts (player-facing + DM-only)
  - `evaluate_consequences` — Evaluate consequence rules against narrative state
  - `FactionService`, `ConsequenceEngine`, `RandomTableService`, `HandoutService`
  - Faction Tracker appendix in PDF (Apéndice E)
- **DM Experience (Fase 4)** — Session prep, flowcharts, roster, hooks
  - `generate_session_prep` — "Previously on...", scenarios, relevant NPCs, reminders
  - `generate_flowchart` — Mermaid syntax + native SVG (3 detail levels)
  - `AdventureRoster` — Master table: NPCs, monsters, encounters per act/area
  - `PlayerHookService` — Template-driven hooks connecting PC backgrounds to plot
  - Session Zero template auto-generated on campaign creation
- **Production Polish (Fase 5)** — Caching, benchmarks, CI/CD, docs, release
  - LRU cache for CanonService (90.9% coverage)
  - Performance benchmarks (LoadCanon, ValidateAct, ProcessBatch)
  - Enhanced CI/CD: Go 1.23/1.24 matrix, lint, coverage gate >60%
  - DM guide (docs/dm-guide.md) and Developer guide (docs/developer-guide.md)
  - Graceful degradation with CANON_LEGACY_MODE
  - Dockerfile, Makefile, release scripts

### Added
- **Narrative Coherence Subsystem** — Complete canon and validation system for campaign consistency
  - `generate_adventure_bible` — Creates canon.json with immutable facts, entities, timeline, and world rules
  - `validate_canon` — Validates content proposals against canon (NPC deaths, lore consistency, entity existence)
  - `update_narrative_state` — Tracks session state (clues, quests, deaths, decisions)
  - `check_consistency` — Full campaign consistency validation before PDF compilation
  - `process_consistency_gate` — Batch validation gate with approve/reject/retry workflow
- **Validation Engine** — 10 validation rules (expanded from initial 4):
  - npc_death_state, entity_existence, world_rule_violation, timeline_order
  - quest_reward_existence, level_encounter_balance, location_existence
  - timeline_consistency, prerequisite_clue_check, faction_reputation_gate
- **Consistency Gate Service** — Atomic batch processing with lock management and retry logic
- **Domain models** — CanonDocument, NarrativeState, GateStatus, BatchProposal, GateResult, LockState
- **Dual repository pattern** — Filesystem + in-memory repositories for testing
- **Migration tool** — `migrate-v1-to-v2` converts existing campaigns to new format
- **Test coverage** — 82.6% coverage on services, strict TDD mode activated

### Changed
- **grimorio-npc** — Added alignment, combat stats, location, quest involvement, secrets; WotC format
- **grimorio-bestiary** — Added tactics, role (skirmisher/tank/controller), encounter groups, MM page references
- **grimorio-encounters** — Added round-by-round development, tactical maps, conditions, alternative resolutions
- **PDF compilation order** — Now follows professional D&D structure: Lore → Areas → Appendices
- **Act template** — Redesigned with Out of the Abyss style sections, read-aloud text, numbered areas
- **Architecture** — Refactored to clean architecture with domain/services/repository layers
- **Pipeline** — New 5-phase flow: Foundation → Areas → Integration → Visuals → Compilation

### Fixed
- Fixed 26 golangci-lint errors: unchecked error returns in tests, unused variables, empty branches
- Fixed `domain.KeyItemUpdate` undefined type in narrative state tests
- Fixed race conditions in cache tests
- All tests pass with `-race` flag (100% pass rate, 75.4% coverage)

## [1.0.0] - 2024-XX-XX

### Added
- Initial release of Grimorio
- MCP server for D&D 5e campaign generation
- AI-powered adventure book generation with styled PDF output
- Support for lore, NPCs, bestiary, encounters, maps, and AI art

[unreleased]: https://github.com/pauvalls/grimorio/compare/v2.0.0...HEAD
[2.0.0]: https://github.com/pauvalls/grimorio/releases/tag/v2.0.0
[1.0.0]: https://github.com/pauvalls/grimorio/releases/tag/v1.0.0
