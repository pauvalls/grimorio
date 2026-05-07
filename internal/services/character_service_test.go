package services

import (
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

func TestCharacterService_CreateCharacter(t *testing.T) {
	repo := repository.NewMemoryCharacterRepository()
	charService := NewCharacterService(repo)

	tests := []struct {
		name       string
		campaignID string
		charName   string
		race       string
		class      string
		level      int
		background string
		alignment  string
		wantErr    bool
	}{
		{
			name:       "create valid character",
			campaignID: "test-campaign",
			charName:   "Gandalf",
			race:       "humano",
			class:      "mago",
			level:      5,
			background: "sabio",
			alignment:  "LG",
			wantErr:    false,
		},
		{
			name:       "create character without optional fields",
			campaignID: "test-campaign",
			charName:   "Aragorn",
			wantErr:    false,
		},
		{
			name:       "create character without name",
			campaignID: "test-campaign",
			wantErr:    true,
		},
		{
			name:     "create character without campaign",
			charName: "Legolas",
			wantErr:  true,
		},
		{
			name:       "invalid level - too high",
			campaignID: "test-campaign",
			charName:   "Powerful",
			level:      21,
			wantErr:    true,
		},
		{
			name:       "invalid race",
			campaignID: "test-campaign",
			charName:   "Unknown",
			race:       "dragon",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			char, err := charService.CreateCharacter(tt.campaignID, tt.charName, tt.race, tt.class, tt.level, tt.background, tt.alignment)
			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateCharacter() expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("CreateCharacter() unexpected error: %v", err)
				return
			}
			if char.Name != tt.charName {
				t.Errorf("CreateCharacter() name = %v, want %v", char.Name, tt.charName)
			}
			if char.CampaignID != tt.campaignID {
				t.Errorf("CreateCharacter() campaign = %v, want %v", char.CampaignID, tt.campaignID)
			}
			if char.Status != "alive" {
				t.Errorf("CreateCharacter() status = %v, want alive", char.Status)
			}
			// Check that stats are calculated (not all 10s)
			if char.Stats.STR == 10 && char.Stats.DEX == 10 && char.Stats.CON == 10 &&
				char.Stats.INT == 10 && char.Stats.WIS == 10 && char.Stats.CHA == 10 {
				t.Errorf("CreateCharacter() all stats are 10, expected calculated values")
			}
			// Check HP is calculated (not default 10)
			if char.HP.Maximum == 10 && char.Class != "" {
				t.Errorf("CreateCharacter() HP = %v, expected calculated value for class %s", char.HP.Maximum, char.Class)
			}
			// Check AC is calculated (not default 10)
			if char.AC == 10 && char.Class != "" {
				t.Errorf("CreateCharacter() AC = %v, expected calculated value for class %s", char.AC, char.Class)
			}
			// Check skills are assigned
			if len(char.Skills) == 0 && (char.Class != "" || char.Background != "") {
				t.Errorf("CreateCharacter() no skills assigned")
			}
			// Check features are assigned for valid class
			if len(char.Features) == 0 && char.Class != "" {
				t.Errorf("CreateCharacter() no features assigned for class %s", char.Class)
			}
		})
	}
}

func TestCharacterService_GetCharacter(t *testing.T) {
	repo := repository.NewMemoryCharacterRepository()
	charService := NewCharacterService(repo)

	// Create a character
	_, err := charService.CreateCharacter("get-test", "Gandalf", "humano", "mago", 5, "sabio", "LG")
	if err != nil {
		t.Fatalf("Failed to create test character: %v", err)
	}

	tests := []struct {
		name       string
		campaignID string
		charName   string
		wantErr    bool
	}{
		{
			name:       "get existing character",
			campaignID: "get-test",
			charName:   "Gandalf",
			wantErr:    false,
		},
		{
			name:       "get non-existent character",
			campaignID: "get-test",
			charName:   "Saruman",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			char, err := charService.GetCharacter(tt.campaignID, tt.charName)
			if tt.wantErr {
				if err == nil {
					t.Errorf("GetCharacter() expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetCharacter() unexpected error: %v", err)
				return
			}
			if char.Name != tt.charName {
				t.Errorf("GetCharacter() name = %v, want %v", char.Name, tt.charName)
			}
		})
	}
}

func TestCharacterService_ListCharacters(t *testing.T) {
	repo := repository.NewMemoryCharacterRepository()
	charService := NewCharacterService(repo)

	// Create some characters
	_, _ = charService.CreateCharacter("list-test", "Gandalf", "humano", "mago", 5, "sabio", "LG")
	_, _ = charService.CreateCharacter("list-test", "Aragorn", "humano", "guerrero", 3, "soldado", "NG")

	characters, err := charService.ListCharacters("list-test")
	if err != nil {
		t.Errorf("ListCharacters() unexpected error: %v", err)
		return
	}

	if len(characters) != 2 {
		t.Errorf("ListCharacters() got %d characters, want 2", len(characters))
	}
}

func TestCharacterService_UpdateCharacter(t *testing.T) {
	repo := repository.NewMemoryCharacterRepository()
	charService := NewCharacterService(repo)

	// Create a character
	char, err := charService.CreateCharacter("update-test", "Gandalf", "humano", "mago", 5, "sabio", "LG")
	if err != nil {
		t.Fatalf("Failed to create test character: %v", err)
	}

	// Update level
	char.Level = 6
	err = charService.UpdateCharacter(char)
	if err != nil {
		t.Errorf("UpdateCharacter() unexpected error: %v", err)
		return
	}

	// Verify update
	updated, err := charService.GetCharacter("update-test", "Gandalf")
	if err != nil {
		t.Errorf("GetCharacter() after update unexpected error: %v", err)
		return
	}

	if updated.Level != 6 {
		t.Errorf("UpdateCharacter() level = %v, want 6", updated.Level)
	}
}

func TestCharacterService_AddRelationship(t *testing.T) {
	repo := repository.NewMemoryCharacterRepository()
	charService := NewCharacterService(repo)

	// Create characters
	_, err := charService.CreateCharacter("rel-test", "Gandalf", "humano", "mago", 5, "sabio", "LG")
	if err != nil {
		t.Fatalf("Failed to create test character: %v", err)
	}

	rel := domain.Relationship{
		EntityID:   "frodo-id",
		EntityName: "Frodo",
		EntityType: "pc",
		Type:       "ally",
		Strength:   8,
	}

	err = charService.AddRelationship("rel-test", "Gandalf", rel)
	if err != nil {
		t.Errorf("AddRelationship() unexpected error: %v", err)
		return
	}

	// Verify
	char, err := charService.GetCharacter("rel-test", "Gandalf")
	if err != nil {
		t.Errorf("GetCharacter() unexpected error: %v", err)
		return
	}

	if len(char.Relationships) != 1 {
		t.Errorf("AddRelationship() got %d relationships, want 1", len(char.Relationships))
		return
	}

	if char.Relationships[0].EntityName != "Frodo" {
		t.Errorf("AddRelationship() entity = %v, want Frodo", char.Relationships[0].EntityName)
	}
}
