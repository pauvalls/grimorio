package fs

import (
	"context"
	"testing"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
)

func TestFilesystemMapRepository(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	repo := NewFilesystemMapRepository(tmpDir)

	m := &domain.Map{
		ID:          "map_1",
		CampaignID:  "campaign-1",
		Name:        "Dungeon Level 1",
		Type:        "dungeon",
		Description: "The first level of the ancient dungeon",
		Labels:      []string{"Entrance", "Trap Room"},
		CreatedAt:   time.Now(),
	}

	// Save
	if err := repo.Save(ctx, "campaign-1", m); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Read
	read, err := repo.Read(ctx, "campaign-1", "Dungeon Level 1")
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if read.Name != "Dungeon Level 1" {
		t.Errorf("Read() Name = %s, want Dungeon Level 1", read.Name)
	}

	// Read not found
	_, err = repo.Read(ctx, "campaign-1", "Nonexistent Map")
	if err == nil {
		t.Error("Read() should error for nonexistent map")
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
	if err := repo.Delete(ctx, "campaign-1", "Dungeon Level 1"); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
	list, _ = repo.List(ctx, "campaign-1")
	if len(list) != 0 {
		t.Errorf("List() after delete len = %d, want 0", len(list))
	}
}
