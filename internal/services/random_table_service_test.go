package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

func TestRandomTableService_GenerateTable(t *testing.T) {
	ctx := context.Background()
	canonRepo := repository.NewMemoryCanonRepository()

	doc := &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "test-campaign",
		Entities: []domain.CanonEntity{
			{
				ID:         "mcguffin-test",
				Name:       "Test McGuffin",
				Type:       domain.EntityTypeItem,
				Role:       "mcguffin",
				Properties: map[string]any{"level_range": "5-8"},
			},
		},
		Facts: []domain.CanonFact{
			{ID: "fact-1", Category: "creature", Statement: "Underdark creatures CR 3-11 roam the depths", Source: "lore"},
			{ID: "fact-2", Category: "creature", Statement: "Surface creatures CR 1-3 are common", Source: "lore"},
			{ID: "fact-3", Category: "weather", Statement: "Underdark has no weather", Source: "lore"},
		},
	}
	_ = canonRepo.Save("test-campaign", doc)

	svc := NewRandomTableService(canonRepo)

	t.Run("contextual encounter table", func(t *testing.T) {
		tbl, err := svc.GenerateTable(ctx, "test-campaign", domain.TableTypeEncounter, domain.TableContext{
			LevelRange:  "5-8",
			SettingType: "underdark",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(tbl.Entries) == 0 {
			t.Fatalf("expected entries, got none")
		}
		// All entries should reference underdark context
		for _, e := range tbl.Entries {
			if e.Weight <= 0 {
				t.Fatalf("expected positive weight, got %d", e.Weight)
			}
			if e.Description == "" {
				t.Fatalf("expected non-empty description")
			}
			if e.SourceFact == "" {
				t.Fatalf("expected source_fact")
			}
		}
	})

	t.Run("empty canon returns empty table", func(t *testing.T) {
		freshCanonRepo := repository.NewMemoryCanonRepository()
		freshDoc := &domain.CanonDocument{
			SchemaVersion: domain.SchemaVersionV2,
			CampaignID:    "empty-campaign",
			Facts:         []domain.CanonFact{},
		}
		_ = freshCanonRepo.Save("empty-campaign", freshDoc)
		freshSvc := NewRandomTableService(freshCanonRepo)

		tbl, err := freshSvc.GenerateTable(ctx, "empty-campaign", domain.TableTypeEncounter, domain.TableContext{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(tbl.Entries) != 0 {
			t.Fatalf("expected empty table, got %d entries", len(tbl.Entries))
		}
	})

	t.Run("invalid table type error", func(t *testing.T) {
		_, err := svc.GenerateTable(ctx, "test-campaign", domain.TableType("invalid"), domain.TableContext{})
		if err == nil {
			t.Fatalf("expected error for invalid table type")
		}
	})

	t.Run("table too large capped", func(t *testing.T) {
		freshCanonRepo := repository.NewMemoryCanonRepository()
		freshDoc := &domain.CanonDocument{
			SchemaVersion: domain.SchemaVersionV2,
			CampaignID:    "large-campaign",
		}
		// Add 150 facts to exceed the 100-entry limit
		for i := 0; i < 150; i++ {
			freshDoc.Facts = append(freshDoc.Facts, domain.CanonFact{
				ID:        fmt.Sprintf("fact-%d", i),
				Category:  "creature",
				Statement: fmt.Sprintf("Creature %d", i),
				Source:    "lore",
			})
		}
		_ = freshCanonRepo.Save("large-campaign", freshDoc)
		freshSvc := NewRandomTableService(freshCanonRepo)

		tbl, err := freshSvc.GenerateTable(ctx, "large-campaign", domain.TableTypeEncounter, domain.TableContext{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(tbl.Entries) > 100 {
			t.Fatalf("expected at most 100 entries, got %d", len(tbl.Entries))
		}
	})
}
