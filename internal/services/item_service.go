package services

import (
	"context"
	"fmt"
	"github.com/pauvalls/grimorio/internal/domain"
	"math/rand"
	"time"
)

// ItemService handles magic item generation and validation.
type ItemService struct {
	itemRepo MagicItemRepository
}

// MagicItemRepository defines the repository interface for magic items.
type MagicItemRepository interface {
	Create(ctx context.Context, campaignID string, item *domain.MagicItem) error
	Read(ctx context.Context, campaignID string, itemID string) (*domain.MagicItem, error)
	Update(ctx context.Context, campaignID string, item *domain.MagicItem) error
	Delete(ctx context.Context, campaignID string, itemID string) error
	GetByRarity(ctx context.Context, campaignID string, rarity domain.MagicItemRarity) ([]*domain.MagicItem, error)
	GetByChapter(ctx context.Context, campaignID string, chapterID string) ([]*domain.MagicItem, error)
}

// NewItemService creates a new ItemService.
func NewItemService(itemRepo MagicItemRepository) *ItemService {
	return &ItemService{itemRepo: itemRepo}
}

// GenerateItem generates a magic item with the specified rarity and type.
func (s *ItemService) GenerateItem(ctx context.Context, campaignID string, rarity domain.MagicItemRarity, itemType domain.MagicItemType) (*domain.MagicItem, error) {
	if !domain.IsValidRarity(rarity) {
		return nil, fmt.Errorf("invalid rarity: %s", rarity)
	}

	rand.Seed(time.Now().UnixNano())
	item := &domain.MagicItem{
		ID:                 fmt.Sprintf("item_%s_%d", campaignID, rand.Intn(10000)),
		Name:               generateItemName(rarity, itemType),
		Type:               itemType,
		Rarity:             rarity,
		AttunementRequired: rand.Intn(2) == 0, // 50% chance
		Properties:         generateProperties(rarity, itemType),
		Lore:               generateLore(rarity),
		Hooks:              generateHooks(),
		ValueGP:            generateValue(rarity),
	}

	if item.AttunementRequired {
		item.AttunementRequirements = generateAttunementRequirements(itemType)
	}

	if err := item.Validate(); err != nil {
		return nil, fmt.Errorf("generated item validation failed: %w", err)
	}

	return item, nil
}

// GenerateCursedItem generates a cursed magic item.
func (s *ItemService) GenerateCursedItem(ctx context.Context, campaignID string, rarity domain.MagicItemRarity) (*domain.MagicItem, error) {
	item, err := s.GenerateItem(ctx, campaignID, rarity, domain.ItemTypeWeapon)
	if err != nil {
		return nil, err
	}

	item.Curse = &domain.ItemCurse{
		Effect:        generateCurseEffect(),
		Trigger:       generateCurseTrigger(),
		RemovalMethod: generateCurseRemoval(),
		Permanent:     rand.Intn(3) == 0, // 33% chance
	}

	if !item.Curse.Permanent {
		dc := 15 + rand.Intn(10)
		item.Curse.DCToRemove = &dc
	}

	if err := item.Validate(); err != nil {
		return nil, fmt.Errorf("generated cursed item validation failed: %w", err)
	}

	return item, nil
}

// GetItemsByRarity retrieves items filtered by rarity.
func (s *ItemService) GetItemsByRarity(ctx context.Context, campaignID string, rarity domain.MagicItemRarity) ([]*domain.MagicItem, error) {
	return s.itemRepo.GetByRarity(ctx, campaignID, rarity)
}

// GetItemsByChapter retrieves items filtered by chapter.
func (s *ItemService) GetItemsByChapter(ctx context.Context, campaignID string, chapterID string) ([]*domain.MagicItem, error) {
	return s.itemRepo.GetByChapter(ctx, campaignID, chapterID)
}

// ValidateItemPower checks if an item's power is appropriate for party level.
func (s *ItemService) ValidateItemPower(ctx context.Context, item *domain.MagicItem, partyLevel int) (bool, error) {
	minLevel := domain.GetMinLevelForRarity(item.Rarity)
	if partyLevel < minLevel-2 {
		return false, nil // Too powerful for party
	}
	maxBonus := domain.GetMaxBonusForRarity(item.Rarity)
	if item.Bonus() > maxBonus {
		return false, fmt.Errorf("item bonus %d exceeds max %d for rarity %s", item.Bonus(), maxBonus, item.Rarity)
	}
	return true, nil
}

// Helper functions

func generateProperties(rarity domain.MagicItemRarity, itemType domain.MagicItemType) []domain.MagicItemProperty {
	props := []domain.MagicItemProperty{}

	// Add bonus property for weapons/armor
	if itemType == domain.ItemTypeWeapon || itemType == domain.ItemTypeArmor {
		bonus := 0
		if rarity != domain.RarityCommon {
			bonus = 1
			if rarity == domain.RarityRare || rarity == domain.RarityVeryRare {
				bonus = 2
			} else if rarity == domain.RarityLegendary || rarity == domain.RarityArtifact {
				bonus = 3
			}
		}
		if bonus > 0 {
			props = append(props, domain.MagicItemProperty{
				Name:   "Enhancement Bonus",
				Description: fmt.Sprintf("Grants a +%d bonus to attack and damage rolls", bonus),
				Bonus:  bonus,
			})
		}
	}

	// Add special properties based on rarity
	if rarity == domain.RarityUncommon || rarity == domain.RarityRare {
		props = append(props, domain.MagicItemProperty{
			Name:        "Special Ability",
			Description: "Has one special ability usable once per day",
			RequiresAction: true,
			Charges:     &domain.ItemCharges{Max: 1, Daily: 1, Recharge: "dawn"},
		})
	}

	if rarity == domain.RarityVeryRare || rarity == domain.RarityLegendary {
		props = append(props, domain.MagicItemProperty{
			Name:        "Major Power",
			Description: "Has a major power usable 3 times per day",
			RequiresAction: true,
			Charges:     &domain.ItemCharges{Max: 3, Daily: 3, Recharge: "dawn"},
		})
	}

	return props
}

func generateLore(rarity domain.MagicItemRarity) string {
	lores := map[domain.MagicItemRarity][]string{
		domain.RarityCommon: {
			"A simple item with minor enchantments, crafted by a local artisan.",
			"This item bears faint magical auras, the work of an apprentice wizard.",
		},
		domain.RarityUncommon: {
			"Forged during a time of great conflict, this item has seen many battles.",
			"A gift from a grateful noble, this item carries their family crest.",
		},
		domain.RarityRare: {
			"Created by a renowned master craftsman, this item is one of few made.",
			"Recovered from an ancient tomb, this item bears the marks of a lost civilization.",
		},
		domain.RarityVeryRare: {
			"Legend speaks of this item's role in a great hero's victory.",
			"This item was blessed by a deity and carries divine power.",
		},
		domain.RarityLegendary: {
			"An artifact of immense power, this item shaped the course of history.",
			"Only one such item exists, created at the dawn of the age.",
		},
		domain.RarityArtifact: {
			"A relic of the gods, this item's power is beyond mortal comprehension.",
			"This item is sentient and has its own agenda.",
		},
	}

	items := lores[rarity]
	if len(items) == 0 {
		return "A mysterious item of unknown origin."
	}
	return items[rand.Intn(len(items))]
}

func generateHooks() []string {
	return []string{
		"The party discovers this item in a treasure hoard.",
		"A dying NPC entrusts this item to the party.",
		"This item is the reward for completing a dangerous quest.",
		"A mysterious merchant offers this item for sale.",
	}
}

func generateAttunementRequirements(itemType domain.MagicItemType) string {
	requirements := []string{
		"a warrior",
		"a spellcaster",
		"a creature of good alignment",
		"a creature proficient in Arcana",
	}
	return requirements[rand.Intn(len(requirements))]
}

func generateValue(rarity domain.MagicItemRarity) int {
	min, max := domain.GetApproximateValueGP(rarity)
	if min == 0 && max == 0 {
		return 0 // Artifact
	}
	return min + rand.Intn(max-min)
}

func generateItemName(rarity domain.MagicItemRarity, itemType domain.MagicItemType) string {
	prefixes := map[domain.MagicItemRarity][]string{
		domain.RarityCommon:    {"Simple", "Ordinary", "Basic"},
		domain.RarityUncommon:  {"Enhanced", "Fine", "Superior"},
		domain.RarityRare:      {"Magnificent", "Glorious", "Exalted"},
		domain.RarityVeryRare:  {"Legendary", "Mythic", "Ancient"},
		domain.RarityLegendary: {"Divine", "Celestial", "Eternal"},
		domain.RarityArtifact:  {"Cosmic", "Primordial", "Transcendent"},
	}

	suffixes := map[domain.MagicItemType]string{
		domain.ItemTypeWeapon:   "Blade",
		domain.ItemTypeArmor:    "Armor",
		domain.ItemTypeWondrous: "Amulet",
		domain.ItemTypeRing:     "Ring",
		domain.ItemTypeRod:      "Rod",
		domain.ItemTypeStaff:    "Staff",
		domain.ItemTypeWand:     "Wand",
		domain.ItemTypePotion:   "Elixir",
		domain.ItemTypeScroll:   "Scroll",
		domain.ItemTypeShield:   "Shield",
	}

	prefixList := prefixes[rarity]
	if len(prefixList) == 0 {
		prefixList = []string{"Mysterious"}
	}
	prefix := prefixList[rand.Intn(len(prefixList))]

	suffix := suffixes[itemType]
	if suffix == "" {
		suffix = "Item"
	}

	return fmt.Sprintf("%s %s", prefix, suffix)
}

func generateCurseEffect() string {
	effects := []string{
		"The wielder cannot voluntarily release the item.",
		"The wielder suffers disadvantage on saving throws against fear.",
		"The item drains 1 level from the wielder each dawn.",
		"The wielder is compelled to attack any creature they see.",
	}
	return effects[rand.Intn(len(effects))]
}

func generateCurseTrigger() string {
	triggers := []string{
		"Upon attuning to the item",
		"Upon drawing blood with the item",
		"Upon speaking a command word",
		"Upon taking damage while wielding the item",
	}
	return triggers[rand.Intn(len(triggers))]
}

func generateCurseRemoval() string {
	removals := []string{
		"Cast remove curse on the wielder",
		"Submerge the item in holy water for 24 hours",
		"Defeat the curse's creator in combat",
		"Complete a quest to atone for past sins",
	}
	return removals[rand.Intn(len(removals))]
}
