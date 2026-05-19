package fs

import (
	"context"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
)

func TestFilesystemMilestoneXPRepository(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	repo := NewFilesystemMilestoneXPRepository(tmpDir)

	table := &domain.ChapterXPTable{
		ChapterID:    "chapter_1",
		ChapterTitle: "The Beginning",
		LevelRange:   domain.LevelRange{Min: 1, Max: 3},
		Milestones: []domain.MilestoneXP{
			{ChapterID: "chapter_1", SessionNumber: 1, XPThreshold: 300, CumulativeXP: 300, LevelAchieved: 2},
			{ChapterID: "chapter_1", SessionNumber: 2, XPThreshold: 600, CumulativeXP: 900, LevelAchieved: 3},
		},
		TotalSessions: 2,
	}

	// Create
	if err := repo.Create(ctx, "campaign-1", table); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	// Read
	read, err := repo.Read(ctx, "campaign-1", "chapter_1")
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if read.ChapterID != "chapter_1" {
		t.Errorf("Read() ChapterID = %s, want chapter_1", read.ChapterID)
	}

	// Read not found
	_, err = repo.Read(ctx, "campaign-1", "nonexistent")
	if err == nil {
		t.Error("Read() should error for nonexistent table")
	}

	// Update
	table.ChapterTitle = "Revised Beginning"
	if err := repo.Update(ctx, "campaign-1", table); err != nil {
		t.Fatalf("Update() error: %v", err)
	}
	updated, _ := repo.Read(ctx, "campaign-1", "chapter_1")
	if updated.ChapterTitle != "Revised Beginning" {
		t.Errorf("Update() Title = %s, want Revised Beginning", updated.ChapterTitle)
	}

	// GetTotalXP
	totalXP, err := repo.GetTotalXP(ctx, "campaign-1", "party_1")
	if err != nil {
		t.Fatalf("GetTotalXP() error: %v", err)
	}
	if totalXP != 1200 {
		t.Errorf("GetTotalXP() = %d, want 1200", totalXP)
	}

	// Delete
	if err := repo.Delete(ctx, "campaign-1", "chapter_1"); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
	_, err = repo.Read(ctx, "campaign-1", "chapter_1")
	if err == nil {
		t.Error("Read() after delete should error")
	}
}
