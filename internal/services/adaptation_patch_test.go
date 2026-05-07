package services

import (
	"context"
	"strings"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

func TestAdaptationPatch_GeneratePatch(t *testing.T) {
	ctx := context.Background()
	actRepo := repository.NewMemoryActRepository()
	canonRepo := repository.NewMemoryCanonRepository()

	_ = actRepo.Save(&domain.Act{
		CampaignID: "test-campaign",
		Number:     1,
		Title:      "Act One",
		Content:    "The merchants-guild controls the trade routes.",
	})
	_ = actRepo.Save(&domain.Act{
		CampaignID: "test-campaign",
		Number:     2,
		Title:      "Act Two",
		Content:    "The merchants-guild leader is a trusted ally.",
	})

	doc := &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "test-campaign",
		Entities: []domain.CanonEntity{
			{ID: "merchants-guild", Name: "Merchants Guild", Type: domain.EntityTypeFaction},
		},
	}
	_ = canonRepo.Save("test-campaign", doc)

	svc := NewAdaptationPatchService(actRepo, canonRepo)

	t.Run("faction coup patch", func(t *testing.T) {
		event := domain.WorldEvent{
			ID:          "merchants-guild",
			TriggerType: "faction-coup",
			Description: "The merchants-guild has undergone a coup",
			SessionNum:  3,
		}
		patch, err := svc.GeneratePatch(ctx, "test-campaign", event)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if patch.IsEmpty {
			t.Fatalf("expected non-empty patch")
		}
		if !strings.Contains(patch.MarkdownDiff, "merchants-guild") {
			t.Fatalf("expected markdown diff to reference merchants-guild")
		}
		if len(patch.AffectedActs) != 2 {
			t.Fatalf("expected 2 affected acts, got %d", len(patch.AffectedActs))
		}
	})

	t.Run("no matches returns empty patch", func(t *testing.T) {
		event := domain.WorldEvent{
			ID:          "nonexistent-faction",
			TriggerType: "faction-coup",
			Description: "A coup in a nonexistent faction",
			SessionNum:  3,
		}
		patch, err := svc.GeneratePatch(ctx, "test-campaign", event)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !patch.IsEmpty {
			t.Fatalf("expected empty patch")
		}
	})

	t.Run("idempotent generation", func(t *testing.T) {
		event := domain.WorldEvent{
			ID:          "merchants-guild",
			TriggerType: "faction-coup",
			Description: "Coup",
			SessionNum:  3,
		}
		patch1, _ := svc.GeneratePatch(ctx, "test-campaign", event)
		patch2, _ := svc.GeneratePatch(ctx, "test-campaign", event)
		if patch1.MarkdownDiff != patch2.MarkdownDiff {
			t.Fatalf("expected idempotent patch generation")
		}
	})
}
