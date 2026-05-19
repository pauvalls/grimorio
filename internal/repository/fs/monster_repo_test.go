package fs

import (
	"context"
	"testing"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
)

func TestFilesystemMonsterRepository(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	repo := NewFilesystemMonsterRepository(tmpDir)

	monster := &domain.Monster{
		ID:          "mon_1",
		CampaignID:  "campaign-1",
		Name:        "Goblin",
		CR:          "1/4",
		Type:        "humanoid",
		Size:        "Small",
		Abilities:   []string{"Nimble Escape"},
		Description: "A small, cunning humanoid",
		CreatedAt:   time.Now(),
	}

	// Save
	if err := repo.Save(ctx, "campaign-1", monster); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Read
	read, err := repo.Read(ctx, "campaign-1", "Goblin")
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if read.Name != "Goblin" {
		t.Errorf("Read() Name = %s, want Goblin", read.Name)
	}

	// Read not found
	_, err = repo.Read(ctx, "campaign-1", "Dragon")
	if err == nil {
		t.Error("Read() should error for nonexistent monster")
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
	if err := repo.Delete(ctx, "campaign-1", "Goblin"); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
	list, _ = repo.List(ctx, "campaign-1")
	if len(list) != 0 {
		t.Errorf("List() after delete len = %d, want 0", len(list))
	}
}

func TestFilesystemMonsterRepository_ListEmpty(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	repo := NewFilesystemMonsterRepository(tmpDir)

	list, err := repo.List(ctx, "campaign-1")
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("List() len = %d, want 0", len(list))
	}
}
