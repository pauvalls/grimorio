<div align="center">

# 📜 Grimorio

**D&D One-shot & Campaign Generator**

[English](#english) · [Español](#español)

</div>

---

<a name="english"></a>

## 🇬🇧 English

AI-powered D&D 5e campaign and one-shot generator. Turn a spark of an idea into a fully-formatted, print-ready PDF adventure book — complete with lore, NPCs, bestiary, encounters, and styled layouts inspired by official Wizards of the Coast manuals.

### Features

- **Full campaign generation** — Lore, acts, NPCs, monsters, encounters, and maps
- **Interactive Q&A flow** — `/grimorio` asks questions first, then generates via parallel subagents
- **Image generation** — Procedural SVG maps (free) + AI cover art/portraits via Pollinations.ai (FREE, no API key) or DALL-E 3 (optional)
- **Multi-provider LLM** — Works with OpenAI, Anthropic, Groq, Ollama (via OpenCode / Claude Code)
- **D&D-styled PDF** — Professional layout inspired by official Wizards of the Coast manuals, with embedded images (maps, AI art, portraits)
- **MCP Server** — Native integration with OpenCode and Claude Code as MCP tools
- **Zero cloud dependencies** — Runs 100% locally, no servers required

### Quick Install

```bash
curl -sSL https://raw.githubusercontent.com/pauvalls/grimorio/main/install.sh | bash
```

**What the installer does:**

| Step | Claude Code | OpenCode |
|------|-------------|----------|
| Go 1.24+ | Installs if missing | Installs if missing |
| wkhtmltopdf | Installs if missing | Installs if missing |
| Build binary | `~/.local/bin/grimorio` | `~/.local/bin/grimorio` |
| Plugin files | `~/.claude/plugins/grimorio/` | `~/.config/opencode/plugins/grimorio/` |
| `.mcp.json` | Uses `${CLAUDE_PLUGIN_ROOT}` | Uses absolute path |
| MCP server config | Auto-discovered via `.mcp.json` | Added to `opencode.json` mcp section |
| Agent config | Auto-discovered from plugin | Added to `opencode.json` agent section |
| `/grimorio` command | Auto-discovered from plugin | Added to `opencode.json` command section |
| Shell PATH | `~/.local/bin` added | `~/.local/bin` added |

### Requirements

| Dependency | Auto-installed | Purpose |
|------------|---------------|---------|
| Go 1.24+ | ✅ Yes | Build the MCP server binary |
| wkhtmltopdf | ✅ Yes | Compile HTML to PDF |
| Git | ❌ Must have | Clone the repository |

### Usage

#### OpenCode / Claude Code

Once installed, type in the chat:

```
/grimorio A sunken city where the nobles are aquatic vampires
```

The `grimorio-architect` agent will:

```
Phase 1: Interactive Q&A
  ├─ Campaign name? (kebab-case)
  ├─ One-shot or full campaign?
  ├─ Player level? (1-20)
  ├─ Desired tone? (heroic, dark, humorous, political intrigue)
  └─ Duration? (one-shot, 3-5 sessions, long campaign)

Phase 2: Create campaign structure (MCP: create_campaign)

Phase 3-13: End-to-end orchestration by grimorio-architect
  └─ grimorio-architect coordinates directly:
      Phase 3: Content subagents (PARALLEL delegate)
        ├─ Lore: world, setting, conflict (MCP: save)
        ├─ NPCs: 5+ NPCs + factions (MCP: save_npcs)
        ├─ Bestiary: 3-5 monsters (MCP: save_bestiary)
        ├─ Encounters: balanced fights (MCP: save_encounters)
        └─ Maps: scene descriptions (MCP: save_maps)
      Phase 4: Report content status to user
      Phase 5: Acts subagent (uses [SCENE: ...] placeholders)
        └─ Acts: referencing NPCs, monsters, encounters (MCP: save_act)
      Phase 6: Report acts status to user
      Phase 7: Visual assets (PARALLEL delegate)
        ├─ grimorio-cartographer: battle maps, dividers (MCP: generate_map + generate_divider)
        └─ grimorio-artist: prepares batch-spec.json with all image prompts
      Phase 8: Report SVGs status to user
      Phase 9: AI Image Generation (architect calls MCP directly, SEQUENTIAL)
        ├─ One by one via generate_image (3s delay between each)
        └─ Fallback: retry failed images individually
      Phase 10: Update References (delegate to grimorio-artist)
        └─ Updates markdown files with actual image references
      Phase 11: Report references status to user
      Phase 12: Compile PDF (MCP: compile_pdf) — embeds all images
      Phase 13: Final report to user

```

> **Important:** The grimorio-architect agent does everything end-to-end: gathers requirements, creates structure, delegates content subagents, generates images directly via MCP, and compiles the PDF. The architect reports progress to the user after every phase.

### Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         OpenCode / Claude Code                          │
│                                                                         │
│  ┌─────────────────┐  ┌──────────────┐  ┌──────────────────────────┐   │
│  │ /grimorio       │  │ grimorio-    │  │ grimorio-artist          │   │
│  │ Command         │──│ architect    │  │ (Image specs + refs)     │   │
│  │ (Entry point)   │  │ (Orchestrator│  └──────────────────────────┘   │
│  └─────────────────┘  │  of ALL      │  ┌──────────────────────────┐   │
│                       │  phases)     │  │ grimorio-cartographer    │   │
│                       └──────┬───────┘  │ (SVG maps + dividers)    │   │
│                              │          └──────────────────────────┘   │
│                              │                                          │
│  Content Sub-agents          │  ┌──────────────────────────┐           │
│  (delegated by architect):   │  │ grimorio-narrative-      │           │
│  ├─ grimorio-lore            │  │ custodian                │           │
│  ├─ grimorio-npc             │  │ (Coherence validation    │           │
│  ├─ grimorio-bestiary        │  │  + state tracking)       │           │
│  ├─ grimorio-encounters      │  └──────────────────────────┘           │
│  ├─ grimorio-maps            │          Skill: dnd-5e-srd              │
  │  ├─ grimorio-areas           │          (D&D 5e rules context)         │
  │  │   (10-15 numbered areas/  │                                          │
  │  │    act, WotC format)       │                                          │
  │  ├─ grimorio-integrator      │                                          │
  │  │   (Validation + cross-ref  │                                          │
  │  │    + balance audit)        │                                          │
  │  ├─ grimorio-quests          │                                          │
  │  └─ grimorio-characters      │                                          │
│                              ▼                                          │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                         MCP Server (Go, stdio)                          │
│                                                                         │
│  ┌────────────────────────────┐  ┌──────────────────────────────────┐  │
│  │   Content Tools (v1)       │  │   Narrative Coherence (v2.0)     │  │
│  ├────────────────────────────┤  ├──────────────────────────────────┤  │
│  │ create_campaign            │  │ generate_adventure_bible         │  │
│  │   → Directory structure    │  │   → Creates canon.json           │  │
│  │ get_template               │  │ validate_canon                   │  │
│  │   → Markdown templates     │  │   → Validates against canon      │  │
│  │ save_act / save_npcs       │  │ update_narrative_state           │  │
│  │ save_bestiary / save_maps  │  │   → Tracks session state         │  │
│  │ save_encounters            │  │ check_consistency                │  │
│  └────────────────────────────┘  │   → Full campaign validation     │  │
│  ┌────────────────────────────┐  │ process_consistency_gate         │  │
│  │   Asset Tools              │  │   → Batch approve/reject/retry   │  │
│  ├────────────────────────────┤  └──────────────────────────────────┘  │
│  │ generate_map               │  ┌──────────────────────────────────┐  │
│  │   → SVG battle maps        │  │   Living World (v2.1)            │  │
│  │ generate_divider           │  ├──────────────────────────────────┤  │
│  │   → SVG dividers           │  │ update_faction_reputation        │  │
│  │ generate_image             │  │   → Propagate rep to allies      │  │
│  │   → AI art (FREE)          │  │ generate_random_tables           │  │
│  └────────────────────────────┘  │   → Contextual encounter tables  │  │
│                                  │ generate_handouts                │  │
│                                  │   → Player + DM versions         │  │
│                                  │ evaluate_consequences            │  │
│                                  │   → Trigger rules from state     │  │
│                                  └──────────────────────────────────┘  │
│                                  ┌──────────────────────────────────┐  │
│                                  │   DM Experience (Phase 4)        │  │
│                                  ├──────────────────────────────────┤  │
│                                  │ generate_session_prep            │  │
│                                  │   → DM prep sheet                │  │
│                                  │ generate_flowchart               │  │
│                                  │   → Campaign flowchart (Mermaid) │  │
│                                  └──────────────────────────────────┘  │
│                                  ┌──────────────────────────────────┐  │
│                                  │   Output                           │  │
│                                  │   compile_pdf                    │  │
│                                  │     → Intro → Lore → Acts →        │  │
│                                  │       Setting → Appendices         │  │
│                                  └──────────────────────────────────┘  │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

**Agent Hierarchy (grimorio-architect orchestrates all):**

```
grimorio-architect (Primary Orchestrator)
    │
    ├─ Phase 1: Requirements gathering (interactive)
    ├─ Phase 2: Campaign creation + Adventure Bible (canon)
    ├─ Phase 3a: grimorio-introduction → Introduction/overview (WotC format)
    │
    ├─ Batch 1 (PARALLEL delegate)
    │   ├─ grimorio-npc         → NPCs + factions
    │   ├─ grimorio-bestiary    → Monster stat blocks
    │   └─ grimorio-maps        → Location descriptions
    │   → grimorio-narrative-custodian → Consistency validation
    │
    ├─ Batch 2 (PARALLEL delegate)
    │   ├─ grimorio-lore        → World backstory
    │   ├─ grimorio-setting-guide → DM-only setting reference (spoilers)
    │   ├─ grimorio-quests      → Personal quests
    │   ├─ grimorio-encounters  → Combat challenges
    │   └─ grimorio-characters  → Pre-gen PCs
    │   → grimorio-narrative-custodian → Consistency validation
    │   → grimorio-narrative-custodian → State update
    │
    ├─ Batch 3: Areas (SEQUENTIAL)
    │   └─ grimorio-areas        → 10-15 numbered areas per act (WotC format)
    │      → grimorio-encounters → Combat templates referenced by areas
    │
    ├─ Phase 5e: grimorio-appendices → Consolidated reference (Items, NPCs, Handouts)
    │
    ├─ Batch 4: Integration (OBLIGATORY gate)
    │   └─ grimorio-integrator   → Cross-reference validation, XP balance,
    │                               treasure audit, consistency checks
    │      → Auto-fixes + approval required before compilation
    │
    ├─ Phase 6: grimorio-artist   → Image batch spec
    ├─ Phase 7: AI image generation (sequential MCP calls)
    ├─ Phase 8: grimorio-artist   → Update markdown references
    ├─ Phase 9: grimorio-narrative-custodian → Final consistency check
    ├─ Phase 10: DM Tools → Session prep + flowchart
    └─ Phase 11: PDF compilation (compiler v2 with TOC + cross-refs)
```

> **Development Rule:** Every new MCP tool must update:
> 1. The relevant agent(s) that use it
> 2. The architecture diagrams above
> 3. The MCP tools table in this README
> 4. The install.sh output (if user-facing)

### Image Generation

Grimorio supports multiple modes of image generation with **automatic fallback**:

#### AI Images (Default — FREE with Fallback)

No API key needed. Images are generated using free providers with automatic fallback:

| Priority | Provider | Description | Fallback |
|----------|----------|-------------|----------|
| 1 | Pollinations.ai | FLUX model, 1024x1024 | ✅ |
| 2 | Raphael AI | raphael.app, fast, unlimited | ✅ |
| 3 | DALL-E (optional) | Highest quality, requires API key | Manual config |

**Sequential Generation**: `generate_images_batch` generates images **one at a time** with a 3-second delay between requests. This prevents rate limiting on free APIs and ensures reliable generation.

**Automatic Fallback**: If Pollinations.ai fails, the system automatically tries Raphael AI. If that also fails, it reports the error. No manual intervention needed.

| Tool | Purpose | Cost |
|------|---------|------|
| `generate_image` | Single image with fallback | **FREE** |
| `generate_images_batch` | Multiple images sequential + fallback | **FREE** |

#### Procedural SVG (100% local, free)

No API key needed. Generates maps and dividers on the fly:

| Tool | Purpose | Example |
|------|---------|---------|
| `generate_map` | Battle maps, dungeon layouts, city maps | `![Dungeon Map](assets/dungeon-map.svg)` |
| `generate_divider` | Decorative section separators | `![Divider](assets/ornate-divider.svg)` |

**Map styles:** `dungeon`, `landscape`, `city`

#### DALL-E API (Optional — higher quality, paid)

For premium image quality, switch to DALL-E 3:

```bash
# Set your API key
export OPENAI_API_KEY="sk-..."

# Or add to config
echo '{"image_provider": "dalle", "dalle_api_key": "sk-..."}' > ~/.config/grimorio/config.json
```

| Tool | Purpose | Cost |
|------|---------|------|
| `generate_image` | Single image | ~$0.04-0.08/image (DALL-E 3) |
| `generate_images_batch` | Multiple images in parallel | ~$0.04-0.08/image (DALL-E 3) |

> **Tip:** OpenAI gives $5 free credit to new accounts (~60-120 images).

**PDF Image Embedding:**

All images (SVG maps, AI-generated PNGs, dividers) are automatically embedded into the PDF:
- Images referenced in Markdown with `![alt](assets/file.png)` appear inline
- Image deduplication — same image only appears once even if referenced multiple times
- SVGs are embedded as vector graphics, PNGs as base64

### Plugin Structure

```
~/.config/opencode/plugins/grimorio/    (OpenCode)
~/.claude/plugins/grimorio/             (Claude Code)
    │
    ├─ grimorio                          # MCP server binary
    ├─ .mcp.json                         # MCP server config (per-tool)
    ├─ .claude-plugin/plugin.json        # Plugin manifest
    ├─ commands/
    │   └─ grimorio.md                   # /grimorio slash command
    ├─ agents/
    │   ├─ grimorio-architect.md         # Campaign designer agent (end-to-end orchestration)
    │   ├─ grimorio-orchestrator.md      # DEPRECATED — architect handles this now
    │   ├─ grimorio-artist.md            # Image specs + markdown reference updates
    │   └─ grimorio-cartographer.md      # SVG battle maps + decorative dividers
    └─ skills/
        └─ dnd-5e-srd/SKILL.md           # D&D 5e rules reference

Source code structure:
    │
    ├─ cmd/
    │   ├─ grimorio/                     # Entry point (stdio MCP server)
    │   └─ migrate-v1-to-v2/             # Migration tool v1→v2
    ├─ internal/
    │   ├── domain/                      # Domain models (Canon, NarrativeState, Validation)
    │   ├── mcp/server.go                # MCP tool definitions + handlers
    │   ├── services/                    # Business logic (CanonService, ValidationEngine)
    │   ├── repository/                  # Persistence layer (filesystem + memory)
    │   ├── compiler/compiler.go         # Markdown → HTML → PDF pipeline
    │   ├── svg/svg.go                   # Procedural SVG generator (maps, dividers)
    │   ├── image/                       # Image provider abstraction (Pollinations.ai, DALL-E)
    │   └── config/config.go             # Configuration management
    └─ internal/compiler/templates/      # Embedded CSS + Markdown templates
```

### OpenCode Configuration

The installer automatically adds the following to `~/.config/opencode/opencode.json`:

```jsonc
{
  "mcp": {
    "grimorio": {
      "command": ["/home/user/.config/opencode/plugins/grimorio/grimorio"],
      "type": "local",
      "enabled": true
    }
  },
  "agent": {
    "grimorio-architect": {
      "mode": "primary",
      "tools": { "delegate": true /* ... */ }
    }
  },
  "command": {
    "grimorio": {
      "agent": "grimorio-architect",
      "subtask": false,
      "template": "Generate a D&D 5e campaign..."
    }
  }
}
```

### Output Structure

Every generated campaign lives in its own directory:

```
~/campaigns/
└── sunken-city/
    ├── README.md
    ├── lore.md
    ├── canon.json              ← NEW: Canonical facts, entities, timeline, rules
    ├── narrative_state.json    ← NEW: Session state (clues, quests, deaths, decisions)
    ├── acts/
    │   ├── act_1_the_descent.md
    │   ├── act_2_the_feast.md
    │   └── act_3_the_ritual.md
    ├── npcs/
    │   └── npcs_and_factions.md
    ├── bestiary/
    │   └── bestiary.md
    ├── encounters/
    │   └── encounters.md
    ├── maps/
    │   └── maps_and_scenes.md
    ├── assets/
    │   ├── cover-art.png            ← AI cover art
    │   ├── npc-eldric.png           ← AI NPC portraits
    │   ├── npc-lira.png             ← AI NPC portraits
    │   ├── scene-act1-boss.png      ← AI scene illustrations
    │   ├── scene-act2-ritual.png    ← AI scene illustrations
    │   ├── dungeon-map.svg          ← Procedural battle maps
    │   └── ornate-divider.svg       ← Decorative section dividers
    ├── campaign.html
    └── campaign.pdf                 ← Final PDF with embedded images
```

**PDF Compilation Order** (professional D&D adventure structure):

```
1. Cover Page (with title, subtitle, cover art)
2. Table of Contents
3. Lore y Ambientación       ← World backstory, setting, history
4. Acts (Chapters)            ← Full narrative with embedded NPCs, 
   │                            read-aloud text, numbered areas,
   │                            maps, encounters, and quests
   ├── Act 1: ...
   ├── Act 2: ...
   └── Act 3: ...
5. Apéndice A: NPCs           ← NPC profiles & factions (reference)
6. Apéndice B: Bestiary       ← Monster stat blocks (reference)
7. Apéndice C: Encounters     ← Combat & challenge summaries
8. Apéndice D: Maps           ← Location & zone reference
```

This structure mirrors official Wizards of the Coast adventure modules: chapters embed all narrative content (NPCs, locations, encounters) inline, with appendices at the end for quick reference.

### Available Templates

The MCP server exposes structured templates for each content type:

| Template   | Description                                    |
|------------|------------------------------------------------|
| `act`      | Act/chapter structure with OotA-style sections, read-aloud, numbered areas |
| `npc`      | NPC sheet with motivation and secret           |
| `monster`  | Full D&D 5e stat block                         |
| `encounter`| Encounter with difficulty balancing            |
| `map`      | Scene description with zones                   |
| `lore`     | World-building and conflicts                   |
| `session-zero` | Session zero guide for DMs                 |

### MCP Tools

All **29 tools** available through the MCP server, organized by category:

#### Content Tools (v1)

| Tool | Type | Description |
|------|------|-------------|
| `create_campaign` | File | Creates campaign directory structure |
| `get_template` | Template | Returns structured Markdown template (types: act, npc, monster, encounter, map, lore, session-zero) |
| `save_act` | File | Saves act as Markdown file |
| `save_npcs` | File | Saves NPCs and factions |
| `save_bestiary` | File | Saves monster stat blocks |
| `save_encounters` | File | Saves combat encounters |
| `save_maps` | File | Saves scene descriptions |
| `save_lore` | File | Saves world lore and history |
| `compile_pdf` | PDF | Compiles all content into styled D&D adventure PDF |

#### Character & Quest Tools

| Tool | Type | Description |
|------|------|-------------|
| `generate_character` | Character | Generates a player character with stats and abilities |
| `get_character` | Character | Retrieves a character sheet |
| `list_characters` | Character | Lists all characters in a campaign |
| `create_personal_quest` | Quest | Creates a personal quest for a character |
| `update_quest_status` | Quest | Updates the status of a quest (active, completed, failed, on_hold) |
| `list_quests` | Quest | Lists all quests in a campaign |

#### Asset Tools

| Tool | Type | Description |
|------|------|-------------|
| `generate_map` | SVG | Procedural battle map generator (dungeon, landscape, city) — free, no API |
| `generate_divider` | SVG | Decorative section dividers — free, no API |
| `generate_image` | AI | Single image generation via Pollinations.ai (FREE) or DALL-E (optional) |
| `generate_images_batch` | AI | Bulk image generation — sequential with fallback |

#### Narrative Coherence (v2.0)

| Tool | Type | Description |
|------|------|-------------|
| `generate_adventure_bible` | Canon | Creates `canon.json` with facts, entities, timeline, and world rules |
| `validate_canon` | Validation | Validates content proposals against canon |
| `update_narrative_state` | State | Updates campaign state after sessions |
| `check_consistency` | Validation | Runs full campaign consistency check |
| `process_consistency_gate` | Gate | Batch validation gate — approve/reject/retry |

#### Living World (v2.1)

| Tool | Type | Description |
|------|------|-------------|
| `update_faction_reputation` | Faction | Updates faction reputation with propagation to allies |
| `generate_random_tables` | Tables | Contextual random tables for improvisation |
| `generate_handouts` | Handout | Player-facing + DM-only handouts |
| `evaluate_consequences` | Consequence | Evaluates consequence rules from player decisions |

#### DM Experience (Phase 4)

| Tool | Type | Description |
|------|------|-------------|
| `generate_session_prep` | DM Tool | Generates DM prep sheet for next session |
| `generate_flowchart` | DM Tool | Campaign flowchart (Mermaid diagram + SVG) |

**How Narrative Coherence works:**
1. **Adventure Bible** (`generate_adventure_bible`) — Creates a `canon.json` with immutable facts, entities (NPCs, locations, items), timeline, and world rules
2. **Content Validation** (`validate_canon`) — Before saving any act, quest, or encounter, the system checks:
   - Are referenced NPCs still alive?
   - Do entities exist in the canon?
   - Does content violate world rules (e.g., "magic is banned in this city")?
3. **State Tracking** (`update_narrative_state`) — After each session, record:
   - Which clues were revealed
   - Which quests were completed
   - Which NPCs died
   - Key decisions made by players
4. **Consistency Check** (`check_consistency`) — Validates the entire campaign before PDF compilation
5. **Batch Gate** (`process_consistency_gate`) — Atomically validates multiple content proposals (e.g., an entire act + encounters + NPCs) and returns:
   - `approved` — All proposals pass validation
   - `rejected` — With detailed feedback on which proposals failed and why
   - `retry` — With specific instructions on how to fix the issues

**How the Consequence System works:**
1. **Track decisions** in `update_narrative_state` with `key_decisions`
2. **Evaluate consequences** with `evaluate_consequences` to see what ripples through the world
3. **Update factions** with `update_faction_reputation` — changes propagate to allied factions automatically
4. **Generate handouts** with `generate_handouts` to give players tangible clues and rewards
5. **Create random tables** with `generate_random_tables` for improvisation based on current world state

### Campaign File Structure (v2.0)

Every generated campaign now includes coherence metadata:

```
~/campaigns/
└── sunken-city/
    ├── README.md
    ├── lore.md
    ├── canon.json              ← NEW: Canonical facts, entities, timeline, rules
    ├── narrative_state.json    ← NEW: Session state (clues, quests, deaths, decisions)
    ├── acts/
    │   ├── act_1_the_descent.md
    │   ├── act_2_the_feast.md
    │   └── act_3_the_ritual.md
    ├── npcs/
    │   └── npcs_and_factions.md
    ├── bestiary/
    │   └── bestiary.md
    ├── encounters/
    │   └── encounters.md
    ├── maps/
    │   └── maps_and_scenes.md
    ├── assets/
    │   ├── cover-art.png
    │   ├── npc-eldric.png
    │   ├── dungeon-map.svg
    │   └── ornate-divider.svg
    ├── campaign.html
    └── campaign.pdf
```

** canon.json structure:**
```json
{
  "schema_version": "2.0",
  "campaign_id": "sunken-city",
  "facts": [
    {
      "id": "fact-001",
      "category": "lore",
      "statement": "The curse comes from the dead god Morbus",
      "immutable": true
    }
  ],
  "entities": [
    {
      "id": "npc-lord-vex",
      "name": "Lord Vex",
      "type": "npc",
      "role": "ally",
      "canon_state": "alive",
      "motivation": "Protect the city at all costs",
      "secret": "His family awakened Morbus 200 years ago"
    }
  ],
  "rules": [
    {
      "id": "rule-001",
      "domain": "magic",
      "statement": "Arcane magic is banned in the city"
    }
  ]
}
```

### Migration from v1 to v2

If you have existing campaigns created before v2.0:

```bash
# Migrate all campaigns in ~/campaigns/
go run ./cmd/migrate-v1-to-v2 ~/campaigns
```

This creates `canon.json` and `narrative_state.json` for each campaign, with a `.v1-backup/` directory for safety.

### What's New in v2.0

**Area-Based Generation (WotC Quality)**
- **Before:** 3-5 narrative scenes per act (350 words each) — reads like a novel synopsis
- **After:** 10-15 numbered areas per act (150-200 words each) — reads like official D&D modules
  - Every area has specific DCs, treasure with XP values, and cross-references
  - 90%+ areas have mechanics (vs 50% before)
  - 70% of combat areas have treasure (vs 30% before)

**grimorio-integrator (New Agent)**
- Cross-reference validation: verifies every creature in areas exists in bestiary
- Balance audit: calculates XP per act, checks difficulty curve
- Consistency checks: NPCs can't be in two places, items can't duplicate
- Auto-fixes: adds missing creatures to bestiary, adds treasure to areas, fixes connections

**Compiler v2**
- Hierarchical table of contents with clickable links
- Cross-references between areas ("see Area C4")
- Inline stat blocks for unique creatures
- Player maps + DM maps with secrets
- Automatic handouts: clue lists, NPC rosters, session recaps

**Backwards Compatibility**
- `--compiler-version=1` flag for legacy PDF generation
- `grimorio-areas` — The ONLY area generator (no legacy version)
- Migration tool converts scene-based acts to area-based format

### Complete User Guide

This guide covers the full Grimorio workflow — from creating a campaign to running sessions and maintaining narrative coherence.

> **Additional documentation:** See [`docs/dm-guide.md`](docs/dm-guide.md) for detailed DM advice and [`docs/developer-guide.md`](docs/developer-guide.md) for contributing to Grimorio.

#### 1. Creating a Campaign

**Step-by-step workflow using `/grimorio`:**

```
User: /grimorio A sunken city where the nobles are aquatic vampires

Grimorio:
  Phase 1: "What's the campaign name? (kebab-case)"
  User: "sunken-city"
  Phase 1: "One-shot or full campaign?"
  User: "Full campaign"
  Phase 1: "Player level range?"
  User: "4-6"
  ... (5 questions total)

  Phase 2: Creating campaign structure... ✅
  Phase 3-13: Generating all content... (reports progress)
  
  ✅ Campaign complete!
  PDF: ~/campaigns/sunken-city/campaign.pdf
```

**What happens behind the scenes:**
1. **Requirements gathering** — 5 interactive questions
2. **Campaign creation** — `create_campaign` builds directory structure
3. **Adventure Bible** — `generate_adventure_bible` creates `canon.json`
4. **Content generation** — Parallel subagents create lore, NPCs, bestiary, encounters, maps
5. **Acts** — Narrative acts referencing all generated content
6. **Visual assets** — SVG maps, dividers, AI-generated images
7. **PDF compilation** — `compile_pdf` produces the final book

#### 2. Playing a Session

**Before the session:**
```
grimorio_generate_session_prep(
  campaign="sunken-city",
  session_num=3,
  focus="The players are heading to the sunken cathedral"
)
```

This generates a DM prep sheet with:
- Active quests and their current status
- Relevant NPCs (alive, their locations, motivations)
- Faction reputation warnings
- Pending consequences from previous sessions
- Random tables for improvisation

**During the session:**
- Use the generated acts as your guide
- Reference encounters, maps, and NPCs inline
- Track player decisions mentally or in notes

**After the session:**
```
grimorio_update_narrative_state(
  campaign_id="sunken-city",
  session_num=3,
  revealed_clues=["clue-cathedral-key", "clue-vampire-weakness"],
  dead_npcs=["npc-guard-captain"],
  completed_quests=["quest-find-cathedral"],
  key_decisions=["Players spared the vampire noble's daughter"],
  xp_awarded=450,
  loot_acquired=["Silver Dagger +1", "Cathedral Key"],
  session_summary="Party reached the cathedral, defeated the guard captain..."
)
```

#### 3. Maintaining Coherence

**How validation works:**

Every piece of content is validated against `canon.json` before being saved:

```
grimorio_validate_canon(
  campaign_id="sunken-city",
  proposal={
    id: "act-3",
    type: "act",
    content: "...",
    entity_references: [
      { entity_id: "npc-guard-captain", location: "act_3" }
    ]
  }
)
```

**Example — Preventing NPC resurrections:**

If `npc-guard-captain` was marked dead in session 2, the validator returns:
```json
{
  "status": "rejected",
  "issues": [
    {
      "rule": "npc_death_state",
      "severity": "critical",
      "message": "NPC 'Guard Captain' is dead (session 2) but appears in Act 3",
      "fix_suggestion": "Replace with Lieutenant Mara or use a written letter"
    }
  ]
}
```

**Pre-PDF consistency check:**
```
grimorio_check_consistency(campaign_id="sunken-city")
```

This validates the entire campaign before compilation — checking for dead NPCs appearing alive, lore contradictions, missing entities, and timeline issues.

#### 4. Consequence System

After each session, evaluate what the players' actions mean for the world:

```
grimorio_evaluate_consequences(
  campaign_id="sunken-city",
  trigger_decisions=["Players killed Lord Vex, the noble"]
)
```

**Example — If players killed a noble:**
- **Immediate:** His daughter swears vengeance (new quest)
- **Faction:** House Vex reputation drops to Hostile
- **Political:** Power vacuum — other nobles scramble for his territory
- **Economic:** Trade routes through his district become dangerous
- **Military:** His guards disband or join mercenary groups

**Faction reputation changes:**
```
grimorio_update_faction_reputation(
  campaign="sunken-city",
  faction="house-vex",
  party="default",
  delta=-30,
  reason="Players killed Lord Vex"
)
```

This propagates to allied factions (House Vex allies also drop) and may trigger new consequences.

#### 5. DM Tools

**Session Prep Sheet (`generate_session_prep`):**
- Lists all active quests with current objectives
- Shows relevant NPCs and their current status/location
- Warns about faction reputation issues
- Includes random tables for improvisation (encounters, rumors, weather)
- Notes pending consequences that may trigger

**Campaign Flowchart (`generate_flowchart`):**
```
grimorio_generate_flowchart(
  campaign="sunken-city",
  title="Sunken City Campaign"
)
```

Generates a visual flowchart (Mermaid diagram + SVG) showing:
- All acts and their decision points
- Quest branches and consequences
- NPC relationship map
- Faction standing overview

**Random Tables (`generate_random_tables`):**
```
grimorio_generate_random_tables(
  campaign="sunken-city",
  context="sunken cathedral district"
)
```

Generates contextual tables for improvisation:
- Random encounters (appropriate to location and level)
- Rumors and overheard conversations
- Environmental events
- NPC reactions based on faction reputation

**Handouts (`generate_handouts`):**
```
grimorio_generate_handouts(
  campaign="sunken-city",
  type="clue",
  subject="The Cathedral Seal"
)
```

Creates both:
- **Player version** — What the players see (aged letter, cryptic note)
- **DM version** — Full context and secrets the players don't know

#### 6. Updating Based on Player Decisions

**Track decisions in narrative state:**

After every session, update the state with key decisions:
```
grimorio_update_narrative_state(
  campaign_id="sunken-city",
  session_num=3,
  key_decisions=[
    "Players allied with the Merfolk instead of the Vampires",
    "Players destroyed the Blood Crystal instead of using it"
  ]
)
```

**Adapt the campaign with consequences:**

The next time you generate content, consequences from previous decisions will affect:
- Which NPCs are available (allies vs enemies)
- Faction reactions to the party
- Available quests and paths
- Random encounter tables

**Update faction standings after diplomatic choices:**
```
grimorio_update_faction_reputation(
  campaign="sunken-city",
  faction="merfolk-alliance",
  party="default",
  delta=+20,
  reason="Players helped rescue merfolk prisoners"
)
```

This automatically updates allied factions too (e.g., the Sea Temple priests also improve).

#### 7. Compiling the Final PDF

**When to compile:**
- After initial campaign generation (complete book)
- Before each session (quick reference with latest state)
- After major content updates (new acts, NPCs, images)

**How to compile:**
```
grimorio_compile_pdf(
  campaign="sunken-city",
  title="The Sunken City"
)
```

**What gets included:**
1. Cover page (with AI-generated cover art)
2. Table of Contents
3. Session Zero guide (if generated)
4. Campaign flowchart (if generated)
5. Lore and World Setting
6. All Acts (with embedded NPCs, encounters, maps, scenes)
7. Appendix A: NPCs & Factions
8. Appendix B: Bestiary
9. Appendix C: Encounters
10. Appendix D: Maps
11. Appendix E: Faction Tracker
12. Appendix F: Adventure Roster

All images (SVG maps, AI-generated PNGs, dividers) are automatically embedded.

---

### Development

#### Manual Build

```bash
go build -o grimorio ./cmd/grimorio
```

#### Run MCP Server (standalone)

```bash
./grimorio
```

The server runs in stdio mode (stdin/stdout) for MCP communication.

---

<a name="español"></a>

## 🇪🇸 Español

Generador de campañas y one-shots de D&D 5e impulsado por IA. Convierte una idea en un libro de aventura completo con formato profesional, listo para imprimir — incluyendo lore, NPCs, bestiario, encuentros y maquetación inspirada en los manuales oficiales de Wizards of the Coast.

### Características

- **Generación completa de campañas** — Lore, actos, NPCs, monstruos, encuentros y mapas
- **Flujo interactivo Q&A** — `/grimorio` pregunta primero, después genera con subagentes en paralelo
- **Generación de imágenes** — Mapas SVG procedurales (gratis) + portadas/retratos IA vía Pollinations.ai (GRATIS, sin API key) o DALL-E 3 (opcional)
- **Múltiples proveedores LLM** — Funciona con OpenAI, Anthropic, Groq, Ollama (vía OpenCode / Claude Code)
- **PDF estilo D&D** — Maquetación profesional con CSS Paged Media y wkhtmltopdf, con imágenes embebidas (mapas, arte IA, retratos)
- **Servidor MCP** — Integración nativa con OpenCode y Claude Code como herramientas MCP
- **Sin dependencias de nube** — Funciona 100% local, sin servidores

### Instalación Rápida

```bash
curl -sSL https://raw.githubusercontent.com/pauvalls/grimorio/main/install.sh | bash
```

**Qué hace el instalador:**

| Paso | Claude Code | OpenCode |
|------|-------------|----------|
| Go 1.24+ | Instala si falta | Instala si falta |
| wkhtmltopdf | Instala si falta | Instala si falta |
| Compilar binario | `~/.local/bin/grimorio` | `~/.local/bin/grimorio` |
| Archivos del plugin | `~/.claude/plugins/grimorio/` | `~/.config/opencode/plugins/grimorio/` |
| `.mcp.json` | Usa `${CLAUDE_PLUGIN_ROOT}` | Usa ruta absoluta |
| Config MCP | Auto-detectado vía `.mcp.json` | Agregado a `opencode.json` sección mcp |
| Config agente | Auto-detectado del plugin | Agregado a `opencode.json` sección agent |
| Comando `/grimorio` | Auto-detectado del plugin | Agregado a `opencode.json` sección command |
| Shell PATH | `~/.local/bin` agregado | `~/.local/bin` agregado |

### Requisitos

| Dependencia | Auto-instalada | Propósito |
|------------|---------------|-----------|
| Go 1.24+ | ✅ Sí | Compilar el binario del servidor MCP |
| wkhtmltopdf | ✅ Sí | Compilar HTML a PDF |
| Git | ❌ Requerido | Clonar el repositorio |

### Uso

#### OpenCode / Claude Code

Una vez instalado, escribí en el chat:

```
/grimorio Una ciudad sumergida donde los nobles son vampiros acuáticos
```

El agente `grimorio-architect` va a:

```
Fase 1: Preguntas interactivas
  ├─ ¿Nombre de la campaña? (kebab-case)
  ├─ ¿One-shot o campaña completa?
  ├─ ¿Nivel de los jugadores? (1-20)
  ├─ ¿Tono deseado? (heroico, oscuro, humorístico, intriga política)
  └─ ¿Duración? (one-shot, 3-5 sesiones, campaña larga)

Fase 2: Crear estructura (MCP: create_campaign)

Fase 3-13: Orquestación completa por grimorio-architect
  └─ grimorio-architect coordina directamente:
      Fase 3: Subagentes de contenido (PARALELO delegate)
        ├─ Lore: mundo, ambientación, conflicto (MCP: save)
        ├─ NPCs: 5+ NPCs + facciones (MCP: save_npcs)
        ├─ Bestiario: 3-5 monstruos (MCP: save_bestiary)
        ├─ Encuentros: combates balanceados (MCP: save_encounters)
        └─ Mapas: descripciones de escenas (MCP: save_maps)
      Fase 4: Reportar estado del contenido al usuario
      Fase 5: Subagente de actos (usa placeholders [ESCENA: ...])
        └─ Actos: referenciando NPCs, monstruos, encuentros (MCP: save_act)
      Fase 6: Reportar estado de los actos al usuario
      Fase 7: Assets visuales (PARALELO delegate)
        ├─ grimorio-cartographer: mapas de batalla, divisores (MCP: generate_map + generate_divider)
        └─ grimorio-artist: prepara batch-spec.json con todos los prompts de imágenes
      Fase 8: Reportar estado de SVGs al usuario
      Fase 9: Generación de imágenes AI (architect llama MCP directamente, SECUENCIAL)
        ├─ Una por una con generate_image (3s de delay entre cada una)
        └─ Fallback: reintenta imágenes fallidas individualmente
      Fase 10: Actualizar referencias (delegate a grimorio-artist)
        └─ Actualiza archivos markdown con referencias reales de imágenes
      Fase 11: Reportar estado de referencias al usuario
      Fase 12: Compilar PDF (MCP: compile_pdf) — embebe todas las imágenes
      Fase 13: Reporte final al usuario
```

> **Importante:** El agente grimorio-architect hace todo de principio a fin: recopila requisitos, crea la estructura, delega subagentes de contenido, genera imágenes directamente via MCP, y compila el PDF. El architect reporta progreso al usuario después de cada fase.

### Arquitectura

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         OpenCode / Claude Code                          │
│                                                                         │
│  ┌─────────────────┐  ┌──────────────┐  ┌──────────────────────────┐   │
│  │ Comando         │  │ grimorio-    │  │ grimorio-artist          │   │
│  │ /grimorio       │──│ architect    │  │ (Specs imágenes + refs)  │   │
│  │ (Punto entrada) │  │ (Orquestador │  └──────────────────────────┘   │
│  └─────────────────┘  │  de TODAS    │  ┌──────────────────────────┐   │
│                       │  las fases)  │  │ grimorio-cartographer    │   │
│                       └──────┬───────┘  │ (Mapas SVG + divisores)  │   │
│                              │          └──────────────────────────┘   │
│                              │                                          │
│  Sub-agentes de contenido    │  ┌──────────────────────────┐           │
│  (delegados por architect):  │  │ grimorio-narrative-      │           │
│  ├─ grimorio-lore            │  │ custodian                │           │
│  ├─ grimorio-npc             │  │ (Validación coherencia   │           │
│  ├─ grimorio-bestiary        │  │  + tracking estado)      │           │
│  ├─ grimorio-encounters      │  └──────────────────────────┘           │
│  ├─ grimorio-maps            │          Skill: dnd-5e-srd              │
│  ├─ grimorio-acts            │          (Contexto reglas D&D 5e)       │
│  ├─ grimorio-quests          │                                          │
│  └─ grimorio-characters      │                                          │
│                              ▼                                          │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                         Servidor MCP (Go, stdio)                        │
│                                                                         │
│  ┌────────────────────────────┐  ┌──────────────────────────────────┐  │
│  │   Herramientas v1          │  │   Coherencia Narrativa (v2.0)    │  │
│  ├────────────────────────────┤  ├──────────────────────────────────┤  │
│  │ create_campaign            │  │ generate_adventure_bible         │  │
│  │   → Estructura directorios │  │   → Crea canon.json              │  │
│  │ get_template               │  │ validate_canon                   │  │
│  │   → Templates Markdown     │  │   → Valida contra canon          │  │
│  │ save_act / save_npcs       │  │ update_narrative_state           │  │
│  │ save_bestiary / save_maps  │  │   → Seguimiento de estado        │  │
│  │ save_encounters            │  │ check_consistency                │  │
│  └────────────────────────────┘  │   → Validación completa          │  │
│  ┌────────────────────────────┐  │ process_consistency_gate         │  │
│  │   Herramientas de Assets   │  │   → Gate aprobar/rechazar/retry  │  │
│  ├────────────────────────────┤  └──────────────────────────────────┘  │
│  │ generate_map               │  ├──────────────────────────────────┤  │
│  │   → Mapas batalla SVG      │  │   Mundo Vivo (v2.1)              │  │
│  │ generate_divider           │  ├──────────────────────────────────┤  │
│  │   → Divisores SVG          │  │ update_faction_reputation        │  │
│  │ generate_image             │  │   → Propaga rep a aliados        │  │
│  │   → Arte IA (GRATIS)       │  │ generate_random_tables           │  │
│  └────────────────────────────┘  │   → Tablas de encuentro          │  │
│                                  │ generate_handouts                │  │
│                                  │   → Versiones jugador + DM       │  │
│                                  │ evaluate_consequences            │  │
│                                  │   → Reglas desde estado          │  │
│                                  └──────────────────────────────────┘  │
│                                  ┌──────────────────────────────────┐  │
│                                  │   Experiencia de DM (Fase 4)     │  │
│                                  ├──────────────────────────────────┤  │
│                                  │ generate_session_prep            │  │
│                                  │   → Hoja prep DM                 │  │
│                                  │ generate_flowchart               │  │
│                                  │   → Flowchart campaña (Mermaid)  │  │
│                                  └──────────────────────────────────┘  │
│                                  ┌──────────────────────────────────┐  │
│                                  │   Output                           │  │
│                                  │ compile_pdf                      │  │
│                                  │   → PDF libro aventura D&D       │  │
│                                  │   → Lore → Actos → Apéndices     │  │
│                                  └──────────────────────────────────┘  │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

**Jerarquía de Agentes (grimorio-architect orquesta todo):**

```
grimorio-architect (Orquestador Principal)
    │
    ├─ Fase 1: Requisitos (interactivo)
    ├─ Fase 2: Crear campaña + Biblia de Aventura (canon)
    │
    ├─ Batch 1 (PARALELO delegate)
    │   ├─ grimorio-npc         → NPCs + facciones
    │   ├─ grimorio-bestiary    → Stat blocks monstruos
    │   └─ grimorio-maps        → Descripciones ubicaciones
    │   → grimorio-narrative-custodian → Validación coherencia
    │
    ├─ Batch 2 (PARALELO delegate)
    │   ├─ grimorio-lore        → Trasfondo mundo
    │   ├─ grimorio-quests      → Misiones personales
    │   ├─ grimorio-encounters  → Desafíos combate
    │   └─ grimorio-characters  → PJs pre-generados
    │   → grimorio-narrative-custodian → Validación coherencia
    │   → grimorio-narrative-custodian → Actualizar estado
    │
    ├─ Batch 3 (PARALELO delegate)
    │   ├─ grimorio-cartographer → Mapas SVG + divisores
    │   └─ grimorio-acts         → Actos narrativos
    │   → grimorio-narrative-custodian → Validación coherencia
    │
    ├─ Fase 6: grimorio-artist   → Especificación batch imágenes
    ├─ Fase 7: Generación imágenes IA (llamadas MCP secuenciales)
    ├─ Fase 8: grimorio-artist   → Actualizar referencias markdown
    ├─ Fase 9: grimorio-narrative-custodian → Check final
    ├─ Fase 10: Herramientas de DM → Prep sesión + flowchart
    └─ Fase 11: Compilación PDF
```

> **Regla de Desarrollo:** Cada nueva herramienta MCP debe actualizar:
> 1. El/los agente(s) relevante(s) que la usen
> 2. Los diagramas de arquitectura de arriba
> 3. La tabla de herramientas MCP en este README
> 4. La salida de install.sh (si es visible para el usuario)

### Estructura del Plugin

```
~/.config/opencode/plugins/grimorio/    (OpenCode)
~/.claude/plugins/grimorio/             (Claude Code)
    │
    ├─ grimorio                          # Binario del servidor MCP
    ├─ .mcp.json                         # Config del servidor MCP (por herramienta)
    ├─ .claude-plugin/plugin.json        # Manifiesto del plugin
    ├─ commands/
    │   └─ grimorio.md                   # Comando slash /grimorio
    ├─ agents/
    │   ├─ grimorio-architect.md           # Agente diseñador (orquestación completa)
    │   ├─ grimorio-narrative-custodian.md # Guardián de coherencia narrativa
    │   ├─ grimorio-orchestrator.md        # DEPRECATED — architect lo maneja ahora
    │   ├─ grimorio-artist.md              # Especificaciones de imágenes + actualización de referencias
    │   ├─ grimorio-cartographer.md        # Mapas de batalla SVG + divisores decorativos
    │   ├─ grimorio-lore.md                # Trasfondo y ambientación
    │   ├─ grimorio-npc.md                 # NPCs y facciones
    │   ├─ grimorio-bestiary.md            # Bestiario y stat blocks
    │   ├─ grimorio-encounters.md          # Encuentros y desafíos
    │   ├─ grimorio-maps.md                # Descripciones de mapas
    │   ├─ grimorio-acts.md                # Actos narrativos
    │   ├─ grimorio-quests.md              # Misiones personales
    │   └─ grimorio-characters.md          # Personajes pre-generados
    └─ skills/
        └─ dnd-5e-srd/SKILL.md           # Referencia de reglas D&D 5e

Estructura del código fuente:
    │
    ├─ cmd/
    │   ├─ grimorio/                     # Punto de entrada (servidor MCP stdio)
    │   └─ migrate-v1-to-v2/             # Herramienta de migración v1→v2
    ├─ internal/
    │   ├── domain/                      # Modelos de dominio (Canon, NarrativeState, Validation)
    │   ├── mcp/server.go                # Definiciones de herramientas MCP + handlers
    │   ├── services/                    # Lógica de negocio (CanonService, ValidationEngine)
    │   ├── repository/                  # Capa de persistencia (filesystem + memory)
    │   ├── compiler/compiler.go         # Pipeline Markdown → HTML → PDF
    │   ├── svg/svg.go                   # Generador SVG procedural (mapas, divisores)
    │   ├── image/                       # Abstracción de proveedores de imagen (Pollinations.ai, DALL-E)
    │   └── config/config.go             # Gestión de configuración
    └─ internal/compiler/templates/      # CSS + templates Markdown embebidos
```

### Configuración de OpenCode

El instalador agrega automáticamente lo siguiente a `~/.config/opencode/opencode.json`:

```jsonc
{
  "mcp": {
    "grimorio": {
      "command": ["/home/user/.config/opencode/plugins/grimorio/grimorio"],
      "type": "local",
      "enabled": true
    }
  },
  "agent": {
    "grimorio-architect": {
      "mode": "primary",
      "tools": { "delegate": true /* ... */ }
    }
  },
  "command": {
    "grimorio": {
      "agent": "grimorio-architect",
      "subtask": false,
      "template": "Generate a D&D 5e campaign..."
    }
  }
}
```

### Estructura de Salida

Cada campaña generada vive en su propio directorio:

```
~/campaigns/
└── ciudad-sumergida/
    ├── README.md
    ├── lore.md
    ├── canon.json              ← NUEVO: Hechos canónicos, entidades, timeline, reglas
    ├── narrative_state.json    ← NUEVO: Estado de sesión (pistas, quests, muertes, decisiones)
    ├── acts/
    │   ├── act_1_el_descenso.md
    │   ├── act_2_el_festin.md
    │   └── act_3_el_ritual.md
    ├── npcs/
    │   └── npcs_and_factions.md
    ├── bestiary/
    │   └── bestiary.md
    ├── encounters/
    │   └── encounters.md
    ├── maps/
    │   └── maps_and_scenes.md
    ├── assets/
    │   ├── cover-art.png            ← Portada IA
    │   ├── npc-eldric.png           ← Retratos IA de NPCs
    │   ├── npc-lira.png             ← Retratos IA de NPCs
    │   ├── scene-act1-boss.png      ← Ilustraciones IA de escenas
    │   ├── scene-act2-ritual.png    ← Ilustraciones IA de escenas
    │   ├── dungeon-map.svg          ← Mapas de batalla procedurales
    │   └── ornate-divider.svg       ← Divisores decorativos
    ├── campaign.html
    └── campaign.pdf                 ← PDF final con imágenes embebidas
```

**Orden de compilación del PDF** (estructura profesional de aventura D&D):

```
1. Portada (título, subtítulo, arte de portada)
2. Índice
3. Lore y Ambientación       ← Trasfondo del mundo, historia
4. Actos (Capítulos)          ← Narrativa completa con NPCs, 
   │                            texto para leer, áreas numeradas,
   │                            mapas, encuentros y misiones
   ├── Acto 1: ...
   ├── Acto 2: ...
   └── Acto 3: ...
5. Apéndice A: NPCs           ← Perfiles de NPCs y facciones (referencia)
6. Apéndice B: Bestiario      ← Stat blocks de monstruos (referencia)
7. Apéndice C: Encuentros     ← Resumen de combates y desafíos
8. Apéndice D: Mapas          ← Referencia de localizaciones y zonas
```

Esta estructura refleja los módulos de aventura oficiales de Wizards of the Coast: los capítulos embeben todo el contenido narrativo (NPCs, localizaciones, encuentros), con apéndices al final para consulta rápida.

### Templates Disponibles

El servidor MCP expone templates estructurados para cada tipo de contenido:

| Template    | Descripción                                         |
|-------------|-----------------------------------------------------|
| `act`       | Estructura de acto estilo OotA con secciones, read-aloud, áreas numeradas |
| `npc`       | Ficha de NPC con motivación y secreto               |
| `monster`   | Stat block completo D&D 5e                          |
| `encounter` | Encuentro con balance de dificultad                 |
| `map`       | Descripción de escena con zonas                     |
| `lore`      | Ambientación y conflictos                           |
| `session-zero` | Guía de sesión cero para DMs                    |

### Herramientas MCP

Todas las **29 herramientas** disponibles a través del servidor MCP, organizadas por categoría:

#### Herramientas de Contenido (v1)

| Herramienta | Tipo | Descripción |
|------------|------|-------------|
| `create_campaign` | Archivo | Crea estructura de directorios de campaña |
| `get_template` | Template | Devuelve template Markdown estructurado (tipos: act, npc, monster, encounter, map, lore, session-zero) |
| `save_act` | Archivo | Guarda acto como archivo Markdown |
| `save_npcs` | Archivo | Guarda NPCs y facciones |
| `save_bestiary` | Archivo | Guarda stat blocks de monstruos |
| `save_encounters` | Archivo | Guarda encuentros de combate |
| `save_maps` | Archivo | Guarda descripciones de escenas |
| `save_lore` | Archivo | Guarda lore e historia del mundo |
| `compile_pdf` | PDF | Compila todo en PDF estilo aventura D&D |

#### Herramientas de Personajes y Misiones

| Herramienta | Tipo | Descripción |
|------------|------|-------------|
| `generate_character` | Personaje | Genera un personaje jugador con stats y habilidades |
| `get_character` | Personaje | Obtiene la ficha de un personaje |
| `list_characters` | Personaje | Lista todos los personajes de una campaña |
| `create_personal_quest` | Misión | Crea una misión personal para un personaje |
| `update_quest_status` | Misión | Actualiza el estado de una misión (active, completed, failed, on_hold) |
| `list_quests` | Misión | Lista todas las misiones de una campaña |

#### Herramientas de Assets

| Herramienta | Tipo | Descripción |
|------------|------|-------------|
| `generate_map` | SVG | Generador de mapas de batalla procedurales (dungeon, landscape, city) — gratis, sin API |
| `generate_divider` | SVG | Divisores decorativos — gratis, sin API |
| `generate_image` | IA | Generación individual de imágenes vía Pollinations.ai (GRATIS) o DALL-E (opcional) |
| `generate_images_batch` | IA | Generación masiva de imágenes — secuencial con fallback |

#### Coherencia Narrativa (v2.0)

| Herramienta | Tipo | Descripción |
|------------|------|-------------|
| `generate_adventure_bible` | Canon | Crea `canon.json` con hechos, entidades, timeline y reglas del mundo |
| `validate_canon` | Validación | Valida propuestas de contenido contra el canon |
| `update_narrative_state` | Estado | Actualiza estado de campaña después de sesiones |
| `check_consistency` | Validación | Ejecuta validación completa de consistencia |
| `process_consistency_gate` | Gate | Gate de validación por lotes — aprobar/rechazar/reintentar |

#### Mundo Vivo (v2.1)

| Herramienta | Tipo | Descripción |
|------------|------|-------------|
| `update_faction_reputation` | Facción | Actualiza reputación de facción con propagación a aliados |
| `generate_random_tables` | Tablas | Tablas aleatorias contextuales para improvisación |
| `generate_handouts` | Handout | Handouts para jugadores y secretos para DM |
| `evaluate_consequences` | Consecuencia | Evalúa reglas de consecuencia de decisiones de jugadores |

#### Experiencia de DM (Fase 4)

| Herramienta | Tipo | Descripción |
|------------|------|-------------|
| `generate_session_prep` | DM Tool | Genera hoja de preparación de sesión para el DM |
| `generate_flowchart` | DM Tool | Diagrama de flujo de campaña (Mermaid + SVG) |

**Cómo funciona la Coherencia Narrativa:**
1. **Biblia de Aventura** (`generate_adventure_bible`) — Crea un `canon.json` con hechos inmutables, entidades (NPCs, localizaciones, items), timeline y reglas del mundo
2. **Validación de Contenido** (`validate_canon`) — Antes de guardar cualquier acto, quest o encuentro, el sistema verifica:
   - ¿Los NPCs referenciados siguen vivos?
   - ¿Las entidades existen en el canon?
   - ¿El contenido viola reglas del mundo (ej., "la magia está prohibida en esta ciudad")?
3. **Seguimiento de Estado** (`update_narrative_state`) — Después de cada sesión, registra:
   - Qué pistas fueron reveladas
   - Qué quests se completaron
   - Qué NPCs murieron
   - Decisiones clave de los jugadores
4. **Verificación de Consistencia** (`check_consistency`) — Valida toda la campaña antes de la compilación del PDF
5. **Gate por Lotes** (`process_consistency_gate`) — Valida atómicamente múltiples propuestas de contenido (ej., un acto completo + encuentros + NPCs) y devuelve:
   - `approved` — Todas las propuestas pasan la validación
   - `rejected` — Con feedback detallado sobre qué propuestas fallaron y por qué
   - `retry` — Con instrucciones específicas sobre cómo corregir los problemas

**Cómo funciona el Sistema de Consecuencias:**
1. **Trackeá decisiones** en `update_narrative_state` con `key_decisions`
2. **Evaluá consecuencias** con `evaluate_consequences` para ver qué repercusiones tiene en el mundo
3. **Actualizá facciones** con `update_faction_reputation` — los cambios se propagan automáticamente a facciones aliadas
4. **Generá handouts** con `generate_handouts` para dar pistas tangibles y recompensas a los jugadores
5. **Creá tablas aleatorias** con `generate_random_tables` para improvisación basada en el estado actual del mundo

### Migración de v1 a v2

Si tenés campañas existentes creadas antes de v2.0:

```bash
# Migra todas las campañas en ~/campaigns/
go run ./cmd/migrate-v1-to-v2 ~/campaigns
```

Esto crea `canon.json` y `narrative_state.json` para cada campaña, con un directorio `.v1-backup/` por seguridad.

### Novedades en v2.0

**Generación Basada en Áreas (Calidad WotC)**
- **Antes:** 3-5 escenas narrativas por acto (350 palabras cada una) — lee como sinopsis de novela
- **Ahora:** 10-15 áreas numeradas por acto (150-200 palabras cada una) — lee como módulos oficiales de D&D
  - Cada área tiene CDs específicos, tesoro con valores de XP, y referencias cruzadas
  - 90%+ de áreas tienen mecánicas (vs 50% antes)
  - 70% de áreas con combate tienen tesoro (vs 30% antes)

**grimorio-integrator (Nuevo Agente)**
- Validación de referencias cruzadas: verifica que cada criatura en áreas existe en bestiary
- Auditoría de balance: calcula XP por acto, verifica curva de dificultad
- Chequeos de consistencia: NPCs no pueden estar en dos lugares, items no pueden duplicarse
- Auto-correcciones: agrega criaturas faltantes al bestiary, agrega tesoro a áreas, corrige conexiones

**Compilador v2**
- Tabla de contenidos jerárquica con links clickeables
- Referencias cruzadas entre áreas ("ver Área C4")
- Stat blocks inline para criaturas únicas
- Mapas de jugador + mapas de DM con secretos
- Handouts automáticos: listas de pistas, roster de NPCs, recaps de sesión

**Compatibilidad Hacia Atrás**
- Flag `--compiler-version=1` para generación de PDF legacy
- `grimorio-areas` es el ÚNICO generador de áreas (sin versión legacy)
- Herramienta de migración convierte actos basados en escenas a formato basado en áreas

**Formato WotC Profesional (v2.1+)**
- `grimorio-introduction` — Introduction/overview con Foreword, Story Overview, Adventure Background, Running the Adventure
- `grimorio-setting-guide` — DM-only setting reference con Geography, History, Culture, Factions, Secrets
- `grimorio-appendices` — Consolidated appendices con Magic Items, NPCs/Monsters, Handouts, Maps, Reference Tables
- Pipeline: Introduction → Lore → Acts → Setting Guide → Individual Appendices → Appendices.md

### Guía de Uso Completa

Esta guía cubre el flujo de trabajo completo de Grimorio — desde crear una campaña hasta jugar sesiones y mantener la coherencia narrativa.

> **Documentación adicional:** Ver [`docs/dm-guide.md`](docs/dm-guide.md) para consejos detallados de DM y [`docs/developer-guide.md`](docs/developer-guide.md) para contribuir a Grimorio.

#### 1. Crear una Campaña

**Flujo de trabajo paso a paso usando `/grimorio`:**

```
Usuario: /grimorio Una ciudad sumergida donde los nobles son vampiros acuáticos

Grimorio:
  Fase 1: "¿Nombre de la campaña? (kebab-case)"
  Usuario: "ciudad-sumergida"
  Fase 1: "¿One-shot o campaña completa?"
  Usuario: "Campaña completa"
  Fase 1: "¿Rango de nivel de los jugadores?"
  Usuario: "4-6"
  ... (5 preguntas en total)

  Fase 2: Creando estructura de campaña... ✅
  Fase 3-13: Generando todo el contenido... (reporta progreso)
  
  ✅ ¡Campaña completa!
  PDF: ~/campaigns/ciudad-sumergida/campaign.pdf
```

**Qué pasa detrás de escena:**
1. **Recopilación de requisitos** — 5 preguntas interactivas
2. **Creación de campaña** — `create_campaign` construye la estructura de directorios
3. **Biblia de Aventura** — `generate_adventure_bible` crea `canon.json`
4. **Generación de contenido** — Subagentes en paralelo crean lore, NPCs, bestiario, encuentros, mapas
5. **Actos** — Actos narrativos que referencian todo el contenido generado
6. **Assets visuales** — Mapas SVG, divisores, imágenes generadas por IA
7. **Compilación PDF** — `compile_pdf` produce el libro final

#### 2. Jugar una Sesión

**Antes de la sesión:**
```
grimorio_generate_session_prep(
  campaign="ciudad-sumergida",
  session_num=3,
  focus="Los jugadores se dirigen a la catedral sumergida"
)
```

Esto genera una hoja de preparación para el DM con:
- Quests activas y su estado actual
- NPCs relevantes (vivos, sus ubicaciones, motivaciones)
- Advertencias de reputación de facciones
- Consecuencias pendientes de sesiones anteriores
- Tablas aleatorias para improvisación

**Durante la sesión:**
- Usá los actos generados como guía
- Referenciá encuentros, mapas y NPCs inline
- Trackeá las decisiones de los jugadores mentalmente o en notas

**Después de la sesión:**
```
grimorio_update_narrative_state(
  campaign_id="ciudad-sumergida",
  session_num=3,
  revealed_clues=["pista-llave-catedral", "pista-debilidad-vampiro"],
  dead_npcs=["npc-capitan-guardia"],
  completed_quests=["quest-encontrar-catedral"],
  key_decisions=["Los jugadores perdonaron a la hija del noble vampiro"],
  xp_awarded=450,
  loot_acquired=["Daga de Plata +1", "Llave de la Catedral"],
  session_summary="El grupo llegó a la catedral, derrotó al capitán de la guardia..."
)
```

#### 3. Mantener la Coherencia

**Cómo funciona la validación:**

Cada pieza de contenido se valida contra `canon.json` antes de guardarse:

```
grimorio_validate_canon(
  campaign_id="ciudad-sumergida",
  proposal={
    id: "act-3",
    type: "act",
    content: "...",
    entity_references: [
      { entity_id: "npc-capitan-guardia", location: "act_3" }
    ]
  }
)
```

**Ejemplo — Previniendo resurrecciones de NPCs:**

Si `npc-capitan-guardia` fue marcado como muerto en la sesión 2, el validador devuelve:
```json
{
  "status": "rejected",
  "issues": [
    {
      "rule": "npc_death_state",
      "severity": "critical",
      "message": "El NPC 'Capitán de la Guardia' está muerto (sesión 2) pero aparece en el Acto 3",
      "fix_suggestion": "Reemplazar con la Teniente Mara o usar una carta escrita"
    }
  ]
}
```

**Verificación de coherencia previa al PDF:**
```
grimorio_check_consistency(campaign_id="ciudad-sumergida")
```

Esto valida toda la campaña antes de la compilación — verificando NPCs muertos que aparecen vivos, contradicciones de lore, entidades faltantes y problemas de timeline.

#### 4. Sistema de Consecuencias

Después de cada sesión, evaluá qué significan las acciones de los jugadores para el mundo:

```
grimorio_evaluate_consequences(
  campaign_id="ciudad-sumergida",
  trigger_decisions=["Los jugadores mataron al Lord Vex, el noble"]
)
```

**Ejemplo — Si los jugadores mataron a un noble:**
- **Inmediato:** Su hija jura venganza (nueva quest)
- **Facción:** La reputación de la Casa Vex cae a Hostil
- **Político:** Vacío de poder — otros nobles compiten por su territorio
- **Económico:** Las rutas comerciales a través de su distrito se vuelven peligrosas
- **Militar:** Sus guardias se disuelven o se unen a grupos mercenarios

**Cambios de reputación de facciones:**
```
grimorio_update_faction_reputation(
  campaign="ciudad-sumergida",
  faction="casa-vex",
  party="default",
  delta=-30,
  reason="Los jugadores mataron a Lord Vex"
)
```

Esto se propaga a facciones aliadas (los aliados de la Casa Vex también caen) y puede desencadenar nuevas consecuencias.

#### 5. Herramientas de DM

**Hoja de Preparación de Sesión (`generate_session_prep`):**
- Lista todas las quests activas con objetivos actuales
- Muestra NPCs relevantes y su estado/ubicación actual
- Advierte sobre problemas de reputación de facciones
- Incluye tablas aleatorias para improvisación (encuentros, rumores, clima)
- Nota consecuencias pendientes que pueden activarse

**Diagrama de Flujo de Campaña (`generate_flowchart`):**
```
grimorio_generate_flowchart(
  campaign="ciudad-sumergida",
  title="Campaña Ciudad Sumergida"
)
```

Genera un diagrama de flujo visual (diagrama Mermaid + SVG) mostrando:
- Todos los actos y sus puntos de decisión
- Ramas de quests y consecuencias
- Mapa de relaciones entre NPCs
- Visión general de facciones

**Tablas Aleatorias (`generate_random_tables`):**
```
grimorio_generate_random_tables(
  campaign="ciudad-sumergida",
  context="distrito de la catedral sumergida"
)
```

Genera tablas contextuales para improvisación:
- Encuentros aleatorios (apropiados a la ubicación y nivel)
- Rumores y conversaciones escuchadas
- Eventos ambientales
- Reacciones de NPCs basadas en reputación de facciones

**Handouts (`generate_handouts`):**
```
grimorio_generate_handouts(
  campaign="ciudad-sumergida",
  type="pista",
  subject="El Sello de la Catedral"
)
```

Crea tanto:
- **Versión para jugadores** — Lo que ven los jugadores (carta envejecida, nota críptica)
- **Versión para DM** — Contexto completo y secretos que los jugadores no saben

#### 6. Actualizar según Jugadores

**Trackeá decisiones en el estado narrativo:**

Después de cada sesión, actualizá el estado con decisiones clave:
```
grimorio_update_narrative_state(
  campaign_id="ciudad-sumergida",
  session_num=3,
  key_decisions=[
    "Los jugadores se aliaron con los Tritones en vez de los Vampiros",
    "Los jugadores destruyeron el Cristal de Sangre en vez de usarlo"
  ]
)
```

**Adaptá la campaña con consecuencias:**

La próxima vez que generés contenido, las consecuencias de decisiones previas afectarán:
- Qué NPCs están disponibles (aliados vs enemigos)
- Reacciones de facciones hacia el grupo
- Quests y caminos disponibles
- Tablas de encuentros aleatorios

**Actualizá reputación de facciones después de elecciones diplomáticas:**
```
grimorio_update_faction_reputation(
  campaign="ciudad-sumergida",
  faction="alianza-triton",
  party="default",
  delta=+20,
  reason="Los jugadores ayudaron a rescatar prisioneros tritones"
)
```

Esto actualiza automáticamente facciones aliadas también (ej. los sacerdotes del Templo del Mar también mejoran).

#### 7. Compilar PDF Final

**Cuándo compilar:**
- Después de la generación inicial de campaña (libro completo)
- Antes de cada sesión (referencia rápida con estado actual)
- Después de actualizaciones importantes de contenido (nuevos actos, NPCs, imágenes)

**Cómo compilar:**
```
grimorio_compile_pdf(
  campaign="ciudad-sumergida",
  title="La Ciudad Sumergida"
)
```

**Qué se incluye:**
1. Portada (con arte de portada generado por IA)
2. Índice
3. Guía de Sesión Cero (si fue generada)
4. Diagrama de flujo de campaña (si fue generado)
5. Lore y Ambientación del Mundo
6. Todos los Actos (con NPCs, encuentros, mapas y escenas embebidos)
7. Apéndice A: NPCs y Facciones
8. Apéndice B: Bestiario
9. Apéndice C: Encuentros
10. Apéndice D: Mapas
11. Apéndice E: Faction Tracker
12. Apéndice F: Adventure Roster

Todas las imágenes (mapas SVG, PNGs generados por IA, divisores) se embeben automáticamente.

---

### Desarrollo

#### Compilar manualmente

```bash
go build -o grimorio ./cmd/grimorio
```

#### Ejecutar servidor MCP (standalone)

```bash
./grimorio
```

El servidor corre en modo stdio (stdin/stdout) para comunicación MCP.

---

## License

MIT
