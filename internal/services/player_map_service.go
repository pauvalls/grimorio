package services

import (
	"context"
	"fmt"
	"github.com/pauvalls/grimorio/internal/domain"
)

// PlayerMapService generates player-facing maps with secrets redacted.
type PlayerMapService struct {
	mapRepo PlayerMapRepository
}

// PlayerMapRepository defines repository interface for player maps.
type PlayerMapRepository interface {
	Create(ctx context.Context, campaignID string, playerMap *domain.PlayerMap) error
	Read(ctx context.Context, campaignID string, mapID string) (*domain.PlayerMap, error)
	Update(ctx context.Context, campaignID string, playerMap *domain.PlayerMap) error
	Delete(ctx context.Context, campaignID string, mapID string) error
	GetByArea(ctx context.Context, campaignID string, areaID string) ([]*domain.PlayerMap, error)
}

// NewPlayerMapService creates a new PlayerMapService.
func NewPlayerMapService(mapRepo PlayerMapRepository) *PlayerMapService {
	return &PlayerMapService{mapRepo: mapRepo}
}

// GeneratePlayerVariant generates a player-facing map from a DM map.
func (s *PlayerMapService) GeneratePlayerVariant(ctx context.Context, dmMapID string, areaID string) (*domain.PlayerMap, error) {
	// TODO: Load DM map and filter secrets
	playerMap := &domain.PlayerMap{
		ID:              fmt.Sprintf("player_map_%s", dmMapID),
		SourceMapID:     dmMapID,
		AreaID:          areaID,
		Title:           "Area Map",
		Description:     "Player-facing map with secret features hidden",
		IncludedFeatures: []domain.MapFeature{},
		ExcludedFeatures: []string{},
		ExportFormats:   []string{"svg", "pdf", "png"},
		GridVisible:     true,
		Scale:           "1 square = 5 feet",
	}

	if err := playerMap.Validate(); err != nil {
		return nil, fmt.Errorf("generated player map validation failed: %w", err)
	}

	return playerMap, nil
}

// ExportMap exports a map in the specified format.
func (s *PlayerMapService) ExportMap(ctx context.Context, mapID string, format string, includeGrid bool) ([]byte, error) {
	if !isValidExportFormat(format) {
		return nil, fmt.Errorf("invalid export format: %s", format)
	}
	// TODO: Implement actual export
	return []byte{}, nil
}

// GetPlayerMapsByArea retrieves player maps for an area.
func (s *PlayerMapService) GetPlayerMapsByArea(ctx context.Context, campaignID string, areaID string) ([]*domain.PlayerMap, error) {
	return s.mapRepo.GetByArea(ctx, campaignID, areaID)
}

// RedactSecretFeatures removes secret features from a DM map.
func (s *PlayerMapService) RedactSecretFeatures(ctx context.Context, dmMap *domain.PlayerMap) (*domain.PlayerMap, error) {
	if dmMap == nil {
		return nil, fmt.Errorf("dmMap cannot be nil")
	}

	playerMap := &domain.PlayerMap{
		ID:               fmt.Sprintf("player_%s", dmMap.ID),
		SourceMapID:      dmMap.ID,
		CampaignID:       dmMap.CampaignID,
		AreaID:           dmMap.AreaID,
		Title:            dmMap.Title,
		Description:      dmMap.Description,
		IncludedFeatures: []domain.MapFeature{},
		ExcludedFeatures: []string{},
		ExportFormats:    dmMap.ExportFormats,
		GridVisible:      dmMap.GridVisible,
		Scale:            dmMap.Scale,
	}

	// Filter out secret features
	for _, feature := range dmMap.IncludedFeatures {
		if feature.IsSecret {
			playerMap.ExcludedFeatures = append(playerMap.ExcludedFeatures, feature.ID)
		} else {
			playerMap.IncludedFeatures = append(playerMap.IncludedFeatures, feature)
		}
	}

	if err := playerMap.Validate(); err != nil {
		return nil, fmt.Errorf("redacted map validation failed: %w", err)
	}

	return playerMap, nil
}

func isValidExportFormat(format string) bool {
	switch format {
	case "svg", "pdf", "png":
		return true
	default:
		return false
	}
}
