package domain

import (
	"errors"
	"fmt"
)

// Area represents a unified WotC-format area with sequential numbering.
type Area struct {
	ID              string         `json:"id"`
	ChapterID       string         `json:"chapter_id"`
	AreaNumber      int            `json:"area_number"` // 1-15 per chapter
	Title           string         `json:"title"`
	Summary         string         `json:"summary"`
	Description     string         `json:"description"`
	LevelRange      LevelRange     `json:"level_range"`
	Features        []AreaFeature  `json:"features"`
	Encounters      []AreaEncounter `json:"encounters"`
	NPCs            []AreaNPC      `json:"npcs"`
	Treasure        []Treasure     `json:"treasure"`
	Development     string         `json:"development"` // What happens after party leaves
	DMSidebars      []DMSidebar    `json:"dm_sidebars"`
	PlayerReadAloud string         `json:"player_read_aloud,omitempty"`
	Maps            []MapReference `json:"maps"`
}

// AreaFeature represents a feature within an area.
type AreaFeature struct {
	Type        string `json:"type"` // room, passage, hazard, trap, clue, treasure
	Name        string `json:"name"`
	Description string `json:"description"`
	DC          *int   `json:"dc,omitempty"`
	Hidden      bool   `json:"hidden"`
}

// AreaEncounter represents an encounter within an area.
type AreaEncounter struct {
	EncounterID string `json:"encounter_id"`
	Trigger     string `json:"trigger"`
	CRTotal     float64 `json:"cr_total"`
	XPValue     int    `json:"xp_value"`
	TacticsRef  string `json:"tactics_ref,omitempty"`
}

// AreaNPC represents an NPC found in an area.
type AreaNPC struct {
	NPCID           string `json:"npc_id"`
	Role            string `json:"role"`
	InteractionNotes string `json:"interaction_notes"`
}

// Treasure represents treasure in an area.
type Treasure struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ValueGP     int    `json:"value_gp,omitempty"`
	MagicItemID string `json:"magic_item_id,omitempty"`
	Hidden      bool   `json:"hidden,omitempty"`
}

// DMSidebar represents a DM advice sidebar.
type DMSidebar struct {
	Title     string `json:"title"`
	Content   string `json:"content"`
	Type      string `json:"type"` // running, combat, roleplay, puzzle
	Placement string `json:"placement"` // After area description, Before encounter
}

// MapReference represents a map reference.
type MapReference struct {
	Filename  string `json:"filename"`
	Alt       string `json:"alt"`
	IsDMOnly  bool   `json:"is_dm_only"`
}

// Validate checks Area validity according to WotC format.
func (a *Area) Validate() error {
	if a.ID == "" {
		return errors.New("id is required")
	}
	if a.ChapterID == "" {
		return errors.New("chapter_id is required")
	}
	if a.AreaNumber < 1 || a.AreaNumber > 15 {
		return fmt.Errorf("area_number must be between 1 and 15, got %d", a.AreaNumber)
	}
	if a.Title == "" {
		return errors.New("title is required")
	}
	if err := ValidateLevelRange(a.LevelRange); err != nil {
		return fmt.Errorf("invalid level range: %w", err)
	}
	if len(a.Features) == 0 {
		return errors.New("must have at least 1 feature")
	}
	if len(a.Encounters) == 0 {
		return errors.New("must have at least 1 encounter")
	}
	return nil
}

// ValidateAreaNumber checks if area number is valid (1-15).
func ValidateAreaNumber(num int) error {
	if num < 1 || num > 15 {
		return fmt.Errorf("area number must be between 1 and 15, got %d", num)
	}
	return nil
}

// ValidateSequentialNumbers checks if area numbers are sequential with no gaps.
func ValidateSequentialNumbers(numbers []int) error {
	if len(numbers) == 0 {
		return errors.New("area numbers cannot be empty")
	}
	for i := 1; i < len(numbers); i++ {
		if numbers[i] != numbers[i-1]+1 {
			return fmt.Errorf("gap in area numbers: %d followed by %d", numbers[i-1], numbers[i])
		}
	}
	return nil
}

// GetTotalXP calculates total XP from all encounters.
func (a *Area) GetTotalXP() int {
	total := 0
	for _, enc := range a.Encounters {
		total += enc.XPValue
	}
	return total
}

// GetTotalCR calculates total CR from all encounters.
func (a *Area) GetTotalCR() float64 {
	total := 0.0
	for _, enc := range a.Encounters {
		total += enc.CRTotal
	}
	return total
}

// HasHiddenFeatures checks if area has hidden features.
func (a *Area) HasHiddenFeatures() bool {
	for _, f := range a.Features {
		if f.Hidden {
			return true
		}
	}
	return false
}

// GetHiddenFeatureCount returns count of hidden features.
func (a *Area) GetHiddenFeatureCount() int {
	count := 0
	for _, f := range a.Features {
		if f.Hidden {
			count++
		}
	}
	return count
}
