package services

import (
	"context"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
)

func TestTreasureService_GenerateIndividualTreasure_CR3(t *testing.T) {
	svc := NewTreasureService()
	result, err := svc.GenerateIndividualTreasure(context.Background(), 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// CR 3 (tier 1: 0-4) should return coins only
	if len(result) == 0 {
		t.Fatal("expected at least one treasure item for CR 3")
	}

	var hasCoins bool
	for _, item := range result {
		if item.Name != "" && item.ValueGP > 0 {
			hasCoins = true
		}
	}
	if !hasCoins {
		t.Error("expected coins in individual treasure for CR 3")
	}
}

func TestTreasureService_GenerateIndividualTreasure_ZeroCR(t *testing.T) {
	svc := NewTreasureService()
	result, err := svc.GenerateIndividualTreasure(context.Background(), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected treasure even for CR 0")
	}
}

func TestTreasureService_GenerateIndividualTreasure_HighCR(t *testing.T) {
	svc := NewTreasureService()
	result, err := svc.GenerateIndividualTreasure(context.Background(), 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected treasure for high CR")
	}

	// High CR should have more valuable treasure
	var totalValue int
	for _, item := range result {
		totalValue += item.ValueGP
	}
	if totalValue < 100 {
		t.Errorf("expected high CR treasure to be valuable, got total %d gp", totalValue)
	}
}

func TestTreasureService_GenerateTreasureHoard_Tier2(t *testing.T) {
	svc := NewTreasureService()
	hoard, err := svc.GenerateTreasureHoard(context.Background(), 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hoard.Tier != 2 {
		t.Errorf("expected Tier 2, got %d", hoard.Tier)
	}

	// Tier 2 hoard should have coins
	if len(hoard.Coins) == 0 {
		t.Error("expected coins in tier 2 hoard")
	}

	// Should have art objects or gems
	if len(hoard.ArtObjects) == 0 && len(hoard.Gems) == 0 {
		t.Error("expected art objects or gems in tier 2 hoard")
	}

	// Should have magic items
	if len(hoard.MagicItems) == 0 {
		t.Error("expected magic items in tier 2 hoard")
	}
}

func TestTreasureService_GenerateTreasureHoard_Tier4(t *testing.T) {
	svc := NewTreasureService()
	hoard, err := svc.GenerateTreasureHoard(context.Background(), 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hoard.Tier != 4 {
		t.Errorf("expected Tier 4, got %d", hoard.Tier)
	}

	// Tier 4 may include legendary items
	var hasLegendary bool
	for _, item := range hoard.MagicItems {
		if item.Rarity == domain.RarityLegendary {
			hasLegendary = true
			break
		}
	}
	if !hasLegendary {
		// Not guaranteed every roll, but verify the structure is valid
		t.Log("no legendary item this roll (probabilistic)")
	}
}

func TestTreasureService_GenerateTreasureHoard_InvalidTier(t *testing.T) {
	svc := NewTreasureService()
	_, err := svc.GenerateTreasureHoard(context.Background(), 0)
	if err == nil {
		t.Error("expected error for invalid tier 0")
	}
	_, err = svc.GenerateTreasureHoard(context.Background(), 5)
	if err == nil {
		t.Error("expected error for invalid tier 5")
	}
}

func TestTreasureService_MagicItemsByRarity(t *testing.T) {
	svc := NewTreasureService()
	rarities := []domain.MagicItemRarity{
		domain.RarityCommon,
		domain.RarityUncommon,
		domain.RarityRare,
		domain.RarityVeryRare,
		domain.RarityLegendary,
	}

	for _, rarity := range rarities {
		items := svc.generateMagicItemsByRarity(rarity, 3)
		if len(items) != 3 {
			t.Errorf("expected 3 %s items, got %d", rarity, len(items))
		}
		for _, item := range items {
			if item.Rarity != rarity {
				t.Errorf("expected rarity %s, got %s", rarity, item.Rarity)
			}
			if item.Name == "" {
				t.Error("expected non-empty magic item name")
			}
		}
	}
}
