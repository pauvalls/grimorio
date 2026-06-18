package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
	"github.com/pauvalls/grimorio/internal/services/consolidation"
	"github.com/pauvalls/grimorio/internal/validators"
)

// ValidationEngine performs rule-based validation of content proposals and campaign consistency
type ValidationEngine struct {
	canonService  *CanonService
	stateService  *NarrativeStateService
	factionRepo   repository.FactionReputationRepository
	campaignDir   string // base directory where campaign subdirs live (e.g. ~/campaigns)
	consolidator  *ConsolidationAdapter
}

// NewValidationEngine creates a new validation engine
func NewValidationEngine(canonService *CanonService, stateService *NarrativeStateService, factionRepo repository.FactionReputationRepository, campaignDir string) *ValidationEngine {
	return &ValidationEngine{
		canonService: canonService,
		stateService: stateService,
		factionRepo:  factionRepo,
		campaignDir:  campaignDir,
		consolidator: NewConsolidationAdapter(campaignDir),
	}
}

// ValidateAct validates an act proposal against canon and narrative state
func (e *ValidationEngine) ValidateAct(ctx context.Context, campaignID string, actID string, content string, refs []domain.EntityReference) (*domain.ValidationReport, error) {
	proposal := domain.ContentProposal{
		ID:               actID,
		Type:             "act",
		Content:          content,
		EntityReferences: refs,
	}
	return e.validate(ctx, campaignID, proposal)
}

// ValidateQuest validates a quest proposal against canon and narrative state
func (e *ValidationEngine) ValidateQuest(ctx context.Context, campaignID string, questID string, content string, refs []domain.EntityReference) (*domain.ValidationReport, error) {
	proposal := domain.ContentProposal{
		ID:               questID,
		Type:             "quest",
		Content:          content,
		EntityReferences: refs,
	}
	return e.validate(ctx, campaignID, proposal)
}

// CheckConsistency performs a campaign-wide consistency check
func (e *ValidationEngine) CheckConsistency(ctx context.Context, campaignID string, scope domain.ConsistencyScope) (*domain.ConsistencyReport, error) {
	doc, err := e.canonService.LoadCanon(ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to load canon: %w", err)
	}

	state, stateErr := e.stateService.Load(ctx, campaignID)

	report := &domain.ConsistencyReport{
		CampaignID: campaignID,
		Issues:     []domain.CheckResult{},
	}

	// Build entity lookup
	entityMap := make(map[string]*domain.CanonEntity)
	for i := range doc.Entities {
		entityMap[doc.Entities[i].ID] = &doc.Entities[i]
	}

	// Rule 1: npc_alive_check — dead NPCs in state must have canon state = dead
	if stateErr == nil {
		for _, death := range state.DeadNPCs {
			entity, exists := entityMap[death.NPCID]
			if !exists {
				report.Issues = append(report.Issues, domain.CheckResult{
					Rule:     "npc_alive_check",
					Passed:   false,
					Severity: "error",
					Message:  fmt.Sprintf("Dead NPC %s not found in canon", death.Name),
				})
			} else if entity.CanonState != domain.EntityStateDead {
				report.Issues = append(report.Issues, domain.CheckResult{
					Rule:     "npc_alive_check",
					Passed:   false,
					Severity: "error",
					Message:  fmt.Sprintf("NPC %s is dead in narrative state but canon state is %s", death.Name, entity.CanonState),
				})
			}
		}
	}

	// Rule 2: mcguffin_continuity — mcguffin entities in canon must match key items in state
	if stateErr == nil {
		var mcguffinEntities []domain.CanonEntity
		for _, entity := range doc.Entities {
			if entity.Role == "mcguffin" {
				mcguffinEntities = append(mcguffinEntities, entity)
			}
		}

		var mcguffinItems []domain.KeyItem
		for _, item := range state.KeyItems {
			if item.IsMcGuffin {
				mcguffinItems = append(mcguffinItems, item)
			}
		}

		// Check that each mcguffin entity has a matching key item
		for _, entity := range mcguffinEntities {
			found := false
			for _, item := range mcguffinItems {
				if item.ID == entity.ID {
					found = true
					break
				}
			}
			if !found {
				report.Issues = append(report.Issues, domain.CheckResult{
					Rule:     "mcguffin_continuity",
					Passed:   false,
					Severity: "warning",
					Message:  fmt.Sprintf("McGuffin entity %s exists in canon but has no matching key item in narrative state", entity.Name),
				})
			}
		}

		// Check that each mcguffin key item has a matching entity
		for _, item := range mcguffinItems {
			found := false
			for _, entity := range mcguffinEntities {
				if entity.ID == item.ID {
					found = true
					break
				}
			}
			if !found {
				report.Issues = append(report.Issues, domain.CheckResult{
					Rule:     "mcguffin_continuity",
					Passed:   false,
					Severity: "warning",
					Message:  fmt.Sprintf("Key item %s is marked as mcguffin in state but no matching entity exists in canon", item.Name),
				})
			}
		}
	}

	// Rule 3: lore_rule_compliance — check if any active quests violate rules (simplified)
	// For full scope, validate across all campaign artifacts
	if scope == domain.ConsistencyScopeFull {
		// Full-scope cross-reference validation
		// Check that all NPCs referenced in quests exist and are alive
		if stateErr == nil {
			for _, quest := range state.ActiveQuests {
				if quest.GiverNPC != "" {
					if entity, exists := entityMap[quest.GiverNPC]; exists {
						if entity.CanonState == domain.EntityStateDead {
							report.Issues = append(report.Issues, domain.CheckResult{
								Rule:     "npc_alive_check",
								Passed:   false,
								Severity: "error",
								Message:  fmt.Sprintf("Quest giver NPC %s is dead but assigned to active quest %s", entity.Name, quest.Name),
							})
						}
					}
				}
			}
			for _, quest := range state.CompletedQuests {
				if quest.GiverNPC != "" {
					if entity, exists := entityMap[quest.GiverNPC]; exists {
						if entity.CanonState == domain.EntityStateDead {
							report.Issues = append(report.Issues, domain.CheckResult{
								Rule:     "npc_alive_check",
								Passed:   false,
								Severity: "warning",
								Message:  fmt.Sprintf("Quest giver NPC %s is dead but was assigned to completed quest %s", entity.Name, quest.Name),
							})
						}
					}
				}
			}
		}

		// Check timeline continuity: events should be in chronological order
		for i := 1; i < len(doc.Timeline); i++ {
			if doc.Timeline[i].Timestamp < doc.Timeline[i-1].Timestamp {
				report.Issues = append(report.Issues, domain.CheckResult{
					Rule:     "timeline_consistency",
					Passed:   false,
					Severity: "warning",
					Message:  fmt.Sprintf("Timeline event %s appears before %s", doc.Timeline[i].ID, doc.Timeline[i-1].ID),
				})
			}
		}

		// Check for entities without relationships (orphaned entities warning)
		if len(doc.Relationships) > 0 {
			connected := make(map[string]bool)
			for _, rel := range doc.Relationships {
				connected[rel.From] = true
				connected[rel.To] = true
			}
			for _, entity := range doc.Entities {
				if entity.Type == domain.EntityTypeNPC && !connected[entity.ID] {
					report.Issues = append(report.Issues, domain.CheckResult{
						Rule:     "entity_isolation",
						Passed:   false,
						Severity: "warning",
						Message:  fmt.Sprintf("NPC %s has no relationships in canon", entity.Name),
					})
				}
			}
		}

		// WotC Format Validations — run on all act files
		e.runWotCConsistencyChecks(report, campaignID)

		// Integration cross-reference: act <-> bestiary <-> npcs
		e.runIntegrationChecks(report, campaignID)

		// Consolidation checks: cross-file coherence via the consolidation engine
		e.runConsolidationChecks(ctx, report, campaignID)
	}

	// Compute report statistics
	e.computeConsistencyStats(report)

	return report, nil
}

// runWotCConsistencyChecks runs WotC format validations on all act files when scope=full.
func (e *ValidationEngine) runWotCConsistencyChecks(report *domain.ConsistencyReport, campaignID string) {
	areasDir := filepath.Join(e.campaignDir, campaignID, "areas")
	info, err := os.Stat(areasDir)
	if err != nil || !info.IsDir() {
		return // no areas directory — nothing to validate
	}

	files, err := os.ReadDir(areasDir)
	if err != nil {
		return
	}

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".md") {
			continue
		}
		content, err := os.ReadFile(filepath.Join(areasDir, f.Name()))
		if err != nil {
			continue
		}
		chapterName := strings.TrimSuffix(f.Name(), ".md")

		// 1. Developments: 3-5 IF-THEN branches
		devResult := validators.ValidateDevelopments(string(content))
		for _, verr := range devResult.Errors {
			report.Issues = append(report.Issues, domain.CheckResult{
				Rule:     "wotc_developments",
				Passed:   false,
				Severity: "warning",
				Message:  fmt.Sprintf("[%s] %s", chapterName, verr.Message),
				Location: chapterName,
			})
		}

		// 2. Multiple solution paths
		solResult := validators.ValidateMultipleSolutions(string(content))
		for _, verr := range solResult.Errors {
			report.Issues = append(report.Issues, domain.CheckResult{
				Rule:     "wotc_multiple_solutions",
				Passed:   false,
				Severity: "warning",
				Message:  fmt.Sprintf("[%s] %s", chapterName, verr.Message),
				Location: chapterName,
			})
		}

		// 3. Character hooks
		hooksResult := validators.ValidateCharacterHooks(string(content))
		for _, verr := range hooksResult.Errors {
			report.Issues = append(report.Issues, domain.CheckResult{
				Rule:     "wotc_character_hooks",
				Passed:   false,
				Severity: "warning",
				Message:  fmt.Sprintf("[%s] %s", chapterName, verr.Message),
				Location: chapterName,
			})
		}

		// 4. Boxed text quality
		boxedResult := validators.ValidateBoxedText(string(content))
		for _, verr := range boxedResult.Errors {
			report.Issues = append(report.Issues, domain.CheckResult{
				Rule:     "wotc_boxed_text",
				Passed:   false,
				Severity: "warning",
				Message:  fmt.Sprintf("[%s] %s", chapterName, verr.Message),
				Location: chapterName,
			})
		}
	}

	// 5. NPC word count — validate against npcs directory
	npcsDir := filepath.Join(e.campaignDir, campaignID, "npcs")
	npcsInfo, npcsErr := os.Stat(npcsDir)
	if npcsErr == nil && npcsInfo.IsDir() {
		npcFiles, _ := os.ReadDir(npcsDir)
		for _, f := range npcFiles {
			if !strings.HasSuffix(f.Name(), ".md") {
				continue
			}
			npcContent, err := os.ReadFile(filepath.Join(npcsDir, f.Name()))
			if err != nil {
				continue
			}
			npcResult := validators.ValidateNPCWordCount(string(npcContent))
			for _, verr := range npcResult.Errors {
				report.Issues = append(report.Issues, domain.CheckResult{
					Rule:     "wotc_npc_word_count",
					Passed:   false,
					Severity: "warning",
					Message:  fmt.Sprintf("[%s] %s", f.Name(), verr.Message),
					Location: f.Name(),
				})
			}
			for _, warn := range npcResult.Warnings {
				report.Issues = append(report.Issues, domain.CheckResult{
					Rule:     "wotc_npc_word_count",
					Passed:   false,
					Severity: "warning",
					Message:  fmt.Sprintf("[%s] %s", f.Name(), warn),
					Location: f.Name(),
				})
			}
		}
	}
}

// runIntegrationChecks validates cross-references between acts, bestiary, and NPCs.
func (e *ValidationEngine) runIntegrationChecks(report *domain.ConsistencyReport, campaignID string) {
	campDir := filepath.Join(e.campaignDir, campaignID)

	// Read bestiary
	bestiaryPath := filepath.Join(campDir, "bestiary")
	bestiaryContent := ""
	if bi, err := os.Stat(bestiaryPath); err == nil && bi.IsDir() {
		var parts []string
		files, _ := os.ReadDir(bestiaryPath)
		for _, f := range files {
			if strings.HasSuffix(f.Name(), ".md") {
				data, err := os.ReadFile(filepath.Join(bestiaryPath, f.Name()))
				if err == nil {
					parts = append(parts, string(data))
				}
			}
		}
		bestiaryContent = strings.Join(parts, "\n")
	}

	// Read NPCs
	npcsPath := filepath.Join(campDir, "npcs")
	npcsContent := ""
	if ni, err := os.Stat(npcsPath); err == nil && ni.IsDir() {
		var parts []string
		files, _ := os.ReadDir(npcsPath)
		for _, f := range files {
			if strings.HasSuffix(f.Name(), ".md") {
				data, err := os.ReadFile(filepath.Join(npcsPath, f.Name()))
				if err == nil {
					parts = append(parts, string(data))
				}
			}
		}
		npcsContent = strings.Join(parts, "\n")
	}

	// Read acts
	areasPath := filepath.Join(campDir, "areas")
	if ai, err := os.Stat(areasPath); err != nil || !ai.IsDir() {
		return
	}

	files, _ := os.ReadDir(areasPath)
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".md") {
			continue
		}
		actContent, err := os.ReadFile(filepath.Join(areasPath, f.Name()))
		if err != nil {
			continue
		}
		chapterName := strings.TrimSuffix(f.Name(), ".md")

		// Integration validation per act
		integResult := validators.ValidateIntegration(string(actContent), bestiaryContent, npcsContent)
		for _, verr := range integResult.Errors {
			report.Issues = append(report.Issues, domain.CheckResult{
				Rule:     "integration",
				Passed:   false,
				Severity: "error",
				Message:  fmt.Sprintf("[%s] %s", chapterName, verr.Message),
				Location: chapterName,
			})
		}
	}
}

func (e *ValidationEngine) validate(ctx context.Context, campaignID string, proposal domain.ContentProposal) (*domain.ValidationReport, error) {
	report := &domain.ValidationReport{
		ArtifactID:    proposal.ID,
		ArtifactType:  proposal.Type,
		OverallStatus: "approved",
		Checks:        []domain.CheckResult{},
		Suggestions:   []domain.Suggestion{},
	}

	doc, err := e.canonService.LoadCanon(ctx, campaignID)
	if err != nil {
		report.AddCheck("canon_load", false, "critical", fmt.Sprintf("Failed to load canon: %v", err), "")
		report.ComputeOverallStatus()
		return report, nil
	}

	state, _ := e.stateService.Load(ctx, campaignID)

	// Build entity lookup
	entityMap := make(map[string]*domain.CanonEntity)
	for i := range doc.Entities {
		entityMap[doc.Entities[i].ID] = &doc.Entities[i]
	}

	// Rule 1: entity_not_found + npc_alive_check
	for _, ref := range proposal.EntityReferences {
		entity, found := entityMap[ref.EntityID]
		if !found {
			report.AddCheck("entity_not_found", false, "critical",
				fmt.Sprintf("Entity %s not found in canon", ref.EntityID),
				ref.Location)
			report.AddSuggestion(
				fmt.Sprintf("Entity %s does not exist", ref.EntityID),
				"Create the entity first or correct the reference",
				"All referenced entities must exist in the canon document")
			continue
		}

		// Check required state
		if ref.RequiredState != "" && entity.CanonState != ref.RequiredState {
			report.AddCheck("entity_state_mismatch", false, "error",
				fmt.Sprintf("Entity %s is %s, required %s", ref.EntityID, entity.CanonState, ref.RequiredState),
				ref.Location)
			report.AddSuggestion(
				fmt.Sprintf("Entity %s state mismatch", ref.EntityID),
				fmt.Sprintf("Adjust narrative to reflect that %s is %s", entity.Name, entity.CanonState),
				"Canon state must match the required state for the proposal")
		}

		// npc_alive_check
		if state != nil && entity.Type == domain.EntityTypeNPC {
			for _, death := range state.DeadNPCs {
				if death.NPCID == ref.EntityID {
					report.AddCheck("npc_alive_check", false, "critical",
						fmt.Sprintf("NPC %s is dead (session %d)", entity.Name, death.Session),
						ref.Location)
					report.AddSuggestion(
						fmt.Sprintf("NPC %s is dead", entity.Name),
						"Replace with a new NPC or use a non-NPC method (letter, vision)",
						"Dead NPCs cannot appear alive in new content")
				}
			}
		}
	}

	// Rule 2: lore_rule_compliance
	for _, rule := range doc.Rules {
		if violation := e.checkRuleViolation(proposal.Content, rule); violation {
			report.AddCheck("lore_rule_compliance", false, "critical",
				fmt.Sprintf("Violates rule %s: %s", rule.ID, rule.Statement),
				"")
			report.AddSuggestion(
				fmt.Sprintf("Lore rule violation: %s", rule.Statement),
				"Adjust content to comply with the established rule",
				"Canon rules are immutable constraints on the world")
		}
	}

	// Rule 4: mcguffin_continuity
	if state != nil {
		for _, item := range state.KeyItems {
			if item.IsMcGuffin {
				mcguffinEntity, found := entityMap[item.ID]
				if found && strings.Contains(strings.ToLower(proposal.Content), strings.ToLower(mcguffinEntity.Name)) {
					// Proposal mentions the mcguffin — check basic consistency
					if item.Holder == "party" && strings.Contains(strings.ToLower(proposal.Content), "villain's lair") {
						report.AddCheck("mcguffin_continuity", false, "error",
							fmt.Sprintf("McGuffin %s is held by the party but proposal places it in villain's lair", mcguffinEntity.Name),
							"")
						report.AddSuggestion(
								fmt.Sprintf("McGuffin %s location mismatch", mcguffinEntity.Name),
								"Ensure the mcguffin location matches the narrative state",
								"McGuffin continuity must be maintained across acts")
					}
				}
			}
		}
	}

	// Rule 4.5: chapter_area_validation — areas in chapter must have required NPCs/locations in canon
	if proposal.Type == "chapter" || proposal.Type == "act" {
		// Extract area references from content (simplified: look for "Area X" patterns)
		areaRefs := e.extractAreaReferences(proposal.Content)
		for _, areaRef := range areaRefs {
			// Check if referenced NPCs exist in canon
			for _, npcRef := range areaRef.NPCs {
				found := false
				for _, entity := range doc.Entities {
					if entity.Type == domain.EntityTypeNPC && strings.Contains(strings.ToLower(entity.Name), strings.ToLower(npcRef)) {
						found = true
						break
					}
				}
				if !found {
					report.AddCheck("chapter_area_validation", false, "warning",
						fmt.Sprintf("NPC '%s' in area %s not found in canon", npcRef, areaRef.AreaNumber),
						"")
					report.AddSuggestion(
						fmt.Sprintf("NPC '%s' referenced but not in canon", npcRef),
						"Add NPC to canon or correct the reference",
						"All NPCs in areas must exist in canon")
				}
			}
		}
	}

	// Rule 5: quest_reward_existence
	if proposal.Type == "quest" {
		rewards := e.extractRewardNames(proposal.Content)
		for _, reward := range rewards {
			found := false
			for _, entity := range doc.Entities {
				if entity.Type == domain.EntityTypeItem && strings.Contains(strings.ToLower(entity.Name), strings.ToLower(reward)) {
					found = true
					break
				}
			}
			if !found {
				report.AddCheck("quest_reward_existence", false, "error",
					fmt.Sprintf("Reward '%s' not found in canon entities", reward),
					"")
				report.AddSuggestion(
					fmt.Sprintf("Quest reward '%s' does not exist", reward),
					"Add the item to canon or replace with an existing canon item",
					"All quest rewards must reference canon entities")
			}
		}
	}

	// Rule 6: level_encounter_balance
	if proposal.Type == "encounter" {
		crs := e.extractCRs(proposal.Content)
		levelRange := e.parseLevelRange(doc)
		maxLevel := levelRange[1]
		for _, cr := range crs {
			if cr > maxLevel+3 {
				report.AddCheck("level_encounter_balance", false, "error",
					fmt.Sprintf("CR %d incompatible with party level %d", cr, maxLevel),
					"")
				report.AddSuggestion(
					fmt.Sprintf("Encounter CR %d exceeds party capability", cr),
					fmt.Sprintf("Reduce CR to at most %d (party max + 3)", maxLevel+3),
					"Encounter difficulty must match campaign level range")
			}
		}
	}

	// Rule 7: location_existence
	if proposal.Type == "act" || proposal.Type == "encounter" || proposal.Type == "quest" {
		locations := e.extractLocationReferences(proposal.Content)
		for _, loc := range locations {
			found := false
			for _, entity := range doc.Entities {
				if entity.Type == domain.EntityTypeLocation && strings.Contains(strings.ToLower(entity.Name), strings.ToLower(loc)) {
					found = true
					break
				}
			}
			if !found {
				report.AddCheck("location_existence", false, "error",
					fmt.Sprintf("Location '%s' not found in canon", loc),
					"")
				report.AddSuggestion(
					fmt.Sprintf("Location '%s' does not exist in canon", loc),
					"Add the location to canon or correct the reference",
					"All referenced locations must exist in the canon document")
			}
		}
	}

	// Rule 8: timeline_consistency
	if proposal.Type == "act" && len(doc.Timeline) > 0 {
		events := e.extractTimelineEvents(proposal.Content)
		for _, event := range events {
			for _, prereqID := range event.Prerequisites {
				prereq := e.findTimelineEvent(doc.Timeline, prereqID)
				if prereq != nil && event.Ordinal <= prereq.Ordinal {
					report.AddCheck("timeline_consistency", false, "error",
						fmt.Sprintf("Event %s occurs before prerequisite %s", event.ID, prereqID),
						"")
					report.AddSuggestion(
						fmt.Sprintf("Timeline event %s violates chronology", event.ID),
						"Reorder events so prerequisites occur first",
						"Timeline must maintain causal ordering")
				}
			}
		}
	}

	// Rule 9: prerequisite_clue_check
	if proposal.Type == "act" && state != nil {
		requiredClues := e.extractRequiredClues(proposal.Content)
		for _, clueID := range requiredClues {
			revealed := false
			for _, clue := range state.RevealedClues {
				if clue.ID == clueID {
					revealed = true
					break
				}
			}
			hasAlt := e.hasAlternativePath(proposal.Content, clueID)
			if !revealed && !hasAlt {
				report.AddCheck("prerequisite_clue_check", false, "error",
					fmt.Sprintf("Act requires clue %s which has not been revealed and has no alternative", clueID),
					"")
				report.AddSuggestion(
					fmt.Sprintf("Clue %s not yet revealed", clueID),
					"Add an alternative path or ensure the clue is revealed in a prior act",
					"Acts must be completable even if players miss optional clues")
			}
		}
	}

	// Rule 10: faction_reputation_gate
	factionRefs := e.extractFactionReferences(proposal.Content)
	if proposal.FactionContext != "" {
		report.AddCheck("faction_context", true, "info",
			fmt.Sprintf("Faction context provided: %s", proposal.FactionContext), "")
	}
	if len(factionRefs) > 0 && e.factionRepo != nil {
		matrix, matrixErr := e.factionRepo.Load(campaignID)
		for _, ref := range factionRefs {
			// Resolve faction name to ID via canon entities
			factionID := ref.FactionID
			for _, entity := range doc.Entities {
				if entity.Type == domain.EntityTypeFaction {
					if strings.EqualFold(entity.ID, ref.FactionID) || strings.EqualFold(entity.Name, ref.FactionID) {
						factionID = entity.ID
						break
					}
				}
			}

			if matrixErr != nil {
				// No reputation matrix yet — just informational
				report.AddCheck("faction_reputation_gate", true, "info",
					fmt.Sprintf("Faction %s referenced — no reputation data available yet", factionID), "")
				continue
			}

			// Find reputation entry for default party
			var entry *domain.ReputationEntry
			for i := range matrix.Entries {
				if matrix.Entries[i].FactionID == factionID {
					entry = &matrix.Entries[i]
					break
				}
			}

			if entry == nil {
				// No reputation recorded — neutral, allow with info
				report.AddCheck("faction_reputation_gate", true, "info",
					fmt.Sprintf("Faction %s has no recorded reputation — assumed neutral", factionID), "")
				continue
			}

			score := entry.Score
			status := domain.ScoreToStatus(score)

			// Check for hostile factions being helpful
			if score <= -80 && isHelpfulReaction(ref.Reaction) {
				report.AddCheck("faction_reputation_gate", false, "error",
					fmt.Sprintf("Faction %s is hostile (score %d) but described as %s without cause", factionID, score, ref.Reaction),
					"")
				report.AddSuggestion(
					fmt.Sprintf("Hostile faction %s cannot be helpful", factionID),
					"Provide a narrative cause for the change or adjust the faction's behavior",
					"Hostile factions require explicit cause to aid the party")
				continue
			}

			// Friendly/allied factions aiding is normal
			if score >= 30 && isHelpfulReaction(ref.Reaction) {
				report.AddCheck("faction_reputation_gate", true, "info",
					fmt.Sprintf("Faction %s is %s (score %d) and described as %s — consistent", factionID, status, score, ref.Reaction),
					"")
				continue
			}

			// Default: check passes with info
			report.AddCheck("faction_reputation_gate", true, "info",
				fmt.Sprintf("Faction %s reputation: %s (%d), described as %s", factionID, status, score, ref.Reaction),
				"")
		}
	}

	// WotC Format Validations (NEW)
	if proposal.Type == "act" {
		e.validateWotCFormat(report, proposal.Content, proposal.Type)
	}

	report.ComputeOverallStatus()
	return report, nil
}

// validateWotCFormat applies WotC professional format validations
func (e *ValidationEngine) validateWotCFormat(report *domain.ValidationReport, content string, proposalType string) {
	// Validation 1: Developments structure (3-5 branches with IF-THEN)
	devResult := validators.ValidateDevelopments(content)
	if !devResult.Valid {
		for _, err := range devResult.Errors {
			report.AddCheck("wotc_developments", false, "error", err.Message, "")
		}
		report.AddSuggestion(
			"Developments section incomplete",
			"Add 3-5 decision branches with **If [condition]**: structure, **Consequence:**, and **Recovery:** paths",
			"WotC standards require multiple decision branches per area")
	}

	// Validation 2: Multiple solution paths (stealth/social/combat)
	solutionsResult := validators.ValidateMultipleSolutions(content)
	if !solutionsResult.Valid {
		for _, err := range solutionsResult.Errors {
			report.AddCheck("wotc_multiple_solutions", false, "error", err.Message, "")
		}
		report.AddSuggestion(
			"Insufficient solution variety",
			"Add at least 2 different solution types per obstacle: Stealth (DC X), Social (DC Y Persuasion), or Combat",
			"WotC adventures offer multiple paths through encounters")
	}

	// Validation 3: Character hooks (2-3 per area)
	hooksResult := validators.ValidateCharacterHooks(content)
	if !hooksResult.Valid {
		for _, err := range hooksResult.Errors {
			report.AddCheck("wotc_character_hooks", false, "warning", err.Message, "")
		}
		report.AddSuggestion(
			"Character hooks missing or insufficient",
			"Add 2-3 character hooks per area tied to background, class, race, or faction",
			"Character hooks increase player engagement")
	}

	// Validation 4: Boxed text quality (100-600 words, second person, present tense)
	boxedResult := validators.ValidateBoxedText(content)
	if !boxedResult.Valid {
		for _, err := range boxedResult.Errors {
			report.AddCheck("wotc_boxed_text", false, "warning", err.Message, "")
		}
		report.AddSuggestion(
			"Boxed text needs improvement",
			"Write 100-600 words in second person present tense (ves, escuchas, sientes) with sensory details only",
			"Boxed text sets the scene for players")
	}

	// Validation 5: NPC word count (for NPC content proposals)
	if proposalType == "npc" {
		npcResult := validators.ValidateNPCWordCount(content)
		if !npcResult.Valid {
			for _, err := range npcResult.Errors {
				report.AddCheck("wotc_npc_word_count", false, "error", err.Message, "")
			}
			report.AddSuggestion(
				"NPC descriptions too short",
				"Major NPCs need 500-800 words (appearance, personality, voice, secrets, hooks, dialogue)",
				"WotC NPCs have detailed descriptions for DM reference")
		}
	}
}

func (e *ValidationEngine) checkRuleViolation(content string, rule domain.CanonRule) bool {
	contentLower := strings.ToLower(content)
	statementLower := strings.ToLower(rule.Statement)

	if strings.Contains(statementLower, "banned") || strings.Contains(statementLower, "prohibited") || strings.Contains(statementLower, "forbidden") {
		parts := strings.Fields(statementLower)
		for _, word := range parts {
			if len(word) > 3 && strings.Contains(contentLower, word) {
				return true
			}
		}
	}
	return false
}

func (e *ValidationEngine) computeConsistencyStats(report *domain.ConsistencyReport) {
	report.TotalChecks = len(report.Issues)
	report.Passed = 0
	report.Warnings = 0
	report.Errors = 0
	report.Criticals = 0

	for _, issue := range report.Issues {
		if issue.Passed {
			report.Passed++
		} else {
			switch issue.Severity {
			case "critical":
				report.Criticals++
			case "error":
				report.Errors++
			case "warning":
				report.Warnings++
			}
		}
	}

	// Passed count includes checks that aren't in issues (implicit passes)
	// For now, only track failures
	switch {
	case report.Criticals > 0:
		report.OverallHealth = "critical"
	case report.Errors > 0:
		report.OverallHealth = "poor"
	case report.Warnings > 0:
		report.OverallHealth = "fair"
	case report.TotalChecks > 0:
		report.OverallHealth = "good"
	default:
		report.OverallHealth = "excellent"
	}
}

// extractRewardNames extracts reward names from quest content
func (e *ValidationEngine) extractRewardNames(content string) []string {
	var rewards []string
	// Match patterns like "Reward: Item Name" or "rewards: Item Name"
	re := regexp.MustCompile(`(?i)(?:reward|rewards)[\s]*[:\-]?\s*([^\.\n]+)`)
	matches := re.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			reward := strings.TrimSpace(match[1])
			if reward != "" {
				rewards = append(rewards, reward)
			}
		}
	}
	return rewards
}

// extractCRs extracts challenge ratings from encounter content
func (e *ValidationEngine) extractCRs(content string) []int {
	var crs []int
	// Match patterns like "CR 5" or "CR: 3" or "(CR 1/2)"
	re := regexp.MustCompile(`(?i)CR\s*[:\-]?\s*(\d+)`)
	matches := re.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			cr, err := strconv.Atoi(match[1])
			if err == nil {
				crs = append(crs, cr)
			}
		}
	}
	return crs
}

// parseLevelRange extracts min and max level from campaign level_range
func (e *ValidationEngine) parseLevelRange(doc *domain.CanonDocument) [2]int {
	var result [2]int
	// Find mcguffin entity with level_range property
	for _, entity := range doc.Entities {
		if entity.Role == "mcguffin" {
			if lr, ok := entity.Properties["level_range"].(string); ok {
				parts := strings.Split(lr, "-")
				if len(parts) == 2 {
					min, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
					max, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
					result[0] = min
					result[1] = max
					return result
				}
			}
		}
	}
	// Default fallback
	result[0] = 1
	result[1] = 20
	return result
}

// extractLocationReferences extracts location names from content
func (e *ValidationEngine) extractLocationReferences(content string) []string {
	var locations []string
	// Match patterns like "at Location Name" or "in Location Name"
	re := regexp.MustCompile(`(?i)(?:at|in|inside|within|near|outside)\s+([A-Z][A-Za-z\s]+?)(?:\.|,|\s+(?:and|to|from|the|a|an)\s+|$)`)
	matches := re.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			loc := strings.TrimSpace(match[1])
			if len(loc) > 2 {
				locations = append(locations, loc)
			}
		}
	}
	return locations
}

// TimelineEvent represents a parsed timeline event reference
type TimelineEvent struct {
	ID            string
	Ordinal       int
	Prerequisites []string
}

// extractTimelineEvents extracts timeline event references from act content
func (e *ValidationEngine) extractTimelineEvents(content string) []TimelineEvent {
	var events []TimelineEvent
	// Match "evt-XXX" or "event XXX" patterns
	re := regexp.MustCompile(`(?i)(evt-[\w-]+)`)
	matches := re.FindAllStringSubmatch(content, -1)
	for i, match := range matches {
		if len(match) > 1 {
			events = append(events, TimelineEvent{
				ID:      match[1],
				Ordinal: i + 1,
			})
		}
	}
	return events
}

// findTimelineEvent finds a timeline event by ID
func (e *ValidationEngine) findTimelineEvent(timeline []domain.CanonTimelineEvent, eventID string) *TimelineEvent {
	for i, evt := range timeline {
		if evt.ID == eventID {
			return &TimelineEvent{
				ID:      evt.ID,
				Ordinal: i + 1,
			}
		}
	}
	return nil
}

// extractRequiredClues extracts clue IDs that are required from content
func (e *ValidationEngine) extractRequiredClues(content string) []string {
	var clues []string
	// Match "clue ID" or "requires clue" patterns
	re := regexp.MustCompile(`(?i)(?:clue|requires|needs)\s+([\w-]+)`)
	matches := re.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			clueID := strings.TrimSpace(match[1])
			if clueID != "" {
				clues = append(clues, clueID)
			}
		}
	}
	return clues
}

// hasAlternativePath checks if content mentions an alternative path for a clue
func (e *ValidationEngine) hasAlternativePath(content string, clueID string) bool {
	// Simple heuristic: check for words like "alternative", "or", "without"
	lowerContent := strings.ToLower(content)
	altPatterns := []string{"alternative", "or else", "without", "bypass", "skip"}
	for _, pattern := range altPatterns {
		if strings.Contains(lowerContent, pattern) {
			return true
		}
	}
	return false
}

// FactionReference represents a faction mention in content
type FactionReference struct {
	FactionID string
	Reaction  string
}

// extractFactionReferences extracts faction references from content
func (e *ValidationEngine) extractFactionReferences(content string) []FactionReference {
	var refs []FactionReference
	// Match "faction X reacts Y" or "the X is Y" patterns
	re := regexp.MustCompile(`(?i)(?:\bthe\s+)?(?:\bfaction\s+)?([A-Z][\w\s]+?)\s+(?:reacts?|is|are)\s+(\w+)`)
	matches := re.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 2 {
			factionName := strings.TrimSpace(match[1])
			// Avoid capturing "faction" as part of the name
			factionName = strings.TrimPrefix(factionName, "faction ")
			factionName = strings.TrimPrefix(factionName, "Faction ")
			refs = append(refs, FactionReference{
				FactionID: factionName,
				Reaction:  strings.ToLower(strings.TrimSpace(match[2])),
			})
		}
	}
	return refs
}

// AreaReference represents an area reference in chapter content
type AreaReference struct {
	AreaNumber string
	NPCs       []string
	Locations  []string
}

// extractAreaReferences extracts area references from chapter content
func (e *ValidationEngine) extractAreaReferences(content string) []AreaReference {
	var refs []AreaReference
	// Match "Area X" or "Area X: Title" patterns
	re := regexp.MustCompile(`(?i)area\s+(\d+)(?::\s*([^\n]+))?`)
	matches := re.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			areaNum := strings.TrimSpace(match[1])
			ref := AreaReference{AreaNumber: areaNum}
			
			// Try to extract NPCs mentioned in this area section (simplified)
			// In a full implementation, would parse the area section content
			npcRe := regexp.MustCompile(`(?i)\b(?:NPC|character|person)\s+named\s+([A-Z][\w]+)`)
			npcMatches := npcRe.FindAllStringSubmatch(content, -1)
			for _, npcMatch := range npcMatches {
				if len(npcMatch) > 1 {
					ref.NPCs = append(ref.NPCs, npcMatch[1])
				}
			}
			
			refs = append(refs, ref)
		}
	}
	return refs
}

func isHelpfulReaction(reaction string) bool {
	switch reaction {
	case "helpful", "friendly", "allied", "supportive", "aiding":
		return true
	default:
		return false
	}
}

// runConsolidationChecks dispatches the seven cross-file validation methods
// that delegate to the consolidation engine. They run only inside the
// ConsistencyScopeFull branch.
func (e *ValidationEngine) runConsolidationChecks(ctx context.Context, report *domain.ConsistencyReport, campaignID string) {
	if e.consolidator == nil {
		return
	}
	e.validateLoreCoherence(ctx, campaignID, report)
	e.validateEntityUniqueness(ctx, campaignID, report)
	e.validateEventCanonicalLocations(ctx, campaignID, report)
	e.validateStatBlockConsistency(ctx, campaignID, report)
	e.validateMapAssetsExist(ctx, campaignID, report)
	e.validateGeneratedFileFreshness(ctx, campaignID, report)
	e.validateNoDuplicateFiles(ctx, campaignID, report)
}

// emitConsolidationCheck writes a single summary CheckResult and one
// per-issue CheckResult. severityDefault applies when the analyzer didn't
// tag an issue with a severity. issues can be nil.
func (e *ValidationEngine) emitConsolidationCheck(report *domain.ConsistencyReport, rule string, res *consolidation.AnalysisResult, severityDefault string) {
	if res == nil {
		return
	}
	if !res.Passed {
		report.Issues = append(report.Issues, domain.CheckResult{
			Rule:     rule,
			Passed:   false,
			Severity: severityDefault,
			Message:  res.Message,
			Location: joinLocations(res.Locations),
		})
	}
	for _, issue := range res.Issues {
		sev := issue.Severity
		if sev == "" {
			sev = severityDefault
		}
		report.Issues = append(report.Issues, domain.CheckResult{
			Rule:     rule,
			Passed:   false,
			Severity: sev,
			Message:  issue.Message,
			Location: joinLocations(issue.Locations),
		})
	}
}

// validateLoreCoherence runs the LoreCoherence analyzer and surfaces
// treaty/event/primordial contradictions as critical CheckResults.
func (e *ValidationEngine) validateLoreCoherence(ctx context.Context, campaignID string, report *domain.ConsistencyReport) {
	res, err := e.consolidator.RunAnalyzer(ctx, campaignID, consolidation.NewLoreCoherence())
	if err != nil || res == nil {
		return
	}
	// Lore contradictions are canon-breaking → critical. The analyzer's
	// Issues carry their own severity, so we override to "critical" unless
	// it was already "warning" (event placement is less severe).
	original := res.Issues
	for i := range original {
		if original[i].Severity != "warning" {
			original[i].Severity = "critical"
		}
	}
	e.emitConsolidationCheck(report, "consolidation_lore_coherence", res, "critical")
}

// validateEntityUniqueness surfaces entity name collisions.
func (e *ValidationEngine) validateEntityUniqueness(ctx context.Context, campaignID string, report *domain.ConsistencyReport) {
	res, err := e.consolidator.RunAnalyzer(ctx, campaignID, consolidation.NewEntityResolver(0.85))
	if err != nil || res == nil {
		return
	}
	e.emitConsolidationCheck(report, "consolidation_entity_uniqueness", res, "warning")
}

// validateEventCanonicalLocations detects events placed in multiple files.
func (e *ValidationEngine) validateEventCanonicalLocations(ctx context.Context, campaignID string, report *domain.ConsistencyReport) {
	res, err := e.consolidator.RunAnalyzer(ctx, campaignID, consolidation.NewEventCanonizer())
	if err != nil || res == nil {
		return
	}
	e.emitConsolidationCheck(report, "consolidation_event_canonical_location", res, "error")
}

// validateStatBlockConsistency detects conflicting boss CR values.
func (e *ValidationEngine) validateStatBlockConsistency(ctx context.Context, campaignID string, report *domain.ConsistencyReport) {
	res, err := e.consolidator.RunAnalyzer(ctx, campaignID, consolidation.NewStatBlockResolver())
	if err != nil || res == nil {
		return
	}
	e.emitConsolidationCheck(report, "consolidation_stat_block_consistency", res, "error")
}

// validateMapAssetsExist checks for missing map/asset references.
func (e *ValidationEngine) validateMapAssetsExist(ctx context.Context, campaignID string, report *domain.ConsistencyReport) {
	campaignDir := filepath.Join(e.campaignDir, campaignID)
	res, err := e.consolidator.RunAnalyzer(ctx, campaignID, consolidation.NewMapReferenceChecker(campaignDir))
	if err != nil || res == nil {
		return
	}
	e.emitConsolidationCheck(report, "consolidation_map_assets_exist", res, "error")
}

// validateGeneratedFileFreshness reports stale campaign.md/INDEX.md.
func (e *ValidationEngine) validateGeneratedFileFreshness(ctx context.Context, campaignID string, report *domain.ConsistencyReport) {
	freshness, err := e.consolidator.VerifyFreshness(ctx, campaignID)
	if err != nil || freshness == nil {
		return
	}
	if freshness.CampaignMDStale {
		report.Issues = append(report.Issues, domain.CheckResult{
			Rule:     "consolidation_generated_file_freshness",
			Passed:   false,
			Severity: "warning",
			Message:  fmt.Sprintf("campaign.md is stale relative to sources (sources newest: %s)", freshness.SourcesNewest.Format("2006-01-02")),
			Location: "campaign.md",
		})
	}
	if freshness.IndexStale {
		report.Issues = append(report.Issues, domain.CheckResult{
			Rule:     "consolidation_generated_file_freshness",
			Passed:   false,
			Severity: "warning",
			Message:  fmt.Sprintf("INDEX.md is stale relative to sources (sources newest: %s)", freshness.SourcesNewest.Format("2006-01-02")),
			Location: "INDEX.md",
		})
	}
}

// validateNoDuplicateFiles detects identical-content files.
func (e *ValidationEngine) validateNoDuplicateFiles(ctx context.Context, campaignID string, report *domain.ConsistencyReport) {
	res, err := e.consolidator.RunAnalyzer(ctx, campaignID, consolidation.NewFileConsolidator())
	if err != nil || res == nil {
		return
	}
	// Only surface duplicate_file issues (stale_generated_file belongs
	// to validateGeneratedFileFreshness).
	filtered := &consolidation.AnalysisResult{
		Rule:      res.Rule,
		Passed:    res.Passed,
		Severity:  res.Severity,
		Message:   res.Message,
		Locations: res.Locations,
	}
	for _, issue := range res.Issues {
		if issue.Rule == "duplicate_file" {
			filtered.Issues = append(filtered.Issues, issue)
		}
	}
	if !res.Passed && len(filtered.Issues) > 0 {
		// Keep the global Passed signal only if duplicates were detected.
		e.emitConsolidationCheck(report, "consolidation_no_duplicate_files", filtered, "warning")
	}
}
