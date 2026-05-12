# Tasks: Consolidate Skills + Agents Architecture

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~5450 (+450 / -5000) |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1: install.sh; PR 2-3: skill merges (8+8); PR 4: cleanup/registry/config; PR 5: verification |
| Delivery strategy | ask-on-risk |
| Chain strategy | pending |

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: pending
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Install script foundation | PR 1 | copy_skills(), verify_skills(), main() wiring, plugin cleanup |
| 2 | Content merges batch A | PR 2 | 8 skills: architect, areas, npc, bestiary, encounters, lore, maps, quests |
| 3 | Content merges batch B | PR 3 | 8 skills: characters, introduction, setting-guide, appendices, artist, cartographer, integrator, narrative-custodian |
| 4 | Cleanup and config | PR 4 | Delete agents, update registry, simplify opencode.json, /grimorio command |
| 5 | Verification and commit | PR 5 | Full reinstall test, git commit/push |

## Phase 1: Install Script Foundation

- [ ] 1.1 Add `copy_skills()` to `scripts/install.sh` after `verify_agents()`: copy `grimorio-*/` and `dnd-5e-srd/` to `~/.config/opencode/skills/`, report line counts, return error if zero copied.
- [ ] 1.2 Enhance `verify_skills()` in `scripts/install.sh`: add `expected_skills` array with all 17 skills, loop to verify each exists, report missing, return non-zero if any missing.
- [ ] 1.3 Update `main()` in `scripts/install.sh`: add `copy_skills()` call before `verify_skills()` in `reinstall` and `full` modes.
- [ ] 1.4 Remove plugin `agents/` and `skills/` subdirectory creation from `setup_plugin()` in `scripts/install.sh`; keep binary, templates, `.mcp.json`.

## Phase 2: Content Merge

- [ ] 2.1 Merge batch A (8 skills): Compare `agents/*.md` vs `skills/*/SKILL.md` for architect, areas, npc, bestiary, encounters, lore, maps, quests. Append unique workflow content to SKILL.md, verify no detail lost.
- [ ] 2.2 Merge batch B (8 skills): Compare `agents/*.md` vs `skills/*/SKILL.md` for characters, introduction, setting-guide, appendices, artist, cartographer, integrator, narrative-custodian. Append unique workflow content to SKILL.md, verify no detail lost.

## Phase 3: Cleanup and Configuration

- [ ] 3.1 Delete all `agents/grimorio-*.md` files (16 files) via `git rm`; remove `agents/` directory if empty.
- [ ] 3.2 Update `.atl/skill-registry.md`: change all 17 skill paths to `~/.config/opencode/skills/{name}/SKILL.md` format.
- [ ] 3.3 Update `~/.config/opencode/opencode.json`: replace 16 grimorio-* agent prompts with one-liner skill loaders; keep tools and mode.
- [ ] 3.4 Update `/grimorio` command template in `~/.config/opencode/opencode.json` to load `grimorio-architect` skill.

## Phase 4: Testing and Verification

- [ ] 4.1 Run `./scripts/install.sh --reinstall`; verify 17 skills exist in `~/.config/opencode/skills/`, verify registry paths, verify agent prompts, test `/grimorio` command.
- [ ] 4.2 Git commit all changes: `scripts/install.sh`, `skills/*/SKILL.md`, `.atl/skill-registry.md`, deleted `agents/*.md`. Message: "refactor: consolidate skills+agents architecture, fix install script". Push.
