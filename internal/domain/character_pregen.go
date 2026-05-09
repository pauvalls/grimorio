package domain

import (
	"errors"
	"fmt"
)

// PregenCharacter represents a pre-generated character with campaign integration.
type PregenCharacter struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Race            string            `json:"race"`
	Class           string            `json:"class"`
	Subclass        string            `json:"subclass,omitempty"`
	Level           int               `json:"level"`
	Background      string            `json:"background"`
	Alignment       string            `json:"alignment"`
	AbilityScores   AbilityScores     `json:"ability_scores"`
	Skills          []SkillProficiency `json:"skills"`
	Features        []ClassFeature    `json:"features"`
	Equipment       []Equipment       `json:"equipment"`
	Spells          *SpellSelection   `json:"spells,omitempty"`
	Personality     PregenPersonality `json:"personality"`
	Backstory       Backstory         `json:"backstory"`
	CampaignTies    []CampaignTie     `json:"campaign_ties"`
	StatBlockURL    string            `json:"stat_block_url,omitempty"`
}

// AbilityScores represents the six D&D ability scores.
type AbilityScores struct {
	STR int `json:"str"`
	DEX int `json:"dex"`
	CON int `json:"con"`
	INT int `json:"int"`
	WIS int `json:"wis"`
	CHA int `json:"cha"`
}

// SkillProficiency represents a skill proficiency.
type SkillProficiency struct {
	Skill     string `json:"skill"`
	Modifier  int    `json:"modifier"`
	Expertise bool   `json:"expertise,omitempty"`
}

// ClassFeature represents a class feature.
type ClassFeature struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Level       int    `json:"level"`
}

// Equipment represents character equipment.
type Equipment struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Quantity    int    `json:"quantity,omitempty"`
}

// SpellSelection represents spell selection for spellcasters.
type SpellSelection struct {
	Cantrips        []string `json:"cantrips"`
	SpellsKnown     []string `json:"spells_known"`
	SpellSlots      []int    `json:"spell_slots"` // Slots per level
	PreparedSpells  []string `json:"prepared_spells,omitempty"`
}

// PregenPersonality represents character personality traits.
type PregenPersonality struct {
	Traits           []string `json:"traits"`            // At least 2
	Bond             string   `json:"bond"`              // Tied to campaign
	Ideal            string   `json:"ideal"`             // Conflict-capable
	Flaw             string   `json:"flaw"`              // Exploitable
	AdditionalBonds  []string `json:"additional_bonds,omitempty"`
	AdditionalIdeals []string `json:"additional_ideals,omitempty"`
	AdditionalFlaws  []string `json:"additional_flaws,omitempty"`
}

// Backstory represents character backstory.
type Backstory struct {
	Summary      string           `json:"summary"`
	KeyEvents    []BackstoryEvent `json:"key_events"`
	OpenQuestions []string        `json:"open_questions"` // For Session Zero
	Secrets      []BackstorySecret `json:"secrets"`
}

// BackstoryEvent represents a key event in backstory.
type BackstoryEvent struct {
	Description string `json:"description"`
	Age         string `json:"age"` // Childhood, Recent, etc.
	Impact      string `json:"impact"`
}

// BackstorySecret represents a character secret.
type BackstorySecret struct {
	Description    string `json:"description"`
	RevealTrigger  string `json:"reveal_trigger"`
	Consequences   string `json:"consequences"`
}

// CampaignTie represents a connection to campaign elements.
type CampaignTie struct {
	Type         string `json:"type"` // npc, quest, faction, location, item
	ReferenceID  string `json:"reference_id"`
	Relationship string `json:"relationship"`
	Stakes       string `json:"stakes"`
}

// Validate checks PregenCharacter validity.
func (pc *PregenCharacter) Validate() error {
	if pc.ID == "" {
		return errors.New("id is required")
	}
	if pc.Name == "" {
		return errors.New("name is required")
	}
	if pc.Race == "" {
		return errors.New("race is required")
	}
	if pc.Class == "" {
		return errors.New("class is required")
	}
	if pc.Level < 1 || pc.Level > 20 {
		return fmt.Errorf("level must be between 1 and 20, got %d", pc.Level)
	}
	if err := pc.Personality.Validate(); err != nil {
		return fmt.Errorf("personality validation failed: %w", err)
	}
	if err := pc.Backstory.Validate(); err != nil {
		return fmt.Errorf("backstory validation failed: %w", err)
	}
	if err := pc.AbilityScores.Validate(); err != nil {
		return fmt.Errorf("ability scores validation failed: %w", err)
	}
	return nil
}

// Validate checks PregenPersonality validity.
func (pp *PregenPersonality) Validate() error {
	if len(pp.Traits) < 2 {
		return errors.New("must have at least 2 personality traits")
	}
	if pp.Bond == "" {
		return errors.New("bond is required")
	}
	if pp.Ideal == "" {
		return errors.New("ideal is required")
	}
	if pp.Flaw == "" {
		return errors.New("flaw is required")
	}
	return nil
}

// Validate checks Backstory validity.
func (b *Backstory) Validate() error {
	if b.Summary == "" {
		return errors.New("summary is required")
	}
	if len(b.KeyEvents) < 2 {
		return errors.New("must have at least 2 key events")
	}
	if len(b.OpenQuestions) < 1 {
		return errors.New("must have at least 1 open question")
	}
	return nil
}

// Validate checks AbilityScores validity.
func (as *AbilityScores) Validate() error {
	scores := []int{as.STR, as.DEX, as.CON, as.INT, as.WIS, as.CHA}
	for i, score := range scores {
		if score < 3 || score > 18 {
			return fmt.Errorf("ability score %d must be between 3 and 18, got %d", i, score)
		}
	}
	return nil
}

// GetAbilityModifier calculates the modifier for an ability score.
func (as *AbilityScores) GetAbilityModifier(score int) int {
	return (score - 10) / 2
}

// GetTotalAbilityScore returns sum of all ability scores.
func (as *AbilityScores) GetTotalAbilityScore() int {
	return as.STR + as.DEX + as.CON + as.INT + as.WIS + as.CHA
}

// HasCampaignTie checks if character has a tie to a specific campaign element.
func (pc *PregenCharacter) HasCampaignTie(referenceID string) bool {
	for _, tie := range pc.CampaignTies {
		if tie.ReferenceID == referenceID {
			return true
		}
	}
	return false
}
