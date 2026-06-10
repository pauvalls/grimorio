package domain

import (
	"encoding/json"
	"testing"
)

func TestTreasure_StructFields(t *testing.T) {
	tr := Treasure{
		Name:        "Gold coins",
		Description: "50 gp",
		ValueGP:     50,
	}
	if tr.Name != "Gold coins" {
		t.Errorf("expected Name 'Gold coins', got %q", tr.Name)
	}
	if tr.ValueGP != 50 {
		t.Errorf("expected ValueGP 50, got %d", tr.ValueGP)
	}
}

func TestTreasureHoard_StructFields(t *testing.T) {
	hoard := TreasureHoard{
		Tier:       2,
		Coins:      []CoinPurse{{CP: 100, SP: 200, GP: 300, PP: 0}},
		ArtObjects: []ArtObject{{Description: "Gold ring", ValueGP: 25}},
		Gems:       []Gem{{Description: "Ruby", ValueGP: 50}},
		MagicItems: []MagicItemRoll{{Name: "Potion of Healing", Rarity: RarityCommon}},
	}

	if hoard.Tier != 2 {
		t.Errorf("expected Tier 2, got %d", hoard.Tier)
	}
	if len(hoard.Coins) != 1 {
		t.Errorf("expected 1 coin purse, got %d", len(hoard.Coins))
	}
	if len(hoard.ArtObjects) != 1 {
		t.Errorf("expected 1 art object, got %d", len(hoard.ArtObjects))
	}
	if len(hoard.Gems) != 1 {
		t.Errorf("expected 1 gem, got %d", len(hoard.Gems))
	}
	if len(hoard.MagicItems) != 1 {
		t.Errorf("expected 1 magic item, got %d", len(hoard.MagicItems))
	}
}

func TestTreasureHoard_JSONRoundTrip(t *testing.T) {
	original := TreasureHoard{
		Tier:  3,
		Coins: []CoinPurse{{CP: 0, SP: 0, GP: 1000, PP: 100}},
		MagicItems: []MagicItemRoll{
			{Name: "Bag of Holding", Rarity: RarityUncommon},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded TreasureHoard
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.Tier != original.Tier {
		t.Errorf("expected Tier %d, got %d", original.Tier, decoded.Tier)
	}
	if len(decoded.MagicItems) != 1 {
		t.Fatalf("expected 1 magic item, got %d", len(decoded.MagicItems))
	}
	if decoded.MagicItems[0].Name != "Bag of Holding" {
		t.Errorf("expected magic item name 'Bag of Holding', got %q", decoded.MagicItems[0].Name)
	}
}

func TestCoinPurse_TotalGP(t *testing.T) {
	cp := CoinPurse{CP: 1000, SP: 100, GP: 50, PP: 2}
	// 1000cp = 10gp, 100sp = 10gp, 50gp = 50gp, 2pp = 20gp -> total 90gp
	expected := 90
	if got := cp.TotalGP(); got != expected {
		t.Errorf("expected TotalGP %d, got %d", expected, got)
	}
}

func TestCoinPurse_TotalGP_Zero(t *testing.T) {
	cp := CoinPurse{}
	if got := cp.TotalGP(); got != 0 {
		t.Errorf("expected TotalGP 0, got %d", got)
	}
}
