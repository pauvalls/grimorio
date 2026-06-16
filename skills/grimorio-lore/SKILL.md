---
name: grimorio-lore
version: "1.0.0"
description: Generate world lore, backstory, history, and setting with narrative depth
---

# grimorio-lore — Lore Master

## Required Template

**BEFORE generating content, READ the template:**

```
get_template(type="lore")
```

The template defines the mandatory WotC format for lore and worldbuilding.

## Available Tools

**MCP Tools (USE to save content):**
- `save_lore` — Save lore
- `validate_canon` — Validate against canon.json
- `check_consistency` — Consistency check
- `process_consistency_gate` — Batch validation with auto-retry
- `get_template` — Get WotC template

**DO NOT use Write for creative content** — The agent's frontmatter no longer includes Write to enforce the use of MCP save tools.

## Mandatory Workflow

```
1. READ context:
   - canon.json (canonical facts, entities, world rules)

2. READ template:
   - get_template(type="lore")

3. GENERATE lore following the template:
   - General synopsis (narrative hook)
   - The world (geography, history, culture)
   - Central conflict (threat, stakeholders, PC role)
   - Themes and tone
   - Narrative inflection points (5-7 milestones)

4. VALIDATE before saving:
   - validate_canon() with entity_references
   - process_consistency_gate() for batch validation
   - Maximum 3 retries on failure

5. SAVE only if validation passes:
   - save_lore(campaign, content)

6. REPORT to the architect
```

## Mandatory WotC Format

```markdown
# Lore and Setting: {Campaign Name}

## General Synopsis

[2-3 paragraphs that hook the DM. Who the PCs are, where they are, what is happening, and why they should care. This is the campaign's elevator pitch.]

---

## The World

### Geography

[Description of the physical environment with atmospheric details. Climate, vegetation, architecture. Use the 5 senses.]

### Recent History

[What happened in the last weeks/months that triggered the current situation. Timeline of relevant events.]

### Culture and Society

[How people live, what they believe, what fears they have, what keeps them united or divided. Customs, traditions, social structure.]

---

## The Central Conflict

### The Threat

[Description of the villain, their motivations, their plan, and why they are a credible threat. The villain must have understandable motivation, they are not evil for the sake of it.]

### The Stakeholders

[Who else has interests in the conflict: potential allies, neutrals, secondary antagonists.]

### The Role of the Players

[Why the PCs are involved and what is expected of them. No "chosen ones" — they are ordinary people in extraordinary circumstances.]

---

## Themes and Tone

### Narrative Themes

- **{Theme 1}:** [Brief explanation of how it manifests]
- **{Theme 2}:** [Brief explanation]
- **{Theme 3}:** [Brief explanation]
- **{Theme 4}:** [Brief explanation]
- **{Theme 5}:** [Brief explanation]
- **{Theme 6}:** [Brief explanation]

### General Tone

[Heroic, dark, humorous, mysterious, etc. Consistent throughout the lore.]

---

## Narrative Inflection Points

[5-7 key moments that structure the story. They are NOT detailed scenes — they are narrative MILESTONES that guide the DM.]

1. **{Milestone 1}:** [Brief description of the inflection moment]
2. **{Milestone 2}:** [Brief description]
3. **{Milestone 3}:** [Brief description]
4. **{Milestone 4}:** [Brief description]
5. **{Milestone 5}:** [Brief description]
6. **{Milestone 6}:** [Brief description]
7. **{Milestone 7}:** [Brief description]

---

## World Rules

[Specific rules of this setting that affect gameplay. E.g., "Arcane magic is forbidden", "The dead cannot be resurrected", "The sun never sets".]

- **R-{001}:** [Specific rule]
- **R-{002}:** [Specific rule]
- **R-{003}:** [Specific rule]
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
        id: "lore-main",
        type: "lore",
        content: "Generated lore summary...",
        entity_references: [
          { entity_id: "fact-001", location: "lore" },
          { entity_id: "entity-001", location: "lore" }
        ]
      }
    )
    
    IF result.status == "approved":
        validation_passed = true
    ELSE:
        retry_count += 1
        Fix issues based on result.feedback
    
IF validation_passed:
    save_lore(campaign="{campaign_name}", content=...)
ELSE:
    Report failure: "Validation failed after 3 retries"
    DO NOT save content
```

## Pre-Save Checklist

- [ ] **General Synopsis:** 2-3 paragraphs, clear narrative hook
- [ ] **Geography:** Description with 5 senses, atmospheric details
- [ ] **Recent History:** Timeline of triggering events
- [ ] **Culture and Society:** Customs, beliefs, fears
- [ ] **The Threat:** Villain with understandable motivation, clear plan
- [ ] **Stakeholders:** Allies, neutrals, secondary antagonists
- [ ] **PC Role:** Why they are involved (not "chosen ones")
- [ ] **Themes:** 4-6 narrative themes with explanation
- [ ] **Tone:** Consistent throughout the document
- [ ] **Inflection Points:** 5-7 numbered milestones
- [ ] **World Rules:** Specific rules that affect gameplay
- [ ] **Show, don't Tell:** Sensory descriptions vs. abstract assertions
- [ ] **Hooks:** Each section gives the DM ideas to develop
- [ ] **Approximate Level:** Deadly threat but not invincible for level 1

## Cross-References Format

**MANDATORY use of markdown links:**

```markdown
❌ WRONG: The villain mentions in the NPCs
✅ RIGHT: The villain [Lord Blackthorn](npcs/npcs_and_factions.md#lord-blackthorn)

❌ WRONG: The main city
✅ RIGHT: The city of [Valdrift](maps/maps.md#valdrift)

❌ WRONG: The creature that haunts the forest
✅ RIGHT: The [Whispering Wraith](bestiary/bestiary.md#whispering-wraith)
```

## Writing Standards

### Showing vs. Telling

**❌ WRONG (Telling):**
> "The village is afraid."

**✅ RIGHT (Showing):**
> "Doors are barred before sunset. Garlic hangs in every window. Silence reigns in the square where children once played."

### Second Person Present

Use second person present for immersive descriptions:

**✅ RIGHT:**
> "You see the shadows lengthen. You hear footsteps in the distance. You feel cold air on the back of your neck."

### Consistent Tone

If the campaign is dark:
- ✅ Maintain somber atmosphere
- ❌ Don't insert jokes or bright elements

If the campaign is heroic:
- ✅ Maintain hope and possibility of triumph
- ❌ Don't make everything hopeless

## WotC Quality Validators

### ValidateWorldBuildingDepth
- ✅ Geography with sensory details (5 senses)
- ✅ Recent history with clear timeline
- ✅ Culture and society with specific customs

### ValidateConflictStructure
- ✅ Villain with understandable motivation
- ✅ Multiple stakeholders (not binary good/evil)
- ✅ PC role justified (not "chosen ones")

### ValidateNarrativePacing
- ✅ 4-6 narrative themes identified
- ✅ 5-7 numbered inflection points
- ✅ Consistent tone throughout the document

### ValidateGameplayIntegration
- ✅ World rules that affect gameplay
- ✅ Threat balanced for the level
- ✅ Hooks for DM development

## Error Handling

If validation fails:

1. **Analyze specific feedback** (e.g., "contradicts canon rule R-005")
2. **Fix concrete issues** (adjust lore to respect world rules)
3. **Re-validate** with corrected content
4. **Maximum 3 retries** — if it fails, abort and report

## Output to the Architect

```markdown
## Generated Lore: {campaign_name}

**Status:** ✅ Complete / ❌ Failed

**Content:**
- Synopsis: {word_count} words
- Geography: {word_count} words
- History: {word_count} words
- Culture: {word_count} words
- Conflict: {word_count} words
- Themes: {count} identified themes
- Inflection Points: {count} milestones

**Validation:**
- validate_canon: ✅ Passed
- process_consistency_gate: ✅ Passed
- ValidateWorldBuildingDepth: ✅ Passed

**Consistency:**
- World rules respected: {count}
- canon.json entities used: {count}
- Cross-references generated: {count}
```
