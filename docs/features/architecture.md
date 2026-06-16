# Architecture

Grimorio is a Go 1.25+ project that follows Clean / Hexagonal /
Screaming Architecture principles. This document gives a high-level tour;
for deep dives see the linked docs.

## Layers

```
┌────────────────────────────────────────────────────────────────────┐
│ cmd/grimorio/             CLI entry point (urfave/cli/v2)          │
│   └── commands/              subcommands: mcp, validate, update,   │
│                              campaign                                │
├────────────────────────────────────────────────────────────────────┤
│ internal/mcp/             MCP server over stdio (mark3labs/mcp-go) │
│   └── handlers/             one file per tool family                │
├────────────────────────────────────────────────────────────────────┤
│ internal/services/        Business logic                             │
│   validation_engine.go     17 consistency rules                       │
│   campaign_service.go      campaign CRUD + batches                   │
│   dm_context_service.go    DM session payload aggregation           │
│   narrative_state_service.go session log + checkpoints              │
│   consistency_gate.go      pre-merge gate                            │
│   asset_service.go         image / map / handout generation         │
├────────────────────────────────────────────────────────────────────┤
│ internal/domain/          Pure types, no I/O                        │
│   campaign.go / canon.go / validation.go / narrative_state.go …     │
├────────────────────────────────────────────────────────────────────┤
│ internal/repository/      Storage interfaces + memory & filesystem  │
│   *.go                       adapters                                │
├────────────────────────────────────────────────────────────────────┤
│ internal/compiler/        Markdown → HTML → PDF pipeline            │
│ internal/validators/      WotC format validators (pure)             │
│ internal/cache/           LRU cache                                 │
│ internal/tts/piper/       Piper TTS client                           │
│ internal/svg/             Procedural SVG generators                  │
│ internal/image/           Pollinations.ai client + cache            │
│ internal/namegen/         Syllable-based name generator              │
│ internal/generators/      Bilingual content generators               │
└────────────────────────────────────────────────────────────────────┘
```

## Dependency Rules

1. `internal/domain` imports nothing from the project.
2. `internal/repository` depends only on `domain`.
3. `internal/services` depends on `domain` and `repository`.
4. `internal/mcp` depends on `services` and `domain`.
5. `cmd/grimorio` depends on `commands`, `mcp`, `services`.

There are no upward arrows. A linter check in CI rejects new
`services → mcp` or `domain → services` imports.

## CLI vs MCP

Both surfaces use the same services. The CLI (urfave/cli/v2) is a thin
shim:

- `grimorio validate <name>` → `services.NewValidationEngine().CheckConsistency()`
- `grimorio update all` → `update.NewUpdateCommand(version)` pipeline
- `grimorio mcp` (default) → `internal/mcp.NewServer(cfg)`

The MCP tools are documented in [MCP Tools](mcp-tools.md).

## Key Abstractions

### `ValidationEngine`
The single source of truth for campaign consistency. 17 rules across
4 rule families: lore integrity, narrative-state parity, faction
reputation, and WotC format.

### `ConsistencyReport`
`json:"..."`-tagged report that doubles as the MCP `check_consistency`
response and the `grimorio validate --json` output. This is the
contract that links the two surfaces.

### `NarrativeState`
Append-only JSON state: `SessionLog`, `ActiveQuests`, `CompletedQuests`,
`DeadNPCs`, `KeyItems`, `RevealedClues`. Updated via
`update_narrative_state` after each session.

### `CanonDocument`
The campaign's source-of-truth: `Facts`, `Entities`, `Rules`,
`Timeline`, `Relationships`. Authored via `save_lore`, `save_npcs`, etc.

### `ImageCache` (v5.0)
Hash-based image generation cache. SHA-256 key derived from
prompt + model + dimensions + provider. LRU(50) in memory + sharded
on-disk dedup at `~/.cache/grimorio/images/v1/<hash[:2]>/<hash>.bin`.
Integrated as a `CachingProvider` at the head of the provider chain.
Bypassed via the `force_regenerate` MCP flag.

### `Exporter` Strategy (v5.0)
`internal/compiler` uses a strategy pattern for output formats.
`Exporter` interface with three implementations:
- `PDFExporter` — Markdown → HTML → PDF via Chromium/Chrome headless
- `MarkdownExporter` — Concatenate canonical files with `---` separators
- `EPUBExporter` — Markdown → XHTML → OPF + NCX → ZIP with `.epub` extension

The `export_campaign` MCP tool and `compile_pdf` CLI both dispatch
through the same exporter registry.

### `CampaignHealthService` (v5.0)
Aggregates `ValidationEngine`, `CanonService`, `FactionService`, and
`NarrativeState` into a 0-100 health score across six axes. Exposed
via the `campaign_health_dashboard` MCP tool. The scoring algorithm
is documented in `internal/services/campaign_health_score.go`.

### `TreasureService` (v5.0)
SRD-compliant treasure generation. Two entry points:
`GenerateIndividualTreasure(cr)` and `GenerateTreasureHoard(tier)`.
4 CR tiers, magic items by rarity (Common → Legendary). Returns
`domain.TreasureHoard` with coins, art objects, gems, and magic items.

### `CampaignTemplate` (v5.0)
Pre-defined campaign archetypes. 5 presets:
Urban Fantasy, Gothic Horror, Maritime Adventure, Dungeon Crawl,
Political Intrigue. Applied via the `template` field in
`create_campaign` MCP tool or the `template` query param in
`generate_adventure_bible`. Falls back to template defaults when
the user leaves tone/themes empty.

## Storage

- Campaigns live under `~/campaigns/<name>/` by default
  (override with `CAMPAIGN_ROOT`).
- Per-campaign files: `canon.json`, `narrative_state.json`,
  `areas/`, `npcs/`, `bestiary/`, `assets/`, `session-prep-*.md`.
- The `examples/` directory contains a fully-fleshed reference
  campaign (`la-hoja-de-vlad/`) used by the WotC consistency checks
  and the [walkthroughs](../walkthroughs/la-hoja-de-vlad.md).

## Where to Read More

- [Developer Guide](../developer-guide.md) — how to add a service, test
  patterns, TDD discipline.
- [Campaign Consistency](../campaign-consistency.md) — gate pipeline,
  checkpoints, rollback, dynamic content.
- [PDF Compiler](pdf-compiler.md) — HTML/PDF pipeline internals.
- [MCP Tools](mcp-tools.md) — full tool catalogue.

## Internationalization (v5.0)

Grimorio is **English-first by default**, with Spanish as a user-selectable
alternative at session start:

- All skills, agents, and templates in `skills/`, `agents/`, and
  `internal/compiler/templates/` are written in English.
- The `examples/la-hoja-de-vlad/` reference campaign stays in Spanish
  on purpose — it's the WotC regression baseline.
- The `grimorio-architect` agent prompts the user for language
  preference at the start of every campaign and passes the choice
  to all sub-agents via a `LANG:` preamble on each `delegate` call.
- The `grimorio-dm` agent has the same language intake pattern (Section 10).
- The `force_regenerate` MCP flag is documented in English with
  parameter description in English.

To add a third language, see the "Adding a language" section in the
[Developer Guide](../developer-guide.md#internationalization).
