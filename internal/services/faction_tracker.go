package services

import (
	"context"
	"fmt"
	"github.com/pauvalls/grimorio/internal/domain"
)

// FactionTracker tracks faction reputation across sessions.
type FactionTracker struct {
	factionRepo FactionRepository
}

// NewFactionTracker creates a new FactionTracker.
func NewFactionTracker(factionRepo FactionRepository) *FactionTracker {
	return &FactionTracker{factionRepo: factionRepo}
}

// GenerateReputationTable generates a faction reputation table.
func (s *FactionTracker) GenerateReputationTable(ctx context.Context, campaignID string) (*FactionReputationTable, error) {
	factions, err := s.factionRepo.GetAll(ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to get factions: %w", err)
	}

	table := &FactionReputationTable{
		CampaignID: campaignID,
		Factions:   []FactionReputationEntry{},
	}

	for _, faction := range factions {
		entry := FactionReputationEntry{
			FactionID:   faction.ID,
			FactionName: faction.Name,
			Reputation:  0, // TODO: Get from session history
			Status:      getReputationStatus(0),
			Allies:      []string{},
			Enemies:     []string{},
		}
		table.Factions = append(table.Factions, entry)
	}

	return table, nil
}

// TrackReputationHistory tracks reputation changes per session.
func (s *FactionTracker) TrackReputationHistory(ctx context.Context, campaignID, factionID string, sessionNum, delta int, reason string) error {
	// TODO: Implement history tracking
	return nil
}

// GetReputationStatus returns status string for reputation value.
func getReputationStatus(rep int) string {
	switch {
	case rep >= 75:
		return "Ally"
	case rep >= 25:
		return "Friendly"
	case rep >= -25:
		return "Neutral"
	case rep >= -75:
		return "Hostile"
	default:
		return "Enemy"
	}
}

// FactionReputationTable represents a faction reputation tracking table.
type FactionReputationTable struct {
	CampaignID string                 `json:"campaign_id"`
	Factions   []FactionReputationEntry `json:"factions"`
}

// FactionReputationEntry represents a single faction's reputation.
type FactionReputationEntry struct {
	FactionID   string   `json:"faction_id"`
	FactionName string   `json:"faction_name"`
	Reputation  int      `json:"reputation"`
	Status      string   `json:"status"`
	Allies      []string `json:"allies"`
	Enemies     []string `json:"enemies"`
}

// ExportReputationTable exports the reputation table as PDF.
func (s *FactionTracker) ExportReputationTable(ctx context.Context, campaignID string) ([]byte, error) {
	table, err := s.GenerateReputationTable(ctx, campaignID)
	if err != nil {
		return nil, err
	}
	// TODO: Implement PDF export
	return []byte{}, nil
}
