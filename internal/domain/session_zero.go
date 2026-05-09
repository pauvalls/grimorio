package domain

import (
	"errors"
	"fmt"
)

// SessionZeroGuide represents a comprehensive Session Zero guide for a campaign.
type SessionZeroGuide struct {
	CampaignID          string                `json:"campaign_id"`
	CampaignPitch       string                `json:"campaign_pitch"`
	Tone                string                `json:"tone"`
	Themes              []string              `json:"themes"`
	ContentWarnings     []ContentWarning      `json:"content_warnings"`
	SessionExpectations SessionExpectations   `json:"session_expectations"`
	SafetyTools         []SafetyTool          `json:"safety_tools"`
	CharacterCreation   CharacterCreationGuide `json:"character_creation"`
	HouseRules          []HouseRule           `json:"house_rules"`
	Agenda              []SessionAgendaItem   `json:"agenda"`
	ShockPoints         []ShockPoint          `json:"shock_points,omitempty"`
}

// ShockPoint represents a content warning with severity and safety tools.
type ShockPoint struct {
	Type         string   `json:"type"`         // Type of content (violence, horror, etc.)
	Severity     string   `json:"severity"`     // mild, moderate, intense
	Description  string   `json:"description"`  // Detailed description
	SafetyTools  []string `json:"safety_tools"` // Recommended safety tools for this shock point
}

// ContentWarning represents a content warning with severity.
type ContentWarning struct {
	Topic       string `json:"topic"`
	Severity    string `json:"severity"` // mild, moderate, intense
	Description string `json:"description"`
	Avoidable   bool   `json:"avoidable"`
}

// SessionExpectations represents table expectations and etiquette.
type SessionExpectations struct {
	SessionLength   string   `json:"session_length"`
	Frequency       string   `json:"frequency"`
	AttendancePolicy string  `json:"attendance_policy"`
	TableEtiquette  []string `json:"table_etiquette"`
}

// SafetyTool represents a safety tool for the table.
type SafetyTool struct {
	Name        string `json:"name"` // X-Card, Lines and Veils, Script Change
	Description string `json:"description"`
	HowToUse    string `json:"how_to_use"`
}

// CharacterCreationGuide represents character creation guidelines.
type CharacterCreationGuide struct {
	Level                 int      `json:"level"`
	AbilityScoreMethod    string   `json:"ability_score_method"` // 4d6 drop lowest, Standard Array, Point Buy
	AllowedSources        []string `json:"allowed_sources"`
	BannedSources         []string `json:"banned_sources"`
	BackgroundGuidance    string   `json:"background_guidance"`
	BondPrompts           []string `json:"bond_prompts"`
	IdealPrompts          []string `json:"ideal_prompts"`
	FlawPrompts           []string `json:"flaw_prompts"`
	PartyCohesionQuestions []string `json:"party_cohesion_questions"`
}

// HouseRule represents a house rule for the campaign.
type HouseRule struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Reason      string `json:"reason"`
	Example     string `json:"example,omitempty"`
}

// SessionAgendaItem represents an agenda item for Session Zero.
type SessionAgendaItem struct {
	Topic            string `json:"topic"`
	DurationMinutes  int    `json:"duration_minutes"`
	Description      string `json:"description"`
}

// Validate checks Session Zero guide validity.
func (s *SessionZeroGuide) Validate() error {
	if s.CampaignID == "" {
		return errors.New("campaign_id is required")
	}
	if s.CampaignPitch == "" {
		return errors.New("campaign_pitch is required")
	}
	if s.Tone == "" {
		return errors.New("tone is required")
	}
	if err := validateAgendaDuration(s.Agenda); err != nil {
		return fmt.Errorf("agenda validation failed: %w", err)
	}
	if err := validateCharacterCreation(s.CharacterCreation); err != nil {
		return fmt.Errorf("character creation validation failed: %w", err)
	}
	return nil
}

// validateAgendaDuration checks if agenda duration is realistic (90-180 minutes).
func validateAgendaDuration(agenda []SessionAgendaItem) error {
	if len(agenda) == 0 {
		return errors.New("agenda cannot be empty")
	}
	totalMinutes := 0
	for _, item := range agenda {
		if item.Topic == "" {
			return errors.New("agenda item topic is required")
		}
		if item.DurationMinutes <= 0 {
			return errors.New("agenda item duration must be positive")
		}
		totalMinutes += item.DurationMinutes
	}
	if totalMinutes < 90 || totalMinutes > 180 {
		return fmt.Errorf("total agenda duration (%d minutes) should be between 90-180 minutes", totalMinutes)
	}
	return nil
}

// validateCharacterCreation checks character creation guide validity.
func validateCharacterCreation(cc CharacterCreationGuide) error {
	if cc.Level < 1 || cc.Level > 20 {
		return fmt.Errorf("character level must be between 1 and 20, got %d", cc.Level)
	}
	if cc.AbilityScoreMethod == "" {
		return errors.New("ability_score_method is required")
	}
	return nil
}

// GetTotalAgendaDuration returns total agenda duration in minutes.
func (s *SessionZeroGuide) GetTotalAgendaDuration() int {
	total := 0
	for _, item := range s.Agenda {
		total += item.DurationMinutes
	}
	return total
}

// GetContentWarningsBySeverity filters warnings by severity level.
func (s *SessionZeroGuide) GetContentWarningsBySeverity(severity string) []ContentWarning {
	var result []ContentWarning
	for _, w := range s.ContentWarnings {
		if w.Severity == severity {
			result = append(result, w)
		}
	}
	return result
}
