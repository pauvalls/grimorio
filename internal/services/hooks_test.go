package services

import (
	"context"
	"strings"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

func TestPlayerHookService_GenerateHooks(t *testing.T) {
	ctx := context.Background()
	charRepo := repository.NewMemoryCharacterRepository()
	canonRepo := repository.NewMemoryCanonRepository()

	svc := NewPlayerHookService(charRepo, canonRepo)

	seedData := func() {
		doc := &domain.CanonDocument{
			SchemaVersion: domain.SchemaVersionV2,
			CampaignID:    "test-campaign",
			Entities: []domain.CanonEntity{
				{ID: "mcguffin", Name: "The Ancient Crown", Type: domain.EntityTypeItem, Role: "mcguffin"},
				{ID: "villain", Name: "Lord Darkmore", Type: domain.EntityTypeNPC, Role: "villain"},
				{ID: "ally", Name: "Sister Maria", Type: domain.EntityTypeNPC, Role: "ally"},
			},
			Relationships: []domain.CanonRelationship{
				{ID: "rel-1", From: "mcguffin", To: "villain", Type: domain.RelationshipTypeEnemy},
			},
		}
		_ = canonRepo.Save("test-campaign", doc)

		_ = charRepo.Save(&domain.Character{
			Name:       "Aric",
			CampaignID: "test-campaign",
			Class:      "guerrero",
			Background: "soldado",
			Level:      1,
		})
		_ = charRepo.Save(&domain.Character{
			Name:       "Lyra",
			CampaignID: "test-campaign",
			Class:      "mago",
			Background: "sabio",
			Level:      1,
		})
		_ = charRepo.Save(&domain.Character{
			Name:       "Kael",
			CampaignID: "test-campaign",
			Class:      "picaro",
			Background: "criminal",
			Level:      1,
		})
		_ = charRepo.Save(&domain.Character{
			Name:       "Mira",
			CampaignID: "test-campaign",
			Class:      "clerigo",
			Background: "acolito",
			Level:      1,
		})
		_ = charRepo.Save(&domain.Character{
			Name:       "Unknown",
			CampaignID: "test-campaign",
			Class:      "barbaro",
			Background: "",
			Level:      1,
		})
	}

	t.Run("generates hooks for all characters", func(t *testing.T) {
		seedData()

		hooks, warnings, err := svc.GenerateHooks(ctx, "test-campaign")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(hooks) != 5 {
			t.Fatalf("expected 5 hooks, got %d", len(hooks))
		}
		if len(warnings) != 1 {
			t.Fatalf("expected 1 warning for missing background, got %d", len(warnings))
		}
	})

	t.Run("hooks reference canon entities", func(t *testing.T) {
		seedData()

		hooks, _, _ := svc.GenerateHooks(ctx, "test-campaign")

		foundCanonRef := false
		for _, h := range hooks {
			if strings.Contains(h.Hook, "Ancient Crown") || strings.Contains(h.Hook, "Darkmore") || strings.Contains(h.Hook, "Sister Maria") {
				foundCanonRef = true
			}
			if h.Hook == "" {
				t.Fatalf("expected non-empty hook for %s", h.CharacterName)
			}
		}
		if !foundCanonRef {
			t.Fatalf("expected at least one hook to reference a canon entity")
		}
	})

	t.Run("generic fallback for missing background", func(t *testing.T) {
		seedData()

		hooks, warnings, _ := svc.GenerateHooks(ctx, "test-campaign")

		var unknownHook *domain.CharacterHook
		for i := range hooks {
			if hooks[i].CharacterName == "Unknown" {
				unknownHook = &hooks[i]
				break
			}
		}
		if unknownHook == nil {
			t.Fatalf("expected hook for Unknown character")
		}
		if unknownHook.Hook == "" {
			t.Fatalf("expected generic fallback hook")
		}
		if !strings.Contains(warnings[0], "background") {
			t.Fatalf("expected background warning, got %q", warnings[0])
		}
	})

	t.Run("no characters returns empty with warning", func(t *testing.T) {
		doc := &domain.CanonDocument{
			SchemaVersion: domain.SchemaVersionV2,
			CampaignID:    "empty-chars",
		}
		_ = canonRepo.Save("empty-chars", doc)

		hooks, warnings, err := svc.GenerateHooks(ctx, "empty-chars")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(hooks) != 0 {
			t.Fatalf("expected 0 hooks, got %d", len(hooks))
		}
		if len(warnings) == 0 {
			t.Fatalf("expected warning for no characters")
		}
	})

	t.Run("missing campaign returns error", func(t *testing.T) {
		_, _, err := svc.GenerateHooks(ctx, "nonexistent")
		if err == nil {
			t.Fatalf("expected error for missing campaign")
		}
	})
}
