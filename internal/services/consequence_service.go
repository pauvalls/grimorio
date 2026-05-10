package services

import (
	"context"
	"fmt"
	"github.com/pauvalls/grimorio/internal/domain"
)

// ConsequenceService handles consequence table generation and propagation.
type ConsequenceService struct {
	conseqRepo ConsequenceTableRepository
	factionRepo FactionRepository
}

// ConsequenceTableRepository defines repository interface for consequence tables.
type ConsequenceTableRepository interface {
	Create(ctx context.Context, campaignID string, table *domain.ConsequenceTable) error
	Read(ctx context.Context, campaignID string, tableID string) (*domain.ConsequenceTable, error)
	Update(ctx context.Context, campaignID string, table *domain.ConsequenceTable) error
	Delete(ctx context.Context, campaignID string, tableID string) error
	GetActive(ctx context.Context, campaignID string) ([]*domain.ConsequenceTable, error)
}

// FactionRepository defines repository interface for factions.
type FactionRepository interface {
	GetAll(ctx context.Context, campaignID string) ([]*domain.Faction, error)
	UpdateReputation(ctx context.Context, campaignID string, factionID string, delta int) error
}

// NewConsequenceService creates a new ConsequenceService.
func NewConsequenceService(conseqRepo ConsequenceTableRepository, factionRepo FactionRepository) *ConsequenceService {
	return &ConsequenceService{
		conseqRepo: conseqRepo,
		factionRepo: factionRepo,
	}
}

// GenerateConsequenceTable generates a consequence table for an act transition.
func (s *ConsequenceService) GenerateConsequenceTable(ctx context.Context, campaignID string, fromAct, toAct int) (*domain.ConsequenceTable, error) {
	if fromAct >= toAct {
		return nil, fmt.Errorf("to_act must be greater than from_act")
	}

	table := &domain.ConsequenceTable{
		ID:                fmt.Sprintf("consequences_act_%d_to_%d", fromAct, toAct),
		CampaignID:        campaignID,
		FromAct:           fromAct,
		ToAct:             toAct,
		QuestOutcomes:     []domain.QuestOutcome{},
		FactionChanges:    []domain.FactionChange{},
		NPCChanges:        []domain.NPCChange{},
		NewOpportunities:  []domain.Opportunity{},
		LockedContent:     []domain.LockedContent{},
		WorldStateChanges: []domain.WorldStateChange{},
	}

	if err := table.Validate(); err != nil {
		return nil, fmt.Errorf("generated consequence table validation failed: %w", err)
	}

	return table, nil
}

// PropagateFactionChanges propagates faction reputation changes through alliances and enmities.
func (s *ConsequenceService) PropagateFactionChanges(ctx context.Context, campaignID string, changes []domain.FactionChange) (*PropagationResult, error) {
	result := &PropagationResult{
		OriginalChanges: changes,
		PropagatedChanges: []domain.FactionChange{},
	}

	// BFS propagation through faction relationships
	// TODO: Implement full propagation logic
	result.PropagatedChanges = append(result.PropagatedChanges, changes...)

	return result, nil
}

// TrackQuestOutcomes tracks quest outcomes and updates reputation.
func (s *ConsequenceService) TrackQuestOutcomes(ctx context.Context, campaignID string, outcomes []domain.QuestOutcome) error {
	for _, outcome := range outcomes {
		for _, repChange := range outcome.ReputationChanges {
			if err := s.factionRepo.UpdateReputation(ctx, campaignID, repChange.FactionID, repChange.Delta); err != nil {
				return fmt.Errorf("failed to update reputation: %w", err)
			}
		}
	}
	return nil
}

// GetActiveConsequences retrieves active consequence tables.
func (s *ConsequenceService) GetActiveConsequences(ctx context.Context, campaignID string) ([]*domain.ConsequenceTable, error) {
	return s.conseqRepo.GetActive(ctx, campaignID)
}

// EvaluateConsequences evaluates consequence state for conflicts.
func (s *ConsequenceService) EvaluateConsequences(ctx context.Context, campaignID string) (*ConsequenceEvaluation, error) {
	// TODO: Implement evaluation logic
	return &ConsequenceEvaluation{
		CampaignID: campaignID,
		Conflicts:  []string{},
		Valid:      true,
	}, nil
}

// PropagationResult holds the result of faction change propagation.
type PropagationResult struct {
	OriginalChanges   []domain.FactionChange
	PropagatedChanges []domain.FactionChange
}

// ConsequenceEvaluation holds the result of consequence evaluation.
type ConsequenceEvaluation struct {
	CampaignID string
	Conflicts  []string
	Valid      bool
}
