package services

import (
	"context"
	"fmt"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

// FactionService handles faction reputation business logic
type FactionService struct {
	canonRepo   repository.CanonRepository
	factionRepo repository.FactionReputationRepository
}

// NewFactionService creates a new faction service
func NewFactionService(canonRepo repository.CanonRepository, factionRepo repository.FactionReputationRepository) *FactionService {
	return &FactionService{
		canonRepo:   canonRepo,
		factionRepo: factionRepo,
	}
}

// UpdateReputation applies a reputation delta and propagates to related factions
func (s *FactionService) UpdateReputation(ctx context.Context, campaignID, factionID, partyID string, delta int8, reason, actionType string) (*domain.ReputationUpdateResult, error) {
	if delta == 0 {
		return nil, fmt.Errorf("zero_delta")
	}
	if delta < -100 || delta > 100 {
		return nil, fmt.Errorf("score_out_of_bounds")
	}

	doc, err := s.canonRepo.Load(campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to load canon: %w", err)
	}

	// Verify faction exists
	found := false
	for _, e := range doc.Entities {
		if e.ID == factionID && e.Type == domain.EntityTypeFaction {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("faction_not_found")
	}

	matrix, err := s.factionRepo.Load(campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to load reputation matrix: %w", err)
	}

	entry := matrix.GetEntry(factionID, partyID)
	// Get current session from narrative state if available, default to 1
	sessionNum := 1
	// Session number is tracked in narrative state, not canon
	// Default to 1 for now; could be enhanced with state service injection
	_ = s.canonRepo
	entry.ApplyDelta(delta, sessionNum, reason, actionType)

	result := &domain.ReputationUpdateResult{
		DirectChange:      *entry,
		PropagatedChanges: []domain.ReputationEntry{},
	}

	// BFS propagation over ally/enemy graph, cap 2 hops
	propagated := s.propagateReputation(doc, matrix, factionID, partyID, delta, 2)
	result.PropagatedChanges = propagated

	if err := s.factionRepo.Save(campaignID, matrix); err != nil {
		return nil, fmt.Errorf("failed to save reputation matrix: %w", err)
	}

	return result, nil
}

// propagateReputation performs BFS propagation of reputation changes
func (s *FactionService) propagateReputation(doc *domain.CanonDocument, matrix *domain.FactionReputationMatrix, sourceFactionID, partyID string, delta int8, maxHops int) []domain.ReputationEntry {
	type queueItem struct {
		factionID string
		hop       int
		delta     int8
	}

	visited := make(map[string]bool)
	visited[sourceFactionID] = true
	var propagated []domain.ReputationEntry
	queue := []queueItem{{factionID: sourceFactionID, hop: 0, delta: delta}}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.hop >= maxHops {
			continue
		}

		for _, rel := range doc.Relationships {
			var target string
			var sign int8 = 1
			if rel.From == current.factionID {
				target = rel.To
			} else if rel.To == current.factionID {
				target = rel.From
			} else {
				continue
			}

			if rel.Type == domain.RelationshipTypeEnemy {
				sign = -1
			} else if rel.Type != domain.RelationshipTypeAlly {
				continue
			}

			if visited[target] {
				continue
			}
			visited[target] = true

			nextDelta := int8(float64(current.delta) * 0.5 * float64(sign))
			if nextDelta == 0 {
				continue
			}

			entry := matrix.GetEntry(target, partyID)
			entry.ApplyDelta(nextDelta, 1, fmt.Sprintf("propagated from %s", current.factionID), "propagation")
			propagated = append(propagated, *entry)
			queue = append(queue, queueItem{factionID: target, hop: current.hop + 1, delta: nextDelta})
		}
	}

	return propagated
}

// GetReputationMatrix returns the full reputation matrix for a campaign
func (s *FactionService) GetReputationMatrix(ctx context.Context, campaignID string) (*domain.FactionReputationMatrix, error) {
	return s.factionRepo.Load(campaignID)
}

// GetPlayerReputationMatrix returns the reputation matrix filtering secret factions
func (s *FactionService) GetPlayerReputationMatrix(ctx context.Context, campaignID string) (*domain.FactionReputationMatrix, error) {
	doc, err := s.canonRepo.Load(campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to load canon: %w", err)
	}

	matrix, err := s.factionRepo.Load(campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to load reputation matrix: %w", err)
	}

	// Build set of secret faction IDs
	secretIDs := make(map[string]bool)
	for _, e := range doc.Entities {
		if e.Type == domain.EntityTypeFaction {
			if isSecret, ok := e.Properties["is_secret"].(bool); ok && isSecret {
				secretIDs[e.ID] = true
			}
		}
	}

	var filtered []domain.ReputationEntry
	for _, entry := range matrix.Entries {
		if !secretIDs[entry.FactionID] {
			filtered = append(filtered, entry)
		}
	}

	return &domain.FactionReputationMatrix{
		CampaignID: matrix.CampaignID,
		Entries:    filtered,
	}, nil
}
