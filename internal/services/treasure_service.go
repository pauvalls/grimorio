package services

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
)

// TreasureService generates SRD-compliant treasure.
type TreasureService struct {
	rng *rand.Rand
}

// NewTreasureService creates a new treasure service with a seeded RNG.
func NewTreasureService() *TreasureService {
	return &TreasureService{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// NewTreasureServiceWithSeed creates a treasure service with a fixed seed for tests.
func NewTreasureServiceWithSeed(seed int64) *TreasureService {
	return &TreasureService{
		rng: rand.New(rand.NewSource(seed)),
	}
}

// GenerateIndividualTreasure rolls individual treasure by CR tier.
func (s *TreasureService) GenerateIndividualTreasure(ctx context.Context, cr int) ([]domain.Treasure, error) {
	tier := crToTier(cr)
	var result []domain.Treasure

	switch tier {
	case 1:
		cp := s.roll(6, 6) * 10
		sp := s.roll(2, 6) * 5
		gp := s.roll(2, 6)
		result = appendCoins(result, cp, sp, gp, 0)
	case 2:
		cp := s.roll(2, 6) * 100
		sp := s.roll(2, 6) * 10
		gp := s.roll(6, 6)
		gp2 := s.roll(3, 6) * 10
		result = appendCoins(result, cp, sp, gp+gp2, 0)
	case 3:
		cp := s.roll(4, 6) * 100
		sp := s.roll(6, 6) * 10
		gp := s.roll(2, 6) * 10
		gp2 := s.roll(4, 6) * 10
		pp := s.roll(2, 6) * 10
		result = appendCoins(result, cp, sp, gp+gp2, pp)
	case 4:
		cp := s.roll(12, 6) * 100
		sp := s.roll(2, 6) * 10
		gp := s.roll(8, 6) * 10
		gp2 := s.roll(6, 6) * 10
		pp := s.roll(6, 6) * 10
		pp2 := s.roll(8, 6) * 10
		result = appendCoins(result, cp, sp, gp+gp2, pp+pp2)
	default:
		return nil, fmt.Errorf("invalid CR tier derived from CR %d", cr)
	}

	return result, nil
}

// GenerateTreasureHoard rolls a treasure hoard by tier.
func (s *TreasureService) GenerateTreasureHoard(ctx context.Context, tier int) (*domain.TreasureHoard, error) {
	if tier < 1 || tier > 4 {
		return nil, fmt.Errorf("invalid treasure tier: %d (must be 1-4)", tier)
	}

	hoard := &domain.TreasureHoard{Tier: tier}

	// Roll coins
	hoard.Coins = append(hoard.Coins, s.rollHoardCoins(tier))

	// Roll art objects and gems
	art, gems := s.rollArtAndGems(tier)
	hoard.ArtObjects = art
	hoard.Gems = gems

	// Roll magic items
	hoard.MagicItems = s.rollMagicItems(tier)

	return hoard, nil
}

func (s *TreasureService) roll(count, sides int) int {
	total := 0
	for i := 0; i < count; i++ {
		total += s.rng.Intn(sides) + 1
	}
	return total
}

func crToTier(cr int) int {
	switch {
	case cr >= 0 && cr <= 4:
		return 1
	case cr >= 5 && cr <= 10:
		return 2
	case cr >= 11 && cr <= 16:
		return 3
	default:
		return 4
	}
}

func appendCoins(result []domain.Treasure, cp, sp, gp, pp int) []domain.Treasure {
	if cp > 0 {
		result = append(result, domain.Treasure{Name: fmt.Sprintf("%d cp", cp), Description: fmt.Sprintf("%d copper pieces", cp), ValueGP: cp / 100})
	}
	if sp > 0 {
		result = append(result, domain.Treasure{Name: fmt.Sprintf("%d sp", sp), Description: fmt.Sprintf("%d silver pieces", sp), ValueGP: sp / 10})
	}
	if gp > 0 {
		result = append(result, domain.Treasure{Name: fmt.Sprintf("%d gp", gp), Description: fmt.Sprintf("%d gold pieces", gp), ValueGP: gp})
	}
	if pp > 0 {
		result = append(result, domain.Treasure{Name: fmt.Sprintf("%d pp", pp), Description: fmt.Sprintf("%d platinum pieces", pp), ValueGP: pp * 10})
	}
	return result
}

func (s *TreasureService) rollHoardCoins(tier int) domain.CoinPurse {
	switch tier {
	case 1:
		return domain.CoinPurse{CP: s.roll(6, 6) * 100, SP: s.roll(3, 6) * 100, GP: s.roll(2, 6) * 10}
	case 2:
		return domain.CoinPurse{CP: s.roll(2, 6) * 100, SP: s.roll(2, 6) * 1000, GP: s.roll(6, 6) * 100, PP: s.roll(3, 6) * 10}
	case 3:
		return domain.CoinPurse{GP: s.roll(4, 6) * 1000, PP: s.roll(5, 6) * 100}
	case 4:
		return domain.CoinPurse{GP: s.roll(12, 6) * 1000, PP: s.roll(8, 6) * 1000}
	}
	return domain.CoinPurse{}
}

func (s *TreasureService) rollArtAndGems(tier int) ([]domain.ArtObject, []domain.Gem) {
	var art []domain.ArtObject
	var gems []domain.Gem

	switch tier {
	case 1:
		if s.rng.Intn(100)+1 <= 50 {
			art = s.rollArtObjects(1, 10)
		} else {
			gems = s.rollGems(1, 10)
		}
	case 2:
		roll := s.rng.Intn(100) + 1
		if roll <= 30 {
			art = s.rollArtObjects(1, 25)
		} else if roll <= 60 {
			gems = s.rollGems(1, 50)
		} else if roll <= 80 {
			art = s.rollArtObjects(1, 250)
		} else {
			gems = s.rollGems(1, 100)
		}
	case 3:
		roll := s.rng.Intn(100) + 1
		if roll <= 20 {
			art = s.rollArtObjects(1, 250)
		} else if roll <= 50 {
			gems = s.rollGems(1, 100)
		} else if roll <= 70 {
			art = s.rollArtObjects(1, 750)
		} else {
			gems = s.rollGems(1, 500)
		}
	case 4:
		roll := s.rng.Intn(100) + 1
		if roll <= 25 {
			art = s.rollArtObjects(1, 2500)
		} else if roll <= 50 {
			gems = s.rollGems(1, 1000)
		} else if roll <= 75 {
			art = s.rollArtObjects(1, 7500)
		} else {
			gems = s.rollGems(1, 5000)
		}
	}

	return art, gems
}

func (s *TreasureService) rollArtObjects(count, value int) []domain.ArtObject {
	var result []domain.ArtObject
	descriptions := []string{
		"Silver ewer", "Carved bone statuette", "Small gold bracelet",
		"Cloth-of-gold vestments", "Black velvet mask", "Copper chalice",
		"Engraved copper ring", "Embroidered silk handkerchief",
		"Gold locket with a painted portrait", "Carved ivory statuette",
	}
	for i := 0; i < count; i++ {
		desc := descriptions[s.rng.Intn(len(descriptions))]
		result = append(result, domain.ArtObject{Description: desc, ValueGP: value})
	}
	return result
}

func (s *TreasureService) rollGems(count, value int) []domain.Gem {
	var result []domain.Gem
	descriptions := []string{
		"Azurite", "Banded agate", "Blue quartz", "Eye agate", "Hematite",
		"Malachite", "Moss agate", "Obsidian", "Rhodochrosite", "Tiger eye",
		"Bloodstone", "Carnelian", "Chalcedony", "Chrysoprase", "Citrine",
		"Jasper", "Moonstone", "Onyx", "Quartz", "Sardonyx",
		"Alexandrite", "Aquamarine", "Black pearl", "Blue spinel", "Peridot",
		"Topaz", "Emerald", "Ruby", "Sapphire", "Diamond",
	}
	for i := 0; i < count; i++ {
		desc := descriptions[s.rng.Intn(len(descriptions))]
		result = append(result, domain.Gem{Description: desc, ValueGP: value})
	}
	return result
}

func (s *TreasureService) rollMagicItems(tier int) []domain.MagicItemRoll {
	var result []domain.MagicItemRoll

	switch tier {
	case 1:
		count := s.roll(1, 6)
		result = append(result, s.generateMagicItemsByRarity(domain.RarityCommon, count)...)
	case 2:
		count := s.roll(1, 4)
		result = append(result, s.generateMagicItemsByRarity(domain.RarityUncommon, count)...)
	case 3:
		count := s.roll(1, 4)
		result = append(result, s.generateMagicItemsByRarity(domain.RarityRare, count)...)
	case 4:
		count := s.roll(1, 4)
		result = append(result, s.generateMagicItemsByRarity(domain.RarityVeryRare, count)...)
		// Chance for legendary
		if s.rng.Intn(100)+1 <= 25 {
			result = append(result, s.generateMagicItemsByRarity(domain.RarityLegendary, 1)...)
		}
	}

	return result
}

// generateMagicItemsByRarity generates count magic items of the given rarity.
func (s *TreasureService) generateMagicItemsByRarity(rarity domain.MagicItemRarity, count int) []domain.MagicItemRoll {
	var result []domain.MagicItemRoll
	items := magicItemTableByRarity(rarity)
	if len(items) == 0 {
		return result
	}
	for i := 0; i < count; i++ {
		idx := s.rng.Intn(len(items))
		result = append(result, items[idx])
	}
	return result
}

func magicItemTableByRarity(rarity domain.MagicItemRarity) []domain.MagicItemRoll {
	switch rarity {
	case domain.RarityCommon:
		return []domain.MagicItemRoll{
			{Name: "Potion of Healing", Rarity: domain.RarityCommon},
			{Name: "Potion of Climbing", Rarity: domain.RarityCommon},
			{Name: "Potion of Water Breathing", Rarity: domain.RarityCommon},
			{Name: "Spell Scroll (Cantrip)", Rarity: domain.RarityCommon},
			{Name: "Spell Scroll (1st Level)", Rarity: domain.RarityCommon},
			{Name: "Potion of Resistance", Rarity: domain.RarityCommon},
		}
	case domain.RarityUncommon:
		return []domain.MagicItemRoll{
			{Name: "Bag of Holding", Rarity: domain.RarityUncommon},
			{Name: "Cloak of Protection", Rarity: domain.RarityUncommon},
			{Name: "Gauntlets of Ogre Power", Rarity: domain.RarityUncommon},
			{Name: "Potion of Greater Healing", Rarity: domain.RarityUncommon},
			{Name: "Ring of Swimming", Rarity: domain.RarityUncommon},
			{Name: "Spell Scroll (2nd Level)", Rarity: domain.RarityUncommon},
			{Name: "Spell Scroll (3rd Level)", Rarity: domain.RarityUncommon},
			{Name: "Wand of Magic Missiles", Rarity: domain.RarityUncommon},
		}
	case domain.RarityRare:
		return []domain.MagicItemRoll{
			{Name: "Amulet of Health", Rarity: domain.RarityRare},
			{Name: "Boots of Speed", Rarity: domain.RarityRare},
			{Name: "Bracers of Defense", Rarity: domain.RarityRare},
			{Name: "Flame Tongue", Rarity: domain.RarityRare},
			{Name: "Potion of Superior Healing", Rarity: domain.RarityRare},
			{Name: "Spell Scroll (4th Level)", Rarity: domain.RarityRare},
			{Name: "Spell Scroll (5th Level)", Rarity: domain.RarityRare},
			{Name: "Wand of Lightning Bolts", Rarity: domain.RarityRare},
		}
	case domain.RarityVeryRare:
		return []domain.MagicItemRoll{
			{Name: "Animated Shield", Rarity: domain.RarityVeryRare},
			{Name: "Belt of Giant Strength (Fire)", Rarity: domain.RarityVeryRare},
			{Name: "Cloak of Displacement", Rarity: domain.RarityVeryRare},
			{Name: "Potion of Supreme Healing", Rarity: domain.RarityVeryRare},
			{Name: "Ring of Regeneration", Rarity: domain.RarityVeryRare},
			{Name: "Spell Scroll (6th Level)", Rarity: domain.RarityVeryRare},
			{Name: "Spell Scroll (7th Level)", Rarity: domain.RarityVeryRare},
			{Name: "Wand of Polymorph", Rarity: domain.RarityVeryRare},
		}
	case domain.RarityLegendary:
		return []domain.MagicItemRoll{
			{Name: "Apparatus of Kwalish", Rarity: domain.RarityLegendary},
			{Name: "Armor of Invulnerability", Rarity: domain.RarityLegendary},
			{Name: "Belt of Giant Strength (Storm)", Rarity: domain.RarityLegendary},
			{Name: "Cloak of Invisibility", Rarity: domain.RarityLegendary},
			{Name: "Crystal Ball", Rarity: domain.RarityLegendary},
			{Name: "Ring of Djinni Summoning", Rarity: domain.RarityLegendary},
			{Name: "Ring of Invisibility", Rarity: domain.RarityLegendary},
			{Name: "Staff of the Magi", Rarity: domain.RarityLegendary},
			{Name: "Sword of Sharpness", Rarity: domain.RarityLegendary},
			{Name: "Vorpal Sword", Rarity: domain.RarityLegendary},
		}
	}
	return nil
}
