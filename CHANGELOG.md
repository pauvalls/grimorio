# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.5.0] - 2026-05-09

### Added
- **Template Coverage** — All content-generating agents now use structured templates via `get_template` MCP tool
  - **grimorio-quests**: `type="quest"` template with fields: title, type, hook, stakes, reward, status
  - **grimorio-characters**: `type="character"` template with fields: name, race, class, level, background, alignment
  - **grimorio-introduction**: `type="introduction"` template with fields: campaign_name, setting, tone, themes, level_range
  - **grimorio-setting-guide**: `type="setting"` template with fields: world_overview, factions, key_locations, timeline
  - **grimorio-appendices**: `type="appendix"` template with fields: items, monsters, handouts, tables
  - All 5 agents updated to call `get_template()` BEFORE content generation

- **Auto-Regeneration Loops** — 7 agents with max 3 retries when `validate_canon` fails
  - **grimorio-areas**: Retry loop with fix-and-regenerate logic
  - **grimorio-npc**: Retry loop with fix-and-regenerate logic
  - **grimorio-bestiary**: Retry loop with fix-and-regenerate logic
  - **grimorio-encounters**: Retry loop with fix-and-regenerate logic
  - **grimorio-lore**: Retry loop with fix-and-regenerate logic
  - **grimorio-maps**: Retry loop with fix-and-regenerate logic
  - **grimorio-quests**: Retry loop with fix-and-regenerate logic
  - **Pseudocode pattern**: `WHILE retry_count < max_retries AND NOT validation_passed: validate → fix → retry`
  - **Failure handling**: After 3 failed retries, abort and report specific issues for manual review

- **Blocking Validation** — All agents call `validate_canon` BEFORE `save_*` with result checking
  - **grimorio-architect**: Blocking checks added to Phase 3c, 4c, 5c (Batch 1, 2, 3 validation gates)
  - **grimorio-areas**: Explicit result check before `save_areas`
  - **grimorio-npc**: Explicit result check before `save_npcs`
  - **grimorio-bestiary**: Explicit result check before `save_bestiary`
  - **grimorio-encounters**: Explicit result check before `save_encounters`
  - **grimorio-lore**: Explicit result check before `save_lore`
  - **grimorio-maps**: Explicit result check before `save_maps`
  - **grimorio-quests**: Explicit result check before `create_personal_quest`
  - **Rule**: NEVER save content without validation approval

- **Architect Auto-Validation** — Batch validation with max 2 retries per batch
  - **Phase 3c (Batch 1)**: Auto-retry loop for NPCs, Bestiary, Maps validation
  - **Phase 4c (Batch 2)**: Auto-retry loop for Lore, Setting Guide, Quests, Encounters, Characters validation
  - **Phase 5c (Batch 3)**: Auto-retry loop for Acts, SVG Maps validation
  - **Retry logic**: `WHILE retry_count <= max_retries AND NOT validation_passed: delegate → read result → fix → retry`
  - **Blocking behavior**: If validation fails after max retries, DO NOT proceed to next phase

### Changed
- **grimorio-architect workflow** — Enhanced validation phases with explicit retry logic
  - Phase 3c, 4c, 5c: Changed from simple validation to auto-retry loops (max 2 retries)
  - Added explicit `delegation_read(id)` calls to inspect validation results
  - Added failure reporting with specific issue details

- **Agent validation patterns** — All content agents now follow consistent validation flow
  - Before: Informal "validate then fix" guidance
  - After: Explicit pseudocode with retry counts, status checks, and abort conditions

- **Template system** — Expanded from 6 template types to 11 template types
  - New: quest, character, introduction, setting, appendix
  - All agents updated to reference correct template type in documentation

### Documentation
- **README.md**: Updated version to v2.5.0, added features section for template enforcement, auto-regen loops, blocking validation, architect auto-validation
- **CHANGELOG.md**: Complete v2.5.0 entry with all changes documented
- **Agent files**: 12 agents updated with explicit validation pseudocode and template calls

### Technical Details
- **Template types**: 11 total (areas, npc, monster, encounter, map, lore, quest, character, introduction, setting, appendix)
- **Retry counts**: 3 retries for content agents, 2 retries for architect batch validation
- **Validation gates**: 3 batch gates (Batch 1, 2, 3) with blocking behavior
- **Agents affected**: 12 (5 template coverage + 7 auto-regen loops + architect validation)

[2.5.0]: https://github.com/pauvalls/grimorio/compare/v2.4.0...v2.5.0

## [2.4.0] - 2026-05-09

### Added
- **Documentation Consolidation** — Comprehensive documentation for WotC validation and campaign management
  - **README.md**: 6 new sections:
    - Story Brief Template (copy-paste template with examples)
    - Timeout Configuration (table with 15 agents, env vars, defaults)
    - Campaign Paths (default + custom path configuration)
    - Folder Structure (complete ASCII diagrams for project + campaign)
    - WotC Validation Checklist (17 checks with numerical thresholds)
    - Pre-flight Scripts (5 validation commands with expected output)
  - **scripts/validate-campaign.sh**: Standalone validation script (see below)
  - **Validation thresholds documented**: All WotC quality gates now explicit (boxed text 100-600 words, 2+ hooks/area, 3+ developments/area, etc.)

- **WotC Validation Script** — `scripts/validate-campaign.sh` for automated quality checks
  - **Structure validation**: Checks all required directories and files exist
  - **WotC format validation**: grep-based checks for boxed text, hooks, developments, sidebars
  - **Cross-reference validation**: Verifies creature/NPC/quest/location references exist
  - **Content completeness**: Ensures all required sections are present
  - **Formatted output**: Human-readable pass/fail with ✅/❌ indicators
  - **Exit codes**: 0=pass, 1=fail (CI/CD integration ready)
  - **Remediation guidance**: Specific fix suggestions for each failure type
  - **Usage**: `./scripts/validate-campaign.sh {campaign-name} [--check=structure|wotc|references|all]`

- **grimorio-architect Phase X** — Explicit WotC validation phase before PDF compilation
  - Runs `validate-campaign.sh` automatically before Phase 11 (PDF compilation)
  - Blocks compilation if validation fails (hard gate)
  - Reports specific failures to user with remediation steps
  - Retry logic: allows re-validation after fixes

- **Validation thresholds** — Numerical quality gates enforced across all content
  - Boxed Text: 100-600 words (enforced by validator, not just guideline)
  - Character Hooks: ≥2 per area (hard requirement)
  - Developments: ≥3 branches per area with recovery paths (100% required)
  - Running Guidance: 150-400 words per area (validated)
  - Sidebars: ≥1 per act (new requirement for rules clarifications)
  - Area Mechanics: ≥90% of areas must have DC checks or mechanics
  - Combat Treasure: ≥70% of combat areas must have treasure
  - Chapter Mode Variety: Max 2 consecutive acts with same game mode
  - Asset Handoff: 100% of acts must pass asset to next act
  - Chapter Objectives: 2-3 per act (validated)

### Changed
- **grimorio-architect workflow** — Added Phase X (WotC Validation) between Phase 9 (Final Consistency Check) and Phase 10 (PDF Compilation)
  - Phase X is mandatory — cannot skip validation
  - Validation report shown to user before compilation proceeds
  - Failed validation blocks PDF compilation until issues are resolved

- **Validation enforcement** — WotC thresholds changed from guidelines to requirements
  - Previously: Narrative custodian checked thresholds as soft guidelines
  - Now: validate-campaign.sh enforces thresholds as hard gates
  - PDF compilation fails if any threshold is not met

- **Error reporting** — Validation failures now include specific remediation steps
  - Previously: Generic "validation failed" message
  - Now: "Expand boxed text in Area A3, A7, B2 (add sensory details)"

### Documentation
- **README.md**: +450 lines of documentation (Story Brief, Timeouts, Paths, Structure, Checklist, Pre-flight)
- **CHANGELOG.md**: Complete v2.4.0 entry with all changes documented
- **grimorio-architect.md**: Phase X added with exact grep commands and thresholds
- **scripts/validate-campaign.sh**: 350-line bash script with full validation logic

### Technical Details
- **validate-campaign.sh**: 350 lines, 4 check types, 17 validation rules
- **grep patterns**: Exact patterns for boxed text (`^>>`), hooks (`Hook:|Gancho:`), developments (`Development:|Desarrollo:`), sidebars (`^> #####`)
- **Exit codes**: 0 (pass), 1 (fail), 2 (error/invalid arguments)
- **CI/CD ready**: Script designed for automation with machine-parseable output

[2.4.0]: https://github.com/pauvalls/grimorio/compare/v2.3.0...v2.4.0

## [2.3.0] - 2026-05-08

### Added
- **WotC Format Improvements** — Enhanced area and NPC generation to match WotC published module quality
  - **Templates**: 5 new sections in `areas.md.tmpl`:
    - Boxed Text (`>>` format, 2-4 párrafos, 100-600 palabras)
    - Character Hooks (2+ por área, targetean backgrounds/clases específicas)
    - Developments (3+ ramas con recovery paths obligatorios)
    - Cómo Dirigir esta Escena (5 subsecciones: Preparación, Ritmo, Señales, Improvisar, Guión)
    - Evolución del Hub (opcional, para áreas hub)
  - **Agents**: Rules 15-19 in `grimorio-areas.md` for boxed text, hooks, developments, scene guidance, NPC descriptions
  - **NPC Agent**: Enhanced requirements in `grimorio-npc.md`:
    - 5+ párrafos de descripción (3-5 apariencia, 2-3 personalidad/voz)
    - 3-5 secretos por NPC (1 trivial, 2 importantes, 1-2 críticos)
    - 3-5 líneas de diálogo sample para NPCs clave
  - **Validation**: Check 13A-E in `grimorio-narrative-custodian.md`:
    - 13A: Boxed Text word count (100-600)
    - 13B: Character Hooks count (2+)
    - 13C: Developments branches (3+ with recovery)
    - 13D: Running the Scene subsections (5)
    - 13E: NPC depth (5+ paragraphs, 3+ secrets, 3+ dialogue)
  - Test area generated and validated with all new features

### Changed
- **Template structure** — New sections added after area description, before DM details
- **Agent instructions** — Stricter quality requirements for immersive narration and NPC depth
- **Validator** — 5 new validation checks for WotC format quality gates
- **Checklist** — Updated pre-save checklist in `grimorio-areas.md` to v2.3

### Documentation
- Test area created: `test-campaign/wotc-format-test/test-area.md`
- All 25 tasks completed across 6 phases

## [2.2.0] - 2026-05-08

### Added
- **Chapter Narrative Structure** — WotC-style chapter openers for cohesive narrative framing
  - Domain model: 7 new fields in `Act` struct (`GameMode`, `GameModeSecondary`, `ChapterObjectives`, `EstimatedDuration`, `Tone`, `RunningGuidance`, `AssetHandoff`)
  - Template: "Apertura del Capítulo" section in `areas.md.tmpl` with mode badge, objectives, duration, running guidance, asset handoff
  - Agent: Rules 13-14 in `grimorio-areas.md` for chapter mode selection, asset handoff, mode variety algorithm
  - Validation: Check 12 in `grimorio-narrative-custodian.md` for mode variety, mode-content alignment, asset chain validation
  - 8 canonical game modes: `investigacion`, `sandbox_urbano`, `dungeon_lineal`, `escape`, `viaje`, `intriga`, `confrontacion`, `downtime`
  - 4 asset types: `objeto`, `información`, `aliado`, `base`
  - Unit tests: 17 test cases for domain validation (96.4% coverage)
  - Test fixtures updated across repository, services, and handlers packages

### Changed
- **Template structure** — Chapter opener inserted after blockquote header, before "Adventure Background"
- **Agent instructions** — Mode selection based on act position, variety enforcement (max 2 consecutive same mode)
- **Validator** — Cross-act validation for mode variety and asset chain continuity
- **CampaignService.SaveAct** — Now includes default values for new required fields

### Fixed
- Duration format regex to properly validate "1 sesión" and "X-Y sesiones" patterns
- Test fixtures in `internal/repository`, `internal/services`, and `internal/mcp/handlers`

## [Unreleased]

### Added
- **Chapter Narrative Structure** — WotC-style chapter openers for cohesive narrative framing
  - Domain model: 7 new fields in `Act` struct (`GameMode`, `GameModeSecondary`, `ChapterObjectives`, `EstimatedDuration`, `Tone`, `RunningGuidance`, `AssetHandoff`)
  - Template: "Apertura del Capítulo" section in `areas.md.tmpl` with mode badge, objectives, duration, running guidance, asset handoff
  - Agent: Rules 13-14 in `grimorio-areas.md` for chapter mode selection, asset handoff, mode variety algorithm
  - Validation: Check 12 in `grimorio-narrative-custodian.md` for mode variety, mode-content alignment, asset chain validation
  - 8 canonical game modes: `investigacion`, `sandbox_urbano`, `dungeon_lineal`, `escape`, `viaje`, `intriga`, `confrontacion`, `downtime`
  - 4 asset types: `objeto`, `información`, `aliado`, `base`
  - Unit tests: 17 test cases for domain validation (96.4% coverage)

### Changed
- **Template structure** — Chapter opener inserted after blockquote header, before "Adventure Background"
- **Agent instructions** — Mode selection based on act position, variety enforcement (max 2 consecutive same mode)
- **Validator** — Cross-act validation for mode variety and asset chain continuity

### Fixed
- Duration format regex to properly validate "1 sesión" and "X-Y sesiones" patterns

## [2.1.0] - 2026-05-08

### Added
- **Professional WotC Format** — Introduction, Setting Guide, and Appendices as formal campaign chapters
  - `grimorio-introduction` — New agent: Foreword, Story Overview, Adventure Background timeline, Running the Adventure, Character Creation Guidelines
  - `grimorio-setting-guide` — New agent (DM-only): Geography, History, Culture, Factions, Secrets and Lies, Economy
  - `grimorio-appendices` — New agent: Magic Items (A), NPCs/Monsters (B), Handouts (C), Maps (D), Reference Tables (E)
  - 3 new MCP handlers: `save_introduction`, `save_setting_guide`, `save_appendices`
  - 3 new Go templates: `introduction.md.tmpl`, `setting-guide.md.tmpl`, `appendices.md.tmpl`
  - Compiler v2 updated: Introduction → Lore → Acts → Setting Guide → Individual Appendices → Appendices
- **Area Enhancements** — Inline NPC stat summaries (*alignment race class*) and sidebar support (> ##### pattern)
  - `grimorio-areas` agent updated: inline stats, sidebars, rule 11 (≥1 sidebar/act), rule 12 (inline NPC stats)
  - Sidebar CSS styling: red left border, dashed edges, Cinzel h5 heading
- **Pipeline Updates** — 12-phase flow with Introduction, Setting Guide, and Appendices phases

### Changed
- **Compiler sections order** — Introduction, Lore, Acts, Setting Guide (DM-only), individual appendices, unified Appendices.md
- **Pipeline** — grimorio-architect now orchestrates 12 phases (was 11)
- **Agent hierarchy** — grimorio-areas is the ONLY area generator (grimorio-acts-legacy.md removed)

### Fixed
- Duplicate "Sesión Cero — Guía para el DM" heading in HTML output (hardcoded h2 removed)
- `grimorio-architect` was referencing non-existent `grimorio-acts` agent — updated to `grimorio-areas`

### Changed
- **grimorio-acts-legacy.md** — REMOVED (replaced by grimorio-areas.md)

## [2.0.0] - 2026-05-08

### Added
- **Area-Based Generation (v2.0 Migration)** — Complete rewrite from scene-based to WotC-style area-based modules
  - `grimorio-areas` — New agent (the ONLY area generator), generates 10-15 numbered areas per act (150-200 words each)
  - `grimorio-integrator` — New mandatory validation agent: cross-reference checks, XP balance audit, consistency validation, auto-fixes
  - Area format: Features → Mechanics (bold) → Treasure → Secrets → Connections, with specific DCs and XP values
  - 90%+ areas have mechanics (vs 50% in v1.x), 70% of combat areas have treasure (vs 30%)
  - Cross-references between areas ("see Area C4"), NPC locations, and bestiary entries
  - `internal/validators/` — New validation package: area format, integration checks, cross-references
  - `internal/services/handout.go` — HandoutGenerator for player maps, clue lists, NPC rosters, session recaps
  - Compiler v2: hierarchical TOC with links, clickable cross-references, inline stat blocks for unique creatures, area number highlighting
  - `--compiler-version={1|2}` flag for backwards compatibility
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
