# 📜 Grimorio

**AI-powered D&D 5e campaign and one-shot generator for OpenCode.**

Turn one idea into a print-ready adventure book: lore, NPCs, bestiary, encounters, maps, cover art
and a WotC-styled PDF. It runs as an MCP server, so the same tools work from any agent that can call
MCP.

**[English](README.md) · [Español](README.es.md)**

---

## Install

**Linux / macOS**

```bash
curl -sSL https://raw.githubusercontent.com/pauvalls/grimorio/main/install.sh | sh
```

**Windows (PowerShell)**

```powershell
irm https://raw.githubusercontent.com/pauvalls/grimorio/main/install.ps1 | iex
```

The installer drops the binary in `~/.grimorio/grimorio` and symlinks it to `~/.local/bin/grimorio`.
On Linux and macOS it also configures OpenCode: MCP server, skills, agents and the `/grimorio` command.

Needs `python3` to merge `opencode.json`. Without it the installer prints the exact entry to paste
instead of writing a broken config.

## Verify

```bash
grimorio --version        # grimorio version vX.Y.Z
opencode mcp list         # ✓ grimorio  connected
```

Inside OpenCode you should also see the `grimorio-architect` and `grimorio-dm` agents, the
`grimorio-*` skills, and the `/grimorio` command.

## Use

**Generate a campaign** — ask for it in plain language:

```
/grimorio A sunken city where the nobles are aquatic vampires
```

The architect asks for language, then six questions (name, one-shot or campaign, level range, tone,
duration), and generates the campaign chapter by chapter. Everything lands in
`~/campaigns/<campaign-name>/` and ends as a PDF.

**Run a live session** — switch to the `grimorio-dm` agent inside a campaign directory. It loads
canon, areas, NPCs and bestiary, tracks session state, hides enemy HP/AC, and keeps narration
consistent across sessions. Optional local voice narration with Piper TTS.

**Validate and export** — validation is a CLI command; output format is an MCP tool:

```bash
grimorio validate sunken-city                  # consistency + WotC format + structure
grimorio validate sunken-city --json           # machine-readable, for CI
grimorio campaign export sunken-city           # tar.gz archive of the campaign directory
```

PDF (default), Markdown and EPUB come from the `export_campaign` MCP tool, so ask the agent for it
(for example, "export sunken-city as epub").

## Update

```bash
grimorio update all      # OpenCode integration + skills + agents + binary (recommended)
grimorio update          # binary only
grimorio update commands # OpenCode config and /grimorio command only
grimorio update skills   # skills only
grimorio update agents   # agents only
```

Updates are idempotent: they back up `opencode.json`, migrate older layouts, and keep every setting
that is not theirs. Re-run the installer with `sh -s -- --update` if an update ever fails.

## Documentation

**Start here**

| Doc | What it covers |
|-----|----------------|
| [Getting Started](docs/getting-started.md) | Install, requirements, first campaign, brief template, `validate`, exports |
| [Session Tutorial](docs/tutorials/session-tutorial.md) | Your first game: prep → play → post-session |
| [Walkthrough: La Hoja de Vlad](docs/walkthroughs/la-hoja-de-vlad.md) | Annotated tour of a finished campaign |

**Generate**

| Doc | What it covers |
|-----|----------------|
| [MCP Tools Reference](docs/features/mcp-tools.md) | Every tool, grouped by job |
| [PDF Compiler](docs/features/pdf-compiler.md) | Layout, CSS, sections, output control |
| [PDF Compiler Enhancements](docs/pdf-compiler-enhancements.md) | Newer layout features |
| [Save Command Format](docs/SAVE_COMMANDS_FORMAT.md) | Markdown shape each `save_*` tool expects |
| [Monster Design Rules](docs/dnd-monster-design-rules.md) | CR math and stat-block format (DMG ch. 9) |
| [Character Creation](docs/tutorials/character-creation.md) | PCs, pre-gens, worksheets |

**Run & play**

| Doc | What it covers |
|-----|----------------|
| [DM Agent Guide](docs/dm-agent-guide.md) | Running live sessions with `grimorio-dm` |
| [DM Guide](docs/dm-guide.md) | General advice for running games |
| [Campaign Consistency](docs/campaign-consistency.md) | Health score, rollback, persistent consequences |
| [Session Generator](docs/tutorials/session-generator.md) | Adapt sessions to specific characters |
| [TTS (experimental)](docs/tts-experimental.md) | Piper setup, voices, narration script |
| [TTS Roadmap](docs/tts-roadmap.md) | Where narration is going |

**Reference & contributing**

| Doc | What it covers |
|-----|----------------|
| [Architecture](docs/features/architecture.md) | Layers, storage, key abstractions |
| [Developer Guide](docs/developer-guide.md) | Build, test, lint, release |
| [Migration v2 → v3](docs/migration-v2-to-v3.md) | Move legacy `areas/` campaigns to `chapters/` |
| [Changelog](CHANGELOG.md) | Release history |

## Requirements

| Dependency | When you need it |
|------------|------------------|
| Chrome / Chromium | PDF compilation (preferred engine) |
| wkhtmltopdf | PDF fallback only, if you cannot install Chrome |
| `python3` | Config merge during install and update |
| Go 1.24+ | Only to build from source |
| `poppler-utils` | Only for PDF text extraction on Windows/WSL |

## Build from source

```bash
git clone https://github.com/pauvalls/grimorio.git
cd grimorio
make build     # ./grimorio
make test      # tests + coverage
make lint      # golangci-lint
```

## Troubleshooting

**`grimorio: command not found`** — add `~/.local/bin` to your `PATH`.

**`opencode mcp list` does not show grimorio** — run `grimorio update commands`, then reopen OpenCode.

**Update failed or drifted** — every merge leaves `~/.config/opencode/opencode.json.backup.*`.
Check `grimorio --version` against `~/.grimorio/grimorio --version`; `grimorio update` replaces
every install dir. Still stuck: re-run the installer with `--update`.

**PDF not generated** — install Chrome/Chromium, or `wkhtmltopdf` as a fallback.

## Contributing

Read the [Developer Guide](docs/developer-guide.md). Conventional commits, tests with the code they
verify, and `make test` before opening a PR.

## License

[LICENSE](LICENSE)
