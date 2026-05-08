package domain

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Campaign represents a tabletop RPG campaign
type Campaign struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`        // kebab-case identifier
	Title       string    `json:"title"`       // Display title
	Setting     string    `json:"setting"`     // Brief setting description
	Description string    `json:"description"` // Full description
	Status      string    `json:"status"`      // active, paused, completed, archived
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Validate checks if the campaign is valid
func (c *Campaign) Validate() error {
	if c.Name == "" {
		return NewValidationError("name", "campaign name is required")
	}
	if !IsValidKebabCase(c.Name) {
		return NewValidationError("name", "campaign name must be kebab-case (lowercase letters, numbers, and hyphens only)")
	}
	if c.Title == "" {
		c.Title = c.Name
	}
	return nil
}

// CampaignSummary provides a lightweight overview of a campaign
type CampaignSummary struct {
	Name       string    `json:"name"`
	Title      string    `json:"title"`
	Setting    string    `json:"setting"`
	Status     string    `json:"status"`
	Acts       int       `json:"acts_count"`
	NPCs       int       `json:"npcs_count"`
	Characters int       `json:"characters_count"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Act represents a chapter/act of a campaign
type Act struct {
	ID                  string    `json:"id"`
	CampaignID          string    `json:"campaign_id"`
	Number              int       `json:"number"`
	Title               string    `json:"title"`
	Content             string    `json:"content"` // Markdown content
	Summary             string    `json:"summary"` // Auto-generated or provided
	KeyEvents           []string  `json:"key_events"`
	
	// Chapter Narrative Structure fields
	GameMode            string    `json:"game_mode"`                        // Primary mode (canonical list)
	GameModeSecondary   string    `json:"game_mode_secondary,omitempty"`    // Optional hybrid mode
	ChapterObjectives   []string  `json:"chapter_objectives"`               // 2-3 objectives
	EstimatedDuration   string    `json:"estimated_duration"`               // "2-3 sesiones"
	Tone                string    `json:"tone"`                             // Canonical tone
	RunningGuidance     string    `json:"running_guidance"`                 // 150-400 words
	AssetHandoff        string    `json:"asset_handoff"`                    // Concrete asset
	
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// Validate checks if the act is valid
func (a *Act) Validate() error {
	if a.CampaignID == "" {
		return NewValidationError("campaign_id", "campaign ID is required")
	}
	if a.Number <= 0 {
		return NewValidationError("number", "act number must be positive")
	}
	if a.Title == "" {
		return NewValidationError("title", "act title is required")
	}
	
	// Chapter Narrative Structure validation
	if a.GameMode == "" {
		return NewValidationError("game_mode", "game mode is required")
	}
	if !isValidGameMode(a.GameMode) {
		return NewValidationError("game_mode", fmt.Sprintf("invalid mode '%s'; must be one of: investigacion, sandbox_urbano, dungeon_lineal, escape, viaje, intriga, confrontacion, downtime", a.GameMode))
	}
	if a.GameModeSecondary != "" && a.GameModeSecondary == a.GameMode {
		return NewValidationError("game_mode_secondary", "secondary mode must differ from primary mode")
	}
	if len(a.ChapterObjectives) < 2 || len(a.ChapterObjectives) > 3 {
		return NewValidationError("chapter_objectives", "must have 2-3 objectives")
	}
	if a.EstimatedDuration == "" {
		return NewValidationError("estimated_duration", "estimated duration is required")
	}
	if !isValidDurationFormat(a.EstimatedDuration) {
		return NewValidationError("estimated_duration", "must match pattern: '1 sesión' or 'X-Y sesiones'")
	}
	if a.Tone == "" {
		return NewValidationError("tone", "tone is required")
	}
	if !isValidTone(a.Tone) {
		return NewValidationError("tone", fmt.Sprintf("invalid tone '%s'; must be one of: grim, whimsical, heroic, horror, political, mystery", a.Tone))
	}
	if a.RunningGuidance == "" {
		return NewValidationError("running_guidance", "running guidance is required")
	}
	wordCount := countWords(a.RunningGuidance)
	if wordCount < 150 || wordCount > 400 {
		return NewValidationError("running_guidance", fmt.Sprintf("must be 150-400 words; got %d", wordCount))
	}
	if a.AssetHandoff == "" {
		return NewValidationError("asset_handoff", "asset handoff is required")
	}
	
	return nil
}

// NPC represents a non-player character
type NPC struct {
	ID          string     `json:"id"`
	CampaignID  string     `json:"campaign_id"`
	Name        string     `json:"name"`
	Role        string     `json:"role"` // ally, enemy, neutral, merchant, etc.
	Faction     string     `json:"faction"`
	Description string     `json:"description"`
	Stats       *StatBlock `json:"stats,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// StatBlock represents creature/character statistics
type StatBlock struct {
	STR int `json:"str"`
	DEX int `json:"dex"`
	CON int `json:"con"`
	INT int `json:"int"`
	WIS int `json:"wis"`
	CHA int `json:"cha"`
	HP  int `json:"hp"`
	AC  int `json:"ac"`
}

// Monster represents a creature in the bestiary
type Monster struct {
	ID          string    `json:"id"`
	CampaignID  string    `json:"campaign_id"`
	Name        string    `json:"name"`
	CR          string    `json:"cr"`   // Challenge Rating
	Type        string    `json:"type"` // beast, humanoid, undead, etc.
	Size        string    `json:"size"` // Tiny, Small, Medium, Large, Huge, Gargantuan
	Stats       StatBlock `json:"stats"`
	Abilities   []string  `json:"abilities"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// Encounter represents a combat encounter or challenge
type Encounter struct {
	ID          string       `json:"id"`
	CampaignID  string       `json:"campaign_id"`
	Name        string       `json:"name"`
	Difficulty  string       `json:"difficulty"` // easy, medium, hard, deadly
	Location    string       `json:"location"`
	Monsters    []MonsterRef `json:"monsters"`
	Rewards     []Reward     `json:"rewards"`
	Description string       `json:"description"`
	CreatedAt   time.Time    `json:"created_at"`
}

// MonsterRef references a monster in an encounter
type MonsterRef struct {
	MonsterID string `json:"monster_id"`
	Name      string `json:"name"`
	Quantity  int    `json:"quantity"`
}

// Reward represents loot or experience
type Reward struct {
	Type        string `json:"type"` // gold, item, xp, reputation
	Description string `json:"description"`
	Value       string `json:"value"`
}

// Map represents a location or scene
type Map struct {
	ID          string    `json:"id"`
	CampaignID  string    `json:"campaign_id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"` // dungeon, city, landscape, battlemap
	Description string    `json:"description"`
	SVGPath     string    `json:"svg_path,omitempty"`
	ImagePath   string    `json:"image_path,omitempty"`
	Labels      []string  `json:"labels"`
	CreatedAt   time.Time `json:"created_at"`
}

// Helper functions for Chapter Narrative Structure validation

var validGameModes = map[string]bool{
	"investigacion":    true,
	"sandbox_urbano":   true,
	"dungeon_lineal":   true,
	"escape":           true,
	"viaje":            true,
	"intriga":          true,
	"confrontacion":    true,
	"downtime":         true,
}

func isValidGameMode(mode string) bool {
	return validGameModes[mode]
}

var validTones = map[string]bool{
	"grim":       true,
	"whimsical":  true,
	"heroic":     true,
	"horror":     true,
	"political":  true,
	"mystery":    true,
}

func isValidTone(tone string) bool {
	return validTones[tone]
}

func isValidDurationFormat(duration string) bool {
	// Pattern: "1 sesión" or "X-Y sesiones" where X and Y are 1+ digits
	matched, _ := regexp.MatchString(`^(\d+ sesión|\d+-\d+ sesiones)$`, duration)
	return matched
}

func countWords(text string) int {
	return len(strings.Fields(text))
}
