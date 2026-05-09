# Campaign Brief Description Support — Implementation Tasks

**Change ID:** campaign-brief-description  
**Status:** Ready for Implementation  
**Created:** 2026-05-09

---

## Overview

Add support for an optional `brief_description` field that allows users to provide a campaign idea/story premise when generating adventures. This brief will be used to generate richer, more targeted facts in the canon document.

---

## Task List

### Go Implementation Tasks

#### TASK-GO-001: Add `BriefDescription` field to `CampaignBrief` struct
**File:** `internal/domain/canon.go`  
**Line:** ~254 (CampaignBrief struct definition)

**Change:**
```go
// CampaignBrief represents the initial brief for campaign generation
type CampaignBrief struct {
    Name           string   `json:"name"`
    LevelRange     string   `json:"level_range"`
    Tone           string   `json:"tone"`
    SettingType    string   `json:"setting_type"`
    Themes         []string `json:"themes"`
    VillainType    string   `json:"villain_type"`
    McGuffinType   string   `json:"mcguffin_type"`
    BriefDescription string `json:"brief_description,omitempty"` // NEW FIELD
}
```

**Completion Criteria:**
- [ ] Field added with `omitempty` tag (optional field)
- [ ] Code compiles without errors
- [ ] Field is exported (capitalized)

---

#### TASK-GO-002: Add `brief_description` parameter to MCP tool definition
**File:** `internal/mcp/server.go`  
**Line:** ~248-260 (generate_adventure_bible tool registration)

**Change:**
```go
s.AddTool(mcp.NewTool("generate_adventure_bible",
    mcp.WithDescription("Generate the canon initial document for a campaign from a brief"),
    mcp.WithString("name", mcp.Required(), mcp.Description("Campaign name (kebab-case)")),
    mcp.WithString("level_range", mcp.Description("Level range (e.g., 1-10)")),
    mcp.WithString("tone", mcp.Description("Campaign tone (grim, whimsical, heroic, horror, political, mystery)")),
    mcp.WithString("setting_type", mcp.Description("Setting type (urban, wilderness, dungeon, maritime, planar)")),
    mcp.WithArray("themes", mcp.Description("Campaign themes")),
    mcp.WithString("villain_type", mcp.Description("Type of main villain")),
    mcp.WithString("mcguffin_type", mcp.Description("Type of McGuffin")),
    mcp.WithString("brief_description", mcp.Description("Optional: Campaign idea or story premise (e.g., 'Los PJs investigan traición en la corte vampírica')")), // NEW PARAMETER
), canonHandlers.HandleGenerateAdventureBible())
```

**Completion Criteria:**
- [ ] Parameter added as optional (not in `mcp.Required()`)
- [ ] Description is clear and provides examples
- [ ] Code compiles without errors

---

#### TASK-GO-003: Parse `brief_description` in handler
**File:** `internal/mcp/handlers/canon.go`  
**Line:** ~38-65 (HandleGenerateAdventureBible function)

**Change:**
```go
// Inside HandleGenerateAdventureBible, after parsing other fields:
brief := domain.CampaignBrief{
    Name:         getStringArg(args, "name"),
    LevelRange:   getStringArg(args, "level_range"),
    Tone:         getStringArg(args, "tone"),
    SettingType:  getStringArg(args, "setting_type"),
    VillainType:  getStringArg(args, "villain_type"),
    McGuffinType: getStringArg(args, "mcguffin_type"),
    BriefDescription: getStringArg(args, "brief_description"), // NEW FIELD
}
```

**Completion Criteria:**
- [ ] Field is parsed from arguments
- [ ] No validation required (optional field)
- [ ] Code compiles without errors

---

#### TASK-GO-004: Use brief description to generate richer facts
**File:** `internal/services/canon_service.go`  
**Line:** ~61-100 (InitializeCanon function)

**Change:**
```go
// Inside InitializeCanon, enhance the initial fact generation:
now := time.Now()
doc := &domain.CanonDocument{
    SchemaVersion: domain.SchemaVersionV2,
    CampaignID:    brief.Name,
    CreatedAt:     now,
    UpdatedAt:     now,
    Facts: []domain.CanonFact{
        {
            ID:        "fact-001",
            Category:  "lore",
            Statement: fmt.Sprintf("The campaign '%s' is set in a %s world with %s tone.", brief.Name, brief.SettingType, brief.Tone),
            Source:    "adventure_bible_v1",
            Immutable: true,
            CreatedAt: now,
        },
    },
    // ... rest of struct
}

// NEW: Add brief description as a fact if provided
if brief.BriefDescription != "" {
    doc.Facts = append(doc.Facts, domain.CanonFact{
        ID:        "fact-002",
        Category:  "premise",
        Statement: brief.BriefDescription,
        Source:    "user_brief",
        Immutable: false,
        CreatedAt: now,
    })
}
```

**Completion Criteria:**
- [ ] Brief description is added as a fact when provided
- [ ] Fact has appropriate category ("premise")
- [ ] Fact is mutable (Immutable: false) to allow narrative evolution
- [ ] Code compiles without errors
- [ ] Unit tests pass (if any exist for InitializeCanon)

---

#### TASK-GO-005: Add tests for brief_description (Optional)
**File:** `internal/services/canon_service_test.go` (create if doesn't exist)  
**Line:** N/A

**Change:**
```go
func TestInitializeCanon_WithBriefDescription(t *testing.T) {
    // Setup
    canonRepo := repository.NewInMemoryCanonRepository()
    stateRepo := repository.NewInMemoryNarrativeStateRepository()
    svc := services.NewCanonService(canonRepo, stateRepo)
    
    brief := domain.CampaignBrief{
        Name:             "test-campaign",
        LevelRange:       "1-3",
        Tone:             "dark",
        SettingType:      "urban",
        BriefDescription: "Los PJs investigan una conspiración vampírica en la corte",
    }
    
    // Execute
    doc, err := svc.InitializeCanon(context.Background(), brief)
    
    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    
    // Verify brief description was added as a fact
    found := false
    for _, fact := range doc.Facts {
        if fact.Category == "premise" && fact.Statement == brief.BriefDescription {
            found = true
            if fact.Source != "user_brief" {
                t.Errorf("expected source 'user_brief', got '%s'", fact.Source)
            }
            if fact.Immutable {
                t.Error("expected brief description fact to be mutable")
            }
            break
        }
    }
    
    if !found {
        t.Error("expected brief description to be added as a fact")
    }
}

func TestInitializeCanon_WithoutBriefDescription(t *testing.T) {
    // Setup
    canonRepo := repository.NewInMemoryCanonRepository()
    stateRepo := repository.NewInMemoryNarrativeStateRepository()
    svc := services.NewCanonService(canonRepo, stateRepo)
    
    brief := domain.CampaignBrief{
        Name:        "test-campaign",
        LevelRange:  "1-3",
        Tone:        "dark",
        SettingType: "urban",
        // No BriefDescription
    }
    
    // Execute
    doc, err := svc.InitializeCanon(context.Background(), brief)
    
    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    
    // Verify no premise fact was added
    for _, fact := range doc.Facts {
        if fact.Category == "premise" {
            t.Error("did not expect premise fact when brief description is empty")
        }
    }
}
```

**Completion Criteria:**
- [ ] Test file created
- [ ] Both tests pass
- [ ] Tests cover both cases: with and without brief description
- [ ] Code coverage maintained or improved

---

### Agent Documentation Tasks

#### TASK-AGENT-001: Phase 1 — Add campaign idea as first question
**File:** `agents/grimorio-architect.md`  
**Line:** ~114-125 (Phase 1: Gather Requirements)

**Change:**
```markdown
### Phase 1: Gather Requirements
Ask the user these questions ONE AT A TIME (interactively):
1. **Campaign idea / story premise:** ¿Qué historia querés contar? (e.g., "Los PJs investigan traición en la corte vampírica", "Viaje a través de tierras salvajes para destruir un artefacto maldito")
2. Campaign name? (kebab-case, e.g., "sunken-city")
3. One-shot or full campaign?
4. Player level range? (1-3, 4-6, 7-10, 11-15, 16-20)
5. Desired tone? (heroic, dark, humorous, political intrigue)
6. Duration? (one-shot, 3-5 sessions, long campaign)
```

**Completion Criteria:**
- [ ] Question added as FIRST question (before campaign name)
- [ ] Question includes examples in Spanish
- [ ] Question is clear and open-ended
- [ ] File validates as proper markdown

---

#### TASK-AGENT-002: Phase 2b — Pass brief_description to generate_adventure_bible
**File:** `agents/grimorio-architect.md`  
**Line:** ~127-142 (Phase 2b: Generate Adventure Bible)

**Change:**
```markdown
### Phase 2b: Generate Adventure Bible (Canon)
**CRITICAL:** Before any content is created, establish the canonical facts:

```
generate_adventure_bible(
  campaign_id="{campaign_name}",
  name="{campaign_title}",
  level_range="{level_range}",
  tone="{tone}",
  setting_type="{setting_type}",
  themes=["theme1", "theme2"],
  villain_type="{villain_type}",
  mcguffin_type="{mcguffin_type}",
  brief_description="{campaign_idea_from_phase_1}" // NEW PARAMETER
)
```

This creates `canon.json` — the single source of truth for the campaign.
```

**Completion Criteria:**
- [ ] Parameter added to tool call
- [ ] Parameter uses value from Phase 1
- [ ] File validates as proper markdown

---

#### TASK-AGENT-003: Batch 1-3 — Include Campaign Concept in delegations
**File:** `agents/grimorio-architect.md`  
**Line:** ~150-200 (Phase 3b: Batch 1 delegations)

**Change:**
```markdown
### Phase 3b: Batch 1 — Contenido Base (PARALLEL)
NPCs, Bestiary y Maps se generan con la premisa base de la campaña (tone, level, setting, campaign_idea):

**1. NPCs — Agent: grimorio-npc**
```
delegate(agent="grimorio-npc", prompt="Generate NPCS for campaign '{campaign_name}' at {campaign_path}.

Setting: {setting}
Tone: {tone}
Level: {level_range}
Campaign Concept: {campaign_idea_from_phase_1} // NEW CONTEXT

Create NPCs that support this story premise.")
```

**2. Bestiary — Agent: grimorio-bestiary**
```
delegate(agent="grimorio-bestiary", prompt="Generate BESTIARY for campaign '{campaign_name}' at {campaign_path}.

Setting: {setting}
Tone: {tone}
Level: {level_range}
Campaign Concept: {campaign_idea_from_phase_1} // NEW CONTEXT

Create monsters appropriate for this story.")
```

**3. Maps — Agent: grimorio-maps**
```
delegate(agent="grimorio-maps", prompt="Generate MAP DESCRIPTIONS for campaign '{campaign_name}' at {campaign_path}.

Setting: {setting}
Tone: {tone}
Campaign Concept: {campaign_idea_from_phase_1} // NEW CONTEXT

Create locations that support this narrative.")
```
```

**Completion Criteria:**
- [ ] All three delegations include campaign concept
- [ ] Context is clearly labeled
- [ ] Instructions guide sub-agents to use the concept
- [ ] File validates as proper markdown

**Similar changes for:**
- Phase 4 (Batch 2): Lore, Quests, Encounters, Characters
- Phase 5 (Batch 3): Areas

---

### Documentation Tasks

#### TASK-DOC-001: Document solution in SDD-SOLUTIONS.md
**File:** `SDD-SOLUTIONS.md`  
**Line:** End of file (before last `---`)

**Change:**
```markdown
---

## 📝 Campaign Brief Description Support

### Feature Overview

Added support for optional `brief_description` parameter in `generate_adventure_bible` tool. This allows users to provide a campaign idea or story premise that gets stored as a canon fact and used to generate more targeted content.

### Usage

**MCP Tool Call:**
```python
generate_adventure_bible(
  name="la-hoja-de-vlad",
  level_range="1-5",
  tone="political_horror",
  setting_type="urban",
  themes=["traición", "poder", "sangre"],
  villain_type="vampiro_noble",
  mcguffin_type="artefacto_maldito",
  brief_description="Los PJs son nobles vampíricos que descubren una conspiración para derrocar al rey. Deben navegar la corte, descubrir traidores, y decidir si salvar al rey o unirse al complot."
)
```

**What Happens:**
1. The brief description is stored as a "premise" fact in `canon.json`
2. The fact is mutable (can evolve as the story progresses)
3. Sub-agents receive the campaign concept in their delegation prompts
4. Generated content (NPCs, areas, quests) is tailored to support this premise

### Implementation Details

**Files Changed:**
- `internal/domain/canon.go` — Added `BriefDescription` field to `CampaignBrief`
- `internal/mcp/server.go` — Added `brief_description` parameter to tool
- `internal/mcp/handlers/canon.go` — Parse parameter in handler
- `internal/services/canon_service.go` — Use brief to generate richer facts
- `agents/grimorio-architect.md` — Updated workflow to collect and use brief

**Canon Fact Structure:**
```json
{
  "id": "fact-002",
  "category": "premise",
  "statement": "Los PJs son nobles vampíricos que descubren una conspiración...",
  "source": "user_brief",
  "immutable": false,
  "created_at": "2026-05-09T..."
}
```

### Benefits

1. **More Targeted Content:** NPCs, areas, and quests are generated with the story premise in mind
2. **User Agency:** DMs can specify the story they want to tell, not just the setting
3. **Canon Tracking:** The premise is tracked as a fact, allowing for evolution and consequences
4. **Better Coherence:** All generated content shares a common narrative thread

### Example Workflow

```
1. User: "Quiero una campaña de vampiros políticos"
2. Architect: "¿Podés describir la historia/trama que querés contar?"
3. User: "Los PJs investigan traición en la corte vampírica"
4. Architect: → Stores as brief_description
5. generate_adventure_bible → Creates canon with premise fact
6. Sub-agents → Generate content aligned with the premise
```

### Testing

```bash
# Test with brief description
generate_adventure_bible(
  name="test-brief",
  brief_description="Test premise"
)

# Verify canon.json contains premise fact
cat /home/pau/campaigns/test-brief/canon.json | jq '.facts[] | select(.category=="premise")'

# Should show:
# {
#   "id": "fact-002",
#   "category": "premise",
#   "statement": "Test premise",
#   "source": "user_brief",
#   "immutable": false
# }
```

---
```

**Completion Criteria:**
- [ ] Section added with clear title
- [ ] Usage examples provided
- [ ] Implementation details documented
- [ ] Benefits explained
- [ ] Testing commands included
- [ ] File validates as proper markdown

---

## Summary

| Task ID | Type | File | Status |
|---------|------|------|--------|
| TASK-GO-001 | Go | `internal/domain/canon.go` | ⏳ Pending |
| TASK-GO-002 | Go | `internal/mcp/server.go` | ⏳ Pending |
| TASK-GO-003 | Go | `internal/mcp/handlers/canon.go` | ⏳ Pending |
| TASK-GO-004 | Go | `internal/services/canon_service.go` | ⏳ Pending |
| TASK-GO-005 | Go (Test) | `internal/services/canon_service_test.go` | ⏳ Optional |
| TASK-AGENT-001 | Docs | `agents/grimorio-architect.md` | ⏳ Pending |
| TASK-AGENT-002 | Docs | `agents/grimorio-architect.md` | ⏳ Pending |
| TASK-AGENT-003 | Docs | `agents/grimorio-architect.md` | ⏳ Pending |
| TASK-DOC-001 | Docs | `SDD-SOLUTIONS.md` | ⏳ Pending |

**Total:** 9 tasks (8 required, 1 optional)

---

## Dependencies

- No external dependencies
- All changes are internal to the codebase
- Backward compatible (brief_description is optional)

## Risk Assessment

- **Low Risk:** All changes are additive
- **No Breaking Changes:** Existing tool calls continue to work
- **Testing:** Optional test task provided for validation

## Next Steps

1. Review and approve this task list
2. Implement Go tasks (TASK-GO-001 through TASK-GO-004)
3. Run tests to ensure compilation
4. Implement Agent documentation tasks (TASK-AGENT-001 through TASK-AGENT-003)
5. Implement documentation task (TASK-DOC-001)
6. Verify end-to-end with test campaign
