package domain

import (
	"errors"
)

// PlayerMap represents a player-facing map variant with secret features redacted.
type PlayerMap struct {
	ID              string       `json:"id"`
	SourceMapID     string       `json:"source_map_id"`
	CampaignID      string       `json:"campaign_id"`
	AreaID          string       `json:"area_id"`
	Title           string       `json:"title"`
	Description     string       `json:"description"`
	IncludedFeatures []MapFeature `json:"included_features"`
	ExcludedFeatures []string     `json:"excluded_features"` // IDs of hidden features
	ExportFormats   []string     `json:"export_formats"`      // svg, pdf, png
	GridVisible     bool         `json:"grid_visible"`
	Scale           string       `json:"scale,omitempty"` // "1 square = 5 feet"
}

// MapFeature represents a feature on a map.
type MapFeature struct {
	ID          string  `json:"id"`
	Type        string  `json:"type"` // room, door, passage, stairs, hazard, furniture, decoration
	Label       string  `json:"label,omitempty"`
	IsSecret    bool    `json:"is_secret"`
	Coordinates *Point  `json:"coordinates,omitempty"`
}

// Point represents a 2D coordinate.
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Validate checks player map validity.
func (m *PlayerMap) Validate() error {
	if m.ID == "" {
		return errors.New("id is required")
	}
	if m.SourceMapID == "" {
		return errors.New("source_map_id is required")
	}
	if m.CampaignID == "" {
		return errors.New("campaign_id is required")
	}
	if m.AreaID == "" {
		return errors.New("area_id is required")
	}
	if m.Title == "" {
		return errors.New("title is required")
	}
	// Ensure no secret features are included
	for _, f := range m.IncludedFeatures {
		if f.IsSecret {
			return errors.New("player map cannot include secret features")
		}
	}
	// Validate export formats
	for _, format := range m.ExportFormats {
		if !isValidExportFormat(format) {
			return errors.New("invalid export format: " + format)
		}
	}
	return nil
}

// isValidExportFormat checks if a format is valid.
func isValidExportFormat(format string) bool {
	switch format {
	case "svg", "pdf", "png":
		return true
	default:
		return false
	}
}

// HasFeature checks if a feature is included in the player map.
func (m *PlayerMap) HasFeature(featureID string) bool {
	for _, f := range m.IncludedFeatures {
		if f.ID == featureID {
			return true
		}
	}
	return false
}

// GetExcludedCount returns the number of excluded (secret) features.
func (m *PlayerMap) GetExcludedCount() int {
	return len(m.ExcludedFeatures)
}
