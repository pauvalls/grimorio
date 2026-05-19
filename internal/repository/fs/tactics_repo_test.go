package fs

import (
	"context"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
)

func TestFilesystemTacticsRepository(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	repo := NewFilesystemTacticsRepository(tmpDir)

	tactics := &domain.Tactics{
		MonsterID:        "mon_1",
		EncounterID:      "enc_1",
		IntelligenceTier: domain.TierTactical,
		OpeningMove:      "Cast fireball on clustered enemies",
		TargetPriority: []domain.TargetPriority{
			{Priority: 1, TargetType: "healer", Reasoning: "Prevent healing"},
			{Priority: 2, TargetType: "squishy", Reasoning: "Eliminate damage dealers"},
		},
		RetreatConditions: []domain.RetreatCondition{
			{Trigger: "HP < 25%", Method: "Disengage and flee"},
		},
	}

	// Create
	if err := repo.Create(ctx, "campaign-1", tactics); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	// Read
	read, err := repo.Read(ctx, "campaign-1", "mon_1")
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if read.MonsterID != "mon_1" {
		t.Errorf("Read() MonsterID = %s, want mon_1", read.MonsterID)
	}

	// Read not found
	_, err = repo.Read(ctx, "campaign-1", "nonexistent")
	if err == nil {
		t.Error("Read() should error for nonexistent tactics")
	}

	// ListByEncounter
	byEnc, err := repo.ListByEncounter(ctx, "campaign-1", "enc_1")
	if err != nil {
		t.Fatalf("ListByEncounter() error: %v", err)
	}
	if len(byEnc) != 1 {
		t.Errorf("ListByEncounter() len = %d, want 1", len(byEnc))
	}
}
