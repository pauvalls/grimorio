package namegen

import "fmt"

// Category represents the type of entity to generate names for.
type Category string

// Style represents the cultural flavor of generated names.
type Style string

// Supported name categories.
const (
	CategoryCharacter    Category = "character"
	CategoryNPC          Category = "npc"
	CategoryCity         Category = "city"
	CategoryTavern       Category = "tavern"
	CategoryMonster      Category = "monster"
	CategoryFaction      Category = "faction"
	CategoryItem         Category = "item"
)

// Supported cultural styles.
const (
	StyleGenericFantasy Style = "generic_fantasy"
	StyleElven          Style = "elven"
	StyleDwarven        Style = "dwarven"
	StyleOrcish         Style = "orcish"
	StyleHumanMedieval  Style = "human_medieval"
)

var (
	validCategories = map[Category]struct{}{
		CategoryCharacter: {},
		CategoryNPC:       {},
		CategoryCity:      {},
		CategoryTavern:    {},
		CategoryMonster:   {},
		CategoryFaction:   {},
		CategoryItem:      {},
	}

	validStyles = map[Style]struct{}{
		StyleGenericFantasy: {},
		StyleElven:          {},
		StyleDwarven:        {},
		StyleOrcish:         {},
		StyleHumanMedieval:  {},
	}
)

// Validate checks whether the given category and style are supported.
func Validate(cat Category, style Style) error {
	if _, ok := validCategories[cat]; !ok {
		return fmt.Errorf("invalid category: %q", cat)
	}
	if _, ok := validStyles[style]; !ok {
		return fmt.Errorf("invalid style: %q", style)
	}
	return nil
}
