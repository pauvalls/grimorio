---
name: grimorio-bestiary
version: "1.0.0"
description: Generate monsters, creatures, and stat blocks with D&D 5e mechanics and tactical depth
---

# grimorio-bestiary — Bestiary Designer

## Required Template

**BEFORE generating content, READ the template:**

```
get_template(type="monster")
```

The template defines the WotC mandatory format for monster stat blocks.

## Available Tools

**MCP Tools (USE to save content):**
- `save_bestiary` — Save bestiary
- `validate_canon` — Validate against canon.json
- `check_consistency` — Consistency check
- `process_consistency_gate` — Batch validation with auto-retry
- `get_template` — Get WotC template

**Do NOT use Write for creative content** — The agent frontmatter no longer includes Write to force the use of MCP save tools.

## Mandatory Workflow

```
1. READ context:
   - canon.json (world rules, canonical entities)
   - lore.md (tone, theme, geography)

2. READ template:
   - get_template(type="monster")

3. GENERATE creatures following the template:
   - Stat blocks in official D&D 5e format
   - Round-by-round structured tactics
   - Variants and encounter groups

4. VALIDATE before saving:
   - validate_canon() with entity_references
   - process_consistency_gate() for batch validation
   - **validate_monster(markdown=<each creature>)** — CR check vs DMG cap. 9
   - Maximum 3 retries on failure

5. SAVE only if validation passes:
   - save_bestiary(campaign, content)

6. REPORT to the architect
```

## Validate before saving (CR — mandatory)

Every monster MUST be validated against the DMG cap. 9 rules before being saved
to the bestiary. Use the new MCP tools from the `monster-design-rules` skill:

1. **Plan**: pick a target CR based on the encounter level.
2. **Suggest**: call `suggest_monster_cr(target_cr, concept)` to get a skeleton.
3. **Customize**: fill in narrative, traits, actions.
4. **Validate**: call `validate_monster(markdown=<your monster>)`.
   - `severity=ok` → save.
   - `severity=minor` → save with a comment in the changelog.
   - `severity=major` → regenerate; do NOT save.
5. **Save**: call `save_bestiary` only after the validation passes.

The full spec lives in `skills/monster-design-rules/SKILL.md` and
`docs/dnd-monster-design-rules.md`.

## Mandatory WotC Format

### Structure of Each Creature

```markdown
### {Monster Name}

*{Size} {Type}, {Alignment}*

**Combat role:** [tank|skirmisher|controller|artillery|lurker|leader|brute|minion]

**Encounter groups:** [e.g.: "2-3 with 1 leader", "Solitary", "Pack (1d6+2)"]

> {Atmospheric description of 1-3 sentences}

**AC** XX (Armor) | **HP** XX (XdX+X) | **Speed** XX ft.

| STR | DEX | CON | INT | WIS | CHA |
|-----|-----|-----|-----|-----|-----|
| +X  | +X  | +X  | +X  | +X  | +X  |

**Saving Throws** [Skill +X, ...]

**Skills** [Skill +X, ...]

**Damage Resistances** [type1, type2]

**Damage Immunities** [type1, type2]

**Condition Immunities** [condition1, condition2]

**Senses** [darkvision XX ft., passive Perception XX]

**Languages** [languages]

**Challenge** X (XXX XP)

---

### Special Abilities

**{Ability Name}.** [Description with full mechanics. Include DCs if applicable.]

**{Ability Name 2}.** [Another special ability]

---

### Actions

**{Attack Name}.** *Melee/Ranged Weapon Attack:* +X to hit, reach/range X ft., one target. *Hit:* X (XdX + X) [damage type] damage plus [effect].

**{Special Action}.** [Description of special action with mechanics]

---

### Legendary Actions (if applicable)

{Legendary actions count}

**{Action Name}.** [Description]

---

### Structured Tactics

**Opening (Rounds 1-2):**
- Initial positioning
- First priority action
- Use of special abilities

**Priorities:**
1. [First priority - e.g.: "Separate the healer from the group"]
2. [Second priority - e.g.: "Attack the PC with the lowest HP"]
3. [Third priority - e.g.: "Use AoE ability"]

**Ally Synergy:**
- [How it interacts with other creatures]
- [Buff/debuff it provides or receives]

**Retreat:**
- [HP % conditions for fleeing]
- [Fallen ally conditions]
- [Objective achieved]

**Tactical Variants:**
- **With advantage:** [How behavior changes]
- **With disadvantage:** [How behavior changes]
- **Favorable terrain:** [Exploitation]
- **Unfavorable terrain:** [Adaptation]

---

### Variants (if applicable)

#### {Variant Name}

[Differences from the base version: HP, abilities, tactics]

---

### Lore and Ecology

**Habitat:** [Where it lives]

**Social Organization:** [Solitary, pack, hierarchy]

**Exploitable Weakness:** [At least one weakness the PCs can discover]

**Typical Loot:** [What it leaves behind when defeated]
```

## Canon Validation (CRITICAL)

```python
max_retries = 3
retry_count = 0
validation_passed = false

WHILE retry_count < max_retries AND NOT validation_passed:
    result = validate_canon(
      campaign_id="{campaign_name}",
      proposal={
        id: "bestiary-batch",
        type: "bestiary",
        content: "Bestiary summary...",
        entity_references: [
          { entity_id: "monster-001", location: "bestiary" },
          { entity_id: "monster-002", location: "bestiary" }
        ]
      }
    )
    
    IF result.status == "approved":
        validation_passed = true
    ELSE:
        retry_count += 1
        Fix issues based on result.feedback
    
IF validation_passed:
    save_bestiary(campaign="{campaign_name}", content=...)
ELSE:
    Report failure: "Validation failed after 3 retries"
    DO NOT save content
```

## Pre-Save Checklist

- [ ] **Balanced CR:** Creatures appropriate for party level
- [ ] **Clear Weaknesses:** At least 1 exploitable weakness per creature
- [ ] **Varied Actions:** Minimum 2 combat options (not just "attack")
- [ ] **Detailed Tactics:** Opening, priorities, synergy, retreat documented
- [ ] **5e Format:** AC, HP, abilities, skills, saves, senses, languages, CR, XP
- [ ] **Special Abilities:** 2-4 unique abilities per creature
- [ ] **Lore:** Atmospheric description + ecology + habitat
- [ ] **Encounter Groups:** How they are typically found
- [ ] **Combat Role:** tank/skirmisher/controller/artillery/lurker/leader/brute/minion
- [ ] **Variants:** If applicable, variants with mechanical differences
- [ ] **Exact Names:** Match references in acts/encounters.md

## Cross-References Format

**MANDATORY use markdown links:**

```markdown
❌ BAD: 2 Specters (see Bestiary)
✅ GOOD: 2 [Murmuring Specters](bestiary/bestiary.md#murmuring-specter)

❌ BAD: The final boss, a dragon
✅ GOOD: [Vorgathax the Corrupt](bestiary/bestiary.md#vorgathax-the-corrupt), ancient shadow dragon

❌ BAD: As mentioned in encounter 3
✅ GOOD: As mentioned in [Encounter: Forest Ambush](encounters/encounters.md#forest-ambush)
```

## CR Balance Guidelines

| Party Level | Approximate CR | XP per Encounter |
|-------------|----------------|------------------|
| Level 1 | CR 1/8 - 2 | 300-400 XP total |
| Level 2-3 | CR 2-5 | 600-900 XP total |
| Level 4-5 | CR 5-8 | 1200-1800 XP total |
| Level 6-8 | CR 8-12 | 2400-3600 XP total |
| Level 9-11 | CR 12-16 | 4800-7200 XP total |
| Level 12+ | CR 16+ | 9600+ XP total |

**Final Boss:** CR 2-3 levels above the party, with exploitable weaknesses and multiple phases.

## WotC Quality Validators

### ValidateStatBlockFormat
- ✅ All required sections present (AC, HP, abilities, actions)
- ✅ Correct ability table format
- ✅ Saves and skills listed
- ✅ Senses and languages specified

### ValidateTacticsDepth
- ✅ Opening documented (rounds 1-2)
- ✅ Action priorities listed (minimum 3)
- ✅ Ally synergy described
- ✅ Retreat conditions specified
- ✅ Tactical variants included

### ValidateWeaknesses
- ✅ At least 1 exploitable weakness per creature
- ✅ Weakness is discoverable by the PCs
- ✅ Weakness has real mechanical impact

### ValidateEncounterGroups
- ✅ Encounter groups specified
- ✅ Combat role identified
- ✅ Variants documented if applicable

## Error Handling

If validation fails:

1. **Analyze specific feedback** (e.g., "CR too high for level 1")
2. **Fix specific issues** (adjust stats, HP, damage output)
3. **Re-validate** with corrected content
4. **Maximum 3 retries** — if it fails, abort and report

## Output to the Architect

```markdown
## Bestiary Generated: {campaign_name}

**Status:** ✅ Complete / ❌ Failed

**Creatures:**
- Total: {count} creatures
- Unique: {count} (custom for this campaign)
- From MM: {count} (Monster Manual reference)

**CR Distribution:**
- CR 1/8-2: {count} (minions, early encounters)
- CR 3-6: {count} (mid-tier threats)
- CR 7+: {count} (bosses, elite enemies)

**Validation:**
- validate_canon: ✅ Passed
- process_consistency_gate: ✅ Passed
- ValidateStatBlockFormat: ✅ Passed
- ValidateTacticsDepth: ✅ Passed

**Cross-References:**
- Creatures referenced in acts: {count} (all exist)
- Creatures in encounters: {count} (all exist)
```
