package services

import (
	"context"
	"fmt"
	"github.com/pauvalls/grimorio/internal/domain"
)

// PlayerMapGenerator generates player-facing map variants.
type PlayerMapGenerator struct {
	mapRepo PlayerMapRepository
}

// NewPlayerMapGenerator creates a new PlayerMapGenerator.
func NewPlayerMapGenerator(mapRepo PlayerMapRepository) *PlayerMapGenerator {
	return &PlayerMapGenerator{mapRepo: mapRepo}
}

// GeneratePlayerVariant generates a player-facing map from a DM map.
func (s *PlayerMapGenerator) GeneratePlayerVariant(ctx context.Context, dmMapID string, areaID string, areaTitle string) (*domain.PlayerMap, error) {
	playerMap := &domain.PlayerMap{
		ID:               fmt.Sprintf("player_map_%s", dmMapID),
		SourceMapID:      dmMapID,
		AreaID:           areaID,
		Title:            areaTitle + " - Player Map",
		Description:      "Player-facing map with secret features hidden",
		IncludedFeatures: []domain.MapFeature{},
		ExcludedFeatures: []string{},
		ExportFormats:    []string{"svg", "pdf", "png"},
		GridVisible:      true,
		Scale:            "1 square = 5 feet",
	}

	if err := playerMap.Validate(); err != nil {
		return nil, fmt.Errorf("generated player map validation failed: %w", err)
	}

	return playerMap, nil
}

// ExportMap exports a map in the specified format.
func (s *PlayerMapGenerator) ExportMap(ctx context.Context, mapID string, format string, includeGrid bool) ([]byte, error) {
	if !isValidExportFormat(format) {
		return nil, fmt.Errorf("invalid export format: %s", format)
	}
	// TODO: Implement actual export with wkhtmltopdf or similar
	return []byte{}, nil
}

// RedactSecretFeatures removes secret features from a DM map.
func (s *PlayerMapGenerator) RedactSecretFeatures(dmFeatures []domain.MapFeature) (included []domain.MapFeature, excluded []string) {
	for _, feature := range dmFeatures {
		if feature.IsSecret {
			excluded = append(excluded, feature.ID)
		} else {
			included = append(included, feature)
		}
	}
	return included, excluded
}

// GenerateBlindMap generates a map with no labels.
func (s *PlayerMapGenerator) GenerateBlindMap(ctx context.Context, areaID string) (*domain.PlayerMap, error) {
	playerMap := &domain.PlayerMap{
		ID:               fmt.Sprintf("blind_map_%s", areaID),
		AreaID:           areaID,
		Title:            "Area Map (Blind)",
		Description:      "Map with no labels for exploration",
		IncludedFeatures: []domain.MapFeature{},
		ExcludedFeatures: []string{},
		ExportFormats:    []string{"svg", "pdf"},
		GridVisible:      false,
	}
	return playerMap, nil
}

// GenerateFogOfWarVariant generates a map with fog of war.
func (s *PlayerMapGenerator) GenerateFogOfWarVariant(ctx context.Context, areaID string) (*domain.PlayerMap, error) {
	playerMap := &domain.PlayerMap{
		ID:               fmt.Sprintf("fog_map_%s", areaID),
		AreaID:           areaID,
		Title:            "Area Map (Fog of War)",
		Description:      "Map with progressive reveal",
		IncludedFeatures: []domain.MapFeature{},
		ExcludedFeatures: []string{},
		ExportFormats:    []string{"svg"},
		GridVisible:      true,
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
