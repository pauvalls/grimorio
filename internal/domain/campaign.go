package domain

import (
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
	ID         string    `json:"id"`
	CampaignID string    `json:"campaign_id"`
	Number     int       `json:"number"`
	Title      string    `json:"title"`
	Content    string    `json:"content"` // Markdown content
	Summary    string    `json:"summary"` // Auto-generated or provided
	KeyEvents  []string  `json:"key_events"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
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
