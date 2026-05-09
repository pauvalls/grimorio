package domain

import (
	"errors"
	"fmt"
)

// MagicItemRarity represents item rarity levels following DMG standards.
type MagicItemRarity string

const (
	RarityCommon    MagicItemRarity = "common"
	RarityUncommon  MagicItemRarity = "uncommon"
	RarityRare      MagicItemRarity = "rare"
	RarityVeryRare  MagicItemRarity = "very_rare"
	RarityLegendary MagicItemRarity = "legendary"
	RarityArtifact  MagicItemRarity = "artifact"
)

// MagicItemType represents the type of magic item.
type MagicItemType string

const (
	ItemTypeWeapon    MagicItemType = "weapon"
	ItemTypeArmor     MagicItemType = "armor"
	ItemTypeWondrous  MagicItemType = "wondrous"
	ItemTypeRing      MagicItemType = "ring"
	ItemTypeRod       MagicItemType = "rod"
	ItemTypeStaff     MagicItemType = "staff"
	ItemTypeWand      MagicItemType = "wand"
	ItemTypePotion    MagicItemType = "potion"
	ItemTypeScroll    MagicItemType = "scroll"
	ItemTypeShield    MagicItemType = "shield"
)

// MagicItem represents a magic item with full WotC stat block format.
type MagicItem struct {
	ID                   string            `json:"id"`
	Name                 string            `json:"name"`
	Type                 MagicItemType     `json:"type"`
	Subtype              string            `json:"subtype,omitempty"` // longsword, plate, etc.
	Rarity               MagicItemRarity   `json:"rarity"`
	AttunementRequired   bool              `json:"attunement_required"`
	AttunementRequirements string          `json:"attunement_requirements,omitempty"`
	Properties           []MagicItemProperty `json:"properties"`
	Lore                 string            `json:"lore"`
	Hooks                []string          `json:"hooks"`
	Curse                *ItemCurse        `json:"curse,omitempty"`
	SourceChapter        string            `json:"source_chapter,omitempty"`
	ValueGP              int               `json:"value_gp,omitempty"`
}

// MagicItemProperty represents a single property or ability of a magic item.
type MagicItemProperty struct {
	Name           string       `json:"name"`
	Description    string       `json:"description"`
	RequiresAction bool         `json:"requires_action"` // e.g., command word, charge expenditure
	Charges        *ItemCharges `json:"charges,omitempty"`
	Damage         *DamageDice  `json:"damage,omitempty"`
	Bonus          int          `json:"bonus,omitempty"` // +1, +2, +3
}

// ItemCharges represents charge-based abilities.
type ItemCharges struct {
	Max      int    `json:"max"`
	Daily    int    `json:"daily"`
	Recharge string `json:"recharge,omitempty"` // "dawn", "d8+4", etc.
}

// DamageDice represents damage dice for weapon properties.
type DamageDice struct {
	Dice      string `json:"dice"` // e.g., "1d6", "2d8"
	DamageType string `json:"damage_type"` // e.g., "fire", "force"
}

// ItemCurse represents a cursed item effect.
type ItemCurse struct {
	Effect        string `json:"effect"`
	Trigger       string `json:"trigger"`
	RemovalMethod string `json:"removal_method"`
	DCToRemove    *int   `json:"dc_to_remove,omitempty"`
	Permanent     bool   `json:"permanent"`
}

// Validate checks magic item validity according to DMG standards.
func (m *MagicItem) Validate() error {
	if m.ID == "" {
		return errors.New("id is required")
	}
	if m.Name == "" {
		return errors.New("name is required")
	}
	if m.Type == "" {
		return errors.New("type is required")
	}
	if m.Rarity == "" {
		return errors.New("rarity is required")
	}
	if !isValidRarity(m.Rarity) {
		return fmt.Errorf("invalid rarity: %s", m.Rarity)
	}
	if m.Curse != nil && m.Curse.RemovalMethod == "" {
		return errors.New("cursed items must have removal method")
	}
	if m.AttunementRequired && m.AttunementRequirements == "" {
		return errors.New("attunement_required is true but attunement_requirements is empty")
	}
	// Validate bonus matches rarity
	maxBonus := GetMaxBonusForRarity(m.Rarity)
	for _, prop := range m.Properties {
		if prop.Bonus > maxBonus {
			return fmt.Errorf("bonus %d exceeds maximum %d for rarity %s", prop.Bonus, maxBonus, m.Rarity)
		}
	}
	return nil
}

// Bonus returns the total bonus of the item (highest property bonus).
func (m *MagicItem) Bonus() int {
	bonus := 0
	for _, p := range m.Properties {
		if p.Bonus > bonus {
			bonus = p.Bonus
		}
	}
	return bonus
}

// IsValidRarity checks if a rarity string is valid.
func isValidRarity(rarity MagicItemRarity) bool {
	switch rarity {
	case RarityCommon, RarityUncommon, RarityRare, RarityVeryRare, RarityLegendary, RarityArtifact:
		return true
	default:
		return false
	}
}

// GetMaxBonusForRarity returns max bonus for rarity per DMG guidelines.
func GetMaxBonusForRarity(rarity MagicItemRarity) int {
	switch rarity {
	case RarityCommon:
		return 0
	case RarityUncommon:
		return 1
	case RarityRare:
		return 2
	default: // Very Rare, Legendary, Artifact
		return 3
	}
}

// GetMinLevelForRarity returns recommended minimum character level for item rarity.
func GetMinLevelForRarity(rarity MagicItemRarity) int {
	switch rarity {
	case RarityCommon:
		return 1
	case RarityUncommon:
		return 3
	case RarityRare:
		return 5
	case RarityVeryRare:
		return 11
	case RarityLegendary:
		return 17
	case RarityArtifact:
		return 17
	default:
		return 1
	}
}

// GetApproximateValueGP returns approximate gold piece value for rarity.
func GetApproximateValueGP(rarity MagicItemRarity) (min, max int) {
	switch rarity {
	case RarityCommon:
		return 50, 100
	case RarityUncommon:
		return 101, 500
	case RarityRare:
		return 501, 5000
	case RarityVeryRare:
		return 5001, 50000
	case RarityLegendary:
		return 50001, 100000
	case RarityArtifact:
		return 0, 0 // Priceless
	default:
		return 0, 0
	}
}
