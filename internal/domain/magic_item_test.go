package domain

import (
	"strings"
	"testing"
)

func TestMagicItem_Validate(t *testing.T) {
	tests := []struct {
		name    string
		item    *MagicItem
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid item",
			item: &MagicItem{
				ID:       "item_1",
				Name:     "Flame Tongue",
				Type:     ItemTypeWeapon,
				Subtype:  "longsword",
				Rarity:   RarityRare,
				Properties: []MagicItemProperty{
					{Name: "Flame Strike", Description: "Deal +2d6 fire damage"},
				},
				Lore: "A sword forged in dragon fire",
				Hooks: []string{"Seek the dragon's lair"},
			},
			wantErr: false,
		},
		{
			name: "missing id",
			item: &MagicItem{
				Name:   "Flame Tongue",
				Type:   ItemTypeWeapon,
				Rarity: RarityRare,
			},
			wantErr: true,
			errMsg:  "id is required",
		},
		{
			name: "missing name",
			item: &MagicItem{
				ID:     "item_1",
				Type:   ItemTypeWeapon,
				Rarity: RarityRare,
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "missing type",
			item: &MagicItem{
				ID:     "item_1",
				Name:   "Flame Tongue",
				Rarity: RarityRare,
			},
			wantErr: true,
			errMsg:  "type is required",
		},
		{
			name: "missing rarity",
			item: &MagicItem{
				ID:   "item_1",
				Name: "Flame Tongue",
				Type: ItemTypeWeapon,
			},
			wantErr: true,
			errMsg:  "rarity is required",
		},
		{
			name: "invalid rarity",
			item: &MagicItem{
				ID:     "item_1",
				Name:   "Flame Tongue",
				Type:   ItemTypeWeapon,
				Rarity: "epic",
			},
			wantErr: true,
			errMsg:  "invalid rarity",
		},
		{
			name: "cursed item without removal method",
			item: &MagicItem{
				ID:     "item_1",
				Name:   "Cursed Sword",
				Type:   ItemTypeWeapon,
				Rarity: RarityUncommon,
				Curse:  &ItemCurse{Effect: "Cannot drop sword"},
			},
			wantErr: true,
			errMsg:  "must have removal method",
		},
		{
			name: "attunement required without requirements",
			item: &MagicItem{
				ID:                 "item_1",
				Name:               "Staff of Power",
				Type:               ItemTypeStaff,
				Rarity:             RarityVeryRare,
				AttunementRequired: true,
			},
			wantErr: true,
			errMsg:  "attunement_requirements is empty",
		},
		{
			name: "bonus exceeds rarity limit",
			item: &MagicItem{
				ID:     "item_1",
				Name:   "+3 Common Sword",
				Type:   ItemTypeWeapon,
				Rarity: RarityCommon,
				Properties: []MagicItemProperty{
					{Name: "Bonus", Bonus: 3},
				},
			},
			wantErr: true,
			errMsg:  "exceeds maximum",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.item.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %v, expected to contain %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestGetMaxBonusForRarity(t *testing.T) {
	tests := []struct {
		rarity MagicItemRarity
		want   int
	}{
		{RarityCommon, 0},
		{RarityUncommon, 1},
		{RarityRare, 2},
		{RarityVeryRare, 3},
		{RarityLegendary, 3},
		{RarityArtifact, 3},
	}

	for _, tt := range tests {
		t.Run(string(tt.rarity), func(t *testing.T) {
			got := GetMaxBonusForRarity(tt.rarity)
			if got != tt.want {
				t.Errorf("GetMaxBonusForRarity(%s) = %d, want %d", tt.rarity, got, tt.want)
			}
		})
	}
}

func TestGetMinLevelForRarity(t *testing.T) {
	tests := []struct {
		rarity MagicItemRarity
		want   int
	}{
		{RarityCommon, 1},
		{RarityUncommon, 3},
		{RarityRare, 5},
		{RarityVeryRare, 11},
		{RarityLegendary, 17},
		{RarityArtifact, 17},
	}

	for _, tt := range tests {
		t.Run(string(tt.rarity), func(t *testing.T) {
			got := GetMinLevelForRarity(tt.rarity)
			if got != tt.want {
				t.Errorf("GetMinLevelForRarity(%s) = %d, want %d", tt.rarity, got, tt.want)
			}
		})
	}
}

func TestGetApproximateValueGP(t *testing.T) {
	tests := []struct {
		rarity   MagicItemRarity
		wantMin  int
		wantMax  int
	}{
		{RarityCommon, 50, 100},
		{RarityUncommon, 101, 500},
		{RarityRare, 501, 5000},
		{RarityVeryRare, 5001, 50000},
		{RarityLegendary, 50001, 100000},
		{RarityArtifact, 0, 0}, // Priceless
	}

	for _, tt := range tests {
		t.Run(string(tt.rarity), func(t *testing.T) {
			min, max := GetApproximateValueGP(tt.rarity)
			if min != tt.wantMin || max != tt.wantMax {
				t.Errorf("GetApproximateValueGP(%s) = (%d, %d), want (%d, %d)", tt.rarity, min, max, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestMagicItem_Bonus(t *testing.T) {
	item := &MagicItem{
		ID:       "item_1",
		Name:     "Holy Avenger",
		Type:     ItemTypeWeapon,
		Rarity:   RarityLegendary,
		Properties: []MagicItemProperty{
			{Name: "Basic Attack", Bonus: 1},
			{Name: "Smite Evil", Bonus: 2},
			{Name: "Divine Power", Bonus: 3},
		},
		Lore: "A sword of divine power",
	}

	bonus := item.Bonus()
	if bonus != 3 {
		t.Errorf("Bonus() = %d, want 3", bonus)
	}
}
