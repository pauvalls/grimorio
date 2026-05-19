package fs

import (
	"context"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
)

func TestFilesystemHandoutRepositoryV3(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	repo := NewFilesystemHandoutRepositoryV3(tmpDir)

	handout := &domain.Handout{
		ID:         "h_1",
		CampaignID: "campaign-1",
		Type:       domain.HandoutTypeLetter,
		Title:      "Mysterious Letter",
		Content:    "Dear adventurer...",
		Format:     domain.FormatText,
		Style:      domain.StyleFormal,
		QuestRefs:  []string{"quest_1"},
		AreaRefs:   []string{"area_1"},
	}

	// Create
	if err := repo.Create(ctx, "campaign-1", handout); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	// Read
	read, err := repo.Read(ctx, "campaign-1", "h_1")
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if read.ID != "h_1" {
		t.Errorf("Read() ID = %s, want h_1", read.ID)
	}

	// Read not found
	_, err = repo.Read(ctx, "campaign-1", "nonexistent")
	if err == nil {
		t.Error("Read() should error for nonexistent handout")
	}

	// GetByQuest
	byQuest, err := repo.GetByQuest(ctx, "campaign-1", "quest_1")
	if err != nil {
		t.Fatalf("GetByQuest() error: %v", err)
	}
	if len(byQuest) != 1 {
		t.Errorf("GetByQuest() len = %d, want 1", len(byQuest))
	}

	// GetByArea
	byArea, err := repo.GetByArea(ctx, "campaign-1", "area_1")
	if err != nil {
		t.Fatalf("GetByArea() error: %v", err)
	}
	if len(byArea) != 1 {
		t.Errorf("GetByArea() len = %d, want 1", len(byArea))
	}

	// Delete
	if err := repo.Delete(ctx, "campaign-1", "h_1"); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
	_, err = repo.Read(ctx, "campaign-1", "h_1")
	if err == nil {
		t.Error("Read() after delete should error")
	}
}
