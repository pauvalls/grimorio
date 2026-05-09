package services

import (
	"context"
	"fmt"
	"github.com/pauvalls/grimorio/internal/domain"
)

// ConsequenceGenerator generates consequence tables for act transitions.
type ConsequenceGenerator struct {
	conseqRepo ConsequenceTableRepository
	factionRepo FactionRepository
	questRepo QuestRepositoryV3
}

// NewConsequenceGenerator creates a new ConsequenceGenerator.
func NewConsequenceGenerator(
	conseqRepo ConsequenceTableRepository,
	factionRepo FactionRepository,
	questRepo QuestRepositoryV3,
) *ConsequenceGenerator {
	return &ConsequenceGenerator{
		conseqRepo: conseqRepo,
		factionRepo: factionRepo,
		questRepo: questRepo,
	}
}

// GenerateConsequenceTable generates a consequence table for an act transition.
func (s *ConsequenceGenerator) GenerateConsequenceTable(ctx context.Context, campaignID string, fromAct, toAct int, questOutcomes []domain.QuestOutcome) (*domain.ConsequenceTable, error) {
	if fromAct >= toAct {
		return nil, fmt.Errorf("to_act must be greater than from_act")
	}

	table := &domain.ConsequenceTable{
		ID:                fmt.Sprintf("consequences_act_%d_to_%d", fromAct, toAct),
		CampaignID:        campaignID,
		FromAct:           fromAct,
		ToAct:             toAct,
		QuestOutcomes:     questOutcomes,
		FactionChanges:    s.calculateFactionChanges(questOutcomes),
		NPCChanges:        []domain.NPCChange{},
		NewOpportunities:  s.generateNewOpportunities(fromAct, toAct),
		LockedContent:     s.generateLockedContent(fromAct, toAct),
		WorldStateChanges: []domain.WorldStateChange{},
	}

	if err := table.Validate(); err != nil {
		return nil, fmt.Errorf("generated consequence table validation failed: %w", err)
	}

	return table, nil
}

// PropagateFactionChanges propagates faction reputation changes through alliances and enmities.
func (s *ConsequenceGenerator) PropagateFactionChanges(ctx context.Context, campaignID string, changes []domain.FactionChange) (*PropagationResult, error) {
	result := &PropagationResult{
		OriginalChanges:   changes,
		PropagatedChanges: []domain.FactionChange{},
	}

	// BFS propagation through faction relationships
	visited := make(map[string]bool)
	queue := make([]domain.FactionChange, len(changes))
	copy(queue, changes)

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current.FactionID] {
			continue
		}
		visited[current.FactionID] = true
		result.PropagatedChanges = append(result.PropagatedChanges, current)

		// TODO: Get allies and enemies from factionRepo and propagate
		// For now, just add the original changes
	}

	return result, nil
}

// TrackQuestOutcomes tracks quest outcomes and updates reputation.
func (s *ConsequenceGenerator) TrackQuestOutcomes(ctx context.Context, campaignID string, outcomes []domain.QuestOutcome) error {
	for _, outcome := range outcomes {
		for _, repChange := range outcome.ReputationChanges {
			if err := s.factionRepo.UpdateReputation(ctx, campaignID, repChange.FactionID, repChange.Delta); err != nil {
				return fmt.Errorf("failed to update reputation: %w", err)
			}
		}
	}
	return nil
}

// ConsolidateOutcomes consolidates multiple quest outcomes into a single table.
func (s *ConsequenceGenerator) ConsolidateOutcomes(campaignID string, fromAct, toAct int, outcomes []domain.QuestOutcome) *domain.ConsequenceTable {
	factionDeltas := make(map[string]int)
	factionReasons := make(map[string][]string)

	for _, outcome := range outcomes {
		for _, repChange := range outcome.ReputationChanges {
			factionDeltas[repChange.FactionID] += repChange.Delta
			factionReasons[repChange.FactionID] = append(factionReasons[repChange.FactionID], repChange.Reason)
		}
	}

	factionChanges := []domain.FactionChange{}
	for factionID, delta := range factionDeltas {
		factionChanges = append(factionChanges, domain.FactionChange{
			FactionID: factionID,
			Delta:     delta,
			Reason:    joinStrings(factionReasons[factionID], "; "),
		})
	}

	return &domain.ConsequenceTable{
		ID:             fmt.Sprintf("consolidated_act_%d_to_%d", fromAct, toAct),
		CampaignID:     campaignID,
		FromAct:        fromAct,
		ToAct:          toAct,
		QuestOutcomes:  outcomes,
		FactionChanges: factionChanges,
	}
}

func (s *ConsequenceGenerator) calculateFactionChanges(questOutcomes []domain.QuestOutcome) []domain.FactionChange {
	changes := []domain.FactionChange{}
	for _, outcome := range questOutcomes {
		for _, repChange := range outcome.ReputationChanges {
			changes = append(changes, domain.FactionChange{
				FactionID: repChange.FactionID,
				Delta:     repChange.Delta,
				Reason:    repChange.Reason,
			})
		}
	}
	return changes
}

func (s *ConsequenceGenerator) generateNewOpportunities(fromAct, toAct int) []domain.Opportunity {
	return []domain.Opportunity{
		{
			Type:            "quest",
			ReferenceID:     fmt.Sprintf("quest_act_%d_new", toAct),
			UnlockCondition: fmt.Sprintf("Complete Act %d", fromAct),
			Description:     "New quests available in next act",
		},
	}
}

func (s *ConsequenceGenerator) generateLockedContent(fromAct, toAct int) []domain.LockedContent {
	return []domain.LockedContent{
		{
			Type:            "area",
			ReferenceID:     fmt.Sprintf("area_act_%d_locked", fromAct),
			LockReason:      "Area from previous act",
			Unlockable:      false,
		},
	}
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// PropagationResult holds the result of faction change propagation.
type PropagationResult struct {
	OriginalChanges   []domain.FactionChange
	PropagatedChanges []domain.FactionChange
}
