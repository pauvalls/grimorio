# Skill Registry

**Delegator use only.** Any agent that launches sub-agents reads this registry to resolve compact rules, then injects them directly into sub-agent prompts. Sub-agents do NOT read this registry or individual SKILL.md files.

See `_shared/skill-resolver.md` for the full resolution protocol.

## User Skills

| Trigger | Skill | Path |
|---------|-------|------|
| When creating a pull request, opening a PR, or preparing changes for review. | branch-pr | /home/pau/.config/opencode/skills/branch-pr/SKILL.md |
| when a PR would exceed 400 changed lines, when planning chained PRs, stacked PRs, or reviewable slices. | chained-pr | /home/pau/.config/opencode/skills/chained-pr/SKILL.md |
| when writing guides, READMEs, RFCs, onboarding docs, architecture docs, or review-facing documentation. | cognitive-doc-design | /home/pau/.config/opencode/skills/cognitive-doc-design/SKILL.md |
| when drafting or posting feedback, review comments, maintainer replies, Slack messages, or GitHub comments. | comment-writer | /home/pau/.config/opencode/skills/comment-writer/SKILL.md |
| When writing Go tests, using teatest, or adding test coverage. | go-testing | /home/pau/.config/opencode/skills/go-testing/SKILL.md |
| When creating a GitHub issue, reporting a bug, or requesting a feature. | issue-creation | /home/pau/.config/opencode/skills/issue-creation/SKILL.md |
| When user says "judgment day", "judgment-day", "review adversarial", "dual review", "doble review", "juzgar", "que lo juzguen". | judgment-day | /home/pau/.config/opencode/skills/judgment-day/SKILL.md |
| When user asks to create a new skill, add agent instructions, or document patterns for AI. | skill-creator | /home/pau/.config/opencode/skills/skill-creator/SKILL.md |
| when implementing a change, preparing commits, splitting PRs, or planning chained or stacked PRs. | work-unit-commits | /home/pau/.config/opencode/skills/work-unit-commits/SKILL.md |
| When the orchestrator launches you to think through a feature, investigate the codebase, or clarify requirements. | sdd-explore | /home/pau/.config/opencode/skills/sdd-explore/SKILL.md |
| When the orchestrator launches you to create or update a proposal for a change. | sdd-propose | /home/pau/.config/opencode/skills/sdd-propose/SKILL.md |
| When the orchestrator launches you to write or update specs for a change. | sdd-spec | /home/pau/.config/opencode/skills/sdd-spec/SKILL.md |
| When the orchestrator launches you to write or update the technical design for a change. | sdd-design | /home/pau/.config/opencode/skills/sdd-design/SKILL.md |
| When the orchestrator launches you to create or update the task breakdown for a change. | sdd-tasks | /home/pau/.config/opencode/skills/sdd-tasks/SKILL.md |
| When the orchestrator launches you to implement one or more tasks from a change. | sdd-apply | /home/pau/.config/opencode/skills/sdd-apply/SKILL.md |
| When the orchestrator launches you to verify a completed (or partially completed) change. | sdd-verify | /home/pau/.config/opencode/skills/sdd-verify/SKILL.md |
| When the orchestrator launches you to archive a change after implementation and verification. | sdd-archive | /home/pau/.config/opencode/skills/sdd-archive/SKILL.md |
| When the orchestrator launches you to onboard a user through the full SDD cycle. | sdd-onboard | /home/pau/.config/opencode/skills/sdd-onboard/SKILL.md |

## Project Skills

| Trigger | Skill | Path |
|---------|-------|------|
| D&D 5e campaign design, encounter balance, narrative coherence, SRD rules | dnd-5e-srd | ~/.config/opencode/skills/dnd-5e-srd/SKILL.md |
| Generate D&D 5e campaigns end-to-end via delegate pattern, WotC standards | grimorio-architect | ~/.config/opencode/skills/grimorio-architect/SKILL.md |
| Consolidate campaign reference material — magic items, stat blocks, handouts, maps, tables | grimorio-appendices | ~/.config/opencode/skills/grimorio-appendices/SKILL.md |
| Generate numbered playable areas (10-15 per act) with WotC format validation | grimorio-areas | ~/.config/opencode/skills/grimorio-areas/SKILL.md |
| Prepare AI image specifications and update markdown references for artwork | grimorio-artist | ~/.config/opencode/skills/grimorio-artist/SKILL.md |
| Generate monsters, creatures, and stat blocks with D&D 5e mechanics | grimorio-bestiary | ~/.config/opencode/skills/grimorio-bestiary/SKILL.md |
| Generate ALL SVG visual assets — battle maps, dividers, flowcharts | grimorio-cartographer | ~/.config/opencode/skills/grimorio-cartographer/SKILL.md |
| Generate pre-generated player characters with balanced builds and narrative hooks | grimorio-characters | ~/.config/opencode/skills/grimorio-characters/SKILL.md |
| Generate combat encounters, social challenges, and exploration scenes | grimorio-encounters | ~/.config/opencode/skills/grimorio-encounters/SKILL.md |
| Integrate, cross-reference, and polish campaign before PDF compilation | grimorio-integrator | ~/.config/opencode/skills/grimorio-integrator/SKILL.md |
| Generate campaign introduction/overview document | grimorio-introduction | ~/.config/opencode/skills/grimorio-introduction/SKILL.md |
| Generate world lore, backstory, history, and setting with narrative depth | grimorio-lore | ~/.config/opencode/skills/grimorio-lore/SKILL.md |
| Generate location descriptions, scene layouts, and zone breakdowns | grimorio-maps | ~/.config/opencode/skills/grimorio-maps/SKILL.md |
| Validate campaign content for narrative coherence, canon consistency | grimorio-narrative-custodian | ~/.config/opencode/skills/grimorio-narrative-custodian/SKILL.md |
| Generate NPCs, factions, and social relationships with WotC-enhanced descriptions | grimorio-npc | ~/.config/opencode/skills/grimorio-npc/SKILL.md |
| Generate personal quests, side missions, and character-specific narrative hooks | grimorio-quests | ~/.config/opencode/skills/grimorio-quests/SKILL.md |
| Generate DM-only campaign setting guide with spoilers, geography, factions | grimorio-setting-guide | ~/.config/opencode/skills/grimorio-setting-guide/SKILL.md |

## Compact Rules

Pre-digested rules per skill. Delegators copy matching blocks into sub-agent prompts as `## Project Standards (auto-resolved)`.

### branch-pr
- Every PR MUST link an approved issue
- Every PR MUST have exactly one `type:*` label
- Branch names: `^(feat|fix|chore|docs|style|refactor|perf|test|build|ci|revert)\/[a-z0-9._-]+$`
- Automated checks must pass before merge

### chained-pr
- Split PRs exceeding 400 changed lines (additions + deletions)
- Each PR ≤60-minute review target
- Every chained PR must state: starts at, ends at, what came before, what comes next
- Each PR must be autonomous: CI green, one deliverable, reasonable rollback
- Child PR diff must contain ONLY the current work unit

### cognitive-doc-design
- Lead with the answer; context comes after
- Progressive disclosure: happy path first, then details and edge cases
- Use chunking: small sections, short flat lists
- Signpost with headings, labels, callouts
- Prefer tables, checklists, examples over prose

### comment-writer
- Start with actionable point; no recap preamble
- Warm, direct voice like a thoughtful teammate
- Keep to 1-3 short paragraphs or tight bullets
- Explain the technical why when requesting changes
- Match thread language; in Spanish use Rioplatense voseo

### go-testing
- Use table-driven tests as standard pattern
- Use `t.Parallel()` when tests are independent
- Prefer `cmp.Diff` or `reflect.DeepEqual` for complex structs
- Use golden files for large expected outputs
- Mock external dependencies; never hit real APIs in unit tests

### issue-creation
- Blank issues are disabled — MUST use template
- Every issue gets `status:needs-review` automatically
- Maintainer MUST add `status:approved` before any PR
- Questions go to Discussions, not issues

### judgment-day
- Launch TWO independent blind judge sub-agents via `delegate` (parallel, never sequential)
- Orchestrator synthesizes verdicts: Confirmed (both) = high confidence; Single = triage; Conflict = discuss
- Apply fixes, then re-judge until both pass or escalate after 2 iterations
- Inject project standards into judge prompts

### skill-creator
- Skills need: SKILL.md, optional assets/, optional references/
- Frontmatter required: name, description with trigger, license, metadata
- Create skills for repeated patterns, not one-offs
- Keep compact rules 5-15 lines; no fluff

### work-unit-commits
- Commit by deliverable work unit, NOT by file type
- Tests and docs stay in the same commit as the code they verify
- Each commit must be a candidate chained PR when change grows
- Commit message explains outcome, not file list

### dnd-5e-srd
- Every official campaign needs: DM-only background, storyline summary, geopolitical context, PC hooks
- Chapters adopt dominant game mode; transitions deliver narrative assets
- Locations broken into numbered areas with read-aloud text, creatures, treasure, connections, secrets
- NPC systems: companions, faction quest chains, political factions
- Random tables are procedural content generators, not filler
- Every new MCP tool MUST update: relevant agents, README diagrams, MCP tools table, install.sh, this skill if new mechanics

### grimorio-architect
- ALWAYS use `delegate` for content generation — NEVER generate creative content inline
- Report progress to user after EVERY phase
- Execute phases SEQUENTIALLY — each phase waits for previous to complete
- Validate via consistency gate before proceeding to next phase
- Read templates before generating: areas.md.tmpl, npc.md.tmpl, monster.md.tmpl, etc.
- Run `./scripts/validate-campaign.sh --check=all` before PDF compilation (BLOCKING GATE)
- WotC standards: boxed text 100-600 words, ≥2 hooks/area, ≥3 developments with recovery, running guidance 150-400 words, ≥1 sidebar/act

### grimorio-areas
- Read template: `internal/compiler/templates/areas.md.tmpl` BEFORE generating
- 10-15 numbered areas per act
- Boxed text: 100-600 words per area (grep: `^>>`), second person present
- Character hooks: ≥2 per area, target specific backgrounds/classes
- Developments: ≥3 per area with 100% recovery paths (grep: `if.*fail|si.*fallan`)
- Running guidance: 150-400 words, 5 subsections (Preparación, Ritmo, Señales, Improvisar, Ceñirse al Guión)
- Sidebars: ≥1 per act (grep: `^> #####`)
- Cross-references: Use EXACT names from bestiary.md and npcs_and_factions.md
- Decision points table with propagation (Affects: Área X, Acto N)

### grimorio-npc
- Read template: `internal/compiler/templates/npc.md.tmpl` BEFORE generating
- 500-800 words per major NPC
- 6 required sections: Apariencia, Personalidad, Motivación, Secreto, Involucramiento en Quests, Conexiones
- Faction reputation system: Score -100 to +100 with tier benefits table
- Faction propagation: List allies (+X reputation) and enemies (-X reputation)
- Cross-references: Quest names match quests/*.md, faction names consistent, locations match maps.md
- Cita típica for each major NPC

### grimorio-narrative-custodian
- ALWAYS run `./scripts/validate-campaign.sh {campaign_path} --check=all`
- Structure check: directories (acts, npcs, bestiary, encounters, maps, assets), files (lore.md, etc.)
- WotC format: boxed text 100-600 words, ≥2 hooks/area, ≥3 developments with recovery, running guidance 150-400 words, ≥1 sidebar/act
- Cross-references: 100% of creature/NPC references must exist in source files
- Narrative consistency via MCP: validate_canon, check_consistency, evaluate_consequences
- Auto-retry logic: Up to 2 retries with fixes, then report failure and STOP

### grimorio-appendices
- Read template before generating content
- Use MCP `save_appendices` tool — NEVER use Write for creative content
- Consolidate: magic items, stat blocks, handouts, maps, random tables
- Validate against canon before saving

### grimorio-artist
- Generate AI image specs for: NPC portraits, monster illustrations, scene artwork, campaign covers
- ALWAYS sequential generation with 3s delay between images (rate limit)
- Use MCP `generate_image` tool (cover/portrait/illustration/scene types)
- Use MCP `generate_divider` for decorative SVGs
- Update markdown references after generation

### grimorio-bestiary
- Read template: `internal/compiler/templates/monster.md.tmpl` BEFORE generating
- Use MCP `save_bestiary` tool — NEVER use Write for creative content
- Stat blocks must follow D&D 5e SRD format
- Include tactical notes for DM usage
- Validate cross-references from encounters and areas

### grimorio-cartographer
- Generate ALL SVG assets locally (no API): battle maps, dividers, flowcharts
- Use MCP `generate_map` (dungeon/landscape/city styles)
- Use MCP `generate_divider` (ornate/simple/double styles)
- Use MCP `generate_flowchart` for campaign structure (Mermaid + SVG)
- Update markdown references in relevant files

### grimorio-characters
- Read template before generating character sheets
- Use MCP `generate_character`, `get_character`, `list_characters`
- Pre-generated characters must have balanced builds
- Include narrative hooks tied to campaign
- Validate against canon before saving

### grimorio-encounters
- Read template: `internal/compiler/templates/encounter.md.tmpl` BEFORE generating
- Use MCP `save_encounters` tool — NEVER use Write for creative content
- Encounters must have balanced difficulty (CR calculations)
- Include combat, social, and exploration challenges
- Validate creature references against bestiary

### grimorio-integrator
- Integration phase BEFORE PDF compilation
- NEVER generate new creative content — only verify, correct, standardize, integrate
- Cross-reference all acts, NPCs, bestiary, encounters, maps, lore
- Auto-fix inconsistencies when possible
- Generate handouts (summary/encounter/quest/lore/faction)
- Run `check_consistency` and `process_consistency_gate` before finalizing

### grimorio-introduction
- Read template before generating
- Use MCP `save_introduction` tool — NEVER use Write for creative content
- Set campaign expectations and hook the DM
- Include: tone, level range, playtime, themes, content warnings

### grimorio-lore
- Read template before generating
- Use MCP `save_lore` tool — NEVER use Write for creative content
- World lore must have narrative depth and internal consistency
- Timeline of events, geopolitical context, major factions
- Validate cross-references with areas, NPCs, quests

### grimorio-maps
- Read template before generating location descriptions
- Use MCP `save_maps` for descriptions, `generate_map` for battle maps
- Location descriptions need spatial detail and zone breakdowns
- Scene layouts for key encounters
- Validate location names across all campaign files

### grimorio-quests
- Read template before generating
- Use MCP `create_personal_quest`, `list_quests`, `update_quest_status`
- Personal quests tied to character backgrounds
- Side missions with clear objectives and rewards
- Quest status tracking: active, completed, failed, on_hold
- Validate cross-references with NPCs and factions

### grimorio-setting-guide
- Read template before generating
- Use MCP `save_setting_guide` tool — NEVER use Write for creative content
- DM-only content with spoilers
- Geography, history, factions, major NPCs
- Running the campaign guidance

## Project Conventions

No project convention files found (agents.md, AGENTS.md, CLAUDE.md, .cursorrules, GEMINI.md, copilot-instructions.md).
