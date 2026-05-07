package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/pauvalls/grimorio/internal/domain"
)

// ValidationEngine performs rule-based validation of content proposals and campaign consistency
type ValidationEngine struct {
	canonService *CanonService
	stateService *NarrativeStateService
}

// NewValidationEngine creates a new validation engine
func NewValidationEngine(canonService *CanonService, stateService *NarrativeStateService) *ValidationEngine {
	return &ValidationEngine{
		canonService: canonService,
		stateService: stateService,
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
	// For full scope, also check acts if they existed in canon
	if scope == domain.ConsistencyScopeFull {
		// Additional full-scope checks can go here
	}

	// Compute report statistics
	e.computeConsistencyStats(report)

	return report, nil
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

	report.ComputeOverallStatus()
	return report, nil
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
