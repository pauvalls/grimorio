package fs

import (
	"context"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
)

func TestFilesystemAreaRepositoryV3(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	repo := NewFilesystemAreaRepositoryV3(tmpDir)

	area := &domain.Area{
		ID:         "area_ch1_1",
		ChapterID:  "chapter_1",
		AreaNumber: 1,
		Title:      "The Entrance",
		Summary:    "A dark cave entrance",
		Description: "A narrow passage leads into darkness",
		LevelRange: domain.LevelRange{Min: 1, Max: 3},
		Features: []domain.AreaFeature{
			{Type: "room", Name: "Cave Mouth", Description: "A wide cavern entrance", Hidden: false},
		},
		Encounters: []domain.AreaEncounter{
			{EncounterID: "enc_1", Trigger: "Upon entering", CRTotal: 1.0, XPValue: 200},
		},
		NPCs:     []domain.AreaNPC{},
		Treasure: []domain.Treasure{},
		Maps:     []domain.MapReference{},
	}

	// Create
	if err := repo.Create(ctx, "campaign-1", area); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	// Read
	read, err := repo.Read(ctx, "campaign-1", "area_ch1_1")
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if read.ID != "area_ch1_1" {
		t.Errorf("Read() ID = %s, want area_ch1_1", read.ID)
	}

	// Read not found
	_, err = repo.Read(ctx, "campaign-1", "nonexistent")
	if err == nil {
		t.Error("Read() should error for nonexistent area")
	}

	// Update
	area.Title = "Narrow Passage"
	if err := repo.Update(ctx, "campaign-1", area); err != nil {
		t.Fatalf("Update() error: %v", err)
	}
	updated, _ := repo.Read(ctx, "campaign-1", "area_ch1_1")
	if updated.Title != "Narrow Passage" {
		t.Errorf("Update() Title = %s, want Narrow Passage", updated.Title)
	}

	// GetByChapter
	byChapter, err := repo.GetByChapter(ctx, "campaign-1", "chapter_1")
	if err != nil {
		t.Fatalf("GetByChapter() error: %v", err)
	}
	if len(byChapter) != 1 {
		t.Errorf("GetByChapter() len = %d, want 1", len(byChapter))
	}

	// Delete
	if err := repo.Delete(ctx, "campaign-1", "area_ch1_1"); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
	_, err = repo.Read(ctx, "campaign-1", "area_ch1_1")
	if err == nil {
		t.Error("Read() after delete should error")
	}
}
