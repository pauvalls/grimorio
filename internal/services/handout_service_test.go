package services

import (
	"context"
	"strings"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

func TestHandoutService_GenerateHandout(t *testing.T) {
	ctx := context.Background()
	questRepo := repository.NewMemoryQuestRepository()
	canonRepo := repository.NewMemoryCanonRepository()

	_ = questRepo.Save(&domain.Quest{
		ID:          "quest-001",
		CampaignID:  "test-campaign",
		Title:       "The Betrayal",
		Description: "A quest with a secret betrayal. [DM] The quest giver will betray the party at the end.",
	})

	svc := NewHandoutService(questRepo, canonRepo)

	t.Run("dual version quest handout", func(t *testing.T) {
		handout, err := svc.GenerateHandout(ctx, "test-campaign", domain.HandoutTypeQuest, []string{"quest-001"}, domain.HandoutVersionBoth)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if handout.PlayerVersion == "" {
			t.Fatalf("expected player version")
		}
		if handout.DMVersion == "" {
			t.Fatalf("expected DM version")
		}
		// Player version should omit [DM] sections
		if strings.Contains(handout.PlayerVersion, "[DM]") {
			t.Fatalf("player version should not contain [DM] sections")
		}
		// DM version should include [DM] annotations
		if !strings.Contains(handout.DMVersion, "[DM]") {
			t.Fatalf("DM version should contain [DM] sections")
		}
	})

	t.Run("player version only", func(t *testing.T) {
		handout, err := svc.GenerateHandout(ctx, "test-campaign", domain.HandoutTypeQuest, []string{"quest-001"}, domain.HandoutVersionPlayer)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if handout.PlayerVersion == "" {
			t.Fatalf("expected player version")
		}
		if handout.DMVersion != "" {
			t.Fatalf("expected empty DM version when requesting player only")
		}
	})

	t.Run("missing content ref", func(t *testing.T) {
		_, err := svc.GenerateHandout(ctx, "test-campaign", domain.HandoutTypeQuest, []string{"nonexistent"}, domain.HandoutVersionBoth)
		if err == nil {
			t.Fatalf("expected error for missing content ref")
		}
	})

	t.Run("empty content refs rejected", func(t *testing.T) {
		_, err := svc.GenerateHandout(ctx, "test-campaign", domain.HandoutTypeQuest, []string{}, domain.HandoutVersionBoth)
		if err == nil {
			t.Fatalf("expected error for empty content refs")
		}
	})
}
