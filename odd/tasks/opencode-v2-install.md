# Feature: OpenCode V2 install & update

## Objective

Make Grimorio's install and self-update flow correct for OpenCode V2, and make the
documentation match the real behavior.

## Problem

The installed Grimorio was completely disconnected from OpenCode V2. Root causes are all
in this repo:

1. `agents/*.md` and `skills/*/SKILL.md` still use V1 shapes: `permission`/`tools` objects
   and calls to a `delegate(agent=..., prompt=...)` tool that does not exist in V2
   (the tool is `subagent`).
2. `install.sh` wrote the V1 config shape (`mcp.grimorio`, `command.grimorio`) and installed
   skills into `~/.config/opencode/plugins/grimorio/skills`, which V2 does not scan.
   Its awk fallback appended a second `"mcp"` key and produced invalid JSON.
3. `cmd/grimorio/commands/update/assets.go` had the same V1 bugs, plus machine-specific
   absolute paths hardcoded in Go source (`/home/pau/...`), which breaks every other machine.
4. `README.md` documented the old plugin layout and did not mention OpenCode V2.

## Why

A user on OpenCode V2.2.6 ended up with a working binary and zero OpenCode integration:
no MCP entry, no skills, no agents, no command. Running `grimorio update` would have
re-introduced the broken V1 layout.

## Scope

- Repo asset templates: `agents/*.md`, `skills/*/SKILL.md`.
- Go updater: `cmd/grimorio/commands/update/assets.go` (+ tests).
- Installer: `install.sh` (+ `scripts/install_test.go`).
- Docs: `README.md` (EN + ES sections).

Out of scope: campaign generation logic, compiler, MCP tool implementations, `install.ps1`.

## Tasks

- [x] T1 — Convert repo agent/skill templates to OpenCode V2 (`permissions` list, drop
      `name`/`version`, `delegate(` → `subagent(`). Commit `69405f5`.
- [x] T2 — Updater targets: skills and agents installed into the V2 global dirs
      (`~/.config/opencode/skills`, `~/.config/opencode/agents`), agents filtered to `*.md`.
- [x] T3 — Updater config writer: writes `mcp.servers.grimorio` with `disabled`/`environment`,
      removes legacy `mcp.grimorio` and `command.grimorio`, derives `environment` from the real
      environment (no hardcoded user paths), writes `~/.config/opencode/commands/grimorio.md`.
      T2+T3 commit `33800f3`.
- [x] T4 — `install.sh`: same V2 merge, idempotent and non-destructive, no invalid-JSON
      fallback; installs skills/agents/command in V2 paths. Commit `f4aafe4`.
- [x] T5 — README: OpenCode V2 install, update semantics, troubleshooting. Commit `f610a69`.
- [x] T6 — Tests: V2 shape assertions, legacy migration, idempotency, custom environment
      preservation, asset targets and Markdown filtering. Commit `935722b`.

## Verification evidence

- `go test -count=1 -p 1 -parallel 1 ./...` → all packages pass (single worker, no `-race`).
- `sh -n install.sh` → POSIX syntax OK.
- Installer merge exercised against a scratch `HOME`: legacy `mcp.grimorio` and
  `command.grimorio` removed, unrelated `command.other` and `mcp.servers.codegraph` preserved,
  second run reports `opencode.json already up to date`, custom `PIPER_*` values survive.
- Rebuilt binary installed at `~/.grimorio/grimorio` (backup:
  `grimorio.backup.pre-v2-fix`), then the real `grimorio update commands` run on the live
  config: V2 shape kept, `environment` preserved, no legacy keys reappeared,
  `/grimorio` command file left byte-identical (mtime unchanged → write-if-changed works).
- `opencode mcp list` → `✓ grimorio connected` alongside agendo/codegraph/context7/engram.
- Skills: 19 installed (17 `grimorio-*` + `dnd-5e-srd` + `monster-design-rules`).
  Agents: 18 `grimorio-*.md`.

## Follow-ups (not done)

- `install.ps1` (Windows) still writes the V1 `mcp.grimorio` shape; it needs the same
  migration and cannot be validated on this machine.
- Releasing this branch is what propagates the fix to every other install; until a release
  includes it, `grimorio update` on another machine still writes the V1 shape.

## Delivery

`ask-on-risk`; five work-unit commits on `feat/opencode-v2-install`.

Actual size: 1,116 insertions / 697 deletions across 25 files. Most of it is mechanical
churn (18 agent files, 2 test files, the shared installer/updater pair). That is above the
~400-line delivery budget, so the branch should go out as a chain rather than one PR:
1. templates (`agents/`, `skills/`),
2. updater + installer,
3. docs + tests.
