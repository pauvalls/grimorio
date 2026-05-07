package services

import (
	"context"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

func TestFactionService_UpdateReputation(t *testing.T) {
	ctx := context.Background()
	canonRepo := repository.NewMemoryCanonRepository()
	factionRepo := repository.NewMemoryFactionRepository()

	// Seed canon with factions
	doc := &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "test-campaign",
		Entities: []domain.CanonEntity{
			{ID: "faction-a", Name: "Faction A", Type: domain.EntityTypeFaction},
			{ID: "faction-b", Name: "Faction B", Type: domain.EntityTypeFaction},
			{ID: "faction-c", Name: "Faction C", Type: domain.EntityTypeFaction},
		},
		Relationships: []domain.CanonRelationship{
			{ID: "rel-1", From: "faction-a", To: "faction-b", Type: domain.RelationshipTypeAlly, Strength: 5},
			{ID: "rel-2", From: "faction-b", To: "faction-c", Type: domain.RelationshipTypeAlly, Strength: 5},
		},
	}
	_ = canonRepo.Save("test-campaign", doc)

	svc := NewFactionService(canonRepo, factionRepo)

	t.Run("direct reputation update", func(t *testing.T) {
		result, err := svc.UpdateReputation(ctx, "test-campaign", "faction-a", "party-1", 20, "helped in battle", "quest")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.DirectChange.Score != 20 {
			t.Fatalf("direct score = %d, want 20", result.DirectChange.Score)
		}
		if result.DirectChange.Status != "neutral" {
			t.Fatalf("direct status = %q, want neutral", result.DirectChange.Status)
		}
		if len(result.DirectChange.History) != 1 {
			t.Fatalf("history len = %d, want 1", len(result.DirectChange.History))
		}
	})

	t.Run("bounds cap upper", func(t *testing.T) {
		freshCanonRepo := repository.NewMemoryCanonRepository()
		freshFactionRepo := repository.NewMemoryFactionRepository()
		freshDoc := &domain.CanonDocument{
			SchemaVersion: domain.SchemaVersionV2,
			CampaignID:    "bounds-campaign",
			Entities: []domain.CanonEntity{
				{ID: "faction-a", Name: "Faction A", Type: domain.EntityTypeFaction},
			},
		}
		_ = freshCanonRepo.Save("bounds-campaign", freshDoc)
		freshSvc := NewFactionService(freshCanonRepo, freshFactionRepo)

		// First set to 95
		_, _ = freshSvc.UpdateReputation(ctx, "bounds-campaign", "faction-a", "party-1", 95, "previous help", "quest")
		// Then try +10 more — should cap at 100 with actual delta 5
		result, err := freshSvc.UpdateReputation(ctx, "bounds-campaign", "faction-a", "party-1", 10, "more help", "quest")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.DirectChange.Score != 100 {
			t.Fatalf("capped score = %d, want 100", result.DirectChange.Score)
		}
		lastEvent := result.DirectChange.History[len(result.DirectChange.History)-1]
		if lastEvent.Delta != 5 {
			t.Fatalf("recorded delta = %d, want 5", lastEvent.Delta)
		}
	})

	t.Run("faction not found", func(t *testing.T) {
		_, err := svc.UpdateReputation(ctx, "test-campaign", "nonexistent", "party-1", 10, "test", "test")
		if err == nil {
			t.Fatalf("expected error for nonexistent faction")
		}
	})

	t.Run("zero delta rejected", func(t *testing.T) {
		_, err := svc.UpdateReputation(ctx, "test-campaign", "faction-a", "party-1", 0, "no change", "test")
		if err == nil {
			t.Fatalf("expected error for zero delta")
		}
	})

	t.Run("propagation to allies", func(t *testing.T) {
		// Use fresh repos to avoid accumulation from previous subtests
		freshCanonRepo := repository.NewMemoryCanonRepository()
		freshFactionRepo := repository.NewMemoryFactionRepository()
		freshDoc := &domain.CanonDocument{
			SchemaVersion: domain.SchemaVersionV2,
			CampaignID:    "prop-campaign",
			Entities: []domain.CanonEntity{
				{ID: "faction-a", Name: "Faction A", Type: domain.EntityTypeFaction},
				{ID: "faction-b", Name: "Faction B", Type: domain.EntityTypeFaction},
				{ID: "faction-c", Name: "Faction C", Type: domain.EntityTypeFaction},
			},
			Relationships: []domain.CanonRelationship{
				{ID: "rel-1", From: "faction-a", To: "faction-b", Type: domain.RelationshipTypeAlly, Strength: 5},
				{ID: "rel-2", From: "faction-b", To: "faction-c", Type: domain.RelationshipTypeAlly, Strength: 5},
			},
		}
		_ = freshCanonRepo.Save("prop-campaign", freshDoc)
		freshSvc := NewFactionService(freshCanonRepo, freshFactionRepo)

		result, err := freshSvc.UpdateReputation(ctx, "prop-campaign", "faction-a", "party-1", 20, "test propagation", "quest")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// faction-b should get +10 (ally, 1 hop)
		// faction-c should get +5 (ally of ally, 2 hops)
		var bFound, cFound bool
		for _, p := range result.PropagatedChanges {
			if p.FactionID == "faction-b" {
				bFound = true
				if p.Score != 10 {
					t.Fatalf("propagated B score = %d, want 10", p.Score)
				}
			}
			if p.FactionID == "faction-c" {
				cFound = true
				if p.Score != 5 {
					t.Fatalf("propagated C score = %d, want 5", p.Score)
				}
			}
		}
		if !bFound {
			t.Fatalf("expected faction-b in propagated changes")
		}
		if !cFound {
			t.Fatalf("expected faction-c in propagated changes")
		}
	})

	t.Run("propagation to enemies flips sign", func(t *testing.T) {
		freshCanonRepo := repository.NewMemoryCanonRepository()
		freshFactionRepo := repository.NewMemoryFactionRepository()
		freshDoc := &domain.CanonDocument{
			SchemaVersion: domain.SchemaVersionV2,
			CampaignID:    "enemy-campaign",
			Entities: []domain.CanonEntity{
				{ID: "faction-a", Name: "Faction A", Type: domain.EntityTypeFaction},
				{ID: "faction-b", Name: "Faction B", Type: domain.EntityTypeFaction},
			},
			Relationships: []domain.CanonRelationship{
				{ID: "rel-1", From: "faction-a", To: "faction-b", Type: domain.RelationshipTypeEnemy, Strength: 5},
			},
		}
		_ = freshCanonRepo.Save("enemy-campaign", freshDoc)
		freshSvc := NewFactionService(freshCanonRepo, freshFactionRepo)

		result, err := freshSvc.UpdateReputation(ctx, "enemy-campaign", "faction-a", "party-1", 20, "test enemy propagation", "quest")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var bFound bool
		for _, p := range result.PropagatedChanges {
			if p.FactionID == "faction-b" {
				bFound = true
				if p.Score != -10 {
					t.Fatalf("propagated enemy B score = %d, want -10", p.Score)
				}
			}
		}
		if !bFound {
			t.Fatalf("expected faction-b in propagated changes")
		}
	})

	t.Run("circular allies safe", func(t *testing.T) {
		freshCanonRepo := repository.NewMemoryCanonRepository()
		freshFactionRepo := repository.NewMemoryFactionRepository()
		freshDoc := &domain.CanonDocument{
			SchemaVersion: domain.SchemaVersionV2,
			CampaignID:    "circular-campaign",
			Entities: []domain.CanonEntity{
				{ID: "faction-a", Name: "Faction A", Type: domain.EntityTypeFaction},
				{ID: "faction-b", Name: "Faction B", Type: domain.EntityTypeFaction},
				{ID: "faction-c", Name: "Faction C", Type: domain.EntityTypeFaction},
			},
			Relationships: []domain.CanonRelationship{
				{ID: "rel-1", From: "faction-a", To: "faction-b", Type: domain.RelationshipTypeAlly, Strength: 5},
				{ID: "rel-2", From: "faction-b", To: "faction-c", Type: domain.RelationshipTypeAlly, Strength: 5},
				{ID: "rel-3", From: "faction-c", To: "faction-a", Type: domain.RelationshipTypeAlly, Strength: 5},
			},
		}
		_ = freshCanonRepo.Save("circular-campaign", freshDoc)
		freshSvc := NewFactionService(freshCanonRepo, freshFactionRepo)

		result, err := freshSvc.UpdateReputation(ctx, "circular-campaign", "faction-a", "party-1", 20, "test circular", "quest")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Should have exactly 2 propagated changes (B and C), each updated once
		if len(result.PropagatedChanges) != 2 {
			t.Fatalf("expected 2 propagated changes, got %d", len(result.PropagatedChanges))
		}
	})
}

func TestFactionService_GetReputationMatrix(t *testing.T) {
	ctx := context.Background()
	canonRepo := repository.NewMemoryCanonRepository()
	factionRepo := repository.NewMemoryFactionRepository()

	doc := &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "test-campaign",
		Entities: []domain.CanonEntity{
			{ID: "faction-public", Name: "Public Faction", Type: domain.EntityTypeFaction, Properties: map[string]any{"is_secret": false}},
			{ID: "faction-secret", Name: "Secret Faction", Type: domain.EntityTypeFaction, Properties: map[string]any{"is_secret": true}},
		},
	}
	_ = canonRepo.Save("test-campaign", doc)

	svc := NewFactionService(canonRepo, factionRepo)
	_, _ = svc.UpdateReputation(ctx, "test-campaign", "faction-public", "party-1", 30, "test", "test")
	_, _ = svc.UpdateReputation(ctx, "test-campaign", "faction-secret", "party-1", -50, "test", "test")

	t.Run("full matrix includes secret factions", func(t *testing.T) {
		matrix, err := svc.GetReputationMatrix(ctx, "test-campaign")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(matrix.Entries) != 2 {
			t.Fatalf("expected 2 entries, got %d", len(matrix.Entries))
		}
	})

	t.Run("player matrix filters secret factions", func(t *testing.T) {
		matrix, err := svc.GetPlayerReputationMatrix(ctx, "test-campaign")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(matrix.Entries) != 1 {
			t.Fatalf("expected 1 entry (secret filtered), got %d", len(matrix.Entries))
		}
		if matrix.Entries[0].FactionID != "faction-public" {
			t.Fatalf("expected only public faction, got %q", matrix.Entries[0].FactionID)
		}
	})
}
