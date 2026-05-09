# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [3.0.0] - 2026-05-09

**WotC Professional Quality** — Major version bump with unified area format, milestone XP, enhanced handouts, and E2E testing

### Added

#### Domain Models (Phase 1 - TASK-001 to TASK-008)
- **MilestoneXP** domain model with PHB threshold tracking ([`c7173a9`](https://github.com/pauvalls/grimorio/commit/c7173a9))
- **MagicItem** domain model with rarity, attunement, and curse support ([`c7173a9`](https://github.com/pauvalls/grimorio/commit/c7173a9))
- **Tactics** domain model with intelligence tiers and environmental tactics ([`c7173a9`](https://github.com/pauvalls/grimorio/commit/c7173a9))
- **PlayerMap** domain model for player-facing maps with secret redaction ([`c7173a9`](https://github.com/pauvalls/grimorio/commit/c7173a9))
- **SessionZeroGuide** domain model with content warnings and safety tools ([`c7173a9`](https://github.com/pauvalls/grimorio/commit/c7173a9))
- **ConsequenceTable** domain model for act transition tracking ([`c7173a9`](https://github.com/pauvalls/grimorio/commit/c7173a9))
- **Area** unified WotC format with sequential numbering 1-15 ([`c7173a9`](https://github.com/pauvalls/grimorio/commit/c7173a9))
- **PregenCharacter** with campaign-specific bonds/ideals/flaws ([`c7173a9`](https://github.com/pauvalls/grimorio/commit/c7173a9))
- Enhanced **Handout** with type/format/style fields (letter, clue, document, journal) ([`c7173a9`](https://github.com/pauvalls/grimorio/commit/c7173a9))
- Enhanced **Quest** with QuestApproach, QuestFailure, QuestClue structures ([`c7173a9`](https://github.com/pauvalls/grimorio/commit/c7173a9))

#### Services (Phase 1 - TASK-009 to TASK-018)
- **MilestoneService** for XP table generation and party level tracking ([`06c22d4`](https://github.com/pauvalls/grimorio/commit/06c22d4))
- **ItemService** for magic item generation with rarity validation ([`06c22d4`](https://github.com/pauvalls/grimorio/commit/06c22d4))
- **TacticsService** for enemy AI and combat guidance ([`06c22d4`](https://github.com/pauvalls/grimorio/commit/06c22d4))
- **PlayerMapService** for player-facing map generation ([`06c22d4`](https://github.com/pauvalls/grimorio/commit/06c22d4))
- **SessionZeroService** for campaign-specific Session Zero guides ([`06c22d4`](https://github.com/pauvalls/grimorio/commit/06c22d4))
- **HandoutServiceV3** for enhanced handout generation ([`06c22d4`](https://github.com/pauvalls/grimorio/commit/06c22d4))
- **ConsequenceService** for act transition consequence tracking ([`06c22d4`](https://github.com/pauvalls/grimorio/commit/06c22d4))
- **AreaService** for unified WotC area generation ([`06c22d4`](https://github.com/pauvalls/grimorio/commit/06c22d4))

#### MCP Handlers (Phase 2 - TASK-021 to TASK-028)
- `grimorio_generate_xp_table` — Generate milestone XP tables ([`23e4559`](https://github.com/pauvalls/grimorio/commit/23e4559))
- `grimorio_track_party_progress` — Track party level progression ([`23e4559`](https://github.com/pauvalls/grimorio/commit/23e4559))
- `grimorio_generate_magic_item` — Generate magic items by rarity ([`23e4559`](https://github.com/pauvalls/grimorio/commit/23e4559))
- `grimorio_generate_tactics` — Generate enemy combat tactics ([`23e4559`](https://github.com/pauvalls/grimorio/commit/23e4559))
- `grimorio_generate_area` — Generate WotC-format areas ([`23e4559`](https://github.com/pauvalls/grimorio/commit/23e4559))
- `grimorio_generate_areas_chapter` — Generate full chapter areas ([`23e4559`](https://github.com/pauvalls/grimorio/commit/23e4559))

#### Templates (Phase 2 - TASK-029, TASK-035)
- **milestone-xp.md.tmpl** — XP table markdown template ([`23e4559`](https://github.com/pauvalls/grimorio/commit/23e4559))
- **area.md.tmpl** — Unified WotC area template ([`23e4559`](https://github.com/pauvalls/grimorio/commit/23e4559))

#### Validators (Phase 2 - TASK-037, TASK-038)
- **AreaValidator** with WotC quality checks ([`23e4559`](https://github.com/pauvalls/grimorio/commit/23e4559))
- **QuestValidator** with 3-approach validation ([`23e4559`](https://github.com/pauvalls/grimorio/commit/23e4559))

#### Phase 3: Assets & Handouts (TASK-041 to TASK-055)
- **PlayerMapGenerator** for player-facing map variants ([`1352898`](https://github.com/pauvalls/grimorio/commit/1352898))
- **HandoutGenerator** for letters, clues, documents ([`1352898`](https://github.com/pauvalls/grimorio/commit/1352898))
- **ConsequenceGenerator** for act transition tables ([`1352898`](https://github.com/pauvalls/grimorio/commit/1352898))
- **SessionZeroGenerator** for campaign-specific guides ([`1352898`](https://github.com/pauvalls/grimorio/commit/1352898))
- **FactionTracker** for reputation tracking ([`1352898`](https://github.com/pauvalls/grimorio/commit/1352898))

#### Phase 4: Testing & Automation (TASK-056 to TASK-070)
- **E2E Test Suite** with 7 comprehensive tests ([`1352898`](https://github.com/pauvalls/grimorio/commit/1352898))
  - e2e_full_campaign_test.go
  - e2e_milestone_test.go
  - e2e_consequence_test.go
  - e2e_session_flow_test.go
  - e2e_handout_test.go
  - e2e_random_tables_test.go
  - e2e_canon_validation_test.go
- **Changelog Automation** script ([`1352898`](https://github.com/pauvalls/grimorio/commit/1352898))
- **Migration Script** v2 to v3 with rollback support ([`1352898`](https://github.com/pauvalls/grimorio/commit/1352898))
- **Migration Guide** documentation ([`1352898`](https://github.com/pauvalls/grimorio/commit/1352898))

### Changed

- Extended `domain.Handout` with V3 types (letter, clue, document, journal, etc.) ([`c7173a9`](https://github.com/pauvalls/grimorio/commit/c7173a9))
- Extended `domain.Quest` with QuestApproach, QuestFailure, QuestClue structures ([`c7173a9`](https://github.com/pauvalls/grimorio/commit/c7173a9))
- Exported `domain.IsValidRarity` for service use ([`06c22d4`](https://github.com/pauvalls/grimorio/commit/06c22d4))
- Updated README.md with v3.0.0 features ([`1352898`](https://github.com/pauvalls/grimorio/commit/1352898))

### Technical Details

**Architecture**: Hexagonal architecture with domain-driven design  
**Test Coverage**: Unit tests for all domain models and core services + E2E test suite  
**Backward Compatibility**: Additive changes only, existing campaigns remain valid  
**Total Lines**: ~6000 lines across 70 tasks  
**Commits**: 8 logical work-unit commits

### Migration

Existing v2.6.0 campaigns remain compatible. New fields are optional/omitted for legacy campaigns.

**Automated Migration**:
```bash
go run scripts/migrate/migrate_v2_to_v3.go <campaign_path>
```

**Documentation**: See `docs/migration-v2-to-v3.md` for complete migration guide.

---

## [2.6.0] - Previous Version

See git history for v2.6.0 changes.

---

**Full Changelog**: [v2.6.0...v3.0.0](https://github.com/pauvalls/grimorio/compare/v2.6.0...v3.0.0)  
**Release Date**: 2026-05-09  
**Total Changes**: ~6000 lines across 70 tasks  
**Phases Completed**: 4 of 4 (100%)
