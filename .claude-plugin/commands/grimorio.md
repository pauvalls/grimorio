# Generate a D&D 5e campaign or one-shot from the user's idea.

**Version:** 5.0.0 — Chapter-Sequential Pipeline

## EXECUTION MODE: Main Thread Orchestration

**You are the orchestrator.** Execute this workflow directly in the main thread using the `grimorio-architect` agent and MCP tools. Sub-agents are launched via `delegate` by the architect — you do NOT launch them from this thread.

---

## 0. Language Intake (Mandatory, First Question)

Before any other question:

> **¿En qué idioma prefieres jugar? / What language do you prefer to play in? [es/en]**

Default to `en` if the user skips. Store as `session_language`.

---

## 1. Gather Requirements (Interactive, one question at a time)

1. **Campaign name** (kebab-case, e.g. "sunken-city")
2. **Type** — one-shot or full campaign?
3. **Brief idea** — 2-3 sentence story description
4. **Player level** — 1-3, 4-6, 7-10, 11-15, 16-20
5. **Tone** — heroic, dark, humorous, political intrigue, horror, mystery
6. **Duration** — one-shot, 3-5 sessions, long campaign

**Optional**: Offer 5 campaign templates that pre-fill defaults:
- Urban Fantasy · Gothic Horror · Maritime Adventure · Dungeon Crawl · Political Intrigue

---

## 2. Workflow (5 Macro-Phases, chapter-sequential)

The campaign is built in 5 macro-phases. Each macro-phase has BLOCKING gates.
Report progress to the user after each macro-phase.

### Macro-Phase 1: Foundation (parallel where possible)

```
- create_campaign(name, setting, title)
- generate_adventure_bible(...) → canon.json
- generate_names (all 7 categories)
- save_introduction
- save_setting_guide
- save_lore
- GATE: validate_canon → approved (BLOCKING)
```

### Macro-Phase 2: Chapters (sequential, 1 at a time)

For each chapter (typically 3):
```
- save_chapter(chapter_number, title, content)
- generate_map + generate_divider for this chapter
- GATE: narrative-custodian validation (BLOCKING)
- GATE: WotC format validation (BLOCKING)
```

**Why chapters first?** Chapters define the spatial and narrative skeleton.
NPCs, bestiary, encounters, quests, and treasure all anchor to specific
chapters and areas. Building them after chapters ensures cross-references
resolve.

### Macro-Phase 3: Bestiary & Characters (parallel, anchored to chapters)

```
- save_npcs (anchored to specific chapter/area)
- save_bestiary (creatures tied to chapter habitats)
- save_encounters (per-chapter, with generate_treasure for hoards)
- save_quests (main + side + personal)
- save_characters (pre-generated PCs with generate_character_hooks)
- save_appendices (consolidated reference)
- GATE: cross-reference validation (BLOCKING)
```

### Macro-Phase 4: Art & Living World (parallel)

```
- grimorio-artist → batch spec (cover + NPCs + scenes + monsters)
- generate_image (sequential, 3s delay, force_regenerate=false to use cache)
- Update markdown references with generated images
- generate_random_tables (encounters, rumors, weather, treasure)
- generate_handouts (summary, quest, lore, faction)
- generate_treasure (per hoard, SRD-compliant)
- update_faction_reputation (initial setup)
- process_consistency_gate (living world batch)
- GATE: campaign_health_dashboard (score ≥ 70 recommended)
```

### Macro-Phase 5: Export & Deliver

```
- grimorio validate {name} --scope=all (BLOCKING CLI gate)
- export_campaign --format=pdf (default) | --format=markdown | --format=epub
- generate_session_prep + generate_flowchart (optional)
- Final report to user
```

---

## 3. Available MCP Tools (v5.0)

| Category | Tools |
|----------|-------|
| **Creation** | `create_campaign`, `generate_adventure_bible`, `generate_names` |
| **Save** | `save_introduction`, `save_setting_guide`, `save_lore`, `save_chapter`, `save_npcs`, `save_bestiary`, `save_encounters`, `save_maps`, `save_quests`, `save_characters`, `save_appendices` |
| **Assets** | `generate_image`, `generate_map`, `generate_divider`, `generate_flowchart`, `generate_random_tables`, `generate_handouts`, `generate_treasure`, `generate_session_prep` |
| **Validation** | `validate_canon`, `check_consistency`, `process_consistency_gate`, `evaluate_consequences` |
| **State** | `update_narrative_state`, `update_faction_reputation`, `update_quest_status` |
| **Quality (v5.0)** | `campaign_health_dashboard`, `export_campaign` |
| **Compilation** | `compile_pdf` |

---

## 4. Final Report

After all gates pass and export completes, report:

```markdown
## Campaign "{title}" Complete

**Artifacts:**
- PDF: {path}/campaign.pdf
- Markdown: {path}/campaign.md (if exported)
- EPUB: {path}/campaign.epub (if exported)

**Generated:**
- Chapters: {n}
- NPCs: {n} | Monsters: {n} | Encounters: {n}
- Quests: {n} | Pre-gens: {n}
- AI Images: {n} | SVG Maps: {n}
- Handouts: {n} | Random tables: {n}

**Validation:**
- `grimorio validate --scope=all`: PASSED
- Campaign Health: {X}/100
- WotC: PASSED

**Status:** ✅ Success
```

**DO NOT launch subagents from the command thread — the architect manages all delegation internally.**
