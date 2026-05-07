# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [Unreleased]

### Added
- **Living World Subsystem (Fase 3)** — Dynamic factions, consequences, random tables, and handouts
  - `update_faction_reputation` — Modify faction reputation with BFS propagation to allies/enemies
  - `generate_random_tables` — Contextualized encounter, rumor, weather, treasure tables
  - `generate_handouts` — Dual-version handouts (player-facing + DM-only)
  - `evaluate_consequences` — Evaluate consequence rules against narrative state
  - `FactionService` — CRUD + ReputationMatrix + propagation (2-hop cap, circular detection)
  - `ConsequenceEngine` — Trigger matching, condition evaluation, delayed effects
  - `RandomTableService` — Canon-seeded contextual table generation
  - `HandoutService` — Player/DM dual-version generation
  - `AdaptationPatch` — WorldEvent → markdown patch for DM application
  - Faction Tracker appendix in PDF (Apéndice E)
- **Validation Engine Rule 10** — `faction_reputation_gate` (hostile factions cannot be helpful without cause)

## [2.0.0] - 2026-05-07

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
- **PDF compilation order** — Now follows professional D&D structure: Lore → Acts → Appendices
- **Act template** — Redesigned with Out of the Abyss style sections, read-aloud text, numbered areas
- **Architecture** — Refactored to clean architecture with domain/services/repository layers

## [1.0.0] - 2024-XX-XX

### Added
- Initial release of Grimorio
- MCP server for D&D 5e campaign generation
- AI-powered adventure book generation with styled PDF output
- Support for lore, NPCs, bestiary, encounters, maps, and AI art

[unreleased]: https://github.com/pauvalls/grimorio/compare/v2.0.0...HEAD
[2.0.0]: https://github.com/pauvalls/grimorio/releases/tag/v2.0.0
[1.0.0]: https://github.com/pauvalls/grimorio/releases/tag/v1.0.0
