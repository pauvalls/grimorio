package services

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

// SessionPrepService synthesizes a prep sheet from NarrativeState + CanonDocument
type SessionPrepService struct {
	canonRepo   repository.CanonRepository
	stateRepo   repository.NarrativeStateRepository
	factionRepo repository.FactionReputationRepository
	generator   *SessionGenerator
}

// NewSessionPrepService creates a new session prep service
func NewSessionPrepService(
	canonRepo repository.CanonRepository,
	stateRepo repository.NarrativeStateRepository,
	factionRepo repository.FactionReputationRepository,
) *SessionPrepService {
	return &SessionPrepService{
		canonRepo:   canonRepo,
		stateRepo:   stateRepo,
		factionRepo: factionRepo,
		generator:   NewSessionGenerator(canonRepo, stateRepo),
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

	// Load faction reputation matrix if available
	var factionMatrix *domain.FactionReputationMatrix
	if s.factionRepo != nil {
		factionMatrix, _ = s.factionRepo.Load(campaignID)
	}

	// PreviouslyOn from last 3 sessions + arc context
	prep.PreviouslyOn = s.generatePreviouslyOn(state)

	// Active quests
	prep.ActiveQuests = s.extractActiveQuests(state)

	// Relevant NPCs (alive, connected to active quests)
	prep.RelevantNPCs = s.findRelevantNPCs(state, doc)

	// Likely scenarios from active quests + decisions + pending effects + factions
	prep.LikelyScenarios = s.generateLikelyScenarios(state, factionMatrix, sessionNum)

	// Reminders
	prep.Reminders = s.generateReminders(state, doc, sessionNum)

	// Populate new fields
	prep.PendingEffects = getPendingEffects(state, sessionNum)
	if factionMatrix != nil {
		prep.FactionSnapshot = factionMatrix.Entries
	}

	// Chapter objectives (NEW)
	if state.CurrentChapter != "" {
		objectives := GetChapterObjectives(doc, state.CurrentChapter)
		if len(objectives) > 0 {
			completedIDs := make([]string, len(state.CompletedQuests))
			for i, q := range state.CompletedQuests {
				completedIDs[i] = q.ID
			}
			completed := CountCompletedObjectives(completedIDs, objectives)
			prep.Reminders = append(prep.Reminders,
				fmt.Sprintf("📊 Chapter %s progress: %d/%d objectives", state.CurrentChapter, completed, len(objectives)))
		}
	}

	if len(state.SessionLog) == 0 {
		warnings = append(warnings, "no previous sessions found")
	}
	if len(state.ActiveQuests) == 0 {
		warnings = append(warnings, "no active quests")
	}

	return prep, warnings, nil
}

func getPendingEffects(state *domain.NarrativeState, sessionNum int) []domain.DelayedEffect {
	if state == nil {
		return nil
	}
	var result []domain.DelayedEffect
	for _, de := range state.PendingEffects {
		if !de.Applied && de.ApplySession <= sessionNum {
			result = append(result, de)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ApplySession < result[j].ApplySession
	})
	return result
}

func (s *SessionPrepService) generatePreviouslyOn(state *domain.NarrativeState) string {
	if len(state.SessionLog) == 0 {
		return "No previous sessions recorded. This is the beginning of the campaign."
	}

	currentSession := state.CurrentSession
	if currentSession == 0 {
		currentSession = len(state.SessionLog)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Arc context: %d sessions recorded, currently in session %d.\n", len(state.SessionLog), currentSession)

	start := len(state.SessionLog) - 3
	if start < 0 {
		start = 0
	}
	for i := len(state.SessionLog) - 1; i >= start; i-- {
		rec := state.SessionLog[i]
		fmt.Fprintf(&b, "\nSession %d: %s", rec.SessionNum, rec.Summary)
		if len(rec.KeyDecisions) > 0 {
			b.WriteString("\n  Key decisions:")
			for _, d := range rec.KeyDecisions {
				fmt.Fprintf(&b, "\n    - %s (chose: %s)", d.Description, d.ChoiceMade)
			}
		}
	}
	return b.String()
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

func (s *SessionPrepService) generateLikelyScenarios(state *domain.NarrativeState, factionMatrix *domain.FactionReputationMatrix, targetSession int) []string {
	const maxScenarios = 7

	var scenarios []string

	// Priority 1: Pending delayed effects due this session
	var pendingEffects []domain.DelayedEffect
	for _, de := range state.PendingEffects {
		if !de.Applied && de.ApplySession <= targetSession {
			pendingEffects = append(pendingEffects, de)
		}
	}
	sort.Slice(pendingEffects, func(i, j int) bool {
		return pendingEffects[i].ApplySession < pendingEffects[j].ApplySession
	})
	for _, de := range pendingEffects {
		scenarios = append(scenarios, fmt.Sprintf("Delayed consequence due: %s (target: %s)", de.Description, de.Target))
	}

	// Priority 2: Unresolved decisions from ALL sessions
	var unresolvedDecisions []domain.Decision
	for _, rec := range state.SessionLog {
		for _, d := range rec.KeyDecisions {
			if d.ImpactScope != "resolved" {
				unresolvedDecisions = append(unresolvedDecisions, d)
			}
		}
	}
	for _, d := range unresolvedDecisions {
		scenarios = append(scenarios, fmt.Sprintf("Unresolved decision: %s (chose: %s)", d.Description, d.ChoiceMade))
	}

	// Priority 3: Faction reputation changes since previous session
	if factionMatrix != nil {
		for _, entry := range factionMatrix.Entries {
			for _, evt := range entry.History {
				if evt.Session == targetSession-1 {
					scenarios = append(scenarios, fmt.Sprintf("Faction change: %s reputation shifted by %d (%s)", entry.FactionID, evt.Delta, evt.Reason))
				}
			}
		}
	}

	// Priority 4: Active quest continuations
	for _, q := range state.ActiveQuests {
		scenarios = append(scenarios, fmt.Sprintf("Continue quest '%s' from Act %s", q.Name, q.SourceAct))
	}

	// Cap at maxScenarios
	if len(scenarios) > maxScenarios {
		scenarios = scenarios[:maxScenarios]
	}
	return scenarios
}

func (s *SessionPrepService) generateReminders(state *domain.NarrativeState, doc *domain.CanonDocument, targetSession int) []string {
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

	// Pending delayed effects due this session
	for _, de := range state.PendingEffects {
		if !de.Applied && de.ApplySession == targetSession {
			reminders = append(reminders, fmt.Sprintf("Delayed consequence due: %s (target: %s)", de.Description, de.Target))
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
