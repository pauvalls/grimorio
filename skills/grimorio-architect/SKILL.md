---
name: grimorio-architect
version: "5.3.0"
description: Orchestrate end-to-end D&D 5e campaign generation via delegate pattern. Chapter-sequential workflow with prologue, sequential parts, consolidation phase, and monster design engine.
---

# Grimorio Architect Skill

**Type:** orchestrator
**Domain:** D&D 5e campaign generation
**Scope:** End-to-end campaign creation via delegate pattern
**Version:** 5.3.0

> This is the **detailed reference**. The shorter agent prompt at
> `agents/grimorio-architect.md` is the version OpenCode actually loads —
> the frontmatter `version: 5.3.0` on both files is the source of truth.

**What changed in v5.3 vs v5.0:**
- **v5.1**: prologue chapter + 7-part sequential chapter generation + bilingual validators + WotC word counts
- **v5.2**: consolidation phase (Macro-Phase 4.5) — `detect_inconsistencies`, `consolidate_campaign`, `resolve_ambiguity`, `regenerate_index`, `verify_campaign_freshness`
- **v5.3**: monster design engine — `validate_monster`, `suggest_monster_cr`, `audit_monster_cr` (DMG cap. 9 + MM 2025). Bestiary generation now BLOCKING-gated by CR/VD validation.

---

## 0. Language Intake (Mandatory, First Question)

Before any other question, ask the user their preferred session language and
default to English if they skip:

> **¿En qué idioma prefieres jugar? / What language do you prefer to play in? [es/en]**

- Store the answer as `session_language` in your conversation state.
- If the user does not answer, set `session_language = "en"`.
- Every `delegate(agent=..., prompt=...)` call you make MUST prepend the
  chosen language to its prompt body using the `LANG:` preamble block. The
  format is a single header line followed by a blank line:

  ```
  LANG: en

  <original prompt body>
  ```

  Sub-agent skills (e.g. `grimorio-npc`, `grimorio-chapters`) read the `LANG:`
  line from the prompt preamble and render their content in the requested
  language. If the preamble is missing, sub-agents default to English.

---

## Architecture: Chapter-Sequential (v5.0)

The v5.0 workflow is **chapter-sequential** — chapters are the campaign's
skeleton, and everything else (NPCs, bestiary, encounters, quests) anchors
to them. This produces more coherent campaigns than the legacy batch pipeline.

```
┌────────────────────────────────────────────────────────────────────┐
│ MACRO-PHASE 1: FOUNDATION                                          │
│   create_campaign → bible → name pool → introduction + setting    │
│   GATE: validate_canon                                             │
├────────────────────────────────────────────────────────────────────┤
│ MACRO-PHASE 2: CHAPTERS (sequential, 1 at a time)                  │
│   For each chapter:                                                │
│     save_chapter → maps + dividers → narrative-custodian → WotC    │
├────────────────────────────────────────────────────────────────────┤
│ MACRO-PHASE 3: BESTIARY & CHARACTERS (parallel)                    │
│   NPCs + bestiary + encounters + quests + characters + appendices  │
│   GATE: narrative-custodian                                        │
├────────────────────────────────────────────────────────────────────┤
│ MACRO-PHASE 4: ART & LIVING WORLD (parallel)                       │
│   Images → markdown refs → random tables → handouts → factions     │
│   → treasure → health dashboard                                    │
├────────────────────────────────────────────────────────────────────┤
│ MACRO-PHASE 5: EXPORT & DELIVER                                    │
│   grimorio validate (CLI) → export_campaign → final report         │
└────────────────────────────────────────────────────────────────────┘
```

**Why chapters first?**
- Chapters define the **spatial and narrative skeleton**.
- NPCs and monsters anchor to specific chapters and areas.
- Encounters live in areas; quests unfold across chapters.
- Treasure appears in hoards tied to specific encounters.
- Cross-references resolve correctly because the destinations exist.

**Backwards compatibility:** Legacy `areas/` campaigns still work. Use the
`migrate-areas-to-chapters` CLI to upgrade.

---

## Phase 1: Gather Requirements (Interactive)

Ask the user these questions **ONE AT A TIME** (do not batch):

1. **Campaign name** (kebab-case, e.g. "sunken-city")
2. **Type** — one-shot or full campaign?
3. **Brief idea** — 2-3 sentence story description
4. **Player level** — 1-3, 4-6, 7-10, 11-15, 16-20
5. **Tone** — heroic, dark, humorous, political intrigue, horror, mystery
6. **Duration** — one-shot, 3-5 sessions, long campaign

**Optional templates** — offer 5 pre-filled archetypes:
- Urban Fantasy (mystery tone, urban setting)
- Gothic Horror (horror tone, dark setting)
- Maritime Adventure (heroic tone, nautical setting)
- Dungeon Crawl (grim tone, dungeon setting)
- Political Intrigue (political tone, urban setting)

If the user picks a template, populate the questions with template defaults
and let them confirm or override.

---

## Phase 2: Initial Configuration Check

```bash
./scripts/validate-opencode.sh --check=all
```

**Critical checks:**
- ✅ All `grimorio-*` agents defined in `opencode.json`
- ✅ Templates exist in `internal/compiler/templates/`
- ✅ `grimorio` binary is built and runnable
- ✅ SDD config present (`delivery_strategy`, `chain_strategy`)

If any validation fails, **DO NOT proceed**. Report to the user and wait.

---

## Macro-Phase 1: Foundation

### Step 1.1: Create Campaign Structure

```
MCP: create_campaign(name="{campaign-name}", setting="{setting}", title="{title}")
```

Save the returned `campaign_path`.

### Step 1.2: Generate Adventure Bible (Canon)

```
MCP: generate_adventure_bible(
  campaign_id="{campaign-name}",
  name="{campaign-title}",
  brief_description="{brief}",
  level_range="{level-range}",
  tone="{tone}",
  setting_type="{setting-type}",
  themes=["theme1", "theme2"],
  villain_type="{villain-type}",
  mcguffin_type="{mcguffin-type}"
)
```

This creates `canon.json` — the single source of truth.

### Step 1.3: Generate Name Pool

Generate culturally-consistent name pools BEFORE delegating content. This
ensures cross-references use exact names and prevents duplicates.

```
MCP: generate_names(category="npc", style="{style}", count=20)
MCP: generate_names(category="monster", style="{style}", count=15)
MCP: generate_names(category="character", style="{style}", count=10)
MCP: generate_names(category="city", style="{style}", count=10)
MCP: generate_names(category="faction", style="{style}", count=10)
MCP: generate_names(category="tavern", style="{style}", count=8)
MCP: generate_names(category="item", style="{style}", count=12)
```

**Phase mapping for injection:**
- Macro-Phase 2 (Chapters): inject `city` and `tavern` names
- Macro-Phase 3 (NPCs + Bestiary): inject `npc` and `monster` names
- Macro-Phase 3 (Characters): inject `character` names
- Macro-Phase 4 (Appendices): inject `item` and `faction` names

**CRITICAL:** When delegating content creation, prepend the name pool:

```
PRE-GENERATED NAME POOL (USE THESE EXACT NAMES — do NOT invent alternatives):
- NPCs: {name1}, {name2}, {name3} ...
- Monsters: {name1}, {name2}, {name3} ...
...

INSTRUCTION: Use names from the pool above. Do NOT create new names.
```

### Step 1.4: Introduction (parallel with 1.5, 1.6)

```
delegate(agent="grimorio-introduction", prompt="LANG: en

Generate INTRODUCTION for campaign '{campaign-name}' at {campaign-path}.

Read canon.json first. This is a {duration} for levels {level-range}. Tone: {tone}.
Brief: {brief-description}

Template: internal/compiler/templates/introduction.md.tmpl")
```

### Step 1.5: Setting Guide (DM-only, parallel)

```
delegate(agent="grimorio-setting-guide", prompt="LANG: en

Generate SETTING GUIDE for '{campaign-name}' at {campaign-path}.

Read canon.json. DM-only with spoilers.
Template: internal/compiler/templates/setting-guide.md.tmpl

Include: Geography, History, Culture, Factions, Secrets")
```

### Step 1.6: Lore (parallel)

```
delegate(agent="grimorio-lore", prompt="LANG: en

Generate LORE for '{campaign-name}' at {campaign-path}.

Setting: {setting}
Tone: {tone}
Level: {level-range}

CRITICAL: Read template: internal/compiler/templates/lore.md.tmpl")
```

### GATE 1: Canon Validation

```
delegate(agent="grimorio-narrative-custodian", prompt="LANG: en

Validate Foundation for '{campaign-name}' at {campaign-path}.

Read canon.json, introduction.md, setting-guide.md, lore.md.

Check:
- World rule violations
- Missing entities
- Tone consistency
- Faction alignment

MCP: validate_canon

Return: status (approved/rejected) + specific fixes")
```

**BLOCKING:** If rejected after 2 retries, STOP and report failure.

---

## Macro-Phase 2: Chapters (Sequential)

This is the core of v5.0. Each chapter is generated and validated BEFORE
moving to the next. This produces self-contained, playable chapters.

### For each chapter (typically 3):

#### Step 2.x.1: Save Chapter

```
delegate(agent="grimorio-chapters", prompt="LANG: en

Generate CHAPTER {N} for '{campaign-name}' at {campaign-path}.

Campaign type: {duration}. Levels: {level-range}. Tone: {tone}.
Brief: {brief-description}

PRE-GENERATED NAME POOL:
- Cities: {name1}, {name2} ...
- Taverns: {name1}, {name2} ...

CRITICAL:
1. Read ALL source files first:
   - canon.json
   - lore.md
   - introduction.md
   - setting-guide.md
   - chapters/chapter-{N-1}/*.md (if exists)

2. Read template: internal/compiler/templates/areas.md.tmpl (now chapter format)

3. Generate 10-15 numbered areas for Chapter {N}

4. WotC STANDARDS (MANDATORY):
   - Boxed text: 100-600 words per area (grep: '^>>')
   - Character hooks: ≥2 per area (tie to backgrounds/classes)
   - Developments: ≥3 per area with 100% recovery paths
   - Running guidance: 150-400 words per section (5 subsections)
   - Sidebars: ≥1 per act (grep: '^> #####')
   - Cross-references: Use EXACT names from canon.json
   - Read-aloud: 2nd person present tense
   - Treasure: per-chapter hoard rolled with `generate_treasure`

5. Reference NPCs by EXACT name from canon.json

MCP: save_chapter(
  campaign='{campaign-name}',
  chapter_number={N},
  title='{chapter-title}',
  content={generated content}
)")
```

#### Step 2.x.2: Generate Maps & Dividers (per chapter)

```
delegate(agent="grimorio-cartographer", prompt="LANG: en

Generate SVG assets for CHAPTER {N} of '{campaign-name}' at {campaign-path}.

Read: chapters/chapter-{N}/*.md (extract location names from areas)

Generate:
- Battle map for EACH major location (style: dungeon/landscape/city)
- 1 ornate divider for chapter {N}

MCP: generate_map (one per location)
MCP: generate_divider (one per chapter)")
```

#### GATE 2.x: Per-Chapter Validation (BLOCKING)

```
delegate(agent="grimorio-narrative-custodian", prompt="LANG: en

Validate CHAPTER {N} for '{campaign-name}'.

MCP: validate_canon for chapter-{N}
MCP: check_consistency scope=wotc

Check:
- NPC consistency (dead NPCs in earlier chapters don't reappear alive)
- Timeline coherence
- Location consistency
- Boxed text word count (100-600)
- Character hooks (≥2 per area)
- Developments (≥3 with 100% recovery paths)
- Running guidance (150-400 words)
- Sidebars (≥1 per chapter)
- Cross-references resolve

Return: status + specific fixes per area")
```

**BLOCKING:** If rejected after 2 retries, STOP and report failure. DO NOT
proceed to the next chapter.

---

## Macro-Phase 3: Bestiary & Characters (Parallel)

Now that chapters exist, generate the creatures and characters that inhabit them.

### Step 3.1: NPCs (anchored to chapters)

```
delegate(agent="grimorio-npc", prompt="LANG: en

Generate NPCs for '{campaign-name}' at {campaign-path}.

Setting: {setting}
Tone: {tone}
Level: {level-range}

PRE-GENERATED NAME POOL:
- NPCs: {name1}, {name2}, {name3} ...

CRITICAL:
1. Read template: internal/compiler/templates/npc.md.tmpl
2. Follow WotC standard: 500-800 words per major NPC
3. Include 6 required sections: Appearance, Personality, Motivation, Secret, Quest Involvement, Connections
4. Include faction reputation system with propagation
5. Anchor each NPC to a SPECIFIC chapter and area (cite chapter-N/area-MM)
6. Use exact names from the pool above

MCP: save_npcs(campaign='{campaign-name}', content={read npcs_and_factions.md})")
```

### Step 3.2: Bestiary (anchored to chapters)

```
delegate(agent="grimorio-bestiary", prompt="LANG: en

Generate BESTIARY for '{campaign-name}' at {campaign-path}.

Setting: {setting}
Tone: {tone}
Level: {level-range}

PRE-GENERATED NAME POOL:
- Monsters: {name1}, {name2}, {name3} ...

CRITICAL:
1. Read template: internal/compiler/templates/monster.md.tmpl
2. Include complete stat blocks (D&D 5e SRD format)
3. Include tactical notes and lore per monster
4. Specify habitat chapter + area for each monster
5. Use exact names from the pool above

MCP: save_bestiary(campaign='{campaign-name}', content={read bestiary.md})")
```

### Step 3.3: Encounters (per chapter, anchored to areas)

```
delegate(agent="grimorio-encounters", prompt="LANG: en

Generate ENCOUNTERS for '{campaign-name}' at {campaign-path}.

Setting: {setting}
Tone: {tone}
Level: {level-range}

CRITICAL:
1. Read template: internal/compiler/templates/encounter.md.tmpl
2. Read chapters/chapter-*.md to know which creatures and areas exist
3. Balance CR for the level range
4. Include treasure, XP, scaling
5. Use MCP generate_treasure for hoard encounters
6. Reference creatures by EXACT name from bestiary.md

MCP: save_encounters(campaign='{campaign-name}', content={read encounters.md})")
```

### Step 3.4: Quests (parallel with characters)

```
delegate(agent="grimorio-quests", prompt="LANG: en

Generate QUESTS for '{campaign-name}' at {campaign-path}.

Setting: {setting}
Tone: {tone}
Brief: {brief-description}

Include: Main quest, side quests, personal quests per character type
Use MCP create_personal_quest for each PC build")
```

### Step 3.5: Pre-Generated Characters (parallel)

```
delegate(agent="grimorio-characters", prompt="LANG: en

Generate PRE-GENERATED CHARACTERS for '{campaign-name}' at {campaign-path}.

Setting: {setting}
Tone: {tone}
Level: {level-range}

PRE-GENERATED NAME POOL:
- Characters: {name1}, {name2} ...

Include: Backstory, bonds, flaws, equipment, balanced builds
MCP: generate_character_hooks(campaign='{campaign-name}')
MCP: save_characters(campaign='{campaign-name}', characters={list})")
```

### Step 3.6: Appendices (consolidated reference)

```
delegate(agent="grimorio-appendices", prompt="LANG: en

Generate APPENDICES for '{campaign-name}' at {campaign-path}.

Read ALL source files. Compile:
- Magic items (use names from item pool)
- NPC full stat blocks (consolidated from bestiary + npcs)
- Handouts (callouts, rumors, sidebars)
- Reference tables (loot, weather, encounters)
- Maps index

Template: internal/compiler/templates/appendices.md.tmpl

MCP: save_appendices(campaign='{campaign-name}', content={read appendices.md})")
```

### GATE 3: Cross-Reference Validation

```
delegate(agent="grimorio-narrative-custodian", prompt="LANG: en

Validate Bestiary & Characters for '{campaign-name}'.

MCP: validate_canon
MCP: check_consistency scope=full

Check:
- Every NPC reference in chapters resolves to npcs_and_factions.md
- Every monster reference resolves to bestiary.md
- Every encounter uses creatures from bestiary
- Every quest involves NPCs from npcs_and_factions.md
- Every personal quest ties to a specific character

Return: status + specific fixes")
```

**BLOCKING:** If rejected after 2 retries, STOP and report failure.

---

## Macro-Phase 4: Art & Living World (Parallel)

### Step 4.1: Artist Batch Spec

```
delegate(agent="grimorio-artist", prompt="LANG: en

Prepare image batch spec for '{campaign-name}' at {campaign-path}.

Read:
- npcs/npcs_and_factions.md (extract ALL major NPCs)
- bestiary/bestiary.md (extract ALL monsters)
- chapters/chapter-*/*.md (extract ALL [SCENE: ...] placeholders)
- lore.md (extract setting for cover)

Batch spec MUST include:
1. cover-art.png — cover image (type: cover) — FIRST entry
2. npc-[name].png — ONE portrait per major NPC (type: portrait)
3. scene-[act]-[description].png — ONE per [SCENE: ...] (type: scene)
4. monster-[name].png — ONE per key monster (type: illustration)

Save to: assets/batch-spec.json")
```

### Step 4.2: Generate AI Images (Sequential, 3s delay)

```
Read: assets/batch-spec.json

FOR each image in batch-spec.json:
  MCP: generate_image(
    campaign="{campaign-name}",
    filename="{filename}",
    prompt="{prompt}",
    type="{type}",
    force_regenerate=false  # set true to bypass cache
  )
  // Automatic 3s delay between each

Verify: ls {campaign-path}/assets/*.png
```

**Note**: The image cache (SHA-256 LRU+disk) makes re-runs instant. Set
`force_regenerate=true` only when you need a fresh result.

### Step 4.3: Update Markdown References

```
delegate(agent="grimorio-artist", prompt="LANG: en

Update ALL image references for '{campaign-name}' at {campaign-path}.

List: ls {campaign-path}/assets/*.png

For EACH PNG:
1. cover-*.png → README.md at top: ![Cover](assets/filename.png)
2. npc-*.png → npcs/npcs_and_factions.md in matching NPC section
3. scene-*.png → chapters/chapter-*/chapter-*.md, replacing [SCENE: ...] placeholders
4. monster-*.png → bestiary/bestiary.md in matching monster section

CRITICAL: Every PNG MUST be referenced in at least one markdown file")
```

### Step 4.4: Living World Tools (parallel)

```
delegate(agent="grimorio-integrator", prompt="LANG: en

Generate Living World content for '{campaign-name}' at {campaign-path}.

MCP: generate_random_tables(
  campaign='{campaign-name}',
  table_type='encounter',
  location_hint='{primary-location}'
)
MCP: generate_random_tables(
  campaign='{campaign-name}',
  table_type='rumor'
)
MCP: generate_handouts(
  campaign='{campaign-name}',
  handout_type='summary',
  content_refs=['npcs', 'quests']
)
MCP: generate_handouts(
  campaign='{campaign-name}',
  handout_type='quest',
  content_refs=['main-quest']
)
MCP: generate_treasure(
  campaign='{campaign-name}',
  type='hoard',
  cr_or_tier={CR}
)
MCP: process_consistency_gate(batch_id='living-world', proposals=[...])")
```

### Step 4.5: Campaign Health Dashboard

```
MCP: campaign_health_dashboard(campaign_id='{campaign-name}')
```

The dashboard scores the campaign 0-100 on six axes. A score below 70
indicates areas needing attention.

### GATE 4: Pre-Export Validation

```
delegate(agent="grimorio-narrative-custodian", prompt="LANG: en

Run PRE-EXPORT validation for '{campaign-name}'.

MCP: check_consistency scope=full
MCP: evaluate_consequences
MCP: campaign_health_dashboard

Validate:
- Cross-act consistency (NPCs dead in earlier chapters don't reappear alive)
- Quest closure (all quests have resolution or continuation)
- Lore coherence (no contradictions)
- Encounter balance (all CRs appropriate for level)
- Treasure balance (loot appropriate for level and economy)
- Faction consistency (reputation changes tracked)
- State completeness (narrative_state.json reflects all content)
- Health score ≥ 70 (recommended)

Return: status + dashboard scores + specific fixes")
```

---

## Macro-Phase 5: Export & Deliver

### Step 5.1: `grimorio validate` CLI (BLOCKING)

```bash
grimorio validate {campaign-name} --scope=all
```

Expected output: `VALIDATION PASSED` with exit code 0.

**Scopes:**
- `--scope=structure` — directory + file structure only
- `--scope=wotc` — WotC format only (boxed text, hooks, developments, sidebars)
- `--scope=references` — cross-references only
- `--scope=all` (default) — all checks

**BLOCKING:** If exit code 1, fix issues and re-run. If exit code 2, usage
problem (campaign name typo, etc.).

### Step 5.2: Compile & Export

Choose the user's preferred export format:

```bash
# PDF (default, most styled)
grimorio export_campaign --campaign {campaign-name} --format=pdf

# Markdown (for version control, Obsidian)
grimorio export_campaign --campaign {campaign-name} --format=markdown

# EPUB (for e-readers)
grimorio export_campaign --campaign {campaign-name} --format=epub
```

Verify the file exists:

```bash
ls -lh {campaign-path}/campaign.{pdf,md,epub}
```

### Step 5.3: DM Experience Tools (optional)

```bash
MCP: generate_session_prep(campaign_id='{campaign-name}', session_num=1, with_scenarios=true)
MCP: generate_flowchart(campaign_id='{campaign-name}', detail_level='act')
```

### Step 5.4: Final Report

```markdown
## Campaign "{title}" Complete

**Artifacts:**
- PDF: {path}/campaign.pdf
- Markdown: {path}/campaign.md
- EPUB: {path}/campaign.epub

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

---

## Templates & WotC Standards

Each sub-agent MUST read its template before generating:

| Agent | Template | WotC Standard |
|-------|----------|---------------|
| `grimorio-chapters` | `internal/compiler/templates/areas.md.tmpl` (chapter format) | Boxed text 100-600w, ≥2 hooks/area, ≥3 developments w/ recovery, running guidance 150-400w, ≥1 sidebar/act |
| `grimorio-npc` | `internal/compiler/templates/npc.md.tmpl` | 500-800w/NPC, 6 sections, faction reputation |
| `grimorio-bestiary` | `internal/compiler/templates/monster.md.tmpl` | Full stat blocks, tactics, lore |
| `grimorio-encounters` | `internal/compiler/templates/encounter.md.tmpl` | CR balance, rewards, scaling |
| `grimorio-lore` | `internal/compiler/templates/lore.md.tmpl` | World history, atmosphere |
| `grimorio-maps` | `internal/compiler/templates/map.md.tmpl` | Location descriptions, zones |
| `grimorio-setting-guide` | `internal/compiler/templates/setting-guide.md.tmpl` | DM-only, spoilers, geography, factions |
| `grimorio-appendices` | `internal/compiler/templates/appendices.md.tmpl` | Consolidated reference material |
| `grimorio-introduction` | `internal/compiler/templates/introduction.md.tmpl` | DM hooks, expectations |

---

## MCP Tools (v5.0)

### Creation
- `create_campaign(name, setting, title)` — Create campaign directory structure
- `generate_adventure_bible(...)` — Generate canon.json
- `generate_names(category, style, count, seed?)` — Generate names by category

### Save
- `save_introduction(campaign, content)`
- `save_setting_guide(campaign, content)`
- `save_lore(campaign, content)`
- `save_chapter(campaign, chapter_number, title, content)` — **v5.0 preferred**
- `save_areas(campaign, chapter_number, title, content)` — Legacy
- `save_npcs(campaign, content)`
- `save_bestiary(campaign, content)`
- `save_encounters(campaign, content)`
- `save_maps(campaign, content)`
- `save_quests(campaign, content)`
- `save_characters(campaign, characters)`
- `save_appendices(campaign, content)`

### Assets
- `generate_image(campaign, filename, prompt, type, force_regenerate?)`
- `generate_map(campaign, filename, rooms, style, labels)`
- `generate_divider(campaign, filename, style, width)`
- `generate_character_hooks(campaign)`
- `generate_random_tables(campaign, table_type, location_hint?, level_range?, party_size?)`
- `generate_handouts(campaign, handout_type, content_refs, version?)`
- `generate_treasure(campaign, type, cr_or_tier)` — **v5.0**
- `generate_flowchart(campaign, detail_level)`
- `generate_session_prep(campaign, session_num, with_scenarios?)`

### Validation
- `validate_canon(campaign_id, proposal_id, proposal_type, content, faction_context?)`
- `check_consistency(campaign_id, scope)` — scope ∈ {full, lore_only, acts_only, npcs_only, quests_only}
- `process_consistency_gate(batch_id, proposals)`
- `evaluate_consequences(campaign_id)`

### State
- `update_narrative_state(campaign_id, session_num, ...)`
- `update_faction_reputation(campaign_id, faction_id, party_id, delta, reason)`
- `update_quest_status(campaign, quest_id, status, notes)`

### Quality (v5.0)
- `campaign_health_dashboard(campaign_id)` — 0-100 score across 6 axes
- `export_campaign(campaign, format)` — format ∈ {pdf, markdown, epub}

### Compilation
- `compile_pdf(campaign, title)`

---

## Common Pitfalls

### 1. Skip Chapters (❌)
Building NPCs/bestiary before chapters means they have nowhere to anchor.
Cross-references float and the narrative coherence gate fails.

### 2. Inline Content (❌)
Never write narrative content directly. Always `delegate` to a sub-agent
that reads its template.

### 3. Skip Validation (❌)
Every macro-phase has a gate. If it fails, fix and retry — do not proceed.

### 4. Invent Names (❌)
Use the pre-generated name pool. Cross-references must resolve.

### 5. Force `compile_pdf` After Failed `grimorio validate` (❌)
The CLI is a BLOCKING gate. If it exits with code 1, fix the issues first.

### 6. Ignore Health Dashboard (❌)
A score below 70 means the campaign needs work. Don't ship it.

---

## Output Format

### After Each Macro-Phase

```markdown
## Phase {N}: {Name} — {Complete | Failed}

{What was generated}
- {Item 1} ({count})
- {Item 2} ({count})

**Gates:**
- Narrative: ✅ PASS / ❌ FAIL
- WotC: ✅ PASS / ❌ FAIL

**Next:** Phase {N+1} — {Name}
```

### Final Report

```markdown
## Campaign "{title}" Complete

**Artifacts:**
- PDF: {path}/campaign.pdf
- Markdown: {path}/campaign.md
- EPUB: {path}/campaign.epub

**Generated:**
- Chapters: {n}
- NPCs: {n} | Monsters: {n} | Encounters: {n}
- Quests: {n} | Pre-gens: {n}
- AI Images: {n} | SVG Maps: {n}
- Handouts: {n} | Random tables: {n}

**Health Score:** {X}/100
**WotC Validation:** PASSED
**Status:** ✅ Success
```

---

## SDD Configuration

```json
{
  "delivery_strategy": "exception-ok",
  "chain_strategy": "stacked-to-main",
  "artifact_store": "engram"
}
```

**Use `/sdd-new`** for structural changes to Grimorio itself:
- New MCP tools
- Template changes
- Validation changes
- New sub-agents

**DO NOT use SDD for campaign generation** — use `/grimorio` directly.

---

## References

- **Templates:** `internal/compiler/templates/`
- **Validation CLI:** `grimorio validate {name} --scope=all`
- **Migration CLI:** `migrate-areas-to-chapters --campaign {name}`
- **Skill Registry:** `.atl/skill-registry.md`
- **SDD Skills:** `~/.config/opencode/skills/sdd-*/SKILL.md`
- **v5.0 Changelog:** see PR #9
- **Consolidator skill:** `skills/grimorio-consolidator/SKILL.md` (runs after macro-phase 4, before macro-phase 5)

---

## v5.1 Addendum: Sequential Chapters + Prologue

**Prologue (Chapter 0) is MANDATORY.** Generated before Chapter 1 with these parts:
- `save_chapter_part(0, "opener")` — prologue intro
- `save_chapter_part(0, "npcs")` — party-meeting NPCs
- `save_chapter_part(0, "encounters")` — social encounters
- `save_chapter_part(0, "areas-1")` — 3-5 social areas (tavern, road, event)
- `save_chapter_part(0, "closing")` — transition to Chapter 1
- `finalize_chapter(0, title="Prologue", is_prologue=true)`

**Chapters 1-N use 7 parts** (not monolithic `save_chapter`):
- `opener` → `general-features` → `npcs` → `encounters` → `areas-1` → `areas-2` → `closing` → `finalize_chapter`

**WotC word counts** (validated by bilingual validators):
- Area: 150-600 words
- Boxed text: 50-400 words
- Areas per chapter: 7-15
- Chapter total: 3000-16000 words
- Inline sub-features: `***Name.***` bold-italic
- "What's Next?": free narrative prose (2-3 paragraphs, 100-400 words)

---

## v5.2 Addendum: Consolidation Phase (Macro-Phase 4.5)

Runs after Art & Living World, before Export:

```
delegate(agent="grimorio-consolidator", prompt="LANG: en

Run consolidation for '{campaign-name}' at {campaign-path}.

  1. detect_inconsistencies(campaign='{campaign-name}')  # read-only scan
  2. consolidate_campaign(campaign='{campaign-name}', auto_fix=true)
     # safe fixes only: exact-duplicate deletion, markdown renames, INDEX link updates
  3. resolve_ambiguity(campaign='{campaign-name}', question_id=..., decision=...)
     # per open question
  4. regenerate_index(campaign='{campaign-name}')
     # INDEX.md with breadcrumbs and verified links
  5. verify_campaign_freshness(campaign='{campaign-name}')
     # compare campaign.md against sources

BLOCKING GATE: consolidation_report.clean == true
  (no critical issues, no open questions)")
```

---

## v5.3 Addendum: Monster Design Engine

Three MCP tools implement DMG 5e cap. 9 + MM 2025 + SRD 5.1 as an engine-level
guard. Reference: `docs/dnd-monster-design-rules.md` (authoritative).

### `validate_monster(markdown, campaign?)` and `validate_monster(monster_name, campaign)`

Validates a single monster stat block against:
- **Stat block format** (MM 2025): required fields, ordering, line conventions
- **Ability scores** (DMG cap. 9): point-buy limits vs CR
- **Hit points** (DMG cap. 9): Hit Dice by Size table
- **Armor Class** (DMG cap. 9): expected AC vs CR
- **Proficiency bonus** (DMG cap. 9): CR → PB table
- **Damage output / Offensive CR** (DMG cap. 9)
- **Defensive CR** (DMG cap. 9)

Returns a `ValidationResult` with `valid: bool` and a list of issues.

**Use after each creature in `save_bestiary` flow. BLOCKING on `valid: false` (max 2 retries).**

### `suggest_monster_cr(target_cr, concept?)`

Returns a balanced stat-block skeleton for a given target CR (0-30, including
sub-integers 0.125, 0.25, 0.5). Optionally biased by `concept` (e.g.
"fire-breathing dragon"). Format: markdown by default; pass `output="json"`
for machine-readable.

**Use during bestiary planning to seed novel creatures.**

### `audit_monster_cr(campaign)`

Scans the entire campaign bestiary. Returns:
- `summary`: total monsters, monsters by CR bucket, violations count
- `per_monster`: per-monster validation status with specific issues

**Use as the FINAL gate of Macro-Phase 3.2 (Bestiary) before Macro-Phase 4 (Art).**

### Integration in Macro-Phase 3.2

For each creature the bestiary agent drafts:
1. `validate_monster` on the draft (BLOCKING on failure — fix and retry, max 2)
2. Append to bestiary.md
3. After all creatures saved: `audit_monster_cr(campaign)` (BLOCKING on any CR/VD violation)
4. For novel creatures not in canon: `suggest_monster_cr` first → skeleton → flesh out → `validate_monster`
