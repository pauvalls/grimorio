# WotC Quality Improvements — Implementation Tasks

**Change Proposal:** wotc-quality-improvements  
**Created:** 2026-05-08  
**Status:** Ready for Implementation  
**Priority:** High  

---

## Overview

This document breaks down the WotC quality improvements into actionable implementation tasks across 6 categories. Each task includes clear acceptance criteria, files to modify/create, and complexity estimates.

**Total Tasks:** 23  
**Estimated Effort:** 8-12 hours  

---

## Batch 1: MCP Tool Implementation (3 tasks)

### Task 1.1: Create `generate_character_hooks` MCP Tool

**Description:**  
Implement a new MCP tool that generates character-specific hooks for areas based on character backgrounds, classes, and motivations. This tool will be used by the areas agent to inject personalized content.

**Files to Create/Modify:**
- `internal/mcp/handlers/character_hooks.go` (new)
- `internal/mcp/server.go` (modify — register tool)
- `internal/services/character_hooks_service.go` (new)

**Acceptance Criteria:**
- [ ] Tool accepts campaign_id, area_content, and character_list parameters
- [ ] Returns array of character hooks with: character_name, hook_text, trigger_condition
- [ ] Hooks target specific backgrounds (acólito, criminal, soldado, etc.)
- [ ] Hooks target specific classes (mago, guerrero, pícaro, etc.)
- [ ] Hooks include emotional/motivational triggers (venganza, redención, etc.)
- [ ] Tool integrates with existing character repository
- [ ] Unit tests cover edge cases (empty character list, missing backgrounds)

**Complexity:** M  
**Estimated Time:** 2-3 hours

---

### Task 1.2: Add `validate_area_format` MCP Tool

**Description:**  
Expose the area validator as an MCP tool so agents can validate areas before saving.

**Files to Create/Modify:**
- `internal/mcp/handlers/validation.go` (new handler file or extend existing)
- `internal/mcp/server.go` (modify — register tool)

**Acceptance Criteria:**
- [ ] Tool accepts campaign_id and area_content parameters
- [ ] Returns ValidationResult with valid boolean and errors array
- [ ] Errors include field and message for each validation failure
- [ ] Validates: area count (10-15), word count (150-200), numeric DCs, treasure with XP, interactive elements, bidirectional connections
- [ ] Unit tests verify all validation rules

**Complexity:** S  
**Estimated Time:** 1-2 hours

---

### Task 1.3: Add `generate_area_content` MCP Tool

**Description:**  
Create a tool that generates WotC-formatted area content with all required sections (read-aloud, creatures, treasure, connections, secrets, decision points).

**Files to Create/Modify:**
- `internal/mcp/handlers/area_generator.go` (new)
- `internal/services/area_generator_service.go` (new)
- `internal/mcp/server.go` (modify — register tool)

**Acceptance Criteria:**
- [ ] Tool accepts: campaign, area_number, area_name, location_context, difficulty_level
- [ ] Returns markdown with all required WotC sections
- [ ] Includes boxed text (read-aloud) in second person present
- [ ] Includes at least 2 character hooks
- [ ] Includes numeric DCs (no "alto/bajo")
- [ ] Includes treasure with XP if creatures present
- [ ] Includes bidirectional connections
- [ ] Includes at least 3 decision points with consequences
- [ ] Word count falls within 150-200 range

**Complexity:** L  
**Estimated Time:** 3-4 hours

---

## Batch 2: Validator Implementation (5 tasks)

### Task 2.1: Implement `ValidateCharacterHooks` Validator

**Description:**  
Validator to ensure character hooks are properly distributed across areas and target diverse backgrounds/classes.

**Files to Create/Modify:**
- `internal/validators/hooks.go` (new)
- `internal/validators/hooks_test.go` (new)

**Validation Rules:**
1. Each area has at least 2 character hooks
2. Hooks target at least 3 different backgrounds across the act
3. Hooks target at least 3 different classes across the act
4. Hooks include emotional triggers (not just mechanical)
5. Hook text is 1-3 sentences (not paragraphs)

**Acceptance Criteria:**
- [ ] Validator parses hooks from area content
- [ ] Returns errors for missing hooks per area
- [ ] Returns errors for lack of diversity in backgrounds/classes
- [ ] Returns errors for hooks without emotional triggers
- [ ] Unit tests cover all validation rules
- [ ] Integration with existing ValidationResult structure

**Complexity:** M  
**Estimated Time:** 2 hours

---

### Task 2.2: Implement `ValidateDecisionPoints` Validator

**Description:**  
Validator to ensure decision points have proper structure with immediate and deferred consequences.

**Files to Create/Modify:**
- `internal/validators/decisions.go` (new)
- `internal/validators/decisions_test.go` (new)

**Validation Rules:**
1. Each act has at least 3 decision points
2. At least 1 decision has immediate consequence (same area or next)
3. At least 1 decision has deferred consequence (next act)
4. At least 1 decision affects faction reputation
5. All decisions use IF/THEN format
6. All decisions specify Affects (areas/acts)
7. All decisions specify World State changes

**Acceptance Criteria:**
- [ ] Validator parses decision points from area content
- [ ] Validates IF/THEN structure
- [ ] Validates consequence timing (immediate vs deferred)
- [ ] Validates faction reputation impact
- [ ] Validates world state tracking (NPCs, factions, clues, quests)
- [ ] Unit tests cover all validation rules

**Complexity:** M  
**Estimated Time:** 2-3 hours

---

### Task 2.3: Implement `ValidateBoxedText` Validator

**Description:**  
Validator to ensure read-aloud text follows WotC standards (second person, present tense, sensory details).

**Files to Create/Modify:**
- `internal/validators/boxed_text.go` (new)
- `internal/validators/boxed_text_test.go` (new)

**Validation Rules:**
1. Each area has boxed text (read-aloud)
2. Boxed text is in second person ("vos", "tú", "ustedes")
3. Boxed text is in present tense
4. Boxed text is 2-4 sentences (100-600 words total per area)
5. Boxed text includes sensory details (sight, sound, smell)
6. Boxed text is atmospheric, not expository
7. Boxed text does not include mechanics or DCs

**Acceptance Criteria:**
- [ ] Validator extracts boxed text from areas
- [ ] Validates second person perspective
- [ ] Validates present tense
- [ ] Validates sentence count and word count
- [ ] Validates presence of sensory language
- [ ] Flags expository or mechanical content
- [ ] Unit tests cover all validation rules

**Complexity:** M  
**Estimated Time:** 2-3 hours

---

### Task 2.4: Implement `ValidatePacing` Validator

**Description:**  
Validator to ensure proper pacing across the act (mix of combat, exploration, social, downtime).

**Files to Create/Modify:**
- `internal/validators/pacing.go` (new)
- `internal/validators/pacing_test.go` (new)

**Validation Rules:**
1. Act has mix of game modes (combat, exploration, social, downtime)
2. No more than 3 consecutive combat-heavy areas
3. No more than 3 consecutive exploration-heavy areas
4. At least 1 social encounter per act
5. At least 1 downtime opportunity per act
6. Climax area (final 2-3) has highest difficulty
7. Opening area (first 2) has clear objectives

**Acceptance Criteria:**
- [ ] Validator identifies game mode per area
- [ ] Validates mode distribution across act
- [ ] Flags consecutive same-mode areas
- [ ] Validates climax difficulty scaling
- [ ] Validates opening area clarity
- [ ] Unit tests cover pacing scenarios

**Complexity:** M  
**Estimated Time:** 2 hours

---

### Task 2.5: Integrate Validators into Validation Engine

**Description:**  
Wire all new validators into the existing validation engine and consistency gate.

**Files to Modify:**
- `internal/services/validation_engine.go` (modify)
- `internal/services/consistency_gate.go` (modify)
- `internal/validators/validators.go` (modify — export all validators)

**Acceptance Criteria:**
- [ ] All 4 new validators registered in validation engine
- [ ] Validators run as part of consistency gate for areas
- [ ] Validation results aggregated properly
- [ ] Errors from all validators included in gate response
- [ ] Integration tests verify end-to-end validation flow

**Complexity:** S  
**Estimated Time:** 1-2 hours

---

## Batch 3: Template Updates (3 tasks)

### Task 3.1: Update `areas.md.tmpl` with Complete WotC Examples

**Description:**  
Replace placeholder examples in the areas template with complete, production-ready WotC-formatted examples.

**Files to Modify:**
- `internal/compiler/templates/areas.md.tmpl`

**Changes:**
1. Replace Área 1 example with complete 150-200 word area
2. Replace Área 2 example with complete 150-200 word area
3. Replace Área 3 example with complete 150-200 word area
4. Add complete decision points table with all required fields
5. Add complete character hooks section with diverse examples
6. Add complete boxed text examples (second person, present tense, sensory)

**Acceptance Criteria:**
- [ ] Each example area is 150-200 words
- [ ] Each area has all required sections (read-aloud, creatures, treasure, connections, secrets, decision points)
- [ ] Boxed text is second person present tense with sensory details
- [ ] Decision points use IF/THEN format with Affects and World State
- [ ] Character hooks target diverse backgrounds/classes
- [ ] All DCs are numeric (no "alto/bajo")
- [ ] Treasure includes XP totals
- [ ] Connections are bidirectional

**Complexity:** M  
**Estimated Time:** 2-3 hours

---

### Task 3.2: Add WotC Style Guide Section to Template

**Description:**  
Add a comprehensive style guide section to the template that documents WotC formatting standards.

**Files to Modify:**
- `internal/compiler/templates/areas.md.tmpl` (add section at end or beginning)

**Content to Add:**
```markdown
## WotC Style Guide

### Boxed Text (Read-Aloud)
- Write in second person present: "Vosotros entráis", "Sentís", "Veis"
- 2-4 sentences, 100-600 words total per area
- Focus on sensory details: sight, sound, smell, touch
- Atmospheric, not expository — don't explain history or mechanics
- Never include DCs, stats, or game mechanics

### Area Structure
- Heading: `### Área X: Descriptive Name`
- Word count: 150-200 words per area
- Required sections: Read-Aloud, DM Description, Creatures, Treasure, Connections, Secrets/Traps, Development

### Decision Points Format
```markdown
**Decision Points:**
- **IF** los PJs [acción concreta], **THEN** [consecuencia explícita]
  - **Affects:** Área X, Acto N
  - **World State:** NPCs: [quién], Facciones: [cambio ±X], Pistas: [clue-id]
```

### Character Hooks
- Target specific backgrounds: acólito, criminal, soldado, noble, etc.
- Target specific classes: mago, guerrero, pícaro, clérigo, etc.
- Include emotional triggers: venganza, redención, protección, descubrimiento
- 1-3 sentences per hook

### DCs and Mechanics
- Always use numeric DCs: DC 10, DC 15, DC 20
- Never use: "DC alto", "DC bajo", "DC difícil"
- Format: "Percepción DC 14", "Investigación DC 15"

### Treasure
- Always include XP total when creatures present
- Format: `**XP total:** 450 XP`
- Include currency: `23 gp, 45 sp`
- Named items: `**Llave de Latón** (abre Área 5)`

### Connections
- Bidirectional only: if Área 1 → Área 2, then Área 2 → Área 1
- Format: `→ Área X (dirección, descripción breve)`
```

**Acceptance Criteria:**
- [ ] Style guide section added to template
- [ ] Covers all WotC formatting requirements
- [ ] Includes examples for each rule
- [ ] Clear and actionable for AI agents

**Complexity:** S  
**Estimated Time:** 1 hour

---

### Task 3.3: Create Template Validation Test Suite

**Description:**  
Create tests that verify the template produces WotC-compliant output when rendered.

**Files to Create:**
- `internal/compiler/templates/areas_test.go` (new)

**Test Cases:**
1. Rendered template has all required sections
2. Example areas are 150-200 words
3. Boxed text is second person present
4. Decision points use IF/THEN format
5. All DCs are numeric
6. Treasure includes XP

**Acceptance Criteria:**
- [ ] Template renders without errors
- [ ] Rendered output passes all validators
- [ ] Tests cover all template sections
- [ ] Tests fail if template structure changes

**Complexity:** S  
**Estimated Time:** 1-2 hours

---

## Batch 4: Agent Updates (6 tasks)

### Task 4.1: Update Grimorio Architect — Phase 4 Integration

**Description:**  
Update the architect agent to integrate character hooks generation into the area generation workflow.

**Files to Modify:**
- `agents/grimorio-architect.md`

**Changes:**
1. Add character hooks generation step before area generation
2. Wire `generate_character_hooks` tool call
3. Pass hooks to areas agent as context
4. Add validation step after area generation

**Acceptance Criteria:**
- [ ] Phase 4 includes character hooks generation
- [ ] Hooks passed to areas agent
- [ ] Validation runs after area generation
- [ ] Agent reports validation results to user

**Complexity:** S  
**Estimated Time:** 1 hour

---

### Task 4.2: Update Grimorio Areas Agent — Character Hooks Integration

**Description:**  
Update the areas agent to use character hooks from the MCP tool and integrate them into areas.

**Files to Modify:**
- `agents/grimorio-areas.md`

**Changes:**
1. Add step to call `generate_character_hooks` before writing areas
2. Integrate hooks into each area's "Ganchos de Personaje" section
3. Ensure hooks target diverse backgrounds/classes
4. Add validation step using `validate_area_format`

**Acceptance Criteria:**
- [ ] Agent calls generate_character_hooks tool
- [ ] Hooks integrated into area content
- [ ] Each area has at least 2 hooks
- [ ] Hooks target diverse backgrounds/classes
- [ ] Agent validates areas before saving

**Complexity:** M  
**Estimated Time:** 2 hours

---

### Task 4.3: Update Grimorio Areas Agent — Decision Points

**Description:**  
Update the areas agent to generate proper decision points with consequences and world state tracking.

**Files to Modify:**
- `agents/grimorio-areas.md`

**Changes:**
1. Add explicit decision points generation step
2. Require IF/THEN format for all decisions
3. Require Affects and World State fields
4. Ensure at least 3 decisions per act
5. Ensure mix of immediate and deferred consequences

**Acceptance Criteria:**
- [ ] Each area with decisions uses IF/THEN format
- [ ] Each decision specifies Affects (areas/acts)
- [ ] Each decision specifies World State changes
- [ ] At least 3 decisions per act
- [ ] At least 1 immediate consequence
- [ ] At least 1 deferred consequence
- [ ] At least 1 faction reputation change

**Complexity:** M  
**Estimated Time:** 2 hours

---

### Task 4.4: Update Grimorio Areas Agent — Boxed Text Standards

**Description:**  
Update the areas agent to generate boxed text that follows WotC standards (second person, present tense, sensory).

**Files to Modify:**
- `agents/grimorio-areas.md`

**Changes:**
1. Add explicit boxed text generation guidelines
2. Require second person present tense
3. Require sensory details (sight, sound, smell)
4. Prohibit expository or mechanical content
5. Enforce 2-4 sentences per boxed text

**Acceptance Criteria:**
- [ ] All boxed text is second person present
- [ ] All boxed text includes sensory details
- [ ] No expository or mechanical content in boxed text
- [ ] Boxed text is 2-4 sentences
- [ ] Agent validates boxed text before saving

**Complexity:** M  
**Estimated Time:** 1-2 hours

---

### Task 4.5: Update Grimorio Quests Agent — Integration with Areas

**Description:**  
Update the quests agent to ensure quests are referenced in areas and vice versa.

**Files to Modify:**
- `agents/grimorio-quests.md`

**Changes:**
1. Add step to read areas content before generating quests
2. Ensure quests reference specific areas
3. Ensure areas reference quest rewards/completions
4. Add validation for quest-area integration

**Acceptance Criteria:**
- [ ] Quests reference specific area numbers
- [ ] Areas reference quest completions
- [ ] Quest rewards are available in areas
- [ ] Validation checks quest-area consistency

**Complexity:** S  
**Estimated Time:** 1 hour

---

### Task 4.6: Update Grimorio Narrator Custodian — New Validators

**Description:**  
Update the narrative custodian to run the new validators as part of consistency checks.

**Files to Modify:**
- `agents/grimorio-narrative-custodian.md`

**Changes:**
1. Add character hooks validation to batch validation
2. Add decision points validation to batch validation
3. Add boxed text validation to batch validation
4. Add pacing validation to batch validation
5. Update validation report format to include new checks

**Acceptance Criteria:**
- [ ] Custodian runs all 4 new validators
- [ ] Validation report includes results from all validators
- [ ] Rejection includes specific fix suggestions
- [ ] Integration with existing consistency gate

**Complexity:** M  
**Estimated Time:** 2 hours

---

## Batch 5: Integration (3 tasks)

### Task 5.1: Wire MCP Tools to Agents

**Description:**  
Ensure all new MCP tools are accessible to agents and properly documented in agent tool lists.

**Files to Modify:**
- `agents/grimorio-architect.md` (add tools to grimorio_mcp list)
- `agents/grimorio-areas.md` (add tools to grimorio_mcp list)
- `agents/grimorio-narrative-custodian.md` (add tools to grimorio_mcp list)

**Tools to Add:**
- `generate_character_hooks`
- `validate_area_format`
- `generate_area_content`

**Acceptance Criteria:**
- [ ] All tools listed in agent grimorio_mcp arrays
- [ ] Agents can call tools successfully
- [ ] Tool documentation in agent prompts

**Complexity:** S  
**Estimated Time:** 30 minutes

---

### Task 5.2: Update Validation Engine Configuration

**Description:**  
Configure the validation engine to run new validators in the correct order with proper error aggregation.

**Files to Modify:**
- `internal/services/validation_engine.go`
- `internal/config/config.go` (if validation config needed)

**Acceptance Criteria:**
- [ ] Validators run in order: format → hooks → decisions → boxed text → pacing → integration
- [ ] Errors aggregated into single ValidationResult
- [ ] Validation engine exposes all validators via MCP
- [ ] Configuration allows enabling/disabling validators

**Complexity:** S  
**Estimated Time:** 1 hour

---

### Task 5.3: End-to-End Integration Test

**Description:**  
Create an end-to-end test that generates a full act and validates it passes all new validators.

**Files to Create:**
- `internal/mcp/handlers/e2e_wotc_quality_test.go` (new)

**Test Flow:**
1. Generate adventure bible (canon)
2. Generate NPCs and bestiary
3. Generate areas with character hooks
4. Run all validators
5. Assert all validators pass
6. Assert areas meet WotC standards

**Acceptance Criteria:**
- [ ] Test generates complete act
- [ ] All validators pass
- [ ] Areas meet word count requirements
- [ ] Character hooks present and diverse
- [ ] Decision points properly formatted
- [ ] Boxed text follows standards
- [ ] Pacing is appropriate

**Complexity:** L  
**Estimated Time:** 3-4 hours

---

## Batch 6: Testing (3 tasks)

### Task 6.1: Unit Tests for Character Hooks Validator

**Description:**  
Write comprehensive unit tests for the ValidateCharacterHooks validator.

**Files to Create:**
- `internal/validators/hooks_test.go` (if not created in Task 2.1)

**Test Cases:**
1. Valid hooks pass validation
2. Missing hooks fail validation
3. Lack of background diversity fails validation
4. Lack of class diversity fails validation
5. Hooks without emotional triggers fail validation
6. Hook text too long fails validation

**Acceptance Criteria:**
- [ ] 100% code coverage for hooks.go
- [ ] All validation rules tested
- [ ] Edge cases covered (empty input, malformed hooks)
- [ ] Tests run in CI pipeline

**Complexity:** S  
**Estimated Time:** 1 hour

---

### Task 6.2: Unit Tests for Decision Points Validator

**Description:**  
Write comprehensive unit tests for the ValidateDecisionPoints validator.

**Files to Create:**
- `internal/validators/decisions_test.go` (if not created in Task 2.2)

**Test Cases:**
1. Valid decision points pass validation
2. Missing IF/THEN format fails validation
3. Missing Affects field fails validation
4. Missing World State field fails validation
5. No immediate consequences fails validation
6. No deferred consequences fails validation
7. No faction reputation changes fails validation

**Acceptance Criteria:**
- [ ] 100% code coverage for decisions.go
- [ ] All validation rules tested
- [ ] Edge cases covered (empty input, malformed decisions)
- [ ] Tests run in CI pipeline

**Complexity:** S  
**Estimated Time:** 1-2 hours

---

### Task 6.3: Unit Tests for Boxed Text Validator

**Description:**  
Write comprehensive unit tests for the ValidateBoxedText validator.

**Files to Create:**
- `internal/validators/boxed_text_test.go` (if not created in Task 2.3)

**Test Cases:**
1. Valid boxed text passes validation
2. Missing boxed text fails validation
3. First/third person fails validation
4. Past tense fails validation
5. Too short/long fails validation
6. Missing sensory details fails validation
7. Expository content fails validation
8. Mechanical content (DCs, stats) fails validation

**Acceptance Criteria:**
- [ ] 100% code coverage for boxed_text.go
- [ ] All validation rules tested
- [ ] Edge cases covered (empty input, malformed text)
- [ ] Tests run in CI pipeline

**Complexity:** S  
**Estimated Time:** 1-2 hours

---

## Implementation Order

**Recommended Batch Order:**
1. **Batch 2** (Validators) — Foundation for everything else
2. **Batch 1** (MCP Tools) — Expose validators and generators
3. **Batch 3** (Templates) — Provide examples for agents
4. **Batch 4** (Agents) — Update agents to use new tools
5. **Batch 5** (Integration) — Wire everything together
6. **Batch 6** (Testing) — Ensure quality and coverage

**Dependencies:**
- Batch 4 depends on Batch 1 (agents need MCP tools)
- Batch 5 depends on Batches 1-4 (integration requires all components)
- Batch 6 depends on Batch 2 (tests need validators implemented)

---

## Success Metrics

**Quality Metrics:**
- ✅ All areas pass validation (0 errors)
- ✅ 100% test coverage for new validators
- ✅ All agents use new tools correctly
- ✅ End-to-end test passes

**Content Metrics:**
- ✅ Each area has 150-200 words
- ✅ Each area has 2+ character hooks
- ✅ Each act has 3+ decision points
- ✅ All boxed text is second person present
- ✅ All DCs are numeric
- ✅ All treasure includes XP

---

## Notes

- All validators should use the existing `ValidationResult` and `ValidationError` structures
- MCP tools should follow existing patterns in `internal/mcp/handlers/`
- Agent updates should maintain existing workflow structure
- Tests should run in the existing CI pipeline (`make test`)
- Documentation should be in Spanish (rioplatense) for agent-facing content
