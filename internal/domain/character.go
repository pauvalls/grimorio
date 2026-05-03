package domain

import "time"

// Character represents a player character (PC)
type Character struct {
	ID            string            `json:"id"`
	CampaignID    string            `json:"campaign_id"`
	Name          string            `json:"name"`
	PlayerName    string            `json:"player_name,omitempty"`
	Race          string            `json:"race"`
	Class         string            `json:"class"`
	Level         int               `json:"level"`
	Background    string            `json:"background"`
	Alignment     string            `json:"alignment"`
	Stats         Stats             `json:"stats"`
	HP            HP                `json:"hp"`
	AC            int               `json:"ac"`
	Proficiency   int               `json:"proficiency_bonus"`
	Skills        map[string]bool   `json:"skills"`
	Inventory     []Item            `json:"inventory"`
	Features      []Feature         `json:"features"`
	Spells        []Spell           `json:"spells,omitempty"`
	Personality   Personality       `json:"personality"`
	Relationships []Relationship    `json:"relationships"`
	Status        string            `json:"status"` // alive, dead, missing, retired
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

// Stats represents ability scores
type Stats struct {
	STR int `json:"str"`
	DEX int `json:"dex"`
	CON int `json:"con"`
	INT int `json:"int"`
	WIS int `json:"wis"`
	CHA int `json:"cha"`
}

// HP represents hit points
type HP struct {
	Current   int `json:"current"`
	Maximum   int `json:"maximum"`
	Temporary int `json:"temporary,omitempty"`
}

// Item represents an inventory item
type Item struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Quantity    int    `json:"quantity"`
	Type        string `json:"type,omitempty"` // weapon, armor, consumable, misc
}

// Feature represents a class/racial/background feature
type Feature struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      string `json:"source,omitempty"` // class, race, background, feat
}

// Spell represents a spell
type Spell struct {
	Name        string `json:"name"`
	Level       int    `json:"level"`
	School      string `json:"school,omitempty"`
	Description string `json:"description,omitempty"`
	Prepared    bool   `json:"prepared,omitempty"`
}

// Personality represents roleplay traits
type Personality struct {
	Traits       []string `json:"traits"`
	Ideals       []string `json:"ideals"`
	Bonds        []string `json:"bonds"`
	Flaws        []string `json:"flaws"`
	Appearance   string   `json:"appearance,omitempty"`
	Backstory    string   `json:"backstory,omitempty"`
}

// Relationship represents a relationship with an NPC or PC
type Relationship struct {
	EntityID   string    `json:"entity_id"`   // ID of NPC or PC
	EntityName string    `json:"entity_name"` // Display name
	EntityType string    `json:"entity_type"` // npc, pc
	Type       string    `json:"type"`        // ally, enemy, neutral, complicated, mentor, student, family
	Strength   int       `json:"strength"`    // -10 to +10
	Notes      string    `json:"notes,omitempty"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Validate checks if the character is valid
func (c *Character) Validate() error {
	if c.CampaignID == "" {
		return NewValidationError("campaign_id", "campaign ID is required")
	}
	if c.Name == "" {
		return NewValidationError("name", "character name is required")
	}
	if c.Race != "" && !Contains(ValidRaces, c.Race) {
		return NewValidationError("race", "invalid race: "+c.Race)
	}
	if c.Class != "" && !Contains(ValidClasses, c.Class) {
		return NewValidationError("class", "invalid class: "+c.Class)
	}
	if c.Level < 1 || c.Level > 20 {
		return NewValidationError("level", "level must be between 1 and 20")
	}
	if c.Background != "" && !Contains(ValidBackgrounds, c.Background) {
		return NewValidationError("background", "invalid background: "+c.Background)
	}
	if c.Alignment != "" && !Contains(ValidAlignments, c.Alignment) {
		return NewValidationError("alignment", "invalid alignment: "+c.Alignment)
	}
	return nil
}

// CalculateModifier calculates an ability modifier from a score
func CalculateModifier(score int) int {
	return (score - 10) / 2
}
