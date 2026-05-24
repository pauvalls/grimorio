# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
## [v4.0.17] - 2026-05-24

### Docs

- Update changelog for v4.0.16

### Fix

- Fix update_narrative_state bugs reported by user (mcp)

## [v4.0.16] - 2026-05-24

### Docs

- Update changelog for v4.0.15

### Feat

- Full update_narrative_state overhaul (mcp)

## [v4.0.15] - 2026-05-24

### Docs

- Add v4.0.14 entry for update_narrative_state fix (changelog)
- Update changelog for v4.0.14

### Fix

- Replace instead of append for state fields (narrative)





## [v4.0.14] - 2026-05-24

### Fix

- `update_narrative_state` now correctly populates all fields:
  - `active_quests` and `key_items` save as string arrays in root state
  - `dm_notes` and `loot_acquired` save in root state (not just session_log)
  - Added `default_source_act`, `default_choice_made`, `default_impact_scope` params
  - `source_act` no longer hardcoded to "unknown" when using string clues
  - `choice_made` and `impact_scope` no longer empty when using string decisions

## [v4.0.13] - 2026-05-24

### Fix

- `grimorio update` now handles GoReleaser's subdirectory wrapping in release archives
- Added `findBinaryInExtractedDir` helper to search for binary in subdirectories
- Fixes "binary not found in extracted archive" error during self-update

## [v4.0.12] - 2026-05-24

### Fix

- `update_narrative_state` now correctly handles `[]string` arrays from MCP (not just `[]any`)
- Added `getStringArray` helper to handle both `[]string` and `[]any` array types
- All array parameters now use the helper: `revealed_clues`, `dead_npcs`, `key_decisions`, `active_quests`, `key_items`, `loot_acquired`

## [v4.0.11] - 2026-05-24

### Fix

- `update_narrative_state` MCP tool now accepts string arrays for `revealed_clues`, `dead_npcs`, and `key_decisions` (not just objects)
- Added missing MCP parameters: `active_quests`, `key_items`, `session_summary`, `xp_awarded`, `loot_acquired`, `dm_notes`
- Arrays are no longer silently skipped when agents send simple strings instead of objects

## [v4.0.10] - 2026-05-24

### Feat

- Add `grimorio update commands` to update opencode.json MCP + command entries
- Add `grimorio update all` to update skills, agents, and commands in one run
- Update campaign generation template with detailed batch workflow (sequential areas to avoid timeout, narrative + WotC validation after each batch, final integration checks)
- Update install.sh instructions to reflect all 4 update subcommands

## [v4.0.8] - 2026-05-22

### Docs

- Update changelog for v4.0.7

### Fix

- Resolve remaining errcheck issues in download.go and update_test.go (lint)

## [v4.0.7] - 2026-05-22

### Docs

- Update changelog for v4.0.6

### Fix

- Resolve errcheck, staticcheck, and unused issues across update package and tests (lint)

## [v4.0.6] - 2026-05-22

### Docs

- Update changelog for v4.0.5
- Advise running update skills and update agents after install (install)

## [v4.0.5] - 2026-05-22

### Docs

- Update changelog for v4.0.4

### Fix

- Replace deprecated tools.mcp array with permission block (agents)

## [v4.0.4] - 2026-05-22

### Docs

- Update changelog for v4.0.3

### Fix

- Include agents in release archive and add update skills/agents commands (goreleaser)

## [v4.0.3] - 2026-05-22

### Docs

- Update changelog for v4.0.2

### Fix

- Copy agents to opencode global agents directory (install)

## [v4.0.2] - 2026-05-22

### Docs

- Update changelog for v4.0.1

### Fix

- Correct MCP binary path to install dir (install)

## [v4.0.1] - 2026-05-22

### Build

- Bundle agents and skills in release archives (goreleaser)

### Docs

- Restructure highlights to feature v4.0.0 DM prominently (readme)
- Update changelog for v4.0.0
- Add troubleshooting section for update failures (readme)

### Feat

- Implement self-updater with platform detection, download, and atomic replacement (update)
- Rewrite install.sh and create install.ps1 for cross-platform binary distribution (install)
- Update install and update targets with binary fallback (make)

### Fix

- Fix update script bugs - git pull errors, agent registration, binary sync (install)
- Detect non-git install dir and corrupted repos during update (install)
- Synchronize narrative_state.json location between repo and handout generator
- Redirect log output to stderr to prevent variable contamination (install)

## [v4.0.0] - 2026-05-20

### Docs

- Update changelog for v3.8.0
- Add MIT license file and prettier README badges
- Simplify badges to version, CI, and license (readme)
- Replace MIT with MPL-2.0 license, fix author name

### Feat

- Add grimorio-dm AI Dungeon Master agent (v4.0) (dm)

### Fix

- Check outFile.Close in import.go (lint)
- Check errcheck in dm_context_service pdf extraction (lint)

## [v3.8.0] - 2026-05-19

### Ci

- Add go mod download before golangci-lint

### Docs

- Update changelog for v3.7.0
- Add v3.6.0 and v3.7.0 highlights (readme)
- Update changelog for v3.8.0
- Update changelog for v3.8.0
- Update changelog for v3.8.0

### Feat

- Add campaign management CLI commands and MCP visualization tools (campaign)

### Fix

- Check fmt.Sscanf error return in milestone_service.go (lint)
- .gitignore grimorio pattern was ignoring cmd/grimorio/commands/
- Resolve 36 errcheck and staticcheck issues across tests (lint)
- Resolve remaining 18 errcheck/staticcheck issues (lint)

## [v3.8.0] - 2026-05-19

### Ci

- Add go mod download before golangci-lint

### Docs

- Update changelog for v3.7.0
- Add v3.6.0 and v3.7.0 highlights (readme)
- Update changelog for v3.8.0
- Update changelog for v3.8.0

### Feat

- Add campaign management CLI commands and MCP visualization tools (campaign)

### Fix

- Check fmt.Sscanf error return in milestone_service.go (lint)
- .gitignore grimorio pattern was ignoring cmd/grimorio/commands/
- Resolve 36 errcheck and staticcheck issues across tests (lint)

## [v3.8.0] - 2026-05-19

### Ci

- Add go mod download before golangci-lint

### Docs

- Update changelog for v3.7.0
- Add v3.6.0 and v3.7.0 highlights (readme)
- Update changelog for v3.8.0

### Feat

- Add campaign management CLI commands and MCP visualization tools (campaign)

### Fix

- Check fmt.Sscanf error return in milestone_service.go (lint)
- .gitignore grimorio pattern was ignoring cmd/grimorio/commands/

## [v3.8.0] - 2026-05-19

### Docs

- Update changelog for v3.7.0
- Add v3.6.0 and v3.7.0 highlights (readme)

### Feat

- Add campaign management CLI commands and MCP visualization tools (campaign)

### Fix

- Check fmt.Sscanf error return in milestone_service.go (lint)

## [v3.7.0] - 2026-05-19

### Cleanup

- Remove vestigial ConsequenceService

### Docs

- Update changelog for v3.6.0
- Update changelog for v3.5.0

### Feat

- Add release-tag target for auto-versioning (make)
- Add filesystem V3 repository implementations
- Implement service TODO stubs for V3
- Wire V3 handlers and register MCP tools
- Update agents and install.sh for V3 tools

### Test

- Add round-trip and service unit tests for V3 changes


## [v3.6.0] - 2026-05-19

### Feat

- Add domain model, template, and CSS for 4-part narrative prologue (prologue)
- Add PrologueService and CampaignService.SavePrologue (prologue)
- Add MCP tool handler and server registration (prologue)

### Fix

- Integrator agent missing consistency tool references
- Replace WriteString(fmt.Sprintf) with fmt.Fprintf in prologue_service.go (lint)

### Test

- Add tests for domain, service, handler, template, and CSS (prologue)

## [v3.4.0] - 2026-05-19

### Chore

- Remove orphaned SDD markdown files from root

### Docs

- Add WotC validation and HTML fix to changelog (README)
- Update version badge to dynamic shield.io + simplify what's new (readme)

### Feat

- Add 16 grimorio skills with WotC standards preserved
- Extend check_consistency with WotC format + integration validation
- Fix nested divs + add narrative prologue support (compiler)
- Automate releases and changelogs with git-cliff

### Fix

- Resolve final 6 linter errors (CI green)
- Update to fix lit ci issues (lint)
- Update skill registry paths to ~/.config/opencode/skills/
- Prevent <p> wrapping around HTML block placeholders (compiler)
- Resolve staticcheck and unused linter errors
- Replace WriteString(fmt.Sprintf) with fmt.Fprintf in flowchart.go
- Replace WriteString(fmt.Sprintf) with fmt.Fprintf in handout.go
- Replace WriteString(fmt.Sprintf) with fmt.Fprintf in handout.go and svg.go
- Replace all WriteString(fmt.Sprintf) with fmt.Fprintf in svg.go
- Auto-close unclosed divs and add missing page-break CSS (compiler)
- Install git-cliff directly instead of using action (ci)
- Checkout main branch for changelog job to avoid detached HEAD (ci)
- Use commit range for git-cliff instead of --unreleased flag (ci)
- Use GH_PAT secret to bypass branch protection for changelog push (ci)

### Refactor

- Delete agents/ directory (consolidated into skills/)
- Consolidate skills+agents architecture, fix install script

### Test

- Add regression test for HTML block wrapping bug (compiler)

## [v3.4.0] - 2026-05-19

### Chore

- Remove orphaned SDD markdown files from root

### Docs

- Add WotC validation and HTML fix to changelog (README)

### Feat

- Add 16 grimorio skills with WotC standards preserved
- Extend check_consistency with WotC format + integration validation
- Fix nested divs + add narrative prologue support (compiler)

### Fix

- Resolve final 6 linter errors (CI green)
- Update to fix lit ci issues (lint)
- Update skill registry paths to ~/.config/opencode/skills/
- Prevent <p> wrapping around HTML block placeholders (compiler)
- Resolve staticcheck and unused linter errors
- Replace WriteString(fmt.Sprintf) with fmt.Fprintf in flowchart.go
- Replace WriteString(fmt.Sprintf) with fmt.Fprintf in handout.go
- Replace WriteString(fmt.Sprintf) with fmt.Fprintf in handout.go and svg.go
- Replace all WriteString(fmt.Sprintf) with fmt.Fprintf in svg.go
- Auto-close unclosed divs and add missing page-break CSS (compiler)

### Refactor

- Delete agents/ directory (consolidated into skills/)
- Consolidate skills+agents architecture, fix install script

### Test

- Add regression test for HTML block wrapping bug (compiler)
## [v3.3.3] - 2026-05-10

### Fix

- Resolve 9 remaining linter errors (6 errcheck, 3 staticcheck)
## [v3.3.2] - 2026-05-10

### Fix

- PDF HTML rendering (preserve HTML blocks, remove &thinsp; spacing issues)
## [v3.3.1] - 2026-05-10

### Feat

- Add narrative prologue requirement for Chapter 1 (400-600 words, 4-part structure)

### Fix

- Preserve HTML blocks and fix spacing in PDF rendering
## [v3.3.0] - 2026-05-10

### Fix

- Resolve remaining 16 linter errors (errcheck, staticcheck)
## [v3.2.3] - 2026-05-10

### Fix

- Increase compilePDF timeout to 600s (10 min) for large campaigns
## [v3.2.2] - 2026-05-10

### Fix

- Remove Write tool from 8 content agents to prevent double-write
## [v3.2.0] - 2026-05-10

### Fix

- PDF generation bugs and sequential act generation
## [v3.1.1] - 2026-05-10

### Fix

- Resolve 49 linter errors (errcheck, govet, staticcheck, ineffassign)
## [v3.1.0] - 2026-05-10

### Fix

- Exclude /scripts/ from go test to prevent covdata error (ci)
## [v3.0.5] - 2026-05-09

### Fix

- Copy_commands function order and missing closing
## [v3.0.4] - 2026-05-09

### Feat

- Add grimorio command file for OpenCode/Claude Code
## [v3.0.3] - 2026-05-09

### Ci

- Update to Go 1.25 and golangci-lint-action@v7
## [v3.0.2] - 2026-05-09

### Test

- Fix CI tests - skip when wkhtmltopdf unavailable
## [v3.0.1] - 2026-05-09

### Chore

- Update install.sh with proper version ldflags

### Docs

- Add comprehensive PDF Compiler Enhancements guide
- Restructure documentation with bilingual support (EN/ES)
- Fix acts/ to areas/ directory references

### Feat

- Add CSS classes for DM sidebars, stat-blocks v2, and session components
- Extend domain models with ShockPoints, BackstoryHooks, and SessionPrep enhancements
- Add Shock Points generation to Session Zero Service
- Add Shock Points section and character worksheet to Session Zero template
- Create SessionGenerator service for contextual session content
- Add GetPrepWithScenarios() method to SessionPrepService
- Create Session Prep template with comprehensive sections
- Expand CharacterService with backstory and narrative generation
- Create Character Sheet template with comprehensive sections
- Extend PDF Compiler with new HTML generation methods
- Update MCP handlers with expanded functionality

### Fix

- /grimorio command executes in main thread (no agent delegation)
- Install.sh now validates JSON and creates backups before modifying opencode.json

### Test

- Add integration tests for PDF compilation with new features
- Add CSS visual regression tests
- Add backward compatibility verification tests
## [v3.0.0] - 2026-05-09

### Docs

- Add CHANGELOG.md v3.0.0 entry
- Update CHANGELOG.md with complete v3.0.0 entry (70 tasks)

### Feat

- Sync command template from grimorio-architect agent
- Clean Installer v2 - Complete MCP installation
- Add V3 domain models (milestone, magic item, tactics, areas, consequences)
- Implement Phase 1 services (milestone, item, tactics, area, consequence)
- Implement Phase 2 MCP handlers, templates, and validators (TASK-021 to TASK-040)
- Complete Phase 3 & 4 - E2E tests, changelog automation, migration (TASK-041 to TASK-070)

### Fix

- Correct sync_command_from_agent function placement
- Restore clean_installation function
- Handle missing files in setup_plugin gracefully
- Remove unused import in E2E tests
- Test updates for v3.0 compatibility

### Refactor

- Organize scripts into subdirectories, remove duplicate generators
## [v2.6.0] - 2026-05-09

### Add

- Campaign PDF for La Hoja de Vlad example

### Docs

- Add SDD solutions reference for common problems
- Add solution for agents not using templates
- Add path issue solution to SDD-SOLUTIONS
- Add story brief requirement to SDD-SOLUTIONS
- Add timeout configuration for grimorio-areas
- Add real campaign structure examples from la-hoja-de-vlad
- Documentation Consolidation v2.4.0

### Feat

- Add WotC quality validators and character hooks integration
- Integrate WotC standards into agents 2)
- Add WotC enhanced NPC standards and validators
- Add JSON→MD conversion script for PDF inclusion
- Add campaign brief_description support
- Add La Hoja de Vlad campaign example
- Template & Validation Enforcement v2.5.0
- WotC Quality Fixes v2.6.0

### Fix

- Add explicit delegation strategy to grimorio-architect
- Add CRITICAL instruction to READ TEMPLATE FIRST
- Add CRITICAL template instruction to all agents
## [v2.3.0] - 2026-05-08

### Chore

- Remove stale campaign data from repo

### Feat

- Add version badge to README and dynamic version in install output
- Narrative quality improvements (decision trees, faction reputation, world state tracking)
- Add explicit Check 12 validation for chapter narrative structure (architect)
- WotC Format Improvements - Complete (v2.3.0)

### Fix

- Guard cp commands in setup_plugin to avoid errors when directories don't exist
- Remove duplicate v prefix in version display
- Reexec from cloned repo to bypass GitHub raw cache 2)
- Sync agents with actual repo state (remove grimorio-acts, add areas/introduction/setting-guide/appendices/integrator)
- Remove grimorio_ prefix from MCP tool names (agents)
- Install.sh copies templates to plugin directory

### Refactor

- Surgical plugin install - only touch Grimorio files, preserve user customizations
- Rename save_act→save_areas, add 4 agents, update install.sh
## [v2.1.0] - 2026-05-08

### Docs

- Update README and CHANGELOG for WotC professional format

### Feat

- Add Introduction, Setting Guide, and Appendices (professional-wotc-format)
- Add sidebar CSS styling for DM tips and notes (compiler)

### Fix

- Resolve all golangci-lint errors
- Resolve remaining golangci-lint errors
- Resolve final batch of golangci-lint errors
- Resolve last golangci-lint error in canon_test.go
- Resolve remaining golangci-lint errors in canon_service_test.go
- Resolve critical bugs in MCP server (mcp)
- Use grimorio-areas instead of non-existent grimorio-acts (architect)
- Remove duplicate Session Zero heading (compiler)

### Test

- Add test cases for new WotC templates (compiler)
- Add WotC professional format E2E test (e2e)
- Update TestGenerateHTML_WithNewSections to use proper WotC area format (compiler)
## [v2.0.0] - 2026-05-08

### Chore

- Add grimorio-sdd-roadmap.md to .gitignore
- Ignore build artifacts and sdd directory (gitignore)

### Docs

- Add migration script and update README for v2.0 coherence tools
- Add process_consistency_gate to narrative coherence docs (readme)
- Update CHANGELOG, ROADMAP, and install.sh for v2.0.0
- Add mandatory development rules for artifact updates (sdd)
- Update roadmap, agents, README, CHANGELOG for Fase 3 completion
- Mark Fase 5 complete, update CHANGELOG for v2.0.0 release
- Add complete user guide to README, update agents with Phase 3-4 MCP tools
- Update CHANGELOG with v2.0 area-based generation

### Feat

- Add narrative coherence subsystem (canon)
- Add narrative coherence tools and server wiring (mcp)
- Implement Fase 2 narrative coherence gates (consistency-gate)
- Update all markdown templates for v2 narrative coherence (templates)
- Update dnd-5e-srd skill with official adventure patterns (skill)
- Add grimorio-narrative-custodian agent for coherence (agents)
- Implement Fase 3 - factions, consequences, random tables, handouts (living-world)
- Implement Fase 4 - session prep, flowcharts, roster, hooks (dm-experience)
- Implement Fase 5 - caching, benchmarks, CI/CD, docs, release (polish)
- Add clean_installation() to wipe previous install before reinstall (install)
- Agent v2 formats + compiler version flag (f1-foundation)
- Area validator v2 + grimorio-areas agent (f2-areas)
- Cross-reference validator + XP budget + treasure checks (f3-integration)
- Handout generator + compiler integration (f4-visuals)
- Compiler v2 + templates + CSS + migrator (f5-compilation)

### Fix

- Resolve golangci-lint errors and update agents for v2 (lint)
- Update command template to Phase 3-13 with Living World + DM Experience (install)
- Add grimorio_mcp to all subagents, install.sh configures narrative-custodian (agents)
- Add missing parameters to consistency gate and related tools (mcp)
- Arregla 3 bugs críticos en herramientas MCP (grimorio)

### Test

- Add missing state error scenario test (session-prep)
## [v0.1.2] - 2026-05-07

### Fix

- Switch release to manual trigger (workflow_dispatch) (ci)
## [v0.1.1] - 2026-05-07

### Fix

- Commit changelog changes before goreleaser (ci)
## [v0.1.0] - 2026-05-07

### Chore

- Update ASCII art in install.sh
- Remove grimorio-orchestrator references from install.sh and README
- Remove last orchestrator mention from print_instructions (install)

### Ci

- Update Go version from 1.21 to 1.25 in workflow
- Downgrade Go to 1.24 and fix all golangci-lint issues

### Docs

- Update workflow to reflect orchestrator pattern (readme)
- List all 8 content subagents in post-install instructions (install)

### Feat

- Initial release - D&D one-shot and campaign generator
- Added grimorio command to generate all easy (add-command)
- Added svg and dalle api vinculation (images)
- Added svg and dalle api vinculation (images)
- Added svg and dalle api vinculation (images)
- Added svg and dalle api vinculation (images)
- Pdf issues (fix)
- Add code block support with styling for ASCII maps (pdf)
- [breaking] Free AI image generation via Pollinations.ai + improved SVG maps
- Reorder generation workflow - acts generated last
- Images generated in parallel from the start, optional
- Emphasize acts are generated last before PDF
- Cover page now fills entire page
- Add grimorio-orchestrator to eliminate main thread polling (orchestrator)
- Add generate_images_batch tool for parallel AI image generation (mcp)
- [breaking] Implement clean architecture with TDD for Phase 0 (architecture)
- Add 5 missing content tools and fix image/PDF generation (mcp)
- Add user-visible progress reporting after each phase (orchestrator)
- Add Raphael AI fallback and sequential batch generation (image)
- Force sequential generation and remove batch tool (image)
- Enforce ALL image types and verify references before PDF (architect)
- Restructure PDF compilation order and redesign act template (compiler)
- Add optional markdown linking with inline image references (generate_image)
- Add optional markdown linking for SVG assets (generate_map/divider)
- Add semantic versioning with GoReleaser and automatic changelog (ci)

### Fix

- Correct repo URL to pauvalls/Grimorio
- Update install URLs in README and add .gitignore
- Install plugin in OpenCode and fix color codes (install)
- Always update opencode config, add visual generation to /grimorio (install)
- 1-page cover, link maps to scenes with zone descriptions (pdf)
- Resolve ../assets paths, fix cover to 1 page without flex (pdf)
- Cover wrapper, blockquote read-aloud, bare asset detection (pdf)
- Simplify cover, handle backtick asset refs (pdf)
- Cover 1 page, no blank pages, readable ASCII maps, SVG in acts (pdf)
- Cover fills page, code blocks span both columns (pdf)
- Delegate image generation to grimorio-cartographer subagent
- SVG to PNG conversion + cover art on first page + remove gallery
- Improve SVG text readability and cover page layout
- SVG rendering in PDF + install.sh syntax error
- Prevent duplicate headings + cover art now mandatory
- Force cartographer subagent + acts last + delegate tool
- Update commands/grimorio.md with new template
- Install.sh always updates commands/grimorio.md
- Remove stale grimorio template files from old locations (install)
- Remove subagent orchestration from architect agent (architect)
- Deduplicate images and ignore horizontal rules (compiler)
- Register grimorio-orchestrator agent in opencode.json (install)
- Make cartographer run AFTER content generation (orchestrator)
- Add explicit file-reading workflow (cartographer)
- Make image linking explicit and mandatory (cartographer)
- Generate ALL images, not minimums (cartographer)
- Complete image generation pipeline with parallel batch + fallback (images)
- GenerateMap and GenerateDivider now write SVG files to disk instead of discarding content (assets)
- Remove deprecated grimorio-orchestrator from opencode.json on re-install (install)
- Add HTTP timeouts to image providers and create content-specific subagents (image)
- Reorder generation phases in correct dependency order (architect)
- Update command template to reflect 3-batch ordering (template)
- Move lore to Batch 2 (architect)
- Search cover-*.png, add bold spacing, and enrich scene structure (compiler)
- Replace blocking Lock with TryLock and fix cover image glob
- Preserve raw HTML img tags and add post-PDF image verification with retry (compiler)
- Add narrative quest types and fix create_personal_quest (quests)
- Make configure_shell idempotent, prevent duplicate PATH entries (install)
- Prevent duplicate PATH entries on reinstall via marked blocks (install)

### Refactor

- Grimorio-architect ahora orquesta directamente todo el flujo de generación (agents)

### Test

- Add image embedding tests for PNG, SVG, missing, dedup, and code asset refs (compiler)
- Add comprehensive test suite and fix formatting
