package fs

import (
	"context"
	"testing"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
)

func TestFilesystemEncounterRepository(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	repo := NewFilesystemEncounterRepository(tmpDir)

	enc := &domain.Encounter{
		ID:          "enc_1",
		CampaignID:  "campaign-1",
		Name:        "Goblin Ambush",
		Difficulty:  "medium",
		Location:    "Forest Road",
		Monsters:    []domain.MonsterRef{{Name: "Goblin", Quantity: 3}},
		Description: "A group of goblins ambush travelers",
		CreatedAt:   time.Now(),
	}

	// Save
	if err := repo.Save(ctx, "campaign-1", enc); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Read
	read, err := repo.Read(ctx, "campaign-1", "Goblin Ambush")
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if read.Name != "Goblin Ambush" {
		t.Errorf("Read() Name = %s, want Goblin Ambush", read.Name)
	}

	// Read not found
	_, err = repo.Read(ctx, "campaign-1", "Dragon Lair")
	if err == nil {
		t.Error("Read() should error for nonexistent encounter")
	}

	// List
	list, err := repo.List(ctx, "campaign-1")
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("List() len = %d, want 1", len(list))
	}

	// Delete
	if err := repo.Delete(ctx, "campaign-1", "Goblin Ambush"); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
	list, _ = repo.List(ctx, "campaign-1")
	if len(list) != 0 {
		t.Errorf("List() after delete len = %d, want 0", len(list))
	}
}
