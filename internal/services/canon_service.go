package services

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pauvalls/grimorio/internal/cache"
	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// CanonService handles canon document business logic
type CanonService struct {
	canonRepo    repository.CanonRepository
	stateRepo    repository.NarrativeStateRepository
	cache        *cache.LRUCache[string, *domain.CanonDocument]
	cacheEnabled bool
	degraded     bool
}

// NewCanonService creates a new canon service
func NewCanonService(canonRepo repository.CanonRepository, stateRepo repository.NarrativeStateRepository) *CanonService {
	svc := &CanonService{
		canonRepo: canonRepo,
		stateRepo: stateRepo,
	}

	// Cache configuration from environment
	if os.Getenv("CANON_CACHE_DISABLED") == "1" {
		svc.cacheEnabled = false
	} else {
		svc.cacheEnabled = true
		size := 100
		if v := os.Getenv("CANON_CACHE_SIZE"); v != "" {
			if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
				size = parsed
			}
		}
		svc.cache = cache.NewLRU[string, *domain.CanonDocument](size)
	}

	return svc
}

// SetDegraded sets the degraded mode flag.
func (s *CanonService) SetDegraded(degraded bool) {
	s.degraded = degraded
}

// IsDegraded returns whether the service is in degraded mode.
func (s *CanonService) IsDegraded() bool {
	return s.degraded
}

// InitializeCanon creates a new CanonDocument from a campaign brief
func (s *CanonService) InitializeCanon(ctx context.Context, brief domain.CampaignBrief) (*domain.CanonDocument, error) {
	if brief.Name == "" {
		return nil, domain.NewValidationError("name", "campaign name is required")
	}
	if !domain.IsValidKebabCase(brief.Name) {
		return nil, domain.NewValidationError("name", "campaign name must be kebab-case")
	}

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
			{
				ID:        "fact-002",
				Category:  "story",
				Statement: brief.BriefDescription,
				Source:    "campaign_brief",
				Immutable: false,
				CreatedAt: now,
			},
		},
		Entities: []domain.CanonEntity{
			{
				ID:         fmt.Sprintf("mcguffin-%s", brief.Name),
				Name:       fmt.Sprintf("The %s McGuffin", cases.Title(language.English).String(brief.McGuffinType)),
				Type:       domain.EntityTypeItem,
				Role:       "mcguffin",
				CanonState: domain.EntityStateAlive,
				Properties: map[string]any{
					"type":        brief.McGuffinType,
					"level_range": brief.LevelRange,
				},
			},
		},
		Timeline:      []domain.CanonTimelineEvent{},
		Rules:         []domain.CanonRule{},
		Relationships: []domain.CanonRelationship{},
	}

	if err := s.canonRepo.Save(brief.Name, doc); err != nil {
		return nil, fmt.Errorf("failed to save initial canon: %w", err)
	}

	// Invalidate cache
	if s.cacheEnabled && s.cache != nil {
		s.cache.Remove(brief.Name)
	}

	// Initialize empty narrative state
	state := &domain.NarrativeState{
		SchemaVersion:  domain.SchemaVersionV2,
		CampaignID:     brief.Name,
		CurrentSession: 0,
		LastUpdated:    now,
	}
	if err := s.stateRepo.Save(brief.Name, state); err != nil {
		return nil, fmt.Errorf("failed to save initial narrative state: %w", err)
	}

	return doc, nil
}

// LoadCanon retrieves the canon document for a campaign
func (s *CanonService) LoadCanon(ctx context.Context, campaignID string) (*domain.CanonDocument, error) {
	if s.degraded {
		return &domain.CanonDocument{
			SchemaVersion: domain.SchemaVersionV2,
			CampaignID:    campaignID,
			Facts:         []domain.CanonFact{},
			Entities:      []domain.CanonEntity{},
			Timeline:      []domain.CanonTimelineEvent{},
			Rules:         []domain.CanonRule{},
			Relationships: []domain.CanonRelationship{},
		}, nil
	}

	// Check cache
	if s.cacheEnabled && s.cache != nil {
		if doc, ok := s.cache.Get(campaignID); ok {
			return doc, nil
		}
	}

	doc, err := s.canonRepo.Load(campaignID)
	if err != nil {
		return nil, err
	}

	// Store in cache
	if s.cacheEnabled && s.cache != nil {
		s.cache.Put(campaignID, doc)
	}

	return doc, nil
}

// SaveCanon persists a canon document
func (s *CanonService) SaveCanon(ctx context.Context, doc *domain.CanonDocument) error {
	if doc == nil {
		return domain.NewValidationError("doc", "canon document is required")
	}
	doc.UpdatedAt = time.Now()
	if err := s.canonRepo.Save(doc.CampaignID, doc); err != nil {
		return err
	}
	// Invalidate cache
	if s.cacheEnabled && s.cache != nil {
		s.cache.Remove(doc.CampaignID)
	}
	return nil
}

// RegisterFact adds a fact to the canon document
func (s *CanonService) RegisterFact(ctx context.Context, campaignID string, fact domain.CanonFact) error {
	doc, err := s.canonRepo.Load(campaignID)
	if err != nil {
		return fmt.Errorf("failed to load canon: %w", err)
	}

	if fact.CreatedAt.IsZero() {
		fact.CreatedAt = time.Now()
	}

	doc.Facts = append(doc.Facts, fact)
	doc.UpdatedAt = time.Now()

	if err := s.canonRepo.Save(campaignID, doc); err != nil {
		return err
	}
	// Invalidate cache
	if s.cacheEnabled && s.cache != nil {
		s.cache.Remove(campaignID)
	}
	return nil
}

// QueryEntity searches entities in the canon by filter
func (s *CanonService) QueryEntity(ctx context.Context, campaignID string, filter domain.EntityFilter) ([]domain.CanonEntity, error) {
	if s.degraded {
		return []domain.CanonEntity{}, nil
	}

	doc, err := s.LoadCanon(ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to load canon: %w", err)
	}

	var results []domain.CanonEntity
	for _, entity := range doc.Entities {
		if filter.Type != "" && entity.Type != filter.Type {
			continue
		}
		if filter.Role != "" && entity.Role != filter.Role {
			continue
		}
		if filter.CanonState != "" && entity.CanonState != filter.CanonState {
			continue
		}
		if filter.NameQuery != "" && !strings.Contains(strings.ToLower(entity.Name), strings.ToLower(filter.NameQuery)) {
			continue
		}
		results = append(results, entity)
	}

	return results, nil
}

// UpdateEntityState changes the canonical state of an entity
func (s *CanonService) UpdateEntityState(ctx context.Context, campaignID string, entityID string, state domain.EntityState) error {
	doc, err := s.canonRepo.Load(campaignID)
	if err != nil {
		return fmt.Errorf("failed to load canon: %w", err)
	}

	found := false
	for i := range doc.Entities {
		if doc.Entities[i].ID == entityID {
			doc.Entities[i].CanonState = state
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("entity not found: %s", entityID)
	}

	doc.UpdatedAt = time.Now()
	if err := s.canonRepo.Save(campaignID, doc); err != nil {
		return err
	}
	// Invalidate cache
	if s.cacheEnabled && s.cache != nil {
		s.cache.Remove(campaignID)
	}
	return nil
}

// ValidateProposal validates a content proposal against the canon
func (s *CanonService) ValidateProposal(ctx context.Context, campaignID string, proposal domain.ContentProposal) (*domain.ValidationReport, error) {
	report := &domain.ValidationReport{
		ArtifactID:    proposal.ID,
		ArtifactType:  proposal.Type,
		OverallStatus: "approved",
		Checks:        []domain.CheckResult{},
		Suggestions:   []domain.Suggestion{},
	}

	if s.degraded {
		return report, nil
	}

	doc, err := s.LoadCanon(ctx, campaignID)
	if err != nil {
		report.AddCheck("canon_load", false, "critical", fmt.Sprintf("Failed to load canon: %v", err), "")
		report.ComputeOverallStatus()
		return report, nil
	}

	state, _ := s.stateRepo.Load(campaignID)

	// Check entity references
	for _, ref := range proposal.EntityReferences {
		entity, found := s.findEntity(doc, ref.EntityID)
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

		// Check if NPC is dead in narrative state
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

	// Check lore rules
	for _, rule := range doc.Rules {
		if violation := s.checkRuleViolation(proposal.Content, rule); violation {
			report.AddCheck("lore_rule_compliance", false, "critical",
				fmt.Sprintf("Violates rule %s: %s", rule.ID, rule.Statement),
				"")
			report.AddSuggestion(
				fmt.Sprintf("Lore rule violation: %s", rule.Statement),
				"Adjust content to comply with the established rule",
				"Canon rules are immutable constraints on the world")
		}
	}

	report.ComputeOverallStatus()
	return report, nil
}

// GetRelationshipGraph builds the relationship graph for a campaign
func (s *CanonService) GetRelationshipGraph(ctx context.Context, campaignID string) (*domain.RelationshipGraph, error) {
	doc, err := s.LoadCanon(ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to load canon: %w", err)
	}

	return &domain.RelationshipGraph{
		CampaignID: doc.CampaignID,
		Nodes:      doc.Entities,
		Edges:      doc.Relationships,
	}, nil
}

func (s *CanonService) findEntity(doc *domain.CanonDocument, entityID string) (*domain.CanonEntity, bool) {
	for i := range doc.Entities {
		if doc.Entities[i].ID == entityID {
			return &doc.Entities[i], true
		}
	}
	return nil, false
}

func (s *CanonService) checkRuleViolation(content string, rule domain.CanonRule) bool {
	// Simple keyword-based check for lore rule compliance
	contentLower := strings.ToLower(content)
	statementLower := strings.ToLower(rule.Statement)

	// If the rule states something is banned/prohibited/forbidden
	if strings.Contains(statementLower, "banned") || strings.Contains(statementLower, "prohibited") || strings.Contains(statementLower, "forbidden") {
		// Extract the subject of the ban
		parts := strings.Fields(statementLower)
		for _, word := range parts {
			if len(word) > 3 && strings.Contains(contentLower, word) {
				// Content mentions something related to the ban - flag for review
				// This is a simplified heuristic; real implementation would use NLP
				return true
			}
		}
	}
	return false
}
