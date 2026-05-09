package services

import (
	"context"
	"fmt"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

// SessionPrepService synthesizes a prep sheet from NarrativeState + CanonDocument
type SessionPrepService struct {
	canonRepo   repository.CanonRepository
	stateRepo   repository.NarrativeStateRepository
	generator   *SessionGenerator
}

// NewSessionPrepService creates a new session prep service
func NewSessionPrepService(canonRepo repository.CanonRepository, stateRepo repository.NarrativeStateRepository) *SessionPrepService {
	return &SessionPrepService{
		canonRepo: canonRepo,
		stateRepo: stateRepo,
		generator: NewSessionGenerator(canonRepo, stateRepo),
	}
}

// GetPrep synthesizes a prep sheet for the given campaign and session number
func (s *SessionPrepService) GetPrep(ctx context.Context, campaignID string, sessionNum int) (*domain.SessionPrep, []string, error) {
	var warnings []string

	state, err := s.stateRepo.Load(campaignID)
	if err != nil || state == nil {
		state = &domain.NarrativeState{
			SchemaVersion:  domain.SchemaVersionV2,
			CampaignID:     campaignID,
			CurrentSession: 0,
			SessionLog:     []domain.SessionRecord{},
			ActiveQuests:   []domain.QuestState{},
			DeadNPCs:       []domain.NPCDeathRecord{},
		}
	}

	if sessionNum <= 0 {
		sessionNum = state.CurrentSession + 1
	}

	doc, err := s.canonRepo.Load(campaignID)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("canon not found: %v", err))
		doc = &domain.CanonDocument{CampaignID: campaignID}
	}

	prep := &domain.SessionPrep{
		CampaignID: campaignID,
		SessionNum: sessionNum,
		PrepDate:   time.Now(),
	}

	// PreviouslyOn from last session
	prep.PreviouslyOn = s.generatePreviouslyOn(state)

	// Active quests
	prep.ActiveQuests = s.extractActiveQuests(state)

	// Relevant NPCs (alive, connected to active quests)
	prep.RelevantNPCs = s.findRelevantNPCs(state, doc)

	// Likely scenarios from active quests + decisions
	prep.LikelyScenarios = s.generateLikelyScenarios(state)

	// Reminders
	prep.Reminders = s.generateReminders(state, doc)

	if len(state.SessionLog) == 0 {
		warnings = append(warnings, "no previous sessions found")
	}
	if len(state.ActiveQuests) == 0 {
		warnings = append(warnings, "no active quests")
	}

	return prep, warnings, nil
}

func (s *SessionPrepService) generatePreviouslyOn(state *domain.NarrativeState) string {
	if len(state.SessionLog) == 0 {
		return "No previous sessions recorded. This is the beginning of the campaign."
	}
	last := state.SessionLog[len(state.SessionLog)-1]
	return fmt.Sprintf("Session %d: %s", last.SessionNum, last.Summary)
}

func (s *SessionPrepService) extractActiveQuests(state *domain.NarrativeState) []string {
	var quests []string
	for _, q := range state.ActiveQuests {
		quests = append(quests, fmt.Sprintf("[%s] %s (Act: %s, Giver: %s)", q.Status, q.Name, q.SourceAct, q.GiverNPC))
	}
	return quests
}

func (s *SessionPrepService) findRelevantNPCs(state *domain.NarrativeState, doc *domain.CanonDocument) []string {
	deadSet := make(map[string]bool)
	for _, d := range state.DeadNPCs {
		deadSet[d.NPCID] = true
	}

	// Build set of quest-giver and related NPCs from active quests
	relevantSet := make(map[string]bool)
	for _, q := range state.ActiveQuests {
		if q.GiverNPC != "" && !deadSet[q.GiverNPC] {
			relevantSet[q.GiverNPC] = true
		}
	}

	// Add canon entities that are alive and have relationships to active quest givers
	for _, e := range doc.Entities {
		if e.Type != domain.EntityTypeNPC {
			continue
		}
		if e.CanonState == domain.EntityStateDead {
			continue
		}
		if deadSet[e.ID] {
			continue
		}
		// Include if they are in the relevant set or connected via relationships
		if relevantSet[e.ID] {
			continue // already in set
		}
		for _, rel := range doc.Relationships {
			if rel.From == e.ID && relevantSet[rel.To] {
				relevantSet[e.ID] = true
				break
			}
			if rel.To == e.ID && relevantSet[rel.From] {
				relevantSet[e.ID] = true
				break
			}
		}
	}

	// Resolve names
	var result []string
	nameMap := make(map[string]string)
	for _, e := range doc.Entities {
		nameMap[e.ID] = e.Name
	}
	for id := range relevantSet {
		name := nameMap[id]
		if name == "" {
			name = id
		}
		result = append(result, name)
	}
	return result
}

func (s *SessionPrepService) generateLikelyScenarios(state *domain.NarrativeState) []string {
	if len(state.ActiveQuests) == 0 {
		return []string{}
	}

	var scenarios []string
	for _, q := range state.ActiveQuests {
		scenarios = append(scenarios, fmt.Sprintf("Continue quest '%s' from Act %s", q.Name, q.SourceAct))
	}

	// Add scenario from last session's decisions
	if len(state.SessionLog) > 0 {
		last := state.SessionLog[len(state.SessionLog)-1]
		for _, d := range last.KeyDecisions {
			scenarios = append(scenarios, fmt.Sprintf("Deal with consequences of decision: %s (chose: %s)", d.Description, d.ChoiceMade))
		}
	}

	// Cap at 5
	if len(scenarios) > 5 {
		scenarios = scenarios[:5]
	}
	return scenarios
}

func (s *SessionPrepService) generateReminders(state *domain.NarrativeState, doc *domain.CanonDocument) []string {
	var reminders []string

	// Check for dead NPCs whose canon state is still alive
	deadSet := make(map[string]bool)
	for _, d := range state.DeadNPCs {
		deadSet[d.NPCID] = true
	}
	for _, e := range doc.Entities {
		if e.Type == domain.EntityTypeNPC && deadSet[e.ID] && e.CanonState == domain.EntityStateAlive {
			reminders = append(reminders, fmt.Sprintf("NPC %s is dead in narrative state but alive in canon — update canon", e.Name))
		}
	}

	// Unresolved DM overrides
	for _, o := range state.DMOverrides {
		reminders = append(reminders, fmt.Sprintf("DM Override pending: %s on %s (%s → %s)", o.Reason, o.TargetID, o.Field, o.NewValue))
	}

	// Pending consequences from active quests
	for _, q := range state.ActiveQuests {
		for _, c := range q.Consequences {
			reminders = append(reminders, fmt.Sprintf("Quest '%s' has pending consequence: %s", q.Name, c))
		}
	}

	return reminders
}

// GetPrepWithScenarios extends GetPrep with encounter recommendations, loot suggestions, and NPC appearances.
func (s *SessionPrepService) GetPrepWithScenarios(ctx context.Context, campaignID string, sessionNum int) (*domain.SessionPrep, []string, error) {
	prep, warnings, err := s.GetPrep(ctx, campaignID, sessionNum)
	if err != nil {
		return nil, warnings, err
	}

	// Get party level for loot generation
	partyLevel := s.getPartyLevel(campaignID)

	// Generate encounter recommendations
	encounters, err := s.generator.GenerateEncounterRecommendations(ctx, campaignID, sessionNum)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("failed to generate encounters: %v", err))
	} else {
		prep.EncounterRecommendations = encounters
	}

	// Generate loot suggestions
	loot, err := s.generator.GenerateLootSuggestions(ctx, campaignID, partyLevel)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("failed to generate loot: %v", err))
	} else {
		prep.LootSuggestions = loot
	}

	// Generate NPC appearances
	npcs, err := s.generator.GenerateNPCAppearances(ctx, campaignID, sessionNum)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("failed to generate NPC appearances: %v", err))
	} else {
		prep.NPCAppearances = npcs
	}

	return prep, warnings, nil
}

// getPartyLevel estimates party level from campaign state.
func (s *SessionPrepService) getPartyLevel(campaignID string) int {
	state, err := s.stateRepo.Load(campaignID)
	if err != nil || state == nil {
		return 1
	}

	// Estimate based on session count (typical progression)
	if state.CurrentSession >= 10 {
		return 10
	}
	if state.CurrentSession >= 5 {
		return 5
	}
	return 1 + state.CurrentSession
}
